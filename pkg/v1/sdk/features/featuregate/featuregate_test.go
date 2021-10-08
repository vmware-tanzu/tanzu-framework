// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package featuregate

import (
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	configv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	stringcmp "github.com/vmware-tanzu/tanzu-framework/pkg/v1/test/cmp/strings"
)

func TestComputeFeatureStates(t *testing.T) {
	testCases := []struct {
		description         string
		features            []configv1alpha1.Feature
		spec                configv1alpha1.FeatureGateSpec
		expectedActivated   []string
		expectedDeactivated []string
		expectedUnavailable []string
	}{
		{
			description: "All combinations of availability and discoverability",
			features: []configv1alpha1.Feature{
				{ObjectMeta: metav1.ObjectMeta{Name: "one"}, Spec: configv1alpha1.FeatureSpec{Activated: true, Discoverable: true}},
				{ObjectMeta: metav1.ObjectMeta{Name: "two"}, Spec: configv1alpha1.FeatureSpec{Activated: true, Discoverable: true}},
				{ObjectMeta: metav1.ObjectMeta{Name: "three"}, Spec: configv1alpha1.FeatureSpec{Activated: false, Discoverable: true}},
				{ObjectMeta: metav1.ObjectMeta{Name: "four"}, Spec: configv1alpha1.FeatureSpec{Activated: false, Discoverable: true}},
				{ObjectMeta: metav1.ObjectMeta{Name: "five"}, Spec: configv1alpha1.FeatureSpec{Activated: false, Discoverable: true}},
				{ObjectMeta: metav1.ObjectMeta{Name: "eleven"}, Spec: configv1alpha1.FeatureSpec{Activated: true, Discoverable: true}},
				{ObjectMeta: metav1.ObjectMeta{Name: "twelve"}, Spec: configv1alpha1.FeatureSpec{Activated: false, Discoverable: true}},
				{ObjectMeta: metav1.ObjectMeta{Name: "thirteen"}, Spec: configv1alpha1.FeatureSpec{Activated: false, Discoverable: true}},
				{ObjectMeta: metav1.ObjectMeta{Name: "hundred"}, Spec: configv1alpha1.FeatureSpec{Activated: false, Discoverable: false}},
				{ObjectMeta: metav1.ObjectMeta{Name: "thousand"}, Spec: configv1alpha1.FeatureSpec{Activated: false, Discoverable: false}},
			},
			spec: configv1alpha1.FeatureGateSpec{
				Features: []configv1alpha1.FeatureReference{
					{Name: "one", Activate: true},
					{Name: "two", Activate: false},
					{Name: "three", Activate: true},
					{Name: "four", Activate: false},
					{Name: "six", Activate: true},
					{Name: "seven", Activate: true},
					{Name: "eight", Activate: false},
					{Name: "nine", Activate: true},
					{Name: "ten", Activate: true},
				}},
			expectedActivated:   []string{"one", "three", "eleven"},
			expectedDeactivated: []string{"two", "four", "five", "twelve", "thirteen"},
			expectedUnavailable: []string{"six", "seven", "eight", "nine", "ten"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			activated, deactivated, unavailable := ComputeFeatureStates(tc.spec, tc.features)

			if diff := stringcmp.SliceDiffIgnoreOrder(activated, tc.expectedActivated); diff != "" {
				t.Errorf("got activated features %v, want %v, diff: %s", activated, tc.expectedActivated, diff)
			}

			if diff := stringcmp.SliceDiffIgnoreOrder(deactivated, tc.expectedDeactivated); diff != "" {
				t.Errorf("got deactivated features %v, want %v, diff: %s", deactivated, tc.expectedDeactivated, diff)
			}

			if diff := stringcmp.SliceDiffIgnoreOrder(unavailable, tc.expectedUnavailable); diff != "" {
				t.Errorf("got unavailable features %v, want %v, diff: %s", unavailable, tc.expectedUnavailable, diff)
			}
		})
	}
}

