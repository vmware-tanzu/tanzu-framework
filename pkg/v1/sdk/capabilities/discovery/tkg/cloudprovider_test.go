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

func nodeFor(cloudProvider CloudProvider) *corev1.Node {
	return &corev1.Node{
		Spec: corev1.NodeSpec{ProviderID: fmt.Sprintf("%s://xxx-xxxx-xxxx", cloudProvider)},
	}
}

func discoveryClientForNode(node *corev1.Node) (*DiscoveryClient, error) {
	var objs []runtime.Object
	if node != nil {
		objs = append(objs, node)
	}
	return newFakeDiscoveryClient([]*metav1.APIResourceList{}, Scheme, objs)
}

func TestHasCloudProvider(t *testing.T) {
	testCases := []struct {
		description   string
		cloudProvider CloudProvider
		nodeFn        func(cloudProvider CloudProvider) *corev1.Node
		err           string
		want          bool
	}{
		{"aws", CloudProviderAWS, nodeFor, "", true},
		{"azure", CloudProviderAzure, nodeFor, "", true},
		{"vsphere", CloudProviderVsphere, nodeFor, "", true},
		{"empty node list", CloudProviderVsphere, func(cloudProvider CloudProvider) *corev1.Node {
			return nil
		}, "node list is empty", false},
		{"unknown", CloudProvider("unknown"), nodeFor, "unsupported cloud provider", false},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			dc, err := discoveryClientForNode(tc.nodeFn(tc.cloudProvider))
			if err != nil {
				t.Error(err)
			}

			got, err := dc.HasCloudProvider(context.Background(), tc.cloudProvider)
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
