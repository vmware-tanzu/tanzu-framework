// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkg

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestHasTanzuRunGroup(t *testing.T) {
	discoveryClientWithTanzuRunGroup := func() (*DiscoveryClient, error) {
		return newFakeDiscoveryClient(tanzuRunAPIResourceList, Scheme, nil)
	}

	discoveryClientWithTanzuRunGroupMultipleVersions := func() (*DiscoveryClient, error) {
		resources := []*metav1.APIResourceList{{GroupVersion: "run.tanzu.vmware.com/v1"}}
		resources = append(resources, tanzuRunAPIResourceList...)
		return newFakeDiscoveryClient(resources, Scheme, nil)
	}

	discoveryClientWithoutTanzuRunGroup := func() (*DiscoveryClient, error) {
		return newFakeDiscoveryClient([]*metav1.APIResourceList{}, Scheme, nil)
	}

	testCases := []struct {
		description       string
		discoveryClientFn func() (*DiscoveryClient, error)
		versions          []string
		errExpected       bool
		want              bool
	}{
		{"Tanzu run exists without version if WithVersion method was omittied", discoveryClientWithTanzuRunGroup, nil, false, true},
		{"Tanzu run exists with correct version", discoveryClientWithTanzuRunGroup, []string{"v1alpha1"}, false, true},
		{"Tanzu run does not exist with incorrect version", discoveryClientWithTanzuRunGroup, []string{"v1"}, false, false},
		{"Tanzu run exists with correct multiple versions", discoveryClientWithTanzuRunGroupMultipleVersions, []string{"v1alpha1", "v1"}, false, true},
		{"Tanzu run does not exist with incorrect multiple versions", discoveryClientWithTanzuRunGroupMultipleVersions, []string{"v1alpha1", "v2"}, false, false},
		{"Tanzu run does not exist", discoveryClientWithoutTanzuRunGroup, nil, false, false},
	}
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			dc, err := tc.discoveryClientFn()
			if err != nil {
				t.Error(err)
			}
			got, err := dc.HasTanzuRunGroup(context.Background(), tc.versions...)
			if err != nil && !tc.errExpected {
				t.Error(err)
			}
			if got != tc.want {
				t.Errorf("got: %t, want %t", got, tc.want)
			}
		})
	}
}

func TestHasTanzuKubernetesCluster(t *testing.T) {
	discoveryClientWithTKC := func() (*DiscoveryClient, error) {
		return newFakeDiscoveryClient(tanzuRunAPIResourceList, Scheme, nil)
	}

	discoveryClientWithoutTKC := func() (*DiscoveryClient, error) {
		return newFakeDiscoveryClient([]*metav1.APIResourceList{}, Scheme, nil)
	}

	testCases := []struct {
		description       string
		discoveryClientFn func() (*DiscoveryClient, error)
		errExpected       bool
		want              bool
	}{
		{"TKR exists", discoveryClientWithTKC, false, true},
		{"TKR does not exist", discoveryClientWithoutTKC, false, false},
	}
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			dc, err := tc.discoveryClientFn()
			if err != nil {
				t.Error(err)
			}
			got, err := dc.HasTanzuKubernetesClusterV1alpha1(context.Background())
			if err != nil && !tc.errExpected {
				t.Error(err)
			}
			if got != tc.want {
				t.Errorf("got: %t, want %t", got, tc.want)
			}
		})
	}
}

func TestHasTanzuKubernetesRelease(t *testing.T) {
	discoveryClientWithTKR := func() (*DiscoveryClient, error) {
		return newFakeDiscoveryClient(tanzuRunAPIResourceList, Scheme, nil)
	}

	discoveryClientWithoutTKR := func() (*DiscoveryClient, error) {
		return newFakeDiscoveryClient([]*metav1.APIResourceList{}, Scheme, nil)
	}

	testCases := []struct {
		description       string
		discoveryClientFn func() (*DiscoveryClient, error)
		errExpected       bool
		want              bool
	}{
		{"TKR exists", discoveryClientWithTKR, false, true},
		{"TKR does not exist", discoveryClientWithoutTKR, false, false},
	}
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			dc, err := tc.discoveryClientFn()
			if err != nil {
				t.Error(err)
			}
			got, err := dc.HasTanzuKubernetesReleaseV1alpha1(context.Background())
			if err != nil && !tc.errExpected {
				t.Error(err)
			}
			if got != tc.want {
				t.Errorf("got: %t, want %t", got, tc.want)
			}
		})
	}
}
