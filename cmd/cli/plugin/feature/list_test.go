// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"testing"

	"k8s.io/client-go/kubernetes/scheme"
	crclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	corev1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/core/v1alpha2"
	"github.com/vmware-tanzu/tanzu-framework/featuregates/client/pkg/featuregateclient"
	"github.com/vmware-tanzu/tanzu-framework/featuregates/client/pkg/featuregateclient/fake"
)

func TestFeatureInfoList(t *testing.T) {
	tests := []struct {
		description string
		featuregate string
		filterVars  map[string]bool
		want        []string
	}{
		{
			description: "retrieve only discoverable, activated features, including experimental",
			featuregate: "",
			filterVars: map[string]bool{
				"activated":            true,
				"deactivated":          false,
				"include-experimental": true,
			},
			want: []string{
				"cloud-event-listener",
				"super-toaster",
				"bazzies",
				"tuner",
			},
		},
		{
			description: "retrieve only discoverable, activated features, including experimental that are gated by specified FeatureGate",
			featuregate: "tkg-system",
			filterVars: map[string]bool{
				"activated":            true,
				"deactivated":          false,
				"include-experimental": true,
			},
			want: []string{
				"cloud-event-listener",
				"super-toaster",
				"bazzies",
			},
		},
		{
			description: "retrieve only discoverable, activated features, excluding experimental",
			featuregate: "",
			filterVars: map[string]bool{
				"activated":            true,
				"deactivated":          false,
				"include-experimental": false,
			},
			want: []string{
				"super-toaster",
				"bazzies",
				"tuner",
			},
		},
		{
			description: "retrieve only discoverable, activated features, excluding experimental that are gated by specified FeatureGate",
			featuregate: "tkg-system",
			filterVars: map[string]bool{
				"activated":            true,
				"deactivated":          false,
				"include-experimental": false,
			},
			want: []string{
				"super-toaster",
				"bazzies",
			},
		},
		{
			description: "retrieve only discoverable, deactivated features, including experimental",
			featuregate: "",
			filterVars: map[string]bool{
				"activated":            false,
				"deactivated":          true,
				"include-experimental": true,
			},
			want: []string{
				"hard-to-get",
				"bar",
				"barries",
				"baz",
				"tuna",
			},
		},
		{
			description: "retrieve only discoverable, deactivated features, including experimental that are gated by specified FeatureGate",
			featuregate: "tkg-system",
			filterVars: map[string]bool{
				"activated":            false,
				"deactivated":          true,
				"include-experimental": true,
			},
			want: []string{
				"hard-to-get",
				"bar",
				"barries",
				"baz",
			},
		},
		{
			description: "retrieve only discoverable, deactivated features, excluding experimental",
			featuregate: "",
			filterVars: map[string]bool{
				"activated":            false,
				"deactivated":          true,
				"include-experimental": false,
			},
			want: []string{
				"bar",
				"barries",
				"baz",
				"baz",
				"tuna",
			},
		},
		{
			description: "retrieve only discoverable, deactivated features, excluding experimental that are gated by specified FeatureGate",
			featuregate: "tkg-system",
			filterVars: map[string]bool{
				"activated":            false,
				"deactivated":          true,
				"include-experimental": false,
			},
			want: []string{
				"bar",
				"barries",
				"baz",
				"hard-to-get",
			},
		},
		{
			description: "retrieve all discoverable features, excluding experimental",
			featuregate: "",
			filterVars: map[string]bool{
				"activated":            false,
				"deactivated":          false,
				"include-experimental": false,
			},
			want: []string{
				"bar",
				"barries",
				"baz",
				"bazzies",
				"hard-to-get",
				"super-toaster",
				"tuna",
				"tuner",
			},
		},

		{
			description: "retrieve all discoverable features, excluding experimental that are gated by specified FeatureGate",
			featuregate: "tkg-system",
			filterVars: map[string]bool{
				"activated":            false,
				"deactivated":          false,
				"include-experimental": false,
			},
			want: []string{
				"bar",
				"barries",
				"baz",
				"bazzies",
				"hard-to-get",
				"super-toaster",
			},
		},
		{
			description: "retrieve all discoverable features, including experimental",
			featuregate: "",
			filterVars: map[string]bool{
				"activated":            false,
				"deactivated":          false,
				"include-experimental": true,
			},
			want: []string{
				"cloud-event-listener",
				"bar",
				"barries",
				"hard-to-get",
				"super-toaster",
				"baz",
				"bazzies",
				"tuna",
				"tuner",
			},
		},
		{
			description: "retrieve all discoverable features, including experimental that are gated by specified FeatureGate",
			featuregate: "tkg-system",
			filterVars: map[string]bool{
				"activated":            false,
				"deactivated":          false,
				"include-experimental": true,
			},
			want: []string{
				"cloud-event-listener",
				"bar",
				"barries",
				"hard-to-get",
				"super-toaster",
				"baz",
				"bazzies",
			},
		},
	}

	objs, _, _ := fake.GetTestObjects()
	s := scheme.Scheme
	if err := corev1alpha2.AddToScheme(s); err != nil {
		t.Fatalf("add config scheme: (%v)", err)
	}

	cl := crclient.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	fgClient, err := featuregateclient.NewFeatureGateClient(featuregateclient.WithClient(cl))
	if err != nil {
		t.Fatalf("get FeatureGate client: (%v)", err)
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			// Set global variables used by Cobra.
			featuregate = tc.featuregate
			activated = tc.filterVars["activated"]
			deactivated = tc.filterVars["deactivated"]
			includeExperimental = tc.filterVars["include-experimental"]

			got, err := featureInfoList(context.Background(), fgClient, featuregate)
			if err != nil {
				t.Errorf("procure featureInfoList: %v", err)
			}

			if len(got) != len(tc.want) {
				t.Errorf("got: %v Features, but want %v", got, tc.want)
			}

			for _, want := range tc.want {
				if !featureInfoSliceContains(got, want) {
					t.Errorf("got: %+v, but list is missing Feature %s", got, want)
				}
			}
		})
	}
}

func featureInfoSliceContains(features []FeatureInfo, name string) bool {
	for _, feat := range features {
		if feat.Name == name {
			return true
		}
	}
	return false
}

func (feature FeatureInfo) String() string {
	return feature.Name
}
