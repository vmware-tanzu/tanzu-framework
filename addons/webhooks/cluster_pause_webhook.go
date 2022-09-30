// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	runtanzuv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

var clusterpauselog = logf.Log.WithName("cluster-pause-webhook")

// +kubebuilder:webhook:verbs=create;update,path=/mutate-cluster-x-k8s-io-v1beta1-cluster,mutating=true,failurePolicy=fail,matchPolicy=Equivalent,groups=cluster.x-k8s.io,resources=clusters,versions=v1beta1,name=default.cluster.cluster.x-k8s.io,sideEffects=None,admissionReviewVersions=v1;v1beta1

// ClusterPause implements a validating and defaulting webhook for Cluster.
type ClusterPause struct {
	Client client.Client
}

// SetupWebhookWithManager sets up Cluster webhooks.
func (wh *ClusterPause) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&clusterv1.Cluster{}).
		WithDefaulter(wh).
		Complete()
}

var (
	_ webhook.CustomDefaulter = &ClusterPause{}
)

// Default satisfies the defaulting webhook interface.
func (wh *ClusterPause) Default(ctx context.Context, obj runtime.Object) error { //nolint
	var tkrVersion, currentTkrVersion string
	var tkrLabelFound, ok bool

	cluster, ok := obj.(*clusterv1.Cluster)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a Cluster but got a %T", obj))
	}

	if cluster.Labels == nil {
		return nil
	}

	if tkrVersion, tkrLabelFound = cluster.Labels[runtanzuv1alpha3.LabelTKR]; !tkrLabelFound {
		return nil
	}

	currentCluster := &clusterv1.Cluster{}
	// Try to get the current cluster CR, so we can compare the version
	key := client.ObjectKey{Name: cluster.Name, Namespace: cluster.Namespace}
	if err := wh.Client.Get(ctx, key, currentCluster); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	if currentCluster.Labels != nil {
		currentTkrVersion, tkrLabelFound = currentCluster.Labels[runtanzuv1alpha3.LabelTKR]
	}

	// want to verify the current TKR object (before upgrade) is not a legacy one
	if tkrLabelFound && currentTkrVersion != "" {
		currentTKR := &runtanzuv1alpha3.TanzuKubernetesRelease{}
		if err := wh.Client.Get(ctx, client.ObjectKey{Name: currentTkrVersion}, currentTKR); err != nil {
			if apierrors.IsNotFound(err) {
				return nil
			}
			clusterpauselog.Error(err, "unable to fetch TKR object", "name", currentTkrVersion)
			return err
		}
		if currentTKR != nil && currentTKR.Labels != nil {
			if _, ok := currentTKR.Labels[constants.TKRLabelLegacyClusters]; ok {
				return nil
			}
		}
	}

	// Add pause to cluster if the cluster.Labels["run.tanzu.vmware.com/tkr"] changes
	// The cluster pause state will be unset by ClusterBootstrap controller after it rolls out package updates
	if currentCluster.Labels == nil || !tkrLabelFound || currentTkrVersion != tkrVersion {
		cluster.Spec.Paused = true
		if cluster.Annotations == nil {
			cluster.Annotations = map[string]string{}
		}
		// Use the desired TKR version as label value, ClusterBootstrap will unset
		cluster.Annotations[constants.ClusterPauseLabel] = tkrVersion
		clusterpauselog.Info(fmt.Sprintf("set '%s' annotation to '%s' for cluster '%s'", constants.ClusterPauseLabel, tkrVersion, cluster.Name))
	}

	return nil
}
