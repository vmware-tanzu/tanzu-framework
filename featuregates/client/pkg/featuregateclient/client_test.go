// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package featuregateclient

import (
	"context"
	"errors"
	"testing"
	"time"

	"k8s.io/client-go/kubernetes/scheme"
	crclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	corev1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/core/v1alpha2"
	"github.com/vmware-tanzu/tanzu-framework/featuregates/client/pkg/featuregateclient/fake"
)

const contextTimeout = 30 * time.Second

func TestGetFeature(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	objs, features, _ := fake.GetTestObjects()
	s := scheme.Scheme
	if err := corev1alpha2.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add config scheme: (%v)", err)
	}
	cl := crclient.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	featureGateClient, err := NewFeatureGateClient(WithClient(cl))
	if err != nil {
		t.Fatalf("Unable to get FeatureGateClient: (%v)", err)
	}

	getFeatureTestCases := []struct {
		description string
		featureName string
		returnErr   bool
	}{
		{
			description: "should successfully return feature",
			featureName: "cloud-event-listener",
			returnErr:   false,
		},
		{
			description: "should return an error when querying for feature that doesn't exist",
			featureName: "xyz",
			returnErr:   true,
		},
	}

	for _, tc := range getFeatureTestCases {
		t.Run(tc.description, func(t *testing.T) {
			feature, err := featureGateClient.GetFeature(ctx, tc.featureName)
			if err != nil {
				if !tc.returnErr {
					t.Errorf("error not expected, but got error: %v", err)
				}
			} else if tc.returnErr {
				if err == nil {
					t.Errorf("error expected, but got nothing")
				}
			}
			if feature != nil && (feature.Name != features[tc.featureName].Name ||
				feature.Status.Activated != features[tc.featureName].Status.Activated ||
				feature.Spec.Stability != features[tc.featureName].Spec.Stability ||
				feature.Spec.Description != features[tc.featureName].Spec.Description) {
				t.Errorf("feature returned is not the correct feature, Expected: %v, Got: %v", features[tc.featureName], feature)
			}
		})
	}
}
func TestGetFeatureList(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	objs, _, _ := fake.GetTestObjects()
	s := scheme.Scheme
	if err := corev1alpha2.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add config scheme: (%v)", err)
	}
	cl := crclient.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	featureGateClient, err := NewFeatureGateClient(WithClient(cl))
	if err != nil {
		t.Fatalf("Unable to get FeatureGateClient: (%v)", err)
	}

	test := struct {
		description string
		want        []string
		returnErr   bool
	}{
		description: "should successfully return features on cluster",
		want: []string{
			"bar",
			"barries",
			"baz",
			"bazzies",
			"biz",
			"cloud-event-listener",
			"cloud-event-speaker",
			"cloud-event-relayer",
			"dodgy-experimental-periscope",
			"foo",
			"super-toaster",
			"tuna",
			"tuner",
		},
		returnErr: false,
	}

	t.Run(test.description, func(t *testing.T) {
		features, err := featureGateClient.GetFeatureList(ctx)
		if err != nil {
			t.Errorf("get FeatureList: %v", err)
		}

		for _, feature := range test.want {
			if !featureListContainsFeature(features, feature) {
				t.Errorf("got: %#v, want: %s feature in list", features.Items, feature)
			}
		}
	})
}

func TestGetFeaturegate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	objs, _, _ := fake.GetTestObjects()
	s := scheme.Scheme
	if err := corev1alpha2.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add config scheme: (%v)", err)
	}
	cl := crclient.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	featureGateClient, err := NewFeatureGateClient(WithClient(cl))
	if err != nil {
		t.Fatalf("Unable to get FeatureGateClient: (%v)", err)
	}

	tests := []struct {
		description  string
		featureGate  string
		wantGate     string
		wantFeatures []string
	}{
		{
			description: "should return specified FeatureGate in cluster",
			featureGate: "tkg-system",
			wantGate:    "tkg-system",
			wantFeatures: []string{
				"cloud-event-listener",
				"dodgy-experimental-periscope",
				"super-toaster",
				"bar",
				"foo",
				"baz",
				"hard-to-get",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			got, err := featureGateClient.GetFeatureGate(ctx, tc.featureGate)
			if err != nil {
				t.Errorf("get Feature %sGate: %v", tc.featureGate, err)
			}

			if got.Name != tc.wantGate {
				t.Errorf("got: %s, want: %s", got.Name, tc.wantGate)
			}

			for _, want := range tc.wantFeatures {
				if !featureGateContainsFeature(got, want) {
					t.Errorf("got: %#v, but missing wanted Feature: %s", got, want)
				}
			}
		})
	}
}

