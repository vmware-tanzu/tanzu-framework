// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"errors"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	clusterapiutil "sigs.k8s.io/cluster-api/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	runtanzuv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

// ClusterToKappControllerConfig returns a list of Requests with KappControllerConfig ObjectKey
func (r *KappControllerConfigReconciler) ClusterToKappControllerConfig(o client.Object) []ctrl.Request {
	cluster, ok := o.(*clusterv1beta1.Cluster)
	if !ok {
		r.Log.Error(errors.New("invalid type"),
			"Expected to receive Cluster resource",
			"actualType", fmt.Sprintf("%T", o))
		return nil
	}

	log := r.Log.WithValues(constants.ClusterNameLogKey, cluster.Name)

	log.V(4).Info("Mapping cluster to KappControllerConfig")

	KappControllerConfigList := &runtanzuv1alpha3.KappControllerConfigList{}

	if err := r.Client.List(context.Background(), KappControllerConfigList); err != nil {
		log.Error(err, "Error listing KappControllerConfig")
		return nil
	}

	var requests []ctrl.Request
	for i := range KappControllerConfigList.Items {
		config := &KappControllerConfigList.Items[i]
		if config.Namespace == cluster.Namespace {
			// corresponding kappControllerConfig should have following ownerRef
			ownerReference := metav1.OwnerReference{
				APIVersion: clusterv1beta1.GroupVersion.String(),
				Kind:       cluster.Kind,
				Name:       cluster.Name,
				UID:        cluster.UID,
			}

			if clusterapiutil.HasOwnerRef(config.OwnerReferences, ownerReference) || config.Name == cluster.Name {
				log.V(4).Info("Adding KappControllerConfig for reconciliation",
					constants.NamespaceLogKey, config.Namespace, constants.NameLogKey, config.Name)

				requests = append(requests, ctrl.Request{
					NamespacedName: clusterapiutil.ObjectKey(config),
				})
			}
		}
	}

	return requests
}
