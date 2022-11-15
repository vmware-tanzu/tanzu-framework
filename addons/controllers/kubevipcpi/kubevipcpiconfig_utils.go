// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	clusterapiutil "sigs.k8s.io/cluster-api/util"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	kvcpiv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/addonconfigs/cpi/v1alpha1"
)

// ClusterToKubevipCPIConfig returns a list of Requests with KubevipCPIConfig ObjectKey based on Cluster events
func (r *KubevipCPIConfigReconciler) ClusterToKubevipCPIConfig(o client.Object) []ctrl.Request {
	cluster, ok := o.(*clusterapiv1beta1.Cluster)
	if !ok {
		r.Log.Error(errors.New("invalid type"),
			"Expected to receive Cluster resource",
			"actualType", fmt.Sprintf("%T", o))
		return nil
	}

	r.Log.V(4).Info("Mapping Cluster to KubevipCPIConfig")

	cs := &kvcpiv1alpha1.KubevipCPIConfigList{}
	_ = r.List(context.Background(), cs)

	requests := []ctrl.Request{}
	for i := 0; i < len(cs.Items); i++ {
		config := &cs.Items[i]
		if config.Namespace == cluster.Namespace {
			// avoid enqueuing reconcile requests for template KubevipCPIConfig CRs in event handler of Cluster CR
			if _, ok := config.Annotations[constants.TKGAnnotationTemplateConfig]; ok && config.Namespace == r.Config.SystemNamespace {
				continue
			}

			// corresponding KubevipCPIConfig should have following ownerRef
			ownerReference := metav1.OwnerReference{
				APIVersion: clusterapiv1beta1.GroupVersion.String(),
				Kind:       cluster.Kind,
				Name:       cluster.Name,
				UID:        cluster.UID,
			}
			if clusterapiutil.HasOwnerRef(config.OwnerReferences, ownerReference) {
				r.Log.V(4).Info("Adding KubevipCPIConfig for reconciliation",
					constants.NamespaceLogKey, config.Namespace, constants.NameLogKey, config.Name)

				requests = append(requests, ctrl.Request{
					NamespacedName: clusterapiutil.ObjectKey(config),
				})
			}
		}
	}

	return requests
}

// mapKubevipCPIConfigToDataValues generates CPI data values for non-paravirtual modes
func (r *KubevipCPIConfigReconciler) mapKubevipCPIConfigToDataValues( // nolint
	ctx context.Context,
	kubevipCPIConfig *kvcpiv1alpha1.KubevipCPIConfig, cluster *clusterapiv1beta1.Cluster) (KubevipCPIDataValues, error,
) { // nolint:whitespace
	// allow API user to override the derived values if he/she specified fields in the KubevipCPIConfig
	dataValue := &KubevipCPIDataValues{}
	config := kubevipCPIConfig.Spec
	dataValue.LoadbalancerCIDRs = tryParseString(dataValue.LoadbalancerCIDRs, config.LoadbalancerCIDRs)
	dataValue.LoadbalancerIPRanges = tryParseString(dataValue.LoadbalancerIPRanges, config.LoadbalancerIPRanges)

	return *dataValue, nil
}

// tryParseString tries to convert a string pointer and return its value, if not nil
func tryParseString(src string, sub *string) string {
	if sub != nil {
		return *sub
	}
	return src
}
