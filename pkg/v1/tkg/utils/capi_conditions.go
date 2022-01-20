// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	corev1 "k8s.io/api/core/v1"
	clusterv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"
)

/*
	Copying some utilities from upstream CAPI conditions. This is one of the consequences
	of the CLI to support both TKGm(CAPI v1beta1) and TKGS(v1alpha3). TKGS recently moved
	to v1alpha2 which has support for CAPI(v1alpha3) conditions.
	TODO: Remove this when TKGm and TKGS are on the same CAPI version (Unified TKG).
*/

// Getter interface defines methods that a Cluster API object should implement in order to
// use the conditions package for getting conditions.
type Getter interface {
	crtclient.Object
	GetConditions() clusterv1alpha3.Conditions
}

// Get returns the condition with the given type, if the condition does not exist,
// it returns nil.
func Get(from Getter, t clusterv1alpha3.ConditionType) *clusterv1alpha3.Condition {
	conditions := from.GetConditions()
	if conditions == nil {
		return nil
	}

	for _, condition := range conditions {
		if condition.Type == t {
			return &condition
		}
	}
	return nil
}

// Has returns true if a condition with the given type exists.
func Has(from Getter, t clusterv1alpha3.ConditionType) bool {
	return Get(from, t) != nil
}

// IsTrue is true if the condition with the given type is True, otherwise it return false
// if the condition is not True or if the condition does not exist (is nil).
func IsTrue(from Getter, t clusterv1alpha3.ConditionType) bool {
	if c := Get(from, t); c != nil {
		return c.Status == corev1.ConditionTrue
	}
	return false
}

// IsFalse is true if the condition with the given type is False, otherwise it return false
// if the condition is not False or if the condition does not exist (is nil).
func IsFalse(from Getter, t clusterv1alpha3.ConditionType) bool {
	if c := Get(from, t); c != nil {
		return c.Status == corev1.ConditionFalse
	}
	return false
}

// IsUnknown is true if the condition with the given type is Unknown or if the condition
// does not exist (is nil).
func IsUnknown(from Getter, t clusterv1alpha3.ConditionType) bool {
	if c := Get(from, t); c != nil {
		return c.Status == corev1.ConditionUnknown
	}
	return true
}

// GetReason returns a nil safe string of Reason for the condition with the given type.
func GetReason(from Getter, t clusterv1alpha3.ConditionType) string {
	if c := Get(from, t); c != nil {
		return c.Reason
	}
	return ""
}

// GetMessage returns a nil safe string of Message.
func GetMessage(from Getter, t clusterv1alpha3.ConditionType) string {
	if c := Get(from, t); c != nil {
		return c.Message
	}
	return ""
}

// GetSeverity returns the condition Severity or nil if the condition
// does not exist (is nil).
func GetSeverity(from Getter, t clusterv1alpha3.ConditionType) *clusterv1alpha3.ConditionSeverity {
	if c := Get(from, t); c != nil {
		return &c.Severity
	}
	return nil
}
