// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
)

// SummarizeAppConditions summarizes the provided conditions slice into a single condition with the following logic:
// - If there be any 'Reconciling' condition type with status 'True', the summary condition type will be 'Reconciling'
// - Otherwise, if there be any 'ReconcileFailed' condition type with status 'True', the summary condition type will be 'ReconcileFailed'
// - Otherwise, if there be any 'ReconcileSucceeded' condition type with status 'True', the summary condition type will be 'ReconcileSucceeded'
// - Otherwise, if there be any 'Deleting' condition type with status 'True', the summary condition type will be 'Deleting'
// - Otherwise, if there be any 'DeleteFailed' condition type with status 'True', the summary condition type will be 'DeleteFailed'
// - Otherwise, the condition type is unknown and nil will be returned
// Note that ReconcileFailed|Reconciling|ReconcileSucceeded|Deleting|DeleteFailed|DeleteSucceeded are mutually exclusive
func SummarizeAppConditions(conditions []v1alpha1.AppCondition) *v1alpha1.AppCondition {
	wantedCondTypes := []v1alpha1.AppConditionType{
		v1alpha1.Reconciling,
		v1alpha1.ReconcileFailed,
		v1alpha1.ReconcileSucceeded,
		v1alpha1.Deleting,
		v1alpha1.DeleteFailed,
	}
	for _, wantedCondType := range wantedCondTypes {
		for _, cond := range conditions {
			if cond.Type == wantedCondType && cond.Status == corev1.ConditionTrue {
				return &cond
			}
		}
	}
	return nil
}

// GetKappUsefulErrorMessage extracts the relevant portion from UsefulErrorMessage
func GetKappUsefulErrorMessage(s string) string {
	var errString string
	n := len(s)
	i := strings.Index(s, "kapp: Error")
	if i != -1 {
		errString = s[i:n]
	}

	return errString
}

// HasSameState returns true if a ClusterBootstrap condition has the same state of another; state is defined
// by the union of following fields: Type, Status, Message, Reason (it excludes LastTransitionTime).
func HasSameState(i, j *clusterapiv1beta1.Condition) bool {
	return i.Type == j.Type &&
		i.Status == j.Status &&
		i.Message == j.Message &&
		i.Reason == j.Reason
}
