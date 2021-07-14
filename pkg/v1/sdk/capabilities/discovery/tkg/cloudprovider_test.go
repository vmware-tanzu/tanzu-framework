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

func nodeListFor(cloudProvider CloudProvider) *corev1.NodeList {
	return &corev1.NodeList{
		Items: []corev1.Node{
			{
				Spec: corev1.NodeSpec{ProviderID: fmt.Sprintf("%s://xxx-xxxx-xxxx", cloudProvider)},
			},
		},
	}
}

func TestHasCloudProvider(t *testing.T) {
	discoveryClientFor := func(nodeList *corev1.NodeList) (*DiscoveryClient, error) {
		return newFakeDiscoveryClient([]*metav1.APIResourceList{}, Scheme, []runtime.Object{nodeList})
	}

	testCases := []struct {
		description   string
		cloudProvider CloudProvider
		nodeListFn    func(cloudProvider CloudProvider) *corev1.NodeList
		err           string
		want          bool
	}{
		{"aws", CloudProviderAWS, nodeListFor, "", true},
		{"azure", CloudProviderAzure, nodeListFor, "", true},
		{"vsphere", CloudProviderVsphere, nodeListFor, "", true},
		{"empty node list", CloudProviderVsphere, func(cloudProvider CloudProvider) *corev1.NodeList {
			return &corev1.NodeList{}
		}, "node list is empty", false},
		{"unknown", CloudProvider("unknown"), nodeListFor, "unsupported cloud provider", false},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			dc, err := discoveryClientFor(tc.nodeListFn(tc.cloudProvider))
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
