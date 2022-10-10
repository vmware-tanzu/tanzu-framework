// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	stringcmp "github.com/vmware-tanzu/tanzu-framework/util/cmp/strings"
)

func TestNamespacesMatchingSelector(t *testing.T) {
	newNamespace := func(name string) *corev1.Namespace {
		return &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name, Labels: map[string]string{"kubernetes.io/metadata.name": name}}}
	}

	testCases := []struct {
		description        string
		existingNamespaces []runtime.Object
		namespaceSelector  *metav1.LabelSelector
		want               []string
		err                string
	}{
		{
			description:        "Empty namespace selector - all namespaces selected",
			existingNamespaces: []runtime.Object{newNamespace("default"), newNamespace("kube-system"), newNamespace("kube-public"), newNamespace("tkg-system")},
			namespaceSelector:  &metav1.LabelSelector{},
			want:               []string{"default", "kube-system", "tkg-system", "kube-public"},
			err:                "",
		},
		{
			// Callers interested in feature gates should not be sending a nil namespaces, but this is a generic function.
			description: "Nil namespace selector - no namespaces selected",
			existingNamespaces: []runtime.Object{
				newNamespace("default"),
				newNamespace("kube-system"),
				newNamespace("kube-public"),
				newNamespace("tkg-system"),
			},
			namespaceSelector: nil,
			want:              []string{},
			err:               "",
		},
		{
			description:        "Selector that matches partially",
			existingNamespaces: []runtime.Object{newNamespace("default"), newNamespace("kube-system"), newNamespace("kube-public"), newNamespace("tkg-system")},
			namespaceSelector: &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{Key: "kubernetes.io/metadata.name", Operator: metav1.LabelSelectorOpIn, Values: []string{"kube-system", "tkg-system", "badlabel"}},
				},
			},
			want: []string{"kube-system", "tkg-system"},
			err:  "",
		},
		{
			description:        "Selector that doesn't match anything",
			existingNamespaces: []runtime.Object{newNamespace("default"), newNamespace("kube-system"), newNamespace("kube-public"), newNamespace("tkg-system")},
			namespaceSelector: &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{Key: "kubernetes.io/metadata.name", Operator: metav1.LabelSelectorOpIn, Values: []string{"badlabel1", "badlabel2"}},
				},
			},
			want: []string{},
			err:  "",
		},
		{
			description:        "Bad namespace selector - error",
			existingNamespaces: []runtime.Object{newNamespace("default"), newNamespace("kube-system"), newNamespace("kube-public"), newNamespace("tkg-system")},
			namespaceSelector: &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{
					{Key: "kubernetes.io/metadata.name", Operator: metav1.LabelSelectorOpExists, Values: []string{"value-should-not-be-here"}},
				},
			},
			want: []string{},
			err:  "failed to get namespaces from NamespaceSelector",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			fakeClient := fake.NewClientBuilder().WithScheme(k8sscheme.Scheme).WithRuntimeObjects(tc.existingNamespaces...).Build()
			got, err := NamespacesMatchingSelector(context.Background(), fakeClient, tc.namespaceSelector)
			if err != nil {
				if tc.err == "" {
					t.Errorf("no error string specified, but got error: %v", err)
				}
				if !strings.Contains(err.Error(), tc.err) {
					t.Errorf("error string=%q doesn't match partial=%q", tc.err, err)
				}
			} else if tc.err != "" {
				t.Errorf("error string=%q specified but error not found", tc.err)
			}
			if diff := stringcmp.SliceDiffIgnoreOrder(got, tc.want); diff != "" {
				t.Errorf("got namespaces: %v, want namespaces: %v, diff: %s", got, tc.want, diff)
			}
		})
	}
}
