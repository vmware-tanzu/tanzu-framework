// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	clusterv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	runtanzuv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

// TKRToClusters returns a list of Requests with Cluster ObjectKey for
func (r *TanzuClusterBootstrapReconciler) TKRToClusters(o client.Object) []ctrl.Request {
	tkr, ok := o.(*runtanzuv1alpha3.TanzuKubernetesRelease)
	if !ok {
		r.Log.Error(errors.New("invalid type"),
			"Expected to receive TKR resource",
			"actualType", fmt.Sprintf("%T", o))
		return nil
	}

	log := r.Log.WithValues(constants.TKRNameLogKey, tkr.Name)

	log.V(4).Info("Mapping TKR to cluster")

	clustersList := &clusterv1beta1.ClusterList{}

	if err := r.Client.List(r.context, clustersList, client.MatchingLabels{constants.TKRLabel: tkr.Name}); err != nil {
		log.Error(err, "Error getting clusters using TKR")
		return nil
	}

	var clusters []*clusterv1beta1.Cluster
	for i := range clustersList.Items {
		clusters = append(clusters, &clustersList.Items[i])
	}

	return util.ClustersToRequests(clusters, log)
}

func (r *TanzuClusterBootstrapReconciler) TanzuClusterBootstrapToClusters(o client.Object) []ctrl.Request {
	bootstrap, ok := o.(*runtanzuv1alpha3.TanzuClusterBootstrap)
	if !ok {
		r.Log.Error(errors.New("invalid type"),
			"Expected to receive TanzuClusterBootstrap resource",
			"actualType", fmt.Sprintf("%T", o))
		return nil
	}

	log := r.Log.WithValues(constants.TanzuClusterBootstrapNameLogKey, bootstrap.Name)

	log.V(4).Info("Mapping TanzuClusterBootstrap to cluster")

	cluster := &clusterv1beta1.Cluster{}

	clusterName := bootstrap.Name
	for _, ownerRef := range o.GetOwnerReferences() {
		if ownerRef.APIVersion == clusterv1beta1.GroupVersion.String() {
			clusterName = ownerRef.Name
			break
		}
	}

	if err := r.Client.Get(r.context, client.ObjectKey{Namespace: bootstrap.Namespace, Name: clusterName}, cluster); err != nil {
		log.Error(err, "Error getting cluster using TanzuClusterBootstrap")
		return nil
	}
	return []ctrl.Request{{NamespacedName: client.ObjectKeyFromObject(cluster)}}
}

func (r *TanzuClusterBootstrapReconciler) ProviderToClusters(o client.Object) []ctrl.Request {
	if o == nil {
		return nil
	}
	for _, ownerRef := range o.GetOwnerReferences() {
		if ownerRef.APIVersion == clusterv1beta1.GroupVersion.String() {
			return []ctrl.Request{{NamespacedName: types.NamespacedName{Namespace: o.GetNamespace(), Name: ownerRef.Name}}}
		}
	}
	return nil
}
