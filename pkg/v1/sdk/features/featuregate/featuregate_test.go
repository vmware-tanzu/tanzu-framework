// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package featuregate

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	configv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/test/cmp/strings"
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

			if diff := strings.SliceDiffIgnoreOrder(activated, tc.expectedActivated); diff != "" {
				t.Errorf("got activated features %v, want %v, diff: %s", activated, tc.expectedActivated, diff)
			}

			if diff := strings.SliceDiffIgnoreOrder(deactivated, tc.expectedDeactivated); diff != "" {
				t.Errorf("got deactivated features %v, want %v, diff: %s", deactivated, tc.expectedDeactivated, diff)
			}

			if diff := strings.SliceDiffIgnoreOrder(unavailable, tc.expectedUnavailable); diff != "" {
				t.Errorf("got unavailable features %v, want %v, diff: %s", unavailable, tc.expectedUnavailable, diff)
			}
		})
	}
}
