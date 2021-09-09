// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"testing"

	"k8s.io/client-go/kubernetes/scheme"
	crclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	configv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/sdk/features/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/sdk/features/client/fake"
)

func TestJoinFeatures(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	objs, _, _ := fake.GetTestObjects()
	s := scheme.Scheme
	if err := configv1alpha1.AddToScheme(s); err != nil {
		t.Fatalf("Unable to add config scheme: (%v)", err)
	}
	cl := crclient.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	featureGateClient, err := client.NewFeatureGateClient(client.WithClient(cl))

	if err != nil {
		t.Fatalf("Unable to get FeatureGateClient: (%v)", err)
	}

	joinFeaturesTestCases := []struct {
		description string
		features    []*FeatureInfo
		exists      bool
	}{
		{
			description: "should successfully fetch additional information of an available feature",
			features:    []*FeatureInfo{{Name: "foo", Available: true}},
			exists:      true,
		},
		{
			description: "should successfully fetch additional information of an unavailable feature",
			features:    []*FeatureInfo{{Name: "baz", Available: false}},
			exists:      true,
		},
		{
			description: "should not fetch additional information of an unavailable feature that doesn't exist",
			features:    []*FeatureInfo{{Name: "xyz", Available: false}},
			exists:      false,
		},
	}

	for _, tc := range joinFeaturesTestCases {
		t.Run(tc.description, func(t *testing.T) {
			_ = joinFeatures(ctx, featureGateClient, tc.features)
			if tc.exists && tc.features[0].Available && tc.features[0].Maturity == "" {
				t.Errorf("expected to fetch available %s feature information", tc.features[0].Name)
			} else if tc.exists && !tc.features[0].Available && tc.features[0].Maturity == "" {
				t.Errorf("expected to fetch unavailable %s feature information", tc.features[0].Name)
			} else if !tc.exists && !tc.features[0].Available && tc.features[0].Maturity != "" {
				t.Errorf("expected not to fetch information of unavailable %s feature that doesn't exist", tc.features[0].Name)
			}
		})
	}
}
