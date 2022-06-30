// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package compatibility provides the TKR Compatibility reconciler.
package compatibility

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
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
	"sigs.k8s.io/yaml"

	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/types"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/util/patchset"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/sets"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/version"
)

const LabelAdditionalTKRs = "run.tanzu.vmware.com/additional-compatible-tkrs"

const fieldTKRVersions = "tkrVersions"

type Reconciler struct {
	version.Compatibility

	Ctx    context.Context
	Log    logr.Logger
	Client client.Client
	Config Config
}

type Compatibility struct {
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
	compatibleSet, err := r.CompatibleVersions(ctx)
	if err != nil {
		return err
	}

	if compatibleSet.Has(tkr.Spec.Version) {
		conditions.MarkTrue(tkr, runv1.ConditionCompatible)
		return nil
	}
	conditions.MarkFalse(tkr, runv1.ConditionCompatible, "", clusterv1.ConditionSeverityWarning, "")
	return nil
}

func (c *Compatibility) CompatibleVersions(ctx context.Context) (sets.StringSet, error) {
	compatibleTKRVersions, err := c.getMCCompatibleTKRVersions(ctx)
	if err != nil {
		return nil, err
	}

	additionalTKRVersions, err := c.getAdditionalCompatibleTKRVersions(ctx)
	if err != nil {
		return nil, err
	}

	return compatibleTKRVersions.Union(additionalTKRVersions), nil
}

func (c *Compatibility) getMCCompatibleTKRVersions(ctx context.Context) (sets.StringSet, error) {
	mgmtClusterVersion, err := c.getManagementClusterVersion(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get the management cluster info")
	}
	if mgmtClusterVersion == "" {
		return sets.Strings(), nil
	}

	metadata, err := c.compatibilityMetadata(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get BOM compatibility metadata")
	}
	if metadata == nil {
		return sets.Strings(), nil
	}

	for _, mgmtVersion := range metadata.ManagementClusterVersions {
		if mgmtClusterVersion == mgmtVersion.TKGVersion {
			return sets.Strings(mgmtVersion.SupportedKubernetesVersions...), nil
		}
	}
	return sets.Strings(), nil
}

// getManagementClusterVersion get the version of the management cluster
func (c *Compatibility) getManagementClusterVersion(ctx context.Context) (string, error) {
	clusterList := &clusterv1.ClusterList{}
	if err := c.Client.List(ctx, clusterList, client.HasLabels{constants.ManagementClusterRoleLabel}); err != nil {
		return "", errors.Wrap(err, "failed to list clusters")
	}

	for i := range clusterList.Items {
		if tkgVersion, ok := clusterList.Items[i].Annotations[constants.TKGVersionKey]; ok {
			return tkgVersion, nil
		}
	}

	c.Log.Info("Could not find the Cluster resource with needed metadata",
		"label", constants.ManagementClusterRoleLabel, "annotation", constants.TKGVersionKey)

	return "", nil
}

func (c *Compatibility) compatibilityMetadata(ctx context.Context) (*types.CompatibilityMetadata, error) {
	cm := &corev1.ConfigMap{}
	cmObjectKey := client.ObjectKey{Namespace: c.Config.TKRNamespace, Name: constants.BOMMetadataConfigMapName}
	if err := c.Client.Get(ctx, cmObjectKey, cm); err != nil {
		err = kerrors.FilterOut(err, apierrors.IsNotFound)
		return nil, err
	}

	metadataContent, ok := cm.BinaryData[constants.BOMMetadataCompatibilityKey]
	if !ok {
		c.Log.Error(errors.New("compatibility key not found in bom-metadata ConfigMap"), "This.")
		return nil, nil
	}

	var metadata types.CompatibilityMetadata
	if err := yaml.Unmarshal(metadataContent, &metadata); err != nil {
		c.Log.Error(err, "Error parsing compatibility data", "ConfigMap", cmObjectKey,
			"key", fmt.Sprintf("binaryData.%s", constants.BOMMetadataCompatibilityKey))
		return nil, nil
	}
	return &metadata, nil
}

func (c *Compatibility) getAdditionalCompatibleTKRVersions(ctx context.Context) (sets.StringSet, error) {
	cmList := &corev1.ConfigMapList{}
	if err := c.Client.List(ctx, cmList, client.InNamespace(c.Config.TKRNamespace), client.HasLabels{LabelAdditionalTKRs}); err != nil {
		return nil, errors.Wrap(err, "error listing additional TKR ConfigMaps")
	}
	return c.additionalTKRVersions(cmList)
}

func (c *Compatibility) additionalTKRVersions(cmList *corev1.ConfigMapList) (sets.StringSet, error) {
	result := sets.StringSet{}
	for i := range cmList.Items {
		cm := &cmList.Items[i]
		if !cm.DeletionTimestamp.IsZero() {
			continue
		}
		var tkrVersions []string
		if err := yaml.Unmarshal([]byte(cm.Data[fieldTKRVersions]), &tkrVersions); err != nil {
			c.Log.Error(err, "Error reading tkrVersions", "ConfigMap", fmt.Sprintf("%s/%s", cm.Namespace, cm.Name))
			continue
		}
		result.Add(tkrVersions...)
	}
	return result, nil
}
