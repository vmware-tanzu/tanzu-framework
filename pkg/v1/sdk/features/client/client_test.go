// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"testing"
	"time"

	"k8s.io/client-go/kubernetes/scheme"
	crclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	configv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/sdk/features/client/fake"
)

const contextTimeout = 30 * time.Second

func TestGetFeature(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	objs, features, _ := fake.GetTestObjects()
	s := scheme.Scheme
	if err := configv1alpha1.AddToScheme(s); err != nil {
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
				feature.Spec.Immutable != features[tc.featureName].Spec.Immutable ||
				feature.Spec.Discoverable != features[tc.featureName].Spec.Discoverable ||
				feature.Spec.Activated != features[tc.featureName].Spec.Activated ||
				feature.Spec.Maturity != features[tc.featureName].Spec.Maturity ||
				feature.Spec.Description != features[tc.featureName].Spec.Description) {
				t.Errorf("feature returned is not the correct feature, Expected: %v, Got: %v", features[tc.featureName], feature)
			}
		})
	}
}

func TestGetFeaturegate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	objs, _, featureGates := fake.GetTestObjects()
	s := scheme.Scheme
	if err := configv1alpha1.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add config scheme: (%v)", err)
	}
	cl := crclient.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	featureGateClient, err := NewFeatureGateClient(WithClient(cl))
	if err != nil {
		t.Fatalf("Unable to get FeatureGateClient: (%v)", err)
	}

	getFeatureGateTestCases := []struct {
		description     string
		featureGateName string
		returnErr       bool
	}{
		{
			description:     "should successfully return featuregate",
			featureGateName: "tkg-system",
			returnErr:       false,
		},
		{
			description:     "should return an error when querying for feature that doesn't exist",
			featureGateName: "bar",
			returnErr:       true,
		},
	}

	for _, tc := range getFeatureGateTestCases {
		t.Run(tc.description, func(t *testing.T) {
			featureGate, err := featureGateClient.GetFeatureGate(ctx, tc.featureGateName)
			if err != nil {
				if !tc.returnErr {
					t.Errorf("error not expected, but got error: %v", err)
				}
			} else if tc.returnErr {
				if err == nil {
					t.Errorf("error expected, but got nothing")
				}
			}
			if featureGate != nil && (featureGate.Name != featureGates[tc.featureGateName].Name ||
				len(featureGate.Spec.Features) != len(featureGates[tc.featureGateName].Spec.Features)) {
				t.Errorf("featuregate returned is not the correct featuregate, Expected: %v, Got: %v", featureGates[tc.featureGateName], featureGate)
			}
		})
	}
}

func TestActivateFeature(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	objs, _, _ := fake.GetTestObjects()
	s := scheme.Scheme
	if err := configv1alpha1.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add config scheme: (%v)", err)
	}
	cl := crclient.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	featureGateClient, err := NewFeatureGateClient(WithClient(cl))
	if err != nil {
		t.Fatalf("Unable to get FeatureGateClient: (%v)", err)
	}

	activateFeatureTestCases := []struct {
		description     string
		featureName     string
		featureGateName string
		returnErr       bool
	}{
		{
			description:     "should successfully activate the feature",
			featureName:     "foo",
			featureGateName: "tkg-system",
			returnErr:       false,
		},
		{
			description:     "should throw an error when changing immutable feature",
			featureName:     "bar",
			featureGateName: "tkg-system",
			returnErr:       true,
		},
		{
			description:     "should throw an error when changing undiscoverable feature",
			featureName:     "baz",
			featureGateName: "tkg-system",
			returnErr:       true,
		},
		{
			description:     "should throw an error when the feature doesn't exist",
			featureName:     "bax",
			featureGateName: "tkg-system",
			returnErr:       true,
		},
		{
			description:     "should throw an error when the featuregate doesn't exist",
			featureName:     "foo",
			featureGateName: "tkg-system-test",
			returnErr:       true,
		},
	}

	for _, tc := range activateFeatureTestCases {
		t.Run(tc.description, func(t *testing.T) {
			err := featureGateClient.ActivateFeature(ctx, tc.featureName, tc.featureGateName)
			if err != nil {
				if !tc.returnErr {
					t.Errorf("error not expected, but got error: %v", err)
				}
			} else if tc.returnErr {
				if err == nil {
					t.Errorf("error expected, but got nothing")
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
	if err := configv1alpha1.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add config scheme: (%v)", err)
	}
	cl := crclient.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	featureGateClient, err := NewFeatureGateClient(WithClient(cl))
	if err != nil {
		t.Fatalf("Unable to get FeatureGateClient: (%v)", err)
	}

	deactivateFeatureTestCases := []struct {
		description     string
		featureName     string
		featureGateName string
		returnErr       bool
	}{
		{
			description:     "should successfully deactivate the feature",
			featureName:     "foo",
			featureGateName: "tkg-system",
			returnErr:       false,
		},
		{
			description:     "should throw an error when changing immutable feature",
			featureName:     "bar",
			featureGateName: "tkg-system",
			returnErr:       true,
		},
		{
			description:     "should throw an error when changing undiscoverable feature",
			featureName:     "baz",
			featureGateName: "tkg-system",
			returnErr:       true,
		},
		{
			description:     "should throw an error when the feature doesn't exist",
			featureName:     "bax",
			featureGateName: "tkg-system",
			returnErr:       true,
		},
		{
			description:     "should throw an error when the featuregate doesn't exist",
			featureName:     "foo",
			featureGateName: "tkg-system-test",
			returnErr:       true,
		},
	}

	for _, tc := range deactivateFeatureTestCases {
		t.Run(tc.description, func(t *testing.T) {
			err := featureGateClient.DeactivateFeature(ctx, tc.featureName, tc.featureGateName)
			if err != nil {
				if !tc.returnErr {
					t.Errorf("error not expected, but got error: %v", err)
				}
			} else if tc.returnErr {
				if err == nil {
					t.Errorf("error expected, but got nothing")
				}
			}
		})
	}
}
