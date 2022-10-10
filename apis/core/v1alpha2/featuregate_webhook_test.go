// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestComputeFeaturesThatVoidSupportWarranty(t *testing.T) {
	testCases := []struct {
		description     string
		featureList     *FeatureList
		featureGateSpec FeatureGateSpec
		want            []string
	}{
		{
			description: "Multiple features in featuregate void support warranty",
			featureList: &FeatureList{
				Items: []Feature{
					{ObjectMeta: metav1.ObjectMeta{Name: "foo"}, Spec: FeatureSpec{Description: "foo", Stability: "Experimental"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "bar"}, Spec: FeatureSpec{Description: "bar", Stability: "Experimental"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "baz"}, Spec: FeatureSpec{Description: "baz", Stability: "Technical Preview"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "bax"}, Spec: FeatureSpec{Description: "bax", Stability: "Stable"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "qux"}, Spec: FeatureSpec{Description: "qux", Stability: "Experimental"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "quux"}, Spec: FeatureSpec{Description: "quux", Stability: "Stable"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "corge"}, Spec: FeatureSpec{Description: "corge", Stability: "Experimental"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "grault"}, Spec: FeatureSpec{Description: "grault", Stability: "Stable"}},
				},
			},
			featureGateSpec: FeatureGateSpec{
				Features: []FeatureReference{
					// voids warranty and violates policy
					{Name: "foo", Activate: true},
					// voids warranty and violates policy
					{Name: "bar", Activate: true},
					// Doesn't void warranty
					{Name: "baz", Activate: true},
					// Doesn't violate policy
					{Name: "bax", Activate: true},
				},
			},
			want: []string{"foo", "bar"},
		},
		{
			description: "No features in featuregate void support warranty",
			featureList: &FeatureList{
				Items: []Feature{
					{ObjectMeta: metav1.ObjectMeta{Name: "foo"}, Spec: FeatureSpec{Description: "foo", Stability: "Experimental"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "bar"}, Spec: FeatureSpec{Description: "bar", Stability: "Experimental"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "baz"}, Spec: FeatureSpec{Description: "baz", Stability: "Technical Preview"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "bax"}, Spec: FeatureSpec{Description: "bax", Stability: "Stable"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "qux"}, Spec: FeatureSpec{Description: "qux", Stability: "Experimental"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "quux"}, Spec: FeatureSpec{Description: "quux", Stability: "Stable"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "corge"}, Spec: FeatureSpec{Description: "corge", Stability: "Experimental"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "grault"}, Spec: FeatureSpec{Description: "grault", Stability: "Stable"}},
				},
			},
			featureGateSpec: FeatureGateSpec{
				Features: []FeatureReference{
					{Name: "bax", Activate: true},
					{Name: "quux", Activate: true},
					{Name: "grault", Activate: true},
				},
			},
			want: []string{},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			got := computeFeaturesThatVoidSupportWarranty(tc.featureGateSpec, tc.featureList)
			if diff := sliceDiffIgnoreOrder(got, tc.want); diff != "" {
				t.Errorf("got invalid features %v, want %v, diff: %s", got, tc.want, diff)
			}
		})
	}
}

