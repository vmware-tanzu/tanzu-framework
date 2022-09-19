// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NamespacesMatchingSelector returns the list of namespaces after applying the NamespaceSelector filter.
// Note that a nil selector selects nothing, while an empty selector selects everything.
// Callers using this function in feature gates context should be sending a pointer to an empty selector instead of nil.
func NamespacesMatchingSelector(ctx context.Context, c client.Client, selector *metav1.LabelSelector) ([]string, error) {
	s, err := metav1.LabelSelectorAsSelector(selector)
	if err != nil {
		return nil, fmt.Errorf("failed to get namespaces from NamespaceSelector: %w", err)
	}

	nsList := &corev1.NamespaceList{}
	err = c.List(ctx, nsList, client.MatchingLabelsSelector{Selector: s})
	if err != nil {
		return nil, fmt.Errorf("failed to get namespaces from NamespaceSelector: %w", err)
	}

	var namespaces []string
	for i := range nsList.Items {
		namespaces = append(namespaces, nsList.Items[i].Name)
	}
	return namespaces, nil
}
