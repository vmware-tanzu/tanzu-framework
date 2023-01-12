// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"errors"
	"fmt"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
)

var (
	// errInvalidWithResourceMethodArgument occurs when an empty string argument is passed into the WithResource method.
	errInvalidWithResourceMethodArgument = errors.New("WithResource method must have non-empty string argument; omit method to indicate any resource")

	// errInvalidWithVersionsMethodArgument occurs when empty string argument(s) are passed into the WithVersions method.
	errInvalidWithVersionsMethodArgument = errors.New("WithVersions method must be comprised of non-empty string argument(s); omit method to indicate any version")
)

// Group represents any API group that may exist on a cluster.
func Group(queryName, group string) *QueryGVR {
	return &QueryGVR{
		name:  queryName,
		group: group,
	}
}

// nullString represents a string that may be null or not explicitly set.
type nullString struct {
	String string
	IsSet  bool
}

func (s *nullString) set(value string) {
	if s != nil {
		s.String = value
		s.IsSet = true
	}
}

// QueryGVR provides insight to the clusters GVRs
type QueryGVR struct {
	name          string
	group         string
	resource      nullString
	versions      []string
	unmatchedGVRs []string
}

// Name returns the name of the query.
func (q *QueryGVR) Name() string {
	return q.name
}

// WithVersions checks if an API group with all the specified versions exist.
// This method can be omitted to query any version.
func (q *QueryGVR) WithVersions(versions ...string) *QueryGVR {
	q.versions = versions
	return q
}

// WithResource checks if an API group with the specified resource exists.
// This method can be omitted to query any resource.
func (q *QueryGVR) WithResource(resource string) *QueryGVR {
	q.resource.set(resource)
	return q
}

// Run discovery.
func (q *QueryGVR) Run(config *clusterQueryClientConfig) (bool, error) {
	if err := q.validate(config); err != nil {
		return false, fmt.Errorf("failed GroupVersionResource API query validation: %w", err)
	}

	var unmatched []string

	switch {
	case q.versions == nil:
		// q.WithVersions method was omitted and q.versions has not been set.
		// Any version that matches group and/or resource is considered a match.
		unmatchedGroupResource, err := q.unmatchedGroupResource(config)
		if err != nil {
			return false, fmt.Errorf("failed to discover unmatched GroupResource: %w", err)
		}
		if unmatchedGroupResource != "" {
			unmatched = append(unmatched, unmatchedGroupResource)
		}
	case q.resource.String == "":
		unmatchedGroupVersions, err := q.unmatchedGroupVersions(config)
		if err != nil {
			return false, fmt.Errorf("failed to discover unmatched GroupVersions: %w", err)
		}
		unmatched = append(unmatched, unmatchedGroupVersions...)
	case q.resource.String != "":
		unmatchedGVRs, err := q.unmatchedGroupVersionResources(config)
		if err != nil {
			return false, fmt.Errorf("failed to discover unmatched GroupVersionResources: %w", err)
		}
		unmatched = append(unmatched, unmatchedGVRs...)
	}

	q.unmatchedGVRs = unmatched
	return len(unmatched) == 0, nil
}

func (q *QueryGVR) validate(cfg *clusterQueryClientConfig) error {
	if cfg == nil {
		return fmt.Errorf("clusterQueryClientConfig must not be nil")
	}

	var errs []error
	if q.resource.IsSet && q.resource.String == "" {
		errs = append(errs, errInvalidWithResourceMethodArgument)
	}
	if q.versions != nil && containsEmpty(q.versions) {
		errs = append(errs, errInvalidWithVersionsMethodArgument)
	}
	return kerrors.NewAggregate(errs)
}

// containsEmpty looks for an empty string or items containing only white space(s)
// in the given slice of strings.
func containsEmpty(strs []string) bool {
	for _, s := range strs {
		s = strings.TrimSpace(s)
		if s == "" {
			return true
		}
	}
	return false
}

func (q *QueryGVR) unmatchedGroupResource(cfg *clusterQueryClientConfig) (string, error) {
	groups, resources, err := cfg.discoveryClientset.ServerGroupsAndResources()
	if err != nil {
		return "", fmt.Errorf("failed to discover server group and resource: %w", err)
	}

	gvr := schema.GroupVersionResource{
		Group:    q.group,
		Resource: q.resource.String,
	}

	if !q.containsGroup(groups, q.group) {
		return gvr.String(), nil
	}

	// Resource was not specified, meaning there are no unmatched resources.
	if q.resource.String == "" {
		return "", nil
	}

	rscList, err := q.resourceListInGroup(resources, q.group)
	if err != nil {
		return "", fmt.Errorf("failed to find resource list in %s group: %w", q.group, err)
	}
	if !q.resourceExists(rscList) {
		return gvr.String(), nil
	}

	return "", nil
}