func TestComputeImmutableFeatures(t *testing.T) {
	testCases := []struct {
		description     string
		featureList     *FeatureList
		featureGateSpec FeatureGateSpec
		want            []string
	}{
		{
			description: "Multiple features in featuregate are immutable",
			featureList: &FeatureList{
				Items: []Feature{
					{ObjectMeta: metav1.ObjectMeta{Name: "foo"}, Spec: FeatureSpec{Description: "foo", Stability: "Stable"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "bar"}, Spec: FeatureSpec{Description: "bar", Stability: "Experimental"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "baz"}, Spec: FeatureSpec{Description: "baz", Stability: "Technical Preview"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "bax"}, Spec: FeatureSpec{Description: "bax", Stability: "Stable"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "qux"}, Spec: FeatureSpec{Description: "qux", Stability: "Experimental"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "quux"}, Spec: FeatureSpec{Description: "quux", Stability: "Stable"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "corge"}, Spec: FeatureSpec{Description: "corge", Stability: "Experimental"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "grault"}, Spec: FeatureSpec{Description: "grault", Stability: "Stable"}},
				},
			},
			featureGateSpec: FeatureGateSpec{
				Features: []FeatureReference{
					// immutable and cannot be toggled
					{Name: "foo", Activate: false},
					// immutable and cannot be toggled
					{Name: "bax", Activate: false},
					// immutable, but set to default activation state
					{Name: "quux", Activate: true},
				},
			},
			want: []string{"foo", "bax"},
		},
		{
			description: "No features in featuregate are immutable",
			featureList: &FeatureList{
				Items: []Feature{
					{ObjectMeta: metav1.ObjectMeta{Name: "foo"}, Spec: FeatureSpec{Description: "foo", Stability: "Stable"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "bar"}, Spec: FeatureSpec{Description: "bar", Stability: "Experimental"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "baz"}, Spec: FeatureSpec{Description: "baz", Stability: "Technical Preview"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "bax"}, Spec: FeatureSpec{Description: "bax", Stability: "Stable"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "qux"}, Spec: FeatureSpec{Description: "qux", Stability: "Experimental"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "quux"}, Spec: FeatureSpec{Description: "quux", Stability: "Stable"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "corge"}, Spec: FeatureSpec{Description: "corge", Stability: "Experimental"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "grault"}, Spec: FeatureSpec{Description: "grault", Stability: "Stable"}},
				},
			},
			featureGateSpec: FeatureGateSpec{
				Features: []FeatureReference{
					// immutable and cannot be toggled
					{Name: "bar", Activate: true},
					// immutable and cannot be toggled
					{Name: "baz", Activate: true},
				},
			},
			want: []string{},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			got := computeImmutableFeatures(tc.featureGateSpec, tc.featureList)
			if diff := sliceDiffIgnoreOrder(got, tc.want); diff != "" {
				t.Errorf("got invalid features %v, want %v, diff: %s", got, tc.want, diff)
			}
		})
	}
}

func TestComputeInvalidStabilityPolicyOverridedFeatures(t *testing.T) {
	testCases := []struct {
		description     string
		featureGateSpec FeatureGateSpec
		oldFeatureGate  *FeatureGate
		want            []string
	}{
		{
			description: "Multiple features in featuregate set PermanentlyVoidAllSupportGuarantees to false after setting it to true initially",
			featureGateSpec: FeatureGateSpec{
				Features: []FeatureReference{
					{Name: "foo", Activate: false},
					{Name: "bar", Activate: true},
					{Name: "baz", Activate: true},
				},
			},
			oldFeatureGate: &FeatureGate{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Spec: FeatureGateSpec{
					Features: []FeatureReference{
						{Name: "foo", Activate: false},
						{Name: "bar", Activate: true, PermanentlyVoidAllSupportGuarantees: true},
						{Name: "baz", Activate: true, PermanentlyVoidAllSupportGuarantees: true},
					},
				},
			},
			want: []string{"bar", "baz"},
		},
		{
			description: "No features in featuregate set PermanentlyVoidAllSupportGuarantees to false after setting it to true initially",
			featureGateSpec: FeatureGateSpec{
				Features: []FeatureReference{
					{Name: "foo", Activate: false},
					{Name: "bar", Activate: true},
				},
			},
			oldFeatureGate: &FeatureGate{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Spec: FeatureGateSpec{
					Features: []FeatureReference{
						{Name: "foo", Activate: false},
						{Name: "bar", Activate: true},
					},
				},
			},
			want: []string{},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			got := computeFeaturesThatOverridedWarranyVoidOverride(tc.featureGateSpec, tc.oldFeatureGate)
			if diff := sliceDiffIgnoreOrder(got, tc.want); diff != "" {
				t.Errorf("got invalid features %v, want %v, diff: %s", got, tc.want, diff)
			}
		})
	}
}

