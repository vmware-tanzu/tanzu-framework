// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	clusterv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	controlplanev1alpha3 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1alpha3"
	clusterapiutil "sigs.k8s.io/cluster-api/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	runtanzuv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
)

// TKRToClusters returns the clusters using TKR
func (r *AddonReconciler) TKRToClusters(o client.Object) []ctrl.Request {
	var tkr *runtanzuv1alpha1.TanzuKubernetesRelease

	r.Log.V(4).Info("TKr to clusters handler")

	switch obj := o.(type) {
	case *runtanzuv1alpha1.TanzuKubernetesRelease:
		tkr = obj
	default:
		r.Log.Error(errors.New("invalid type"),
			"Expected to receive TKr resource",
			"actualType", fmt.Sprintf("%T", o))
		return nil
	}

	log := r.Log.WithValues(constants.TKRNameLogKey, tkr.Name)

	log.V(4).Info("Mapping TKr to cluster")

	clusters, err := util.GetClustersByTKR(context.TODO(), r.Client, tkr)
	if err != nil {
		log.Error(err, "Error getting clusters using TKr")
		return nil
	}

	return r.clustersToRequests(clusters, log)
}

func (r *AddonReconciler) clustersToRequests(clusters []*clusterv1alpha3.Cluster, log logr.Logger) []ctrl.Request {
	var requests []ctrl.Request

	for _, cluster := range clusters {
		log.V(4).Info("Adding cluster for reconciliation",
			constants.ClusterNamespaceLogKey, cluster.Namespace, constants.ClusterNameLogKey, cluster.Name)

		requests = append(requests, ctrl.Request{
			NamespacedName: clusterapiutil.ObjectKey(cluster),
		})
	}

	return requests
}

// AddonSecretToClusters returns the clusters on which the addon needs to be installed
func (r *AddonReconciler) AddonSecretToClusters(o client.Object) []ctrl.Request {
	var secret *corev1.Secret

	r.Log.V(4).Info("Addon secret to clusters handler")

	switch obj := o.(type) {
	case *corev1.Secret:
		secret = obj
	default:
		r.Log.Error(errors.New("invalid type"),
			"Expected to receive Secret resource",
			"actualType", fmt.Sprintf("%T", o))
		return nil
	}

	log := r.Log.WithValues(constants.AddonSecretNamespaceLogKey, secret.Namespace, constants.AddonSecretNameLogKey, secret.Name)

	log.V(4).Info("Mapping Addon Secret to cluster")

	clusterName := util.GetClusterNameFromAddonSecret(secret)
	if clusterName == "" {
		log.Info("Cluster name label not found on secret")
	}

	cluster, err := util.GetClusterByName(context.TODO(), r.Client, secret.Namespace, clusterName)
	if err != nil {
		log.Error(err, "Error getting cluster object",
			constants.ClusterNamespaceLogKey, secret.Namespace, constants.ClusterNameLogKey, clusterName)
		return nil
	}

	if cluster == nil {
		log.Info("Cluster not found for addon secret")
		return nil
	}

	if !cluster.GetDeletionTimestamp().IsZero() {
		log.Info("Cluster is getting deleted, so skipping request for cluster",
			constants.ClusterNamespaceLogKey, secret.Namespace, constants.ClusterNameLogKey, clusterName)
		return nil
	}

	log.V(4).Info("Adding cluster for reconciliation",
		constants.ClusterNamespaceLogKey, cluster.Namespace, constants.ClusterNameLogKey, cluster.Name)

	return []ctrl.Request{{
		NamespacedName: clusterapiutil.ObjectKey(cluster),
	}}
}

// BOMConfigMapToClusters returns the clusters using the BOM
func (r *AddonReconciler) BOMConfigMapToClusters(o client.Object) []ctrl.Request {
	var configmap *corev1.ConfigMap

	r.Log.V(4).Info("BOM configmap to clusters handler")

	switch obj := o.(type) {
	case *corev1.ConfigMap:
		configmap = obj
	default:
		r.Log.Error(errors.New("invalid type"),
			"Expected to receive ConfigMap resource",
			"actualType", fmt.Sprintf("%T", o))
		return nil
	}

	log := r.Log.WithValues(constants.BOMNamespaceLogKey, configmap.Namespace, constants.BOMNameLogKey, configmap.Name)
	log.V(4).Info("Mapping BOM configmap to cluster")

	tkrName := util.GetTKRNameFromBOMConfigMap(configmap)
	if tkrName == "" {
		log.Info("TKr label not found on BOM configmap")
		return nil
	}

	tkr, err := util.GetTKRByName(context.TODO(), r.Client, tkrName)
	if err != nil || tkr == nil {
		log.Error(err, "Error getting TKr", constants.TKRNameLogKey, tkrName)
		return nil
	}

	clusters, err := util.GetClustersByTKR(context.TODO(), r.Client, tkr)
	if err != nil {
		log.Error(err, "Error getting clusters using TKr", constants.TKRNameLogKey, tkr.Name)
		return nil
	}

	return r.clustersToRequests(clusters, log)
}

// KubeadmControlPlaneToClusters returns the cluster where kcp is present
func (r *AddonReconciler) KubeadmControlPlaneToClusters(o client.Object) []ctrl.Request {
	var kcp *controlplanev1alpha3.KubeadmControlPlane

	r.Log.V(4).Info("Kubeadm control plane to clusters handler")

	switch obj := o.(type) {
	case *controlplanev1alpha3.KubeadmControlPlane:
		kcp = obj
	default:
		r.Log.Error(errors.New("invalid type"),
			"Expected to receive kubeadm control plane resource",
			"actualType", fmt.Sprintf("%T", o))
		return nil
	}

	log := r.Log.WithValues(constants.KCPNamespaceLogKey, kcp.Namespace, constants.KCPNameLogKey, kcp.Name)

	log.V(4).Info("Mapping kubeadm control plane to cluster")

	cluster, err := util.GetOwnerCluster(context.TODO(), r.Client, &kcp.ObjectMeta)
	if err != nil || cluster == nil {
		log.Error(err, "Failed to get cluster owning kcp")
		return nil
	}

	if cluster == nil {
		log.Info("Cluster not found for kcp")
		return nil
	}

	log.V(4).Info("Adding cluster for reconciliation",
		constants.ClusterNamespaceLogKey, cluster.Namespace, constants.ClusterNameLogKey, cluster.Name)

	return []ctrl.Request{{
		NamespacedName: clusterapiutil.ObjectKey(cluster),
	}}
}