func (q *QueryGVR) unmatchedGroupVersions(cfg *clusterQueryClientConfig) ([]string, error) {
	groupList, err := cfg.discoveryClientset.ServerGroups()
	if err != nil {
		return nil, fmt.Errorf("failed to discover server groups: %w", err)
	}

	var unmatched []string

	group := q.groupFromGroupList(groupList)
	if group == nil {
		// No group version matches because group could not be found.
		for _, version := range q.versions {
			unmatched = append(unmatched, schema.GroupVersionResource{
				Group:    q.group,
				Version:  version,
				Resource: q.resource.String,
			}.String())
		}
		return unmatched, nil
	}

	// Group was found. Find which query versions are not.
	for _, queryVersion := range q.versions {
		var matched bool
		for _, discovery := range group.Versions {
			if strings.EqualFold(discovery.Version, queryVersion) {
				matched = true
				continue
			}
		}
		if !matched {
			unmatched = append(unmatched, schema.GroupVersionResource{
				Group:    q.group,
				Version:  queryVersion,
				Resource: q.resource.String,
			}.String())
		}
	}
	return unmatched, nil
}

func (q *QueryGVR) unmatchedGroupVersionResources(cfg *clusterQueryClientConfig) ([]string, error) {
	var unmatched []string
	for _, ver := range q.versions {
		gvr := schema.GroupVersionResource{
			Group:    q.group,
			Version:  ver,
			Resource: q.resource.String,
		}
		resources, err := cfg.discoveryClientset.ServerResourcesForGroupVersion(gvr.GroupVersion().String())
		if err != nil {
			// Second condition is because fake discovery client does not return a proper NotFound error.
			if apierrors.IsNotFound(err) || strings.Contains(
				err.Error(),
				fmt.Sprintf("GroupVersion %q not found", gvr.GroupVersion().String()),
			) {
				unmatched = append(unmatched, gvr.String())
			} else {
				return nil, err
			}
		}
		if !q.resourceExists([]*metav1.APIResourceList{resources}) {
			unmatched = append(unmatched, gvr.String())
		}
	}
	return unmatched, nil
}

// containsGroup looks for a group in a slice of APIGroup.
func (q *QueryGVR) containsGroup(groups []*metav1.APIGroup, group string) bool {
	for _, grp := range groups {
		if strings.EqualFold(grp.Name, group) {
			return true
		}
	}
	return false
}

// resourceListInGroup looks within a slice of APIResourceList for an
// APIResourceList that is in a specified group.
func (q *QueryGVR) resourceListInGroup(resourceLists []*metav1.APIResourceList, group string) ([]*metav1.APIResourceList, error) {
	var list []*metav1.APIResourceList
	for _, rscList := range resourceLists {
		gv, err := schema.ParseGroupVersion(rscList.GroupVersion)
		if err != nil {
			return nil, err
		}

		if strings.EqualFold(gv.Group, group) {
			list = append(list, rscList)
		}
	}
	return list, nil
}

// groupFromGroupList looks for a particular group from an APIGroupList.
func (q *QueryGVR) groupFromGroupList(groupList *metav1.APIGroupList) *metav1.APIGroup {
	for i := range groupList.Groups {
		if strings.EqualFold(groupList.Groups[i].Name, q.group) {
			return &groupList.Groups[i]
		}
	}
	return nil
}

// resourceExists iterates over a slice of APIResourceList and checks if the resource in the query exists
func (q *QueryGVR) resourceExists(resources []*metav1.APIResourceList) bool {
	if len(resources) == 0 {
		return false
	}
	for i := range resources {
		if resources[i] != nil {
			for j := range resources[i].APIResources {
				if resources[i].APIResources[j].Name == q.resource.String {
					return true
				}
			}
		}
	}
	return false
}

// Reason surfaces what didn't match.
func (q *QueryGVR) Reason() string {
	return fmt.Sprintf("GVRs=%v status=unmatched presence=true", q.unmatchedGVRs)
}
