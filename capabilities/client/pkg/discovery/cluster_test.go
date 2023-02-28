// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"errors"
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

var testGVR = Group("carpResource", testapigroup.SchemeGroupVersion.Group).
	WithVersions(testapigroup.SchemeGroupVersion.Version).
	WithResource("carps")

var testPartialSchemaNotFound = Schema("partialSchemaQuery", "partial schema")
var testPartialSchemaFound = Schema("partialSchemaQuery", "example schema for test")

var testObjects = []runtime.Object{
	&testapigroup.Carp{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "test14",
			Namespace:   "testns",
			Annotations: testAnnotations,
		},
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

func queryClientWithSchema() (*ClusterQueryClient, error) {
	scheme, _ := apitest.TestScheme()
	return NewFakeClusterQueryClientWithSchema(nil, scheme, nil)
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
		{
			description:       "partial schema query not found",
			discoveryClientFn: queryClientWithSchema,
			queryTargets:      []QueryTarget{testPartialSchemaNotFound},
			want:              false,
			err:               "",
		},
		{
			description:       "partial schema query found",
			discoveryClientFn: queryClientWithSchema,
			queryTargets:      []QueryTarget{testPartialSchemaFound},
			want:              true,
			err:               "",
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

// TestGVRQueries tests combinations of GVR queries using WithVersions and
// WithResource methods.
func TestGVRQueries(t *testing.T) {
	// apiResources has information similar to an actual Kubernetes cluster.
	apiResources := []*metav1.APIResourceList{
		{
			GroupVersion: "autoscaling/v1",
			APIResources: []metav1.APIResource{
				{
					Name:       "horizontalpodautoscalers",
					Kind:       "HorizontalPodAutoscaler",
					Namespaced: true,
					ShortNames: []string{"hpa"},
				},
			},
		},
		{
			GroupVersion: "autoscaling/v2",
			APIResources: []metav1.APIResource{
				{
					Name:       "horizontalpodautoscalers",
					Kind:       "HorizontalPodAutoscaler",
					Namespaced: true,
					ShortNames: []string{"hpa"},
				},
			},
		},
		{
			GroupVersion: "autoscaling/v2beta2",
			APIResources: []metav1.APIResource{
				{
					Name:       "horizontalpodautoscalers",
					Kind:       "HorizontalPodAutoscaler",
					Namespaced: true,
					ShortNames: []string{"hpa"},
				},
			},
		},
		{
			GroupVersion: "v1",
			APIResources: []metav1.APIResource{
				{
					Name:       "pods",
					Kind:       "Pod",
					Namespaced: true,
					ShortNames: []string{"po"},
				},
			},
		},
	}

	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("adding core objects to scheme: %v", err)
	}

	testObjects := []runtime.Object{
		&corev1.Pod{
			TypeMeta: metav1.TypeMeta{
				APIVersion: corev1.SchemeGroupVersion.String(),
				Kind:       "Pod",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "testPod",
			},
		},
	}

	// Create client that provides fake cluster information.
	testClient, err := NewFakeClusterQueryClient(apiResources, scheme, testObjects)
	if err != nil {
		t.Fatalf("initiating test client: %v", err)
	}

	testCases := []struct {
		description string
		query       *QueryGVR
		want        bool
		err         error
	}{
		{
			description: "existing core group is found",
			query:       Group("test", ""),
			want:        true,
		},
		{
			description: "existing autoscaling group is found",
			query:       Group("test", "autoscaling"),
			want:        true,
		},
		{
			description: "group not found",
			query:       Group("test", "hotscalers"),
			want:        false,
		},
		{
			description: "group not found, resource found",
			query:       Group("test", "hotscalers").WithResource("horizontalpodautoscalers"),
			want:        false,
		},
		{
			description: "group found, resource not found",
			query:       Group("test", "autoscaling").WithResource("verticalautoscaling"),
			want:        false,
		},
		{
			description: "group and resource found",
			query:       Group("test", "autoscaling").WithResource("horizontalpodautoscalers"),
			want:        true,
		},
		{
			description: "group exists, but empty string resource returns error",
			query:       Group("test", "autoscaling").WithResource(""),
			err:         errInvalidWithResourceMethodArgument,
		},
		{
			description: "group not found, version found",
			query:       Group("test", "hotscaling").WithVersions("v1"),
			want:        false,
		},
		{
			description: "group and version found",
			query:       Group("test", "autoscaling").WithVersions("v1"),
			want:        true,
		},
		{
			description: "group and versions found",
			query:       Group("test", "autoscaling").WithVersions("v2", "v1", "v2beta2"),
			want:        true,
		},
		{
			description: "group found, but empty string version returns error",
			query:       Group("test", "autoscaling").WithVersions(""),
			err:         errInvalidWithVersionsMethodArgument,
		},
		{
			description: "core group and version found",
			query:       Group("test", "").WithVersions("v1"),
			want:        true,
		},
		{
			description: "group found and version not found",
			query:       Group("test", "autoscaling").WithVersions("v1", "v2beta1", "v2beta2"),
			want:        false,
		},
		{
			description: "group, versions, resource found",
			query: Group("test", "autoscaling").
				WithVersions("v1", "v2", "v2beta2").
				WithResource("horizontalpodautoscalers"),
			want: true,
		},
		{
			description: "group, versions found, resource not found",
			query: Group("test", "autoscaling").
				WithVersions("v1", "v2beta1", "v2beta2").
				WithResource("verticalscaler"),
			want: false,
		},
		{
			description: "group, resource found, version not found",
			query: Group("test", "autoscaling").
				WithVersions("v3").
				WithResource("horizontalpodautoscalers"),
			want: false,
		},
		{
			description: "versions, resource found, group not found",
			query: Group("test", "hotscaling").
				WithVersions("v2beta1").
				WithResource("horizontalpodautoscalers"),
			want: false,
		},
		{
			description: "group, versions, resource not found",
			query: Group("test", "hotscaling").
				WithVersions("v3").
				WithResource("verticalscaler"),
			want: false,
		},
		{
			description: "multiple empty versions returns error even if resource exists",
			query: Group("test", "").
				WithVersions("", "", "", "").
				WithResource("pods"),
			err: errInvalidWithVersionsMethodArgument,
		},
		{
			description: "multiple empty versions and empty resource returns invalid version error",
			query: Group("test", "").
				WithVersions("", "", "", "").
				WithResource(""),
			err: errInvalidWithVersionsMethodArgument,
		},
		{
			description: "multiple empty versions and empty resource returns invalid resource error",
			query: Group("test", "autoscaling").
				WithVersions("", "", "", "").
				WithResource(""),
			err: errInvalidWithResourceMethodArgument,
		},
		{
			description: "having empty string version among many returns an error",
			query:       Group("test", "").WithVersions("", "v1", "v2", "v3"),
			err:         errInvalidWithVersionsMethodArgument,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			query := testClient.Query(tc.query)

			_, err := query.Execute()
			if tc.err == nil && err != nil {
				t.Fatalf("want: no error, got: %v", err)
			}

			if tc.err != nil {
				if !errors.Is(err, tc.err) {
					t.Errorf("want error: %v, got: %v", tc.err, err)
				}
			}

			if tc.err == nil {
				got := query.Results().ForQuery("test")

				if got.Found != tc.want {
					t.Errorf("want: found %t, got: found %t", tc.want, got.Found)
				}

				// API resources that are not found should have empty reason.
				if got.Found {
					if got.NotFoundReason != "" {
						t.Errorf("want: empty not found reason, got: %s", got.NotFoundReason)
					}
				}

				// API resources that are found should have reason(s).
				if !got.Found {
					if len(got.NotFoundReason) == 0 {
						t.Errorf("want: not found reason, got empty reason")
					}
					t.Logf("want: not found reason, got: %s", got.NotFoundReason)
				}
			}
		})
	}
}
