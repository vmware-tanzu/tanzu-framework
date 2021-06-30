// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkg

import (
	"context"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterctl "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
)

func providerListFor(infraProvider InfrastructureProvider) *clusterctl.ProviderList {
	return &clusterctl.ProviderList{
		Items: []clusterctl.Provider{
			{
				ProviderName: string(infraProvider),
				Type:         string(clusterctl.InfrastructureProviderType),
			},
		},
	}
}

func TestHasInfrastructureProvider(t *testing.T) {
	discoveryClientFor := func(providerList *clusterctl.ProviderList) (*DiscoveryClient, error) {
		return newFakeDiscoveryClient([]*metav1.APIResourceList{}, Scheme, []runtime.Object{providerList})
	}

	testCases := []struct {
		description    string
		infraProvider  InfrastructureProvider
		providerListFn func(infraProvider InfrastructureProvider) *clusterctl.ProviderList
		err            string
		want           bool
	}{
		{"aws", InfrastructureProviderAWS, providerListFor, "", true},
		{"azure", InfrastructureProviderAzure, providerListFor, "", true},
		{"vsphere", InfrastructureProviderVsphere, providerListFor, "", true},
		{"unknown", InfrastructureProvider("unknown"), providerListFor, "unsupported infrastructure provider", false},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			dc, err := discoveryClientFor(tc.providerListFn(tc.infraProvider))
			if err != nil {
				t.Error(err)
			}

			got, err := dc.HasInfrastructureProvider(context.Background(), tc.infraProvider)
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
				t.Errorf("got=%t, want=%t", got, tc.want)
			}
		})
	}
}
