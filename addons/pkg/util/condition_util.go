// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	corev1 "k8s.io/api/core/v1"

	"github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	runtanzuv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

const UnknownCondition = v1alpha1.AppConditionType("UnknownCondition")

type state struct {
	stateExists bool
	status      corev1.ConditionStatus
}

func (s *state) setState(cond v1alpha1.AppCondition) {
	s.stateExists = true
	if s.status == "" || cond.Status == corev1.ConditionTrue {
		s.status = cond.Status
	} else if s.status == corev1.ConditionUnknown && cond.Status == corev1.ConditionFalse {
		s.status = cond.Status
	}
}

// SummarizeAppConditions summarizes the provided conditions slice into a single condition with the following logic:
// - If there be any 'Deleting' condition type, the summary condition type will be 'Deleting'
// - Otherwise, if there be any 'Reconciling' condition type, the summary condition type will be 'Reconciling'
// - Otherwise, if there be any 'DeleteFailed' condition type, the summary condition type will be 'DeleteFailed'
// - Otherwise, if there be any 'ReconcileFailed' condition type, the summary condition type will be 'ReconcileFailed'
// - Otherwise, if there be any 'ReconcileSucceeded' condition type, the summary condition type will be 'ReconcileSucceeded'
// - Otherwise, the summary condition type will be 'UnknownCondition'
// For the summary condition's Status field, it uses the logic in setState(), that is if there be any 'True' Status for a
// given condition type, it overrides other existing 'False' or 'Unknown' Status entries for the same condition type and the
// resulting condition will have 'True' Status. Similarly, existence of 'False' Status overrides 'Unknown' Status
func SummarizeAppConditions(conditions []v1alpha1.AppCondition) v1alpha1.AppCondition {
	summaryCondition := v1alpha1.AppCondition{}
	var (
		reconciling        = &state{}
		reconcileFailed    = &state{}
		reconcileSucceeded = &state{}
		deleting           = &state{}
		deleteFailed       = &state{}
	)

	for _, cond := range conditions {
		if cond.Type == v1alpha1.Reconciling {
			reconciling.setState(cond)
		}
		if cond.Type == v1alpha1.ReconcileFailed {
			reconcileFailed.setState(cond)
		}
		if cond.Type == v1alpha1.ReconcileSucceeded {
			reconcileSucceeded.setState(cond)
		}
		if cond.Type == v1alpha1.Deleting {
			deleting.setState(cond)
		}
		if cond.Type == v1alpha1.DeleteFailed {
			deleteFailed.setState(cond)
		}
	}

	if deleting.stateExists {
		summaryCondition.Type = v1alpha1.Deleting
		summaryCondition.Status = deleting.status
	} else if reconciling.stateExists {
		summaryCondition.Type = v1alpha1.Reconciling
		summaryCondition.Status = reconciling.status
	} else if deleteFailed.stateExists {
		summaryCondition.Type = v1alpha1.DeleteFailed
		summaryCondition.Status = deleteFailed.status
	} else if reconcileFailed.stateExists {
		summaryCondition.Type = v1alpha1.ReconcileFailed
		summaryCondition.Status = reconcileFailed.status
	} else if reconcileSucceeded.stateExists {
		summaryCondition.Type = v1alpha1.ReconcileSucceeded
		summaryCondition.Status = reconcileSucceeded.status
	} else {
		summaryCondition.Type = UnknownCondition
	}

	return summaryCondition
}

// HasSameState returns true if a ClusterBootstrap condition has the same state of another; state is defined
// by the union of following fields: Type, Status, UsefulErrorMessage (it excludes LastTransitionTime).
func HasSameState(i, j *runtanzuv1alpha3.Condition) bool {
	return i.Type == j.Type &&
		i.Status == j.Status &&
		i.UsefulErrorMessage == j.UsefulErrorMessage
}
