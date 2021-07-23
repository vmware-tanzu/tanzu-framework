// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"context"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake" //nolint:staticcheck

	stringcmp "github.com/vmware-tanzu/tanzu-framework/pkg/v1/test/cmp/strings"
)

func TestValidateFeatureImmutability(t *testing.T) {
	testCases := []struct {
		description     string
		currentFeatures []Feature
		featureGateSpec FeatureGateSpec
		oldObj          *FeatureGate
		want            []string
	}{
		{
			description: "Multiple immutable features changed",
			currentFeatures: []Feature{
				{ObjectMeta: metav1.ObjectMeta{Name: "one"}, Spec: FeatureSpec{Activated: true, Discoverable: true, Immutable: true}},
				{ObjectMeta: metav1.ObjectMeta{Name: "two"}, Spec: FeatureSpec{Activated: true, Discoverable: true, Immutable: true}},
				{ObjectMeta: metav1.ObjectMeta{Name: "three"}, Spec: FeatureSpec{Activated: false, Discoverable: true, Immutable: false}},
				{ObjectMeta: metav1.ObjectMeta{Name: "four"}, Spec: FeatureSpec{Activated: false, Discoverable: true, Immutable: false}},
				{ObjectMeta: metav1.ObjectMeta{Name: "five"}, Spec: FeatureSpec{Activated: false, Discoverable: true, Immutable: false}},
				{ObjectMeta: metav1.ObjectMeta{Name: "eleven"}, Spec: FeatureSpec{Activated: true, Discoverable: true, Immutable: true}},
				{ObjectMeta: metav1.ObjectMeta{Name: "twelve"}, Spec: FeatureSpec{Activated: false, Discoverable: true, Immutable: true}},
				{ObjectMeta: metav1.ObjectMeta{Name: "thirteen"}, Spec: FeatureSpec{Activated: false, Discoverable: true, Immutable: false}},
				{ObjectMeta: metav1.ObjectMeta{Name: "hundred"}, Spec: FeatureSpec{Activated: false, Discoverable: false, Immutable: true}},
				{ObjectMeta: metav1.ObjectMeta{Name: "thousand"}, Spec: FeatureSpec{Activated: false, Discoverable: false, Immutable: false}},
			},
			featureGateSpec: FeatureGateSpec{
				Features: []FeatureReference{
					// Immutable, previously activated, now deactivated (disallowed).
					{Name: "one", Activate: false},
					// Immutable, previously deactivated, now activated (disallowed).
					{Name: "two", Activate: true},
					// Non-immutable, previously activated, no state change (allowed).
					{Name: "three", Activate: true},
					// Non-immutable, previously deactivated, no state change (allowed).
					{Name: "four", Activate: false},
					// Unavailable feature, no state change (allowed).
					{Name: "six", Activate: true},
					// Unavailable feature, state changed (allowed).
					{Name: "seven", Activate: false},
					// Unavailable feature, no state change (allowed).
					{Name: "eight", Activate: false},
					// Unavailable feature, no state change (allowed).
					{Name: "nine", Activate: true},
					// Unavailable feature, no state change (allowed).
					{Name: "ten", Activate: true},
					// Immutable, previously activated, now deactivated (disallowed).
					{Name: "eleven", Activate: false},
					// Immutable, previously deactivated, no state change (allowed).
					// Not explicitly specified in spec.
					// {Name: "twelve", Activate: false},
					// Immutable, non-discoverable, no state change (allowed).
					// Not explicitly specified in spec.
					// {Name: "hundred", Activate: false},
				},
			},
			oldObj: &FeatureGate{Status: FeatureGateStatus{
				ActivatedFeatures:   []string{"one", "three", "eleven"},
				DeactivatedFeatures: []string{"two", "four", "five", "twelve", "thirteen"},
			}},
			want: []string{"one", "two", "eleven"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			got := computeChangedImmutableFeatures(tc.featureGateSpec, tc.currentFeatures, tc.oldObj)
			if diff := stringcmp.SliceDiffIgnoreOrder(got, tc.want); diff != "" {
				t.Errorf("got immutable features %v, want %v, diff: %s", got, tc.want, diff)
			}
		})
	}
}

