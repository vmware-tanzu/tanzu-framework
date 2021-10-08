// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testapigroup "k8s.io/apimachinery/pkg/apis/testapigroup/v1"
	"k8s.io/apimachinery/pkg/runtime"
	apitest "k8s.io/apimachinery/pkg/test"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

var testScheme = runtime.NewScheme()

func init() {
	utilruntime.Must(testapigroup.AddToScheme(testScheme))
}

var carp = corev1.ObjectReference{
	Kind:       "Carp",
	Name:       "test14",
	Namespace:  "testns",
	APIVersion: testapigroup.SchemeGroupVersion.String(),
}

var testAnnotations = map[string]string{
	"cluster.x-k8s.io/provider": "infrastructure-fake",
}

var testObject = Object("carpObj", &carp).WithAnnotations(testAnnotations) // .WithConditions(field)

var testGVR = Group("carpResource", testapigroup.SchemeGroupVersion.Group).WithVersions(testapigroup.SchemeGroupVersion.Version).WithResource("carps")

var testObjects = []runtime.Object{
	&testapigroup.Carp{
		ObjectMeta: metav1.ObjectMeta{Name: "test14", Namespace: "testns", Annotations: testAnnotations},
	},
}

var apiResources = []*metav1.APIResourceList{
	{
		GroupVersion: testapigroup.SchemeGroupVersion.String(),
		APIResources: []metav1.APIResource{
			{
				Name:       "carps",
				Kind:       "Carp",
				Group:      testapigroup.SchemeGroupVersion.Group,
				Version:    testapigroup.SchemeGroupVersion.Version,
				Namespaced: true,
			},
		},
	},
}

func queryClientWithResourcesAndObjects() (*ClusterQueryClient, error) {
	return NewFakeClusterQueryClient(apiResources, testScheme, testObjects)
}

func queryClientWithNoResources() (*ClusterQueryClient, error) {
	return NewFakeClusterQueryClient([]*metav1.APIResourceList{}, testScheme, testObjects)
}

func queryClientWithResourcesAndNoObjects() (*ClusterQueryClient, error) {
	scheme, _ := apitest.TestScheme()
	return NewFakeClusterQueryClient(apiResources, scheme, []runtime.Object{})
}

func TestClusterQueries(t *testing.T) {
	testCases := []struct {
		description       string
		discoveryClientFn func() (*ClusterQueryClient, error)
		queryTargets      []QueryTarget
		want              bool
		err               string
	}{
		{
			description:       "all queries successful",
			discoveryClientFn: queryClientWithResourcesAndObjects,
			queryTargets:      []QueryTarget{testGVR, testObject},
			want:              true,
		},
		{
			description:       "resource not found",
			discoveryClientFn: queryClientWithNoResources,
			queryTargets:      []QueryTarget{testGVR, testObject},
			want:              false,
			err:               "no matches for kind \"Carp\" in version \"testapigroup.apimachinery.k8s.io/v1\"",
		},
		{
			description:       "resource found but not object",
			discoveryClientFn: queryClientWithResourcesAndNoObjects,
			queryTargets:      []QueryTarget{testGVR, testObject},
			want:              false,
			err:               "",
		},
		{
			description:       "no query targets",
			discoveryClientFn: queryClientWithResourcesAndObjects,
			queryTargets:      []QueryTarget{},
			want:              true,
			err:               "",
		},
		{
			description:       "query target names contain duplicates",
			discoveryClientFn: queryClientWithResourcesAndObjects,
			queryTargets:      []QueryTarget{testGVR, testObject, Object("carpObj", &carp)},
			want:              false,
			err:               "query target names must be unique",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			c, err := tc.discoveryClientFn()
			if err != nil {
				t.Error(err)
			}

			q := c.Query(tc.queryTargets...).Prepare()

			got, err := q()
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
