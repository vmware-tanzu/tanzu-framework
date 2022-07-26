// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package clusterstatus provides the reconciler for the Cluster UpdatesAvailable status controller.
package clusterstatus

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/util/patchset"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/resolver"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/resolver/data"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/resolution"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/topology"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/version"
)

const LegacyClusterTKRLabel = "tanzuKubernetesRelease"

type Reconciler struct {
	Log         logr.Logger
	Client      client.Client
	TKRResolver resolver.CachingResolver
	Context     context.Context
}

var hasTKRLabel = predicate.NewPredicateFuncs(func(o client.Object) bool {
	ls := labels.Set(o.GetLabels())
	return ls.Has(runv1.LabelTKR) || ls.Has(LegacyClusterTKRLabel)
})

const indexCanUpdateToVersion = ".index.canUpdateToVersion"

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(
		r.Context, &clusterv1.Cluster{}, indexCanUpdateToVersion, versionsClusterCanUpdateTo,
	); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named("cluster_updates_available").
		For(&clusterv1.Cluster{}, builder.WithPredicates(hasTKRLabel, predicate.LabelChangedPredicate{})).
		Watches(
			&source.Kind{Type: &runv1.TanzuKubernetesRelease{}},
			handler.EnqueueRequestsFromMapFunc(r.clustersUpdatingToTKR),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{})).
		Complete(r)
}

func versionsClusterCanUpdateTo(o client.Object) []string {
	cluster := o.(*clusterv1.Cluster)
	if !conditions.IsTrue(cluster, runv1.ConditionUpdatesAvailable) {
		return nil
	}
	message := conditions.GetMessage(cluster, runv1.ConditionUpdatesAvailable)
	return strings.Split(strings.TrimSuffix(strings.TrimPrefix(message, "["), "]"), " ")
}

// clustersUpdatingToTKR, for any given tkr, returns requests for clusters that have tkr.Spec.Kubernetes.Version
// or tkr.Spec.Version in their UpdatesAvailable condition message.
func (r *Reconciler) clustersUpdatingToTKR(o client.Object) []ctrl.Request {
	tkr := o.(*runv1.TanzuKubernetesRelease)
	var result []ctrl.Request
	for _, v := range []string{tkr.Spec.Kubernetes.Version, tkr.Spec.Version} {
		result = append(result, r.clustersUpdatingToVersion(v)...)
	}
	return result
}

// clustersUpdatingToVersion produces requests for clusters that have the given version in their UpdatesAvailable condition message.
func (r *Reconciler) clustersUpdatingToVersion(v string) []ctrl.Request {
	clusterList := &clusterv1.ClusterList{}
	if err := r.Client.List(r.Context, clusterList, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(indexCanUpdateToVersion, v),
	}); err != nil {
		r.Log.Error(err, "error listing clusters updating to", "version", v)
		return nil
	}
	result := make([]reconcile.Request, len(clusterList.Items))
	for i := range clusterList.Items {
		cluster := &clusterList.Items[i]
		result[i].NamespacedName = types.NamespacedName{Namespace: cluster.Namespace, Name: cluster.Name}
	}
	return result
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, retErr error) {
	cluster := &clusterv1.Cluster{}
	if err := r.Client.Get(ctx, req.NamespacedName, cluster); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil // dropping request: cluster is not found
		}
		return ctrl.Result{}, err
	}

	ps := patchset.New(r.Client)
	defer func() {
		// apply patches unless an error is being returned
		if retErr != nil {
			return
		}
		if err := ps.Apply(ctx); err != nil {
			if err = kerrors.FilterOut(err, apierrors.IsConflict); err == nil {
				// retry if someone updated an object we wanted to patch
				result = ctrl.Result{Requeue: true}
			}
			retErr = errors.Wrap(err, "applying patches to TKRs")
		}
	}()

	ps.Add(cluster)

	return ctrl.Result{}, r.calculateAndSetUpdatesAvailable(ctx, cluster)
}

func (r *Reconciler) calculateAndSetUpdatesAvailable(ctx context.Context, cluster *clusterv1.Cluster) error {
	tkrName := cluster.Labels[runv1.LabelTKR]
	clusterClass, err := topology.GetClusterClass(ctx, r.Client, cluster)
	if err != nil {
		return err
	}
	if clusterClass == nil {
		tkrName = cluster.Labels[LegacyClusterTKRLabel]
	}

	tkrVersion, err := version.ParseSemantic(version.FromLabel(tkrName))
	if err != nil {
		conditions.MarkUnknown(cluster, runv1.ConditionUpdatesAvailable, runv1.ReasonCannotParseTKR, "Cannot parse TKR version from TKR name '%s': %s", tkrName, err)
		return nil
	}
	updates, err := r.updatesAvailable(ctx, tkrVersion, cluster, clusterClass)
	if err != nil {
		return err
	}
	setUpdatesAvailable(cluster, updates)
	return nil
}