//nolint:funlen
func TestNamespaceConflicts(t *testing.T) {
	scheme, err := getScheme()
	if err != nil {
		t.Fatal(err)
	}

	newNamespace := func(name string) *corev1.Namespace {
		return &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name, Labels: map[string]string{"kubernetes.io/metadata.name": name}}}
	}

	compareConflicts := func(got, want map[string][]string) string {
		for k := range got {
			sort.Strings(got[k])
		}
		for k := range want {
			sort.Strings(want[k])
		}
		return cmp.Diff(got, want, cmpopts.SortMaps(func(x, y string) bool { return x < y }), cmpopts.EquateEmpty())
	}

	testCases := []struct {
		description          string
		existingNamespaces   []runtime.Object
		existingFeatureGates []runtime.Object
		createObj            *FeatureGate
		want                 map[string][]string
		err                  string
	}{
		{
			description: "Has single conflicting namespace",
			existingNamespaces: []runtime.Object{
				newNamespace("default"),
				newNamespace("kube-system"),
				newNamespace("kube-public"),
				newNamespace("tkg-system"),
			},
			existingFeatureGates: []runtime.Object{
				&FeatureGate{
					ObjectMeta: metav1.ObjectMeta{
						Name: "system-feature-gates",
					},
					Status: FeatureGateStatus{
						Namespaces: []string{"tkg-system"},
					},
				},
			},
			createObj: &FeatureGate{
				ObjectMeta: metav1.ObjectMeta{Name: "foo"},
				Spec: FeatureGateSpec{
					NamespaceSelector: metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{Key: "kubernetes.io/metadata.name", Operator: metav1.LabelSelectorOpIn, Values: []string{"bar", "baz", "default", "tkg-system"}},
						},
					},
				},
			},
			want: map[string][]string{"system-feature-gates": {"tkg-system"}},
			err:  "",
		},
		{
			description: "Multiple conflicting namespaces",
			existingNamespaces: []runtime.Object{
				newNamespace("default"),
				newNamespace("kube-system"),
				newNamespace("kube-public"),
				newNamespace("tkg-system"),
			},
			existingFeatureGates: []runtime.Object{
				&FeatureGate{
					ObjectMeta: metav1.ObjectMeta{
						Name: "system-feature-gates",
					},
					Status: FeatureGateStatus{
						Namespaces: []string{"default", "kube-system", "tkg-system"},
					},
				},
			},
			createObj: &FeatureGate{
				ObjectMeta: metav1.ObjectMeta{Name: "foo"},
				Spec: FeatureGateSpec{
					NamespaceSelector: metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{Key: "kubernetes.io/metadata.name", Operator: metav1.LabelSelectorOpIn, Values: []string{"bar", "baz", "default", "tkg-system"}},
						},
					},
				},
			},
			want: map[string][]string{"system-feature-gates": {"tkg-system", "default"}},
			err:  "",
		},
		{
			description: "Conflicting namespaces across multiple CRs",
			existingNamespaces: []runtime.Object{
				newNamespace("default"),
				newNamespace("kube-system"),
				newNamespace("kube-public"),
				newNamespace("tkg-system"),
				newNamespace("foo"),
				newNamespace("bar"),
				newNamespace("baz"),
			},
			existingFeatureGates: []runtime.Object{
				&FeatureGate{
					ObjectMeta: metav1.ObjectMeta{
						Name: "system-feature-gates",
					},
					Status: FeatureGateStatus{
						Namespaces: []string{"default", "kube-system", "tkg-system"},
					},
				},
				&FeatureGate{
					ObjectMeta: metav1.ObjectMeta{
						Name: "experimental-feature-gates",
					},
					Status: FeatureGateStatus{
						Namespaces: []string{"foo", "bar"},
					},
				},
			},
			createObj: &FeatureGate{
				ObjectMeta: metav1.ObjectMeta{Name: "foo"},
				Spec: FeatureGateSpec{
					NamespaceSelector: metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{Key: "kubernetes.io/metadata.name", Operator: metav1.LabelSelectorOpIn, Values: []string{"foo", "bar", "baz", "default", "tkg-system", "kube-public"}},
						},
					},
				},
			},
			want: map[string][]string{"system-feature-gates": {"tkg-system", "default"}, "experimental-feature-gates": {"bar", "foo"}},
			err:  "",
		},
		{
			description: "No conflicting namespaces",
			existingNamespaces: []runtime.Object{
				newNamespace("default"),
				newNamespace("kube-system"),
				newNamespace("kube-public"),
				newNamespace("tkg-system"),
				newNamespace("foo"),
				newNamespace("bar"),
				newNamespace("baz"),
			},
			existingFeatureGates: []runtime.Object{
				&FeatureGate{
					ObjectMeta: metav1.ObjectMeta{
						Name: "system-feature-gates",
					},
					Status: FeatureGateStatus{
						Namespaces: []string{"default", "kube-system", "tkg-system"},
					},
				},
				&FeatureGate{
					ObjectMeta: metav1.ObjectMeta{
						Name: "experimental-feature-gates",
					},
					Status: FeatureGateStatus{
						Namespaces: []string{"foo", "bar"},
					},
				},
			},
			createObj: &FeatureGate{
				ObjectMeta: metav1.ObjectMeta{Name: "foo"},
				Spec: FeatureGateSpec{
					NamespaceSelector: metav1.LabelSelector{
						// Bad label selector borrowed from kubernetes/apimachinery/pkg/apis/meta/v1/helpers_test.go
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{Key: "kubernetes.io/metadata.name", Operator: metav1.LabelSelectorOpIn, Values: []string{"baz", "kube-public"}},
						},
					},
				},
			},
			want: map[string][]string{},
			err:  "",
		},
		{
			description: "Error occurred due to bad namespace selector in spec",
			existingNamespaces: []runtime.Object{
				newNamespace("default"),
				newNamespace("kube-system"),
				newNamespace("kube-public"),
				newNamespace("tkg-system"),
				newNamespace("foo"),
				newNamespace("bar"),
				newNamespace("baz"),
			},
			existingFeatureGates: []runtime.Object{
				&FeatureGate{
					ObjectMeta: metav1.ObjectMeta{
						Name: "system-feature-gates",
					},
					Status: FeatureGateStatus{
						Namespaces: []string{"default", "kube-system", "tkg-system"},
					},
				},
				&FeatureGate{
					ObjectMeta: metav1.ObjectMeta{
						Name: "experimental-feature-gates",
					},
					Status: FeatureGateStatus{
						Namespaces: []string{"foo", "bar"},
					},
				},
			},
			createObj: &FeatureGate{
				ObjectMeta: metav1.ObjectMeta{Name: "foo"},
				Spec: FeatureGateSpec{
					NamespaceSelector: metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							// Bad selector - Exists operator must have empty values.
							{Key: "kubernetes.io/metadata.name", Operator: metav1.LabelSelectorOpExists, Values: []string{"qux", "norf"}},
						},
					},
				},
			},
			want: map[string][]string{},
			err:  "failed to get namespaces from NamespaceSelector",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			var objs []runtime.Object
			objs = append(objs, tc.existingNamespaces...)
			objs = append(objs, tc.existingFeatureGates...)
			fakeClient := fake.NewFakeClientWithScheme(scheme, objs...)
			got, err := tc.createObj.computeConflictingNamespaces(context.Background(), fakeClient)
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
			if diff := compareConflicts(got, tc.want); diff != "" {
				t.Errorf("got conflicts %v, want %v, diff: %s", got, tc.want, diff)
			}
		})
	}
}