func TestComputeConflictingFeatures(t *testing.T) {
	testCases := []struct {
		description     string
		featureGate     *FeatureGate
		featureGateList *FeatureGateList
		want            []string
	}{
		{
			description: "Multiple conflicting features in featuregate",
			featureGate: &FeatureGate{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-featuregate",
				},
				Spec: FeatureGateSpec{
					Features: []FeatureReference{
						{Name: "foo", Activate: false},
						{Name: "bar", Activate: true},
						{Name: "baz", Activate: true},
					},
				},
			},
			featureGateList: &FeatureGateList{
				TypeMeta: metav1.TypeMeta{},
				ListMeta: metav1.ListMeta{},
				Items: []FeatureGate{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "my-featuregate",
						},
						Spec: FeatureGateSpec{
							Features: []FeatureReference{
								{Name: "foo", Activate: false},
								{Name: "bar", Activate: true},
								{Name: "baz", Activate: true},
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "my-other-featuregate",
						},
						Spec: FeatureGateSpec{
							Features: []FeatureReference{
								{Name: "foo", Activate: false},
								{Name: "bar", Activate: true},
							},
						},
					},
				},
			},
			want: []string{"foo", "bar"},
		},
		{
			description: "No conflicting features in featuregate",
			featureGate: &FeatureGate{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-featuregate",
				},
				Spec: FeatureGateSpec{
					Features: []FeatureReference{
						{Name: "foo", Activate: false},
						{Name: "bar", Activate: true},
					},
				},
			},
			featureGateList: &FeatureGateList{
				TypeMeta: metav1.TypeMeta{},
				ListMeta: metav1.ListMeta{},
				Items: []FeatureGate{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "my-featuregate",
						},
						Spec: FeatureGateSpec{
							Features: []FeatureReference{
								{Name: "foo", Activate: false},
								{Name: "bar", Activate: true},
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "my-other-featuregate",
						},
						Spec: FeatureGateSpec{
							Features: []FeatureReference{
								{Name: "baz", Activate: true},
							},
						},
					},
				},
			},
			want: []string{},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			got := computeConflictingFeatures(tc.featureGate, tc.featureGateList)
			if diff := sliceDiffIgnoreOrder(got, tc.want); diff != "" {
				t.Errorf("got invalid features %v, want %v, diff: %s", got, tc.want, diff)
			}
		})
	}
}

func TestComputeFeaturesThatDoNotExist(t *testing.T) {
	testCases := []struct {
		description     string
		featureGateSpec FeatureGateSpec
		featureList     *FeatureList
		want            []string
	}{
		{
			description: "Multiple features do not exist in cluster",
			featureGateSpec: FeatureGateSpec{
				Features: []FeatureReference{
					{Name: "foo", Activate: false},
					{Name: "bar", Activate: true},
					{Name: "baz", Activate: true},
				},
			},
			featureList: &FeatureList{
				TypeMeta: metav1.TypeMeta{},
				ListMeta: metav1.ListMeta{},
				Items: []Feature{
					{ObjectMeta: metav1.ObjectMeta{Name: "bar"}, Spec: FeatureSpec{Description: "bar", Stability: "Technical Preview"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "bax"}, Spec: FeatureSpec{Description: "bax", Stability: "Stable"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "qux"}, Spec: FeatureSpec{Description: "qux", Stability: "Stable"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "quux"}, Spec: FeatureSpec{Description: "one", Stability: "Stable"}},
				},
			},
			want: []string{"foo", "baz"},
		},
		{
			description: "All features exist in cluster",
			featureGateSpec: FeatureGateSpec{
				Features: []FeatureReference{
					{Name: "foo", Activate: false},
					{Name: "bar", Activate: true},
					{Name: "baz", Activate: true},
				},
			},
			featureList: &FeatureList{
				TypeMeta: metav1.TypeMeta{},
				ListMeta: metav1.ListMeta{},
				Items: []Feature{
					{ObjectMeta: metav1.ObjectMeta{Name: "foo"}, Spec: FeatureSpec{Description: "foo", Stability: "Technical Preview"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "bar"}, Spec: FeatureSpec{Description: "bar", Stability: "Stable"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "baz"}, Spec: FeatureSpec{Description: "baz", Stability: "Stable"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "bax"}, Spec: FeatureSpec{Description: "bax", Stability: "Stable"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "qux"}, Spec: FeatureSpec{Description: "qux", Stability: "Deprecated"}},
					{ObjectMeta: metav1.ObjectMeta{Name: "quux"}, Spec: FeatureSpec{Description: "one", Stability: "Stable"}},
				},
			},
			want: []string{},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			got := computeFeaturesThatDoNotExist(tc.featureGateSpec, tc.featureList)
			if diff := sliceDiffIgnoreOrder(got, tc.want); diff != "" {
				t.Errorf("got invalid features %v, want %v, diff: %s", got, tc.want, diff)
			}
		})
	}
}

// sliceDiffIgnoreOrder returns a human-readable diff of two string slices.
// Two slices are considered equal when they have the same length and same elements. The order of the elements is
// ignored while comparing. Nil and empty slices are considered equal.
//
// This function is intended to be used in tests for comparing expected and actual values, and printing the diff for
// users to debug:
//
//	if diff := sliceDiffIgnoreOrder(got, want); diff != "" {
//	    t.Errorf("got: %v, want: %v, diff: %s", got, want, diff)
//	}
func sliceDiffIgnoreOrder(a, b []string) string {
	return cmp.Diff(a, b, cmpopts.EquateEmpty(), cmpopts.SortSlices(func(x, y string) bool { return x < y }))
}
