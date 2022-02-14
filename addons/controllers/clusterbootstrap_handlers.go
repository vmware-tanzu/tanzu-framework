// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	clusterv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	addontypes "github.com/vmware-tanzu/tanzu-framework/addons/pkg/types"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	runtanzuv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

// TKRToClusters returns a list of Requests with Cluster ObjectKey for
func (r *ClusterBootstrapReconciler) TKRToClusters(o client.Object) []ctrl.Request {
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

func (r *ClusterBootstrapReconciler) ClusterBootstrapToClusters(o client.Object) []ctrl.Request {
	bootstrap, ok := o.(*runtanzuv1alpha3.ClusterBootstrap)
	if !ok {
		r.Log.Error(errors.New("invalid type"),
			"Expected to receive ClusterBootstrap resource",
			"actualType", fmt.Sprintf("%T", o))
		return nil
	}

	log := r.Log.WithValues(constants.ClusterBootstrapNameLogKey, bootstrap.Name)

	log.V(4).Info("Mapping ClusterBootstrap to cluster")

	cluster := &clusterv1beta1.Cluster{}

	clusterName := bootstrap.Name
	for _, ownerRef := range o.GetOwnerReferences() {
		if ownerRef.APIVersion == clusterv1beta1.GroupVersion.String() {
			clusterName = ownerRef.Name
			break
		}
	}

	if err := r.Client.Get(r.context, client.ObjectKey{Namespace: bootstrap.Namespace, Name: clusterName}, cluster); err != nil {
		log.Error(err, "Error getting cluster using ClusterBootstrap")
		return nil
	}
	return []ctrl.Request{{NamespacedName: client.ObjectKeyFromObject(cluster)}}
}

// SecretsToClusters is the Map Function for watching Secrets that filters the events on
// objects of type secret and returns requests for reconcile if required
func (r *ClusterBootstrapReconciler) SecretsToClusters(o client.Object) []ctrl.Request {
	secret, ok := o.(*corev1.Secret)
	if !ok {
		r.Log.Error(errors.New("invalid type"),
			"Expected to receive Secret resource",
			"actualType", fmt.Sprintf("%T", o))
		return nil
	}

	log := r.Log.WithValues(constants.SecretNameLogKey, secret.Name)
	log.V(4).Info("Mapping secrets to cluster")

	// Here we filter based on two criteria
	// 1. Secrets having the type ClusterBootstrapManagedSecret, OR
	// 2. Secrets with ClusterNameLabel set
	// For secrets that are cloned by ClusterBootstrap, we set the type and the first filter is used. We also
	// set the ClusterNameLabel on these cloned secrets
	// For other secrets such as those that we get from provider status (extensible provider model), the second
	// filter is used because it is not possible to patch the Type of these secrets because the Type may be immutable
	//
	// Need to confirm: We can just use the second filter that is based on the cluster label for all secrets.
	if secret.Type == constants.ClusterBootstrapManagedSecret {
		for _, ownerRef := range o.GetOwnerReferences() {
			if ownerRef.APIVersion == clusterv1beta1.GroupVersion.String() {
				return []ctrl.Request{{NamespacedName: types.NamespacedName{Namespace: o.GetNamespace(), Name: ownerRef.Name}}}
			}
		}
	}
	cluster := &clusterv1beta1.Cluster{}
	clusterName := ""
	if secret.GetLabels() != nil {
		clusterName = secret.GetLabels()[addontypes.ClusterNameLabel]
		if clusterName != "" {
			if err := r.Client.Get(r.context, client.ObjectKey{Namespace: secret.Namespace, Name: clusterName}, cluster); err != nil {
				log.Error(err, "Error getting cluster using Secret")
				return nil
			}
			return []ctrl.Request{{NamespacedName: client.ObjectKeyFromObject(cluster)}}
		}
	}

	return nil
}

func (r *ClusterBootstrapReconciler) ProviderToClusters(o client.Object) []ctrl.Request {
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
