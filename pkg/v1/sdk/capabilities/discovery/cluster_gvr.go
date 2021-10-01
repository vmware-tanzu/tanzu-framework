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
	name            string
	group           string
	resource        string
	versions        []string
	fieldPaths      []string
	unmatchedGVRs   []string
	unmatchedFields []string
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

// WithFields checks if field path(s) exist in the GVR's schema.
// Field paths must be dot-separated identifiers (e.g. spec.containers).
// This check is done for each version of the resource, if multiple versions are specified in the query.
func (q *QueryGVR) WithFields(paths ...string) *QueryGVR {
	q.fieldPaths = append(q.fieldPaths, paths...)
	return q
}

// Run discovery
func (q *QueryGVR) Run(config *clusterQueryClientConfig) (bool, error) {
	if q.group != "" && q.resource != "" && len(q.versions) == 0 {
		return false, fmt.Errorf("cannot check for group and resource existence without version info; use WithVersion method")
	}

	// Return error if fieldPaths are specifed when no versions or resource are specified. Group can be empty in case
	// core k8s resources.
	if len(q.fieldPaths) > 0 && (len(q.versions) == 0 || q.resource == "") {
		return false, fmt.Errorf("all of group, versions and resource must be specified to check for field existence")
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

	var unmatchedGVRs []string
	var unmatchedFields []string
	for _, gvr := range gvrs {
		// Check group.
		if !q.groupExists(gvr.Group, groupList) {
			unmatchedGVRs = append(unmatchedGVRs, gvr.String())
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
				unmatchedGVRs = append(unmatchedGVRs, gvr.String())
				continue
			}
			return false, err
		}

		// Check resource if specified.
		if gvr.Resource == "" {
			continue
		}
		if !q.resourceExists(resources) {
			unmatchedGVRs = append(unmatchedGVRs, gvr.String())
			continue
		} else if len(q.fieldPaths) > 0 {
			for _, p := range q.fieldPaths {
				found, err := config.openAPISchemaHelper().fieldExistsInGVR(gvr, p)
				if err != nil {
					return false, err
				}
				if !found {
					unmatchedFields = append(unmatchedFields, p)
				}
			}
		}
	}

	q.unmatchedGVRs = unmatchedGVRs
	q.unmatchedFields = unmatchedFields
	return len(unmatchedGVRs) == 0 && len(unmatchedFields) == 0, nil
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
	return fmt.Sprintf("GVRs=%v fields=%v status=unmatched presence=true", q.unmatchedGVRs, q.unmatchedFields)
}