func TestGetFeaturegateList(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	objs, _, _ := fake.GetTestObjects()
	s := scheme.Scheme
	if err := corev1alpha2.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add config scheme: (%v)", err)
	}
	cl := crclient.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	featureGateClient, err := NewFeatureGateClient(WithClient(cl))
	if err != nil {
		t.Fatalf("Unable to get FeatureGateClient: (%v)", err)
	}

	tests := []struct {
		description string
		wantGates   []string
	}{
		{
			description: "should return all featuregates in cluster",
			wantGates: []string{
				"tkg-system",
				"empty-fg",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			got, err := featureGateClient.GetFeatureGateList(ctx)
			if err != nil {
				t.Errorf("get FeatureGateList: %v", err)
			}

			for _, want := range tc.wantGates {
				if !featureGateListContainsFeatureGate(got, want) {
					t.Errorf("got: %#v, but missing wanted FeatureGate: %s", got, want)
				}
			}
		})
	}
}

func TestActivateFeature(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	objs, _, _ := fake.GetTestObjects()
	testScheme := scheme.Scheme
	if err := corev1alpha2.AddToScheme(testScheme); err != nil {
		t.Fatalf("Unable to add config scheme: (%v)", err)
	}
	cl := crclient.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	featureGateClient, err := NewFeatureGateClient(WithClient(cl))
	if err != nil {
		t.Fatalf("Unable to get FeatureGateClient: (%v)", err)
	}

	tests := []struct {
		description       string
		featureName       string
		allowWarrantyVoid bool
		wantErr           error
		wantGateName      string
		wantActivated     bool
		wantVoidWarranty  bool
	}{
		{
			description:       "should successfully activate a technical preview Feature",
			featureName:       "bar",
			allowWarrantyVoid: false,
			wantErr:           nil,
			wantGateName:      "tkg-system",
			wantActivated:     true,
			wantVoidWarranty:  false,
		},
		{
			description:       "should do nothing if Feature is already activated and a previously voided warranty stays voided",
			featureName:       "cloud-event-listener",
			allowWarrantyVoid: false,
			wantErr:           nil,
			wantGateName:      "tkg-system",
			wantActivated:     true,
			wantVoidWarranty:  true,
		},
		{
			description:       "should throw an error when warranty will be voided without user permission",
			featureName:       "foo",
			allowWarrantyVoid: false,
			wantErr:           ErrTypeForbidden,
			wantGateName:      "tkg-system",
			wantActivated:     false,
			wantVoidWarranty:  false,
		},
		{
			description:       "should not throw an error if warranty would've been voided but is already void",
			featureName:       "cloud-event-listener",
			allowWarrantyVoid: false,
			wantErr:           nil,
			wantGateName:      "tkg-system",
			wantActivated:     true,
			wantVoidWarranty:  true,
		},
		{
			description:       "should set warranty void if warranty is allowed and activating voids warranty",
			featureName:       "dodgy-experimental-periscope",
			allowWarrantyVoid: true,
			wantErr:           nil,
			wantGateName:      "tkg-system",
			wantActivated:     true,
			wantVoidWarranty:  true,
		},
		{
			description:       "should throw an error when the feature doesn't exist",
			featureName:       "bax",
			allowWarrantyVoid: false,
			wantErr:           ErrTypeNotFound,
		},
		{
			description:       "should throw an error when the Feature is not referenced in a FeatureGate",
			featureName:       "specialized-toaster",
			allowWarrantyVoid: false,
			wantErr:           ErrTypeNotFound,
			wantActivated:     false,
			wantVoidWarranty:  false,
		},
		{
			description:       "should throw an error when the Feature is referenced in more than one FeatureGate",
			featureName:       "baz",
			allowWarrantyVoid: false,
			wantErr:           ErrTypeTooMany,
			wantActivated:     false,
			wantVoidWarranty:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			err := featureGateClient.ActivateFeature(ctx, tc.featureName, tc.allowWarrantyVoid)

			// Error is expected for ActivateFeature.
			if tc.wantErr != nil {
				if err == nil {
					t.Errorf("no error, want: %v", tc.wantErr)
				}

				if !errors.Is(err, tc.wantErr) {
					t.Errorf("%v, want: %v", err, tc.wantErr)
				}

				return
			}

			// Error is not expected for ActivateFeature.
			if err != nil {
				t.Error(err)
			} else {
				gateList, err := featureGateClient.GetFeatureGateList(ctx)
				if err != nil {
					t.Error(err)
				}

				gateName, featRef := FeatureRefFromGateList(gateList, tc.featureName)
				if gateName != tc.wantGateName {
					t.Errorf("got FeatureGate %s, want: %s", gateName, tc.wantGateName)
				}

				if featRef.Name != tc.featureName {
					t.Errorf("got Feature %s, want: %s", featRef.Name, tc.featureName)
				}

				if featRef.Activate != tc.wantActivated {
					t.Errorf("got Feature %s activated %t, want: %t", tc.featureName, featRef.Activate, tc.wantActivated)
				}

				if featRef.PermanentlyVoidAllSupportGuarantees != tc.wantVoidWarranty {
					t.Errorf("got Feature %s warranty voided: %t, want: %t", featRef.Name, featRef.PermanentlyVoidAllSupportGuarantees, tc.wantVoidWarranty)
				}
			}
		})
	}
}

