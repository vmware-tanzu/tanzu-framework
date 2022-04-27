// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package clusterstatus provides the reconciler for the Cluster UpdatesAvailable status controller.
package clusterstatus

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	utilversion "k8s.io/apimachinery/pkg/util/version"
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
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/resolution"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/topology"
)

type Reconciler struct {
	Log         logr.Logger
	Client      client.Client
	TKRResolver resolver.CachingResolver
	Context     context.Context
}

var hasTKRLabel = func() predicate.Predicate {
	selector, _ := labels.Parse(runv1.LabelTKR)
	return predicate.NewPredicateFuncs(func(o client.Object) bool {
		return selector.Matches(labels.Set(o.GetLabels()))
	})
}()

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

func (r *Reconciler) clustersUpdatingToTKR(o client.Object) []ctrl.Request {
	tkr := o.(*runv1.TanzuKubernetesRelease)
	var result []ctrl.Request
	for _, v := range []string{tkr.Spec.Kubernetes.Version, tkr.Spec.Version} {
		result = append(result, r.clustersUpdatingToVersion(v)...)
	}
	return result
}

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
	clusterClass, err := topology.GetClusterClass(ctx, r.Client, cluster)
	if err != nil {
		return err
	}
	if clusterClass == nil {
		return nil // do not set UpdatesAvailable condition if clusterClass is not set
	}

	tkrName := cluster.Labels[runv1.LabelTKR]
	tkr := r.TKRResolver.Get(tkrName, &runv1.TanzuKubernetesRelease{}).(*runv1.TanzuKubernetesRelease)
	if tkr == nil {
		conditions.MarkUnknown(cluster, runv1.ConditionUpdatesAvailable, runv1.ReasonTKRNotFound, "TKR '%s' is not found", tkrName)
		return nil
	}
	updates, err := r.updatesAvailable(tkr, cluster, clusterClass)
	if err != nil {
		return err
	}
	if len(updates) == 0 {
		conditions.MarkFalse(cluster, runv1.ConditionUpdatesAvailable, runv1.ReasonAlreadyUpToDate, clusterv1.ConditionSeverityInfo, "")
		return nil
	}
	updatesAvailableCondition := conditions.TrueCondition(runv1.ConditionUpdatesAvailable)
	updatesAvailableCondition.Message = fmt.Sprintf("%v", updates)
	conditions.Set(cluster, updatesAvailableCondition)
	return nil
}

func (r *Reconciler) updatesAvailable(tkr *runv1.TanzuKubernetesRelease, cluster *clusterv1.Cluster, clusterClass *clusterv1.ClusterClass) ([]string, error) {
	result := make([]string, 0, 2)

	sv, err := utilversion.ParseSemantic(tkr.Spec.Kubernetes.Version)
	if err != nil {
		return nil, err
	}
	major, minor := sv.Major(), sv.Minor()

	for _, versionPrefix := range []string{
		vLabelMinor(major, minor),
		vLabelMinor(major, minor+1),
	} {
		updateVersion, err := r.findUpdateVersion(tkr, cluster, clusterClass, versionPrefix)
		if err != nil {
			return nil, err
		}
		if updateVersion != "" {
			result = append(result, updateVersion)
		}
	}

	return result, nil
}

func (r *Reconciler) findUpdateVersion(tkr *runv1.TanzuKubernetesRelease, cluster *clusterv1.Cluster, clusterClass *clusterv1.ClusterClass, versionPrefix string) (string, error) {
	query, err := resolution.ConstructQuery(versionPrefix, cluster, clusterClass)
	if err != nil {
		return "", err
	}
	tkrResult := r.TKRResolver.Resolve(*query)
	if tkrResult.ControlPlane.K8sVersion != "" && tkrResult.ControlPlane.TKRName != tkr.Name {
		if tkrResult.ControlPlane.K8sVersion == cluster.Spec.Topology.Version {
			resolvedTKR := tkrResult.ControlPlane.TKRsByK8sVersion[tkrResult.ControlPlane.K8sVersion][tkrResult.ControlPlane.TKRName]
			return resolvedTKR.Spec.Version, nil
		}
		return tkrResult.ControlPlane.K8sVersion, nil
	}
	return "", nil
}

func vLabelMinor(major, minor uint) string {
	return fmt.Sprintf("v%v.%v", major, minor)
}
