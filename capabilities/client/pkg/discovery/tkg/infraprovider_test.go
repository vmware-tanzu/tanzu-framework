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

func providerFor(infraProvider InfrastructureProvider) *clusterctl.Provider {
	return &clusterctl.Provider{
		ProviderName: string(infraProvider),
		Type:         string(clusterctl.InfrastructureProviderType),
	}
}

func discoveryClientForProvider(provider *clusterctl.Provider) (*DiscoveryClient, error) {
	var objs []runtime.Object
	if provider != nil {
		objs = append(objs, provider)
	}
	return newFakeDiscoveryClient([]*metav1.APIResourceList{}, Scheme, objs)
}

func TestHasInfrastructureProvider(t *testing.T) {
	testCases := []struct {
		description   string
		infraProvider InfrastructureProvider
		providerFn    func(infraProvider InfrastructureProvider) *clusterctl.Provider
		err           string
		want          bool
	}{
		{"aws", InfrastructureProviderAWS, providerFor, "", true},
		{"azure", InfrastructureProviderAzure, providerFor, "", true},
		{"vsphere", InfrastructureProviderVsphere, providerFor, "", true},
		{"unknown", InfrastructureProvider("unknown"), providerFor, "unsupported infrastructure provider", false},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			dc, err := discoveryClientForProvider(tc.providerFn(tc.infraProvider))
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
