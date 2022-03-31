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
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
)

// +kubebuilder:webhook:verbs=create;update,path=/mutate-cluster-x-k8s-io-v1beta1-cluster,mutating=true,failurePolicy=fail,matchPolicy=Equivalent,groups=cluster.x-k8s.io,resources=clusters,versions=v1beta1,name=default.cluster.cluster.x-k8s.io,sideEffects=None,admissionReviewVersions=v1;v1beta1

// ClusterPause implements a validating and defaulting webhook for Cluster.
type ClusterPause struct {
	Client client.Reader
}

// SetupWebhookWithManager sets up Cluster webhooks.
func (wh *ClusterPause) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&clusterv1.Cluster{}).
		WithDefaulter(wh).
		Complete()
}

var _ webhook.CustomDefaulter = &ClusterPause{}

// Default satisfies the defaulting webhook interface.
func (wh *ClusterPause) Default(ctx context.Context, obj runtime.Object) error {
	cluster, ok := obj.(*clusterv1.Cluster)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a Cluster but got a %T", obj))
	}

	// Try to get the current cluster CR so we can compare the version
	currentCluster := &clusterv1.Cluster{}
	key := client.ObjectKey{Name: cluster.Name, Namespace: cluster.Namespace}
	err := wh.Client.Get(ctx, key, currentCluster)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	if cluster.Spec.Topology != nil {
		// Add pause to cluster if the topology.version changes
		// The cluster pause state will be unset by ClusterBootstrap controller after it rolls out package updates
		if currentCluster.Spec.Topology == nil || cluster.Spec.Topology.Version != currentCluster.Spec.Topology.Version {
			cluster.Spec.Paused = true
			if cluster.Annotations == nil {
				cluster.Annotations = map[string]string{}
			}
			// Use the desired TKR version as label value, ClusterBootstrap will unset
			cluster.Annotations[constants.ClusterPauseLabel] = cluster.Spec.Topology.Version
		}
	}

	return nil
}
