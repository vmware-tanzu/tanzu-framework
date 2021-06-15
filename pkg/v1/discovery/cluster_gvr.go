// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GVR represents any resource that may exist on a cluster, with ability to specify:
// todo: add WithFields()
func GVR(gvr *schema.GroupVersionResource) *QueryGVR {
	return &QueryGVR{
		gvr:      gvr,
		presence: true,
	}
}

// QueryGVR provides insight to the clusters GVRs
type QueryGVR struct {
	gvr      *schema.GroupVersionResource
	presence bool
}

// Run discovery
func (q *QueryGVR) Run(config *clusterQueryClientConfig) (bool, error) {
	groupVersion := q.gvr.GroupVersion().String()
	_, err := config.discoveryClientset.ServerGroups()
	if err != nil {
		return false, err
	}
	resources, err := config.discoveryClientset.ServerResourcesForGroupVersion(groupVersion)
	if err != nil {
		return false, err
	}

	// Ensure the state of the resource matches intent
	if q.presence != q.gvrExists(resources) {
		return false, nil
	}

	return true, nil
}

func (q *QueryGVR) gvrExists(resources *metav1.APIResourceList) bool {
	for i := range resources.APIResources {
		if resources.APIResources[i].Name == q.gvr.Resource {
			return true
		}
	}
	return false
}

// Reason surfaces what didnt match
func (q *QueryGVR) Reason() string {
	return fmt.Sprintf("resource=%s status=unmatched presence=%t", q.gvr.Resource, q.presence)
}