func TestDeactivateFeature(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	objs, _, _ := fake.GetTestObjects()
	s := scheme.Scheme
	if err := corev1alpha2.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add config scheme: (%v)", err)
	}
	cl := crclient.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	featureGateClient, err := NewFeatureGateClient(WithClient(cl))
	if err != nil {
		t.Fatalf("Unable to get FeatureGateClient: (%v)", err)
	}

	tests := []struct {
		description       string
		featureName       string
		allowWarrantyVoid bool
		wantErr           error
		wantGateName      string
		wantActivated     bool
		wantVoidWarranty  bool
	}{
		{
			description:       "should successfully deactivate a technical preview Feature",
			featureName:       "barries",
			allowWarrantyVoid: false,
			wantErr:           nil,
			wantGateName:      "tkg-system",
			wantActivated:     false,
			wantVoidWarranty:  false,
		},
		{
			description:       "should throw an error when deactivating a stable (immutable) Feature",
			featureName:       "super-toaster",
			allowWarrantyVoid: false,
			wantErr:           ErrTypeForbidden,
			wantGateName:      "tkg-system",
			wantActivated:     true,
			wantVoidWarranty:  false,
		},
		{
			description:       "should throw an error when the feature doesn't exist",
			featureName:       "bax",
			allowWarrantyVoid: false,
			wantErr:           ErrTypeNotFound,
		},
		{
			description:       "should throw an error when the feature doesn't exist",
			featureName:       "bax",
			allowWarrantyVoid: false,
			wantErr:           ErrTypeNotFound,
		},
		{
			description:       "should throw an error when the Feature is not referenced in a FeatureGate",
			featureName:       "specialized-toaster",
			allowWarrantyVoid: false,
			wantErr:           ErrTypeNotFound,
			wantActivated:     false,
			wantVoidWarranty:  false,
		},
		{
			description:       "should throw an error when the Feature is referenced in more than one FeatureGate",
			featureName:       "bazzies",
			allowWarrantyVoid: false,
			wantErr:           ErrTypeTooMany,
			wantActivated:     true,
			wantVoidWarranty:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			_, err := featureGateClient.DeactivateFeature(ctx, tc.featureName)

			// Error is expected for DeactivateFeature.
			if tc.wantErr != nil {
				if err == nil {
					t.Errorf("no error, want: %v", tc.wantErr)
				}

				if !errors.Is(err, tc.wantErr) {
					t.Errorf("%v, want: %v", err, tc.wantErr)
				}

				return
			}

			// Error is not expected for DeactivateFeature.
			if err != nil {
				t.Error(err)
			} else {
				gateList, err := featureGateClient.GetFeatureGateList(ctx)
				if err != nil {
					t.Error(err)
				}

				gateName, featRef := FeatureRefFromGateList(gateList, tc.featureName)
				if gateName != tc.wantGateName {
					t.Errorf("got FeatureGate %s, want: %s", gateName, tc.wantGateName)
				}

				if featRef.Name != tc.featureName {
					t.Errorf("got Feature %s, want: %s", featRef.Name, tc.featureName)
				}

				if featRef.Activate != tc.wantActivated {
					t.Errorf("got Feature %s activated %t, want: %t", tc.featureName, featRef.Activate, tc.wantActivated)
				}

				if featRef.PermanentlyVoidAllSupportGuarantees != tc.wantVoidWarranty {
					t.Errorf("got Feature %s warranty voided: %t, want: %t", featRef.Name, featRef.PermanentlyVoidAllSupportGuarantees, tc.wantVoidWarranty)
				}
			}
		})
	}
}

func featureGateContainsFeature(gate *corev1alpha2.FeatureGate, feature string) bool {
	for _, feat := range gate.Spec.Features {
		if feature == feat.Name {
			return true
		}
	}
	return false
}

func featureGateListContainsFeatureGate(gates *corev1alpha2.FeatureGateList, feature string) bool {
	for _, gate := range gates.Items {
		if feature == gate.Name {
			return true
		}
	}
	return false
}

func featureListContainsFeature(features *corev1alpha2.FeatureList, feature string) bool {
	for _, feat := range features.Items {
		if feature == feat.Name {
			return true
		}
	}
	return false
}