func (r *Reconciler) updatesAvailable(ctx context.Context, tkrVersion *version.Version, cluster *clusterv1.Cluster, clusterClass *clusterv1.ClusterClass) ([]string, error) {
	var result []string
	major, minor := tkrVersion.Major(), tkrVersion.Minor()

	for _, versionPrefix := range []string{
		vLabelMinor(major, minor),
		vLabelMinor(major, minor+1),
	} {
		updateVersions, err := r.findUpdateVersions(ctx, tkrVersion, cluster, clusterClass, versionPrefix)
		if err != nil {
			return nil, err
		}
		result = append(result, updateVersions...)
	}

	return result, nil
}

func vLabelMinor(major, minor uint) string {
	return fmt.Sprintf("v%v.%v", major, minor)
}

func (r *Reconciler) findUpdateVersions(ctx context.Context, tkrVersion *version.Version, cluster *clusterv1.Cluster, clusterClass *clusterv1.ClusterClass, versionPrefix string) ([]string, error) {
	if clusterClass == nil {
		return r.legacyUpdateTKRVersions(ctx, tkrVersion, versionPrefix)
	}

	query, err := resolution.ConstructQuery(versionPrefix, cluster, clusterClass)
	if err != nil {
		return nil, err
	}
	tkrResult := r.TKRResolver.Resolve(*query)
	if tkrResult.ControlPlane.K8sVersion != "" && tkrResult.ControlPlane.TKRName != version.Label(tkrVersion.String()) {
		return resolvedTKRVersions(tkrVersion, tkrResult)
	}
	return nil, nil
}

var activeAndCompatible = func() labels.Selector {
	selector, err := labels.Parse("!deactivated,!incompatible")
	if err != nil {
		panic(err)
	}
	return selector
}()

func (r *Reconciler) legacyUpdateTKRVersions(ctx context.Context, currentTKRVersion *version.Version, versionPrefix string) ([]string, error) {
	tkrList := &runv1.TanzuKubernetesReleaseList{}
	hasVersionPrefix, _ := labels.NewRequirement(version.Label(versionPrefix), selection.Exists, nil)
	doesNotHaveLabel, _ := labels.NewRequirement(version.Label(currentTKRVersion.String()), selection.DoesNotExist, nil)
	selector := activeAndCompatible.Add(*hasVersionPrefix, *doesNotHaveLabel)
	if err := r.Client.List(ctx, tkrList, client.MatchingLabelsSelector{Selector: selector}); err != nil {
		return nil, errors.Wrapf(err, "error listing TKRs with version prefix '%s'", versionPrefix)
	}
	var result []string
	for i := range tkrList.Items {
		tkr := &tkrList.Items[i]
		tkrVersion, _ := version.ParseSemantic(tkr.Spec.Version)
		if currentTKRVersion.LessThan(tkrVersion) {
			result = append(result, tkr.Spec.Version)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		vi, _ := version.ParseSemantic(result[i])
		vj, _ := version.ParseSemantic(result[j])
		return vi.LessThan(vj)
	})
	return result, nil
}

func resolvedTKRVersions(currentTKRVersion *version.Version, tkrResult data.Result) ([]string, error) {
	var result []string
	for _, tkrs := range tkrResult.ControlPlane.TKRsByK8sVersion {
		for _, tkr := range tkrs {
			tkrVersion, _ := version.ParseSemantic(tkr.Spec.Version)
			if currentTKRVersion.LessThan(tkrVersion) {
				result = append(result, tkr.Spec.Version)
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		vi, _ := version.ParseSemantic(result[i])
		vj, _ := version.ParseSemantic(result[j])
		return vi.LessThan(vj)
	})
	return result, nil
}

func setUpdatesAvailable(cluster *clusterv1.Cluster, updates []string) {
	if len(updates) == 0 {
		conditions.MarkFalse(cluster, runv1.ConditionUpdatesAvailable, runv1.ReasonAlreadyUpToDate, clusterv1.ConditionSeverityInfo, "")
		return
	}
	updatesAvailableCondition := conditions.TrueCondition(runv1.ConditionUpdatesAvailable)
	updatesAvailableCondition.Message = fmt.Sprintf("%v", updates)
	conditions.Set(cluster, updatesAvailableCondition)
}
