// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	clusterapiutil "sigs.k8s.io/cluster-api/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/config"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
)

type CPIConfigAction func(client.Object)

// performActionForCPIConfigIfOwnedByCluster performs the specified action for the cluster owns the CPIConfig
// For a CPIConfig to be owned by a Cluster, it should
//  1. in the same namespace as the cluster
//  2. not a template CPIConfig, a.k.a. has annotation tkg.tanzu.vmware.com/template-config
//  3. has the Cluster as its OwnerReference
func performActionForCPIConfigIfOwnedByCluster(cpiConfig client.Object, cluster *clusterapiv1beta1.Cluster,
	controllerConfig config.ConfigControllerConfig, action CPIConfigAction) {

	namespace := cpiConfig.GetNamespace()
	annotations := cpiConfig.GetAnnotations()
	ownerReferences := cpiConfig.GetOwnerReferences()

	if namespace == cluster.Namespace {
		// avoid enqueuing reconcile requests for template CPIConfig CRs in event handler of Cluster CR
		if _, ok := annotations[constants.TKGAnnotationTemplateConfig]; ok && cpiConfig.GetNamespace() == controllerConfig.SystemNamespace {
			return
		}
	}

	// corresponding CPIConfig should have following ownerRef
	ownerReference := metav1.OwnerReference{
		APIVersion: clusterapiv1beta1.GroupVersion.String(),
		Kind:       cluster.Kind,
		Name:       cluster.Name,
		UID:        cluster.UID,
	}

	if clusterapiutil.HasOwnerRef(ownerReferences, ownerReference) {
		action(cpiConfig)
	}
}

// performEnqueueForCPIConfigIfOwnedByCluster enqueues the request for the cluster that owns this CPIConfig
func performEnqueueForCPIConfigIfOwnedByCluster(cpiConfig client.Object, cluster *clusterapiv1beta1.Cluster,
	controllerConfig config.ConfigControllerConfig, logger logr.Logger, requests []ctrl.Request) []ctrl.Request {

	performActionForCPIConfigIfOwnedByCluster(cpiConfig, cluster, controllerConfig, func(cpiConfig client.Object) {
		logger.V(4).Info("Adding "+cpiConfig.GetObjectKind().GroupVersionKind().Kind+" for reconciliation",
			constants.NamespaceLogKey, cpiConfig.GetNamespace(), constants.NameLogKey, cpiConfig.GetName())

		requests = append(requests, ctrl.Request{
			NamespacedName: clusterapiutil.ObjectKey(cpiConfig),
		})
	})

	return requests
}
