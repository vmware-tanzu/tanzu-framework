// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package conditions

import (
	"context"

	corev1 "k8s.io/api/core/v1"

	corev1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/core/v1alpha2"
	capabilitiesdiscovery "github.com/vmware-tanzu/tanzu-framework/capabilities/client/pkg/discovery"
)

// NewResourceExistenceConditionFunc returns a function for evaluating evaluate a ResourceExistenceCondition
func NewResourceExistenceConditionFunc() func(context.Context, *capabilitiesdiscovery.ClusterQueryClient, *corev1alpha2.ResourceExistenceCondition, string) (corev1alpha2.ReadinessConditionState, string) {
	return func(ctx context.Context, queryClient *capabilitiesdiscovery.ClusterQueryClient, c *corev1alpha2.ResourceExistenceCondition, conditionName string) (corev1alpha2.ReadinessConditionState, string) {
		if c == nil {
			return corev1alpha2.ConditionFailureState, "resourceExistenceCondition is not defined"
		}

		var err error
		var resourceToFind corev1.ObjectReference

		if c.Namespace == nil {
			resourceToFind = corev1.ObjectReference{
				Kind:       c.Kind,
				Name:       c.Name,
				APIVersion: c.APIVersion,
			}
		} else {
			resourceToFind = corev1.ObjectReference{
				Kind:       c.Kind,
				Name:       c.Name,
				Namespace:  *(c.Namespace),
				APIVersion: c.APIVersion,
			}
		}

		queryObject := capabilitiesdiscovery.Object(conditionName, &resourceToFind)
		ok, err := queryClient.PreparedQuery(queryObject)()
		if err != nil {
			return corev1alpha2.ConditionFailureState, err.Error()
		}

		if !ok {
			return corev1alpha2.ConditionFailureState, "resource not found"
		}
		return corev1alpha2.ConditionSuccessState, "resource found"
	}
}
