// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package compatibility provides the TKR Compatibility reconciler.
package compatibility

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/types"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/util/patchset"
)

type Reconciler struct {
	Ctx    context.Context
	Log    logr.Logger
	Client client.Client

	Config Config
}

type Config struct {
	TKRNamespace string
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&runv1.TanzuKubernetesRelease{}).
		Watches(
			&source.Kind{Type: &corev1.ConfigMap{}},
			handler.EnqueueRequestsFromMapFunc(r.enqueueAllTKRs),
			builder.WithPredicates(predicate.NewPredicateFuncs(func(o client.Object) bool {
				return o.GetNamespace() == r.Config.TKRNamespace && o.GetName() == constants.BOMMetadataConfigMapName
			}))).
		Watches(
			&source.Kind{Type: &clusterv1.Cluster{}},
			handler.EnqueueRequestsFromMapFunc(r.enqueueAllTKRs),
			builder.WithPredicates(hasManagementClusterRoleLabel, predicate.AnnotationChangedPredicate{})).
		Named("tkr_compatibility").
		Complete(r)
}

// enqueueAllTKRs returns reconcile.Request for all existing TKRs.
func (r *Reconciler) enqueueAllTKRs(o client.Object) []ctrl.Request {
	r.Log.Info("enqueueing all TKRs, triggered by object",
		"kind", o.GetObjectKind().GroupVersionKind().Kind, "namespace", o.GetNamespace(), "name", o.GetName())

	tkrs := &runv1.TanzuKubernetesReleaseList{}
	if err := r.Client.List(r.Ctx, tkrs); err != nil {
		r.Log.Error(err, "error listing TKRs")
		return nil
	}
	result := make([]ctrl.Request, len(tkrs.Items))
	for i := range tkrs.Items {
		result[i].NamespacedName.Name = tkrs.Items[i].Name
	}
	return result
}

var hasManagementClusterRoleLabel = func() predicate.Predicate {
	selector, _ := labels.Parse(constants.ManagementClusterRoleLabel)
	return predicate.NewPredicateFuncs(func(o client.Object) bool {
		return selector.Matches(labels.Set(o.GetLabels()))
	})
}()

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, retErr error) {
	tkr := &runv1.TanzuKubernetesRelease{}
	if err := r.Client.Get(ctx, req.NamespacedName, tkr); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil // do nothing if the TKR does not exist
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

	ps.Add(tkr)

	return ctrl.Result{}, r.updateTKRCompatibleCondition(ctx, tkr)
}

func (r *Reconciler) updateTKRCompatibleCondition(ctx context.Context, tkr *runv1.TanzuKubernetesRelease) error {
	compatibleSet, err := r.getCompatibleSet(ctx)
	if err != nil {
		return err
	}

	if _, isCompatible := compatibleSet[tkr.Spec.Version]; isCompatible {
		conditions.MarkTrue(tkr, runv1.ConditionCompatible)
		return nil
	}
	conditions.MarkFalse(tkr, runv1.ConditionCompatible, "", clusterv1.ConditionSeverityWarning, "")
	return nil
}

func (r *Reconciler) getCompatibleSet(ctx context.Context) (map[string]struct{}, error) {
	mgmtClusterVersion, err := r.getManagementClusterVersion(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get the management cluster info")
	}

	metadata, err := r.compatibilityMetadata(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get BOM compatibility metadata")
	}

	for _, mgmtVersion := range metadata.ManagementClusterVersions {
		if mgmtClusterVersion == mgmtVersion.TKGVersion {
			return stringSet(mgmtVersion.SupportedKubernetesVersions), nil
		}
	}

	return stringSet(nil), nil
}

func stringSet(ss []string) map[string]struct{} {
	result := make(map[string]struct{}, len(ss))
	for _, s := range ss {
		result[s] = struct{}{}
	}
	return result
}

// getManagementClusterVersion get the version of the management cluster
func (r *Reconciler) getManagementClusterVersion(ctx context.Context) (string, error) {
	clusterList := &clusterv1.ClusterList{}
	if err := r.Client.List(ctx, clusterList, client.HasLabels{constants.ManagementClusterRoleLabel}); err != nil {
		return "", errors.Wrap(err, "failed to list clusters")
	}

	for i := range clusterList.Items {
		if tkgVersion, ok := clusterList.Items[i].Annotations[constants.TKGVersionKey]; ok {
			return tkgVersion, nil
		}
	}

	return "", errors.New("failed to get management cluster info")
}

func (r *Reconciler) compatibilityMetadata(ctx context.Context) (*types.CompatibilityMetadata, error) {
	cm := &corev1.ConfigMap{}
	cmObjectKey := client.ObjectKey{Namespace: r.Config.TKRNamespace, Name: constants.BOMMetadataConfigMapName}
	if err := r.Client.Get(ctx, cmObjectKey, cm); err != nil {
		return nil, err
	}

	metadataContent, ok := cm.BinaryData[constants.BOMMetadataCompatibilityKey]
	if !ok {
		return nil, errors.New("compatibility key not found in bom-metadata ConfigMap")
	}

	var metadata types.CompatibilityMetadata
	if err := yaml.Unmarshal(metadataContent, &metadata); err != nil {
		return nil, err
	}
	return &metadata, nil
}
