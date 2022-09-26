// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkg

import (
	"context"
	"fmt"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestHasNSX(t *testing.T) {
	discoveryClientFor := func(ns *corev1.Namespace) (*DiscoveryClient, error) {
		return newFakeDiscoveryClient(coreAPIResourceList, Scheme, []runtime.Object{ns})
	}

	testCases := []struct {
		description string
		ns          *corev1.Namespace
		want        bool
	}{
		{"NSX found", &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespaceNSX}}, true},
		{"NSX not found", &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "foo"}}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			dc, err := discoveryClientFor(tc.ns)
			if err != nil {
				t.Error(err)
			}

			got, err := dc.HasNSX(context.Background())
			if err != nil {
				t.Error(err)
			}

			if got != tc.want {
				t.Errorf("got=%t, want=%t", got, tc.want)
			}
		})
	}
}

func TestIsTKGS(t *testing.T) {
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
		{"TKGS exists", discoveryClientWithTKC, false, true},
		{"TKGS does not exist", discoveryClientWithoutTKC, false, false},
	}
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			dc, err := tc.discoveryClientFn()
			if err != nil {
				t.Error(err)
			}
			got, err := dc.IsTKGS(context.Background())
			if err != nil && !tc.errExpected {
				t.Error(err)
			}
			if got != tc.want {
				t.Errorf("got: %t, want %t", got, tc.want)
			}
		})
	}
}

var clusterInfo = `
cluster:
  name: tkg-cluster-wc-765
  type: %s
  plan: dev
  kubernetesProvider: VMware Tanzu Kubernetes Grid
  tkgVersion: 1.2.1
  infrastructure:
    provider: vsphere
`

func metadataConfigMapFor(clusterType string) *corev1.ConfigMap {
	metadata := fmt.Sprintf(clusterInfo, clusterType)
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      metadataConfigMapName,
			Namespace: metadataConfigMapNamespace,
		},
		Data: map[string]string{
			"metadata.yaml": metadata,
		},
	}
}

func discoveryClientForConfigMap(cm *corev1.ConfigMap) (*DiscoveryClient, error) {
	return newFakeDiscoveryClient([]*metav1.APIResourceList{}, Scheme, []runtime.Object{cm})
}

func TestIsManagementCluster(t *testing.T) {
	testCases := []struct {
		description string
		clusterType string
		configMapFn func(clusterType string) *corev1.ConfigMap
		want        bool
		err         string
	}{
		{"management cluster", clusterTypeManagement, metadataConfigMapFor, true, ""},
		{"not management cluster", clusterTypeWorkload, metadataConfigMapFor, false, ""},
		{"unknown cluster type", "foo", metadataConfigMapFor, false, "unknown cluster type: foo"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			dc, err := discoveryClientForConfigMap(tc.configMapFn(tc.clusterType))
			if err != nil {
				t.Error(err)
			}

			got, err := dc.IsManagementCluster(context.Background())
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

func TestIsWorkloadCluster(t *testing.T) {
	testCases := []struct {
		description string
		clusterType string
		configMapFn func(clusterType string) *corev1.ConfigMap
		want        bool
		err         string
	}{
		{"workload cluster", clusterTypeWorkload, metadataConfigMapFor, true, ""},
		{"not management cluster", clusterTypeManagement, metadataConfigMapFor, false, ""},
		{"unknown cluster type", "foo", metadataConfigMapFor, false, "unknown cluster type: foo"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			dc, err := discoveryClientForConfigMap(tc.configMapFn(tc.clusterType))
			if err != nil {
				t.Error(err)
			}

			got, err := dc.IsWorkloadCluster(context.Background())
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
