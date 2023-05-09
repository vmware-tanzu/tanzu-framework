// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package conditions

import (
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"

	corev1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/core/v1alpha2"
)

func NewResourceExistenceConditionFunc(dynamicClient *dynamic.DynamicClient, discoveryClient *discovery.DiscoveryClient) func(context.Context, *corev1alpha2.ResourceExistenceCondition) (corev1alpha2.ReadinessConditionState, string) {
	return func(ctx context.Context, c *corev1alpha2.ResourceExistenceCondition) (corev1alpha2.ReadinessConditionState, string) {
		var gv schema.GroupVersion
		var err error

		if gv, err = schema.ParseGroupVersion(c.APIVersion); err != nil {
			return corev1alpha2.ConditionFailureState, err.Error()
		}

		groupResources, err := restmapper.GetAPIGroupResources(discoveryClient)
		if err != nil {
			// TODO: These are retriable failures; retries should be done instead of failing in the first attempt
			return corev1alpha2.ConditionFailureState, err.Error()
		}

		restMapper := restmapper.NewDiscoveryRESTMapper(groupResources)
		restMapping, err := restMapper.RESTMapping(schema.GroupKind{
			Group: gv.Group,
			Kind:  c.Kind,
		}, gv.Version)

		if err != nil {
			return corev1alpha2.ConditionFailureState, err.Error()
		}

		if c.Namespace == nil {
			_, err = dynamicClient.Resource(restMapping.Resource).Get(ctx, c.Name, v1.GetOptions{})
		} else {
			_, err = dynamicClient.Resource(restMapping.Resource).Namespace(*c.Namespace).
				Get(context.TODO(), c.Name, v1.GetOptions{})
		}
		if err != nil {
			return corev1alpha2.ConditionFailureState, err.Error()
		}
		return corev1alpha2.ConditionSuccessState, ""
	}
}