//nolint:funlen
func TestFeaturesActivatedInNamespacesMatchingSelector(t *testing.T) {
	scheme, err := configv1alpha1.SchemeBuilder.Build()
	if err != nil {
		t.Fatal(err)
	}
	if err := k8sscheme.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}

	newNamespace := func(name string) *corev1.Namespace {
		return &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name, Labels: map[string]string{"kubernetes.io/metadata.name": name}}}
	}

	newFeatureGate := func(name string, namespaces, activated, deactivated, unavailable []string) *configv1alpha1.FeatureGate {
		return &configv1alpha1.FeatureGate{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Status: configv1alpha1.FeatureGateStatus{
				Namespaces:          namespaces,
				ActivatedFeatures:   activated,
				DeactivatedFeatures: deactivated,
				UnavailableFeatures: unavailable,
			}}
	}

	testCases := []struct {
		description     string
		existingObjects []runtime.Object
		selector        metav1.LabelSelector
		features        []string
		want            bool
		err             string
	}{
		{
			description: "Namespace matched by selector is gated and all features are activated",
			existingObjects: []runtime.Object{
				newNamespace("kube-system"),
				newNamespace("kube-public"),
				newNamespace("tkg-system"),
				newFeatureGate("tkg", []string{"kube-system", "tkg-system"}, []string{"one", "two"}, []string{"three", "four"}, []string{"five", "six"}),
			},
			selector: metav1.LabelSelector{MatchExpressions: []metav1.LabelSelectorRequirement{
				{Key: "kubernetes.io/metadata.name", Operator: metav1.LabelSelectorOpIn, Values: []string{"tkg-system"}},
			}},
			features: []string{"one", "two"},
			want:     true,
			err:      "",
		},
		{
			description: "Namespace matched by selector is gated and one feature is deactivated",
			existingObjects: []runtime.Object{
				newNamespace("kube-system"),
				newNamespace("ns-no-feature-gates"),
				newNamespace("tkg-system"),
				newFeatureGate("tkg", []string{"kube-system", "tkg-system"}, []string{"one", "two"}, []string{"three", "four"}, []string{"five", "six"}),
			},
			selector: metav1.LabelSelector{MatchExpressions: []metav1.LabelSelectorRequirement{
				{Key: "kubernetes.io/metadata.name", Operator: metav1.LabelSelectorOpIn, Values: []string{"tkg-system"}},
			}},
			features: []string{"one", "three"},
			want:     false,
			err:      "",
		},
		{
			description: "Namespaces matched by selector are gated and all features are activated",
			existingObjects: []runtime.Object{
				newNamespace("kube-system"),
				newNamespace("ns-no-feature-gates"),
				newNamespace("tkg-system"),
				newFeatureGate("tkg", []string{"kube-system", "tkg-system"}, []string{"one", "two"}, []string{"three", "four"}, []string{"five", "six"}),
			},
			selector: metav1.LabelSelector{MatchExpressions: []metav1.LabelSelectorRequirement{
				{Key: "kubernetes.io/metadata.name", Operator: metav1.LabelSelectorOpIn, Values: []string{"tkg-system", "kube-system"}},
			}},
			features: []string{"one", "two"},
			want:     true,
			err:      "",
		},
		{
			description: "Namespaces matched by selector are gated and all features are activated across multiple feature gates",
			existingObjects: []runtime.Object{
				newNamespace("kube-system"),
				newNamespace("ns-no-feature-gates"),
				newNamespace("tkg-system"),
				newFeatureGate("tkg", []string{"kube-system", "tkg-system"}, []string{"one", "two"}, []string{"three", "four"}, []string{"five", "six"}),
				newFeatureGate("dev", []string{"kube-public"}, []string{"one", "two"}, []string{"three"}, []string{}),
			},
			selector: metav1.LabelSelector{MatchExpressions: []metav1.LabelSelectorRequirement{
				{Key: "kubernetes.io/metadata.name", Operator: metav1.LabelSelectorOpIn, Values: []string{"tkg-system", "kube-system", "kube-public"}},
			}},
			features: []string{"one", "two"},
			want:     true,
			err:      "",
		},
		{
			description: "Namespaces are gated and all features are either deactivated or unavailable",
			existingObjects: []runtime.Object{
				newNamespace("kube-system"),
				newNamespace("kube-public"),
				newNamespace("tkg-system"),
				newFeatureGate("tkg", []string{"kube-system", "tkg-system"}, []string{"one", "two"}, []string{"three", "four"}, []string{"five", "six"}),
			},
			selector: metav1.LabelSelector{MatchExpressions: []metav1.LabelSelectorRequirement{
				{Key: "kubernetes.io/metadata.name", Operator: metav1.LabelSelectorOpIn, Values: []string{"tkg-system"}},
			}},
			features: []string{"five", "three"},
			want:     false,
			err:      "",
		},
		{
			description: "Label selector matches only one namespace with all features activated",
			existingObjects: []runtime.Object{
				newNamespace("kube-system"),
				newNamespace("kube-public"),
				newNamespace("tkg-system"),
				newFeatureGate("tkg", []string{"kube-system", "tkg-system"}, []string{"one", "two"}, []string{"three", "four"}, []string{"five", "six"}),
			},
			selector: metav1.LabelSelector{MatchExpressions: []metav1.LabelSelectorRequirement{
				{Key: "kubernetes.io/metadata.name", Operator: metav1.LabelSelectorOpIn, Values: []string{"tkg-system", "ns-no-feature-gates"}},
			}},
			features: []string{"one", "two"},
			want:     true,
			err:      "",
		},
		{
			description: "Label selector matches two namespaces where a feature is in different activation states",
			existingObjects: []runtime.Object{
				newNamespace("kube-system"),
				newNamespace("kube-public"),
				newNamespace("tkg-system"),
				newFeatureGate("tkg", []string{"kube-system", "tkg-system"}, []string{"one"}, []string{}, []string{}),
				newFeatureGate("dev", []string{"kube-public"}, []string{}, []string{"one"}, []string{}),
			},
			selector: metav1.LabelSelector{MatchExpressions: []metav1.LabelSelectorRequirement{
				{Key: "kubernetes.io/metadata.name", Operator: metav1.LabelSelectorOpIn, Values: []string{"tkg-system", "kube-public"}},
			}},
			features: []string{"one"},
			want:     false,
			err:      "",
		},
		{
			description: "Label selector matches no namespaces",
			existingObjects: []runtime.Object{
				newNamespace("kube-system"),
				newNamespace("kube-public"),
				newNamespace("tkg-system"),
				newFeatureGate("tkg", []string{"kube-system", "tkg-system"}, []string{"one", "two"}, []string{"three", "four"}, []string{"five", "six"}),
			},
			selector: metav1.LabelSelector{MatchExpressions: []metav1.LabelSelectorRequirement{
				// LabelSelectorRequirement are ANDed.
				{Key: "kubernetes.io/metadata.name", Operator: metav1.LabelSelectorOpIn, Values: []string{"tkg-system"}},
				{Key: "kubernetes.io/metadata.name", Operator: metav1.LabelSelectorOpIn, Values: []string{"ns-no-feature-gates"}},
			}},
			features: []string{"one", "two"},
			want:     false,
			err:      "",
		},
		{
			description: "Features are not gated in any of the namespaces",
			existingObjects: []runtime.Object{
				newNamespace("kube-system"),
				newNamespace("kube-public"),
				newNamespace("tkg-system"),
				newFeatureGate("tkg", []string{"kube-system", "tkg-system"}, []string{"one", "two"}, []string{"three", "four"}, []string{"five", "six"}),
			},
			selector: metav1.LabelSelector{MatchExpressions: []metav1.LabelSelectorRequirement{
				{Key: "kubernetes.io/metadata.name", Operator: metav1.LabelSelectorOpIn, Values: []string{"ns-no-feature-gates"}},
				{Key: "kubernetes.io/metadata.name", Operator: metav1.LabelSelectorOpIn, Values: []string{"kube-public"}},
			}},
			features: []string{"one", "two"},
			want:     false,
			err:      "",
		},
		{
			description: "No features specified",
			existingObjects: []runtime.Object{
				newNamespace("kube-system"),
				newNamespace("kube-public"),
				newNamespace("tkg-system"),
				newFeatureGate("tkg", []string{"kube-system", "tkg-system"}, []string{"one", "two"}, []string{"three", "four"}, []string{"five", "six"}),
			},
			selector: metav1.LabelSelector{MatchExpressions: []metav1.LabelSelectorRequirement{
				{Key: "kubernetes.io/metadata.name", Operator: metav1.LabelSelectorOpIn, Values: []string{"ns-no-feature-gates"}},
				{Key: "kubernetes.io/metadata.name", Operator: metav1.LabelSelectorOpIn, Values: []string{"kube-public"}},
			}},
			features: []string{},
			want:     false,
			err:      "",
		},
		{
			description: "Error due to bad namespace selector",
			existingObjects: []runtime.Object{
				newNamespace("kube-system"),
				newNamespace("kube-public"),
				newNamespace("tkg-system"),
				newFeatureGate("tkg", []string{"kube-system", "tkg-system"}, []string{"one", "two"}, []string{"three", "four"}, []string{"five", "six"}),
			},
			selector: metav1.LabelSelector{MatchExpressions: []metav1.LabelSelectorRequirement{
				{Key: "kubernetes.io/metadata.name", Operator: metav1.LabelSelectorOpExists, Values: []string{"bad-value"}},
			}},
			features: []string{},
			want:     false,
			err:      "failed to get namespaces from NamespaceSelector",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(tc.existingObjects...).Build()
			got, err := FeaturesActivatedInNamespacesMatchingSelector(context.Background(), fakeClient, tc.selector, tc.features)
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
			if got != tc.want {
				t.Errorf("feature activation: got %t, want %t", got, tc.want)
			}
		})
	}
}
