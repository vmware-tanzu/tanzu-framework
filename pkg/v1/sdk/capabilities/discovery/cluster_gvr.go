// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"fmt"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Group represents any API group that may exist on a cluster, with ability to specify versions and resource to check for GVR.
func Group(queryName, group string) *QueryGVR {
	return &QueryGVR{
		name:  queryName,
		group: group,
	}
}

// QueryGVR provides insight to the clusters GVRs
type QueryGVR struct {
	name          string
	group         string
	resource      string
	versions      []string
	unmatchedGVRs []string
}

// Name returns the name of the query.
func (q *QueryGVR) Name() string {
	return q.name
}

// WithVersions checks if an API group with all the specified versions exist.
func (q *QueryGVR) WithVersions(versions ...string) *QueryGVR {
	q.versions = versions
	return q
}

// WithResource checks if an API group with the specified resource exists.
// WithVersions needs to be used before calling this function.
func (q *QueryGVR) WithResource(resource string) *QueryGVR {
	q.resource = resource
	return q
}

// Run discovery
func (q *QueryGVR) Run(config *clusterQueryClientConfig) (bool, error) {
	if q.group != "" && q.resource != "" && len(q.versions) == 0 {
		return false, fmt.Errorf("cannot check for group and resource existence without version info; use WithVersion method")
	}

	var gvrs []schema.GroupVersionResource
	if len(q.versions) > 0 {
		for _, v := range q.versions {
			gvrs = append(gvrs, schema.GroupVersionResource{Group: q.group, Version: v, Resource: q.resource})
		}
	} else {
		gvrs = append(gvrs, schema.GroupVersionResource{Group: q.group})
	}

	groupList, err := config.discoveryClientset.ServerGroups()
	if err != nil {
		return false, err
	}

	var unmatched []string
	for _, gvr := range gvrs {
		// Check group.
		if !q.groupExists(gvr.Group, groupList) {
			unmatched = append(unmatched, gvr.String())
			continue
		}

		// Check version if specified.
		if gvr.Version == "" {
			continue
		}

		resources, err := config.discoveryClientset.ServerResourcesForGroupVersion(gvr.GroupVersion().String())
		if err != nil {
			// Second condition is because fake discovery client does not return a proper NotFound error.
			if apierrors.IsNotFound(err) || strings.Contains(err.Error(), fmt.Sprintf("GroupVersion %q not found", gvr.GroupVersion().String())) {
				unmatched = append(unmatched, gvr.String())
				continue
			}
			return false, err
		}

		// Check resource if specified.
		if gvr.Resource == "" {
			continue
		}
		if !q.resourceExists(resources) {
			unmatched = append(unmatched, gvr.String())
			continue
		}
	}

	q.unmatchedGVRs = unmatched
	return len(unmatched) == 0, nil
}

func (q *QueryGVR) groupExists(group string, groupList *metav1.APIGroupList) bool {
	for i := range groupList.Groups {
		if strings.EqualFold(groupList.Groups[i].Name, group) {
			return true
		}
	}
	return false
}

func (q *QueryGVR) resourceExists(resources *metav1.APIResourceList) bool {
	for i := range resources.APIResources {
		if resources.APIResources[i].Name == q.resource {
			return true
		}
	}
	return false
}

// Reason surfaces what didnt match
func (q *QueryGVR) Reason() string {
	return fmt.Sprintf("GVRs=%v status=unmatched presence=true", q.unmatchedGVRs)
}
