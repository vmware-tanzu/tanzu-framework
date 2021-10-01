// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/restmapper"
	"k8s.io/kubectl/pkg/util/openapi"
	openapitesting "k8s.io/kubectl/pkg/util/openapi/testing"
)

var resources = openapitesting.NewFakeResources("testdata/test-swagger.json")

type fakeOpenAPIGetter struct{}

func (f *fakeOpenAPIGetter) Get() (openapi.Resources, error) {
	return resources, nil
}

func TestFieldExistsInGVR(t *testing.T) {
	groupResources := []*restmapper.APIGroupResources{
		{
			Group: metav1.APIGroup{
				Versions: []metav1.GroupVersionForDiscovery{
					{Version: "v1"},
				},
				PreferredVersion: metav1.GroupVersionForDiscovery{Version: "v1"},
			},
			VersionedResources: map[string][]metav1.APIResource{
				"v1": {
					{Name: "onekinds", Namespaced: true, Kind: "OneKind"},
				},
			},
		},
	}
	oneKindGVR := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "onekinds"}
	o := &openAPISchemaHelper{
		openAPIGetter: &fakeOpenAPIGetter{},
		restMapper:    restmapper.NewDiscoveryRESTMapper(groupResources),
	}

	testCases := []struct {
		description string
		fieldPath   string
		want        bool
		err         string
	}{
		{
			description: "top level field found",
			fieldPath:   "field1",
			want:        true,
		},
		{
			description: "second level field found",
			fieldPath:   "field1.int",
			want:        true,
		},
		{
			description: "field not found",
			fieldPath:   "foo",
			want:        false,
		},
		{
			description: "invalid field path",
			fieldPath:   "field1$int",
			want:        false,
		},
		{
			description: "field path with leading dot",
			fieldPath:   ".field1.int",
			want:        true,
		},
		{
			description: "empty field path",
			fieldPath:   "",
			want:        false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			got, err := o.fieldExistsInGVR(oneKindGVR, tc.fieldPath)
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
				t.Errorf("fieldExistsInGVR for fieldpath %q: got %t, want %t", tc.fieldPath, got, tc.want)
			}
		})
	}
}
