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

	stringcmp "github.com/vmware-tanzu/tanzu-framework/pkg/v1/test/cmp/strings"
)

var resources = openapitesting.NewFakeResources("testdata/test-swagger.json")

type fakeOpenAPIParser struct{}

func (f *fakeOpenAPIParser) Parse() (openapi.Resources, error) {
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
					{Name: "nonstructuralkinds", Namespaced: true, Kind: "NonStructuralKind"},
				},
			},
		},
	}
	oneKindGVR := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "onekinds"}
	nonStructuralKindGVR := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "nonstructuralkinds"}

	o := &openAPISchemaHelper{
		openAPIParser: &fakeOpenAPIParser{},
		restMapper:    restmapper.NewDiscoveryRESTMapper(groupResources),
	}

	testCases := []struct {
		description       string
		gvr               schema.GroupVersionResource
		fieldPaths        []string
		want              bool
		wantInvalidFields []string
		err               string
	}{
		{
			description: "top level field found", gvr: oneKindGVR, fieldPaths: []string{"field1"}, want: true,
		},
		{
			description: "second level field found", gvr: oneKindGVR, fieldPaths: []string{"field1.int"}, want: true,
		},
		{
			description: "field not found", gvr: oneKindGVR, fieldPaths: []string{"foo"}, want: false, wantInvalidFields: []string{"foo"},
		},
		{
			description: "invalid field path", gvr: oneKindGVR, fieldPaths: []string{"field1$int"}, want: false, wantInvalidFields: []string{"field1$int"},
		},
		{
			description: "field path with leading dot", gvr: oneKindGVR, fieldPaths: []string{".field1.int"}, want: true,
		},
		{
			description: "empty field path", gvr: oneKindGVR, fieldPaths: []string{""}, want: false, wantInvalidFields: []string{""},
		},
		{
			description:       "multiple fields",
			gvr:               oneKindGVR,
			fieldPaths:        []string{"field1", "field1.int", "field1.string", ".field1.int", "foo", "field1$int"},
			want:              false,
			wantInvalidFields: []string{"foo", "field1$int"},
		},
		{
			description:       "non-structural schema GVR",
			gvr:               nonStructuralKindGVR,
			fieldPaths:        []string{"foo"},
			want:              false,
			wantInvalidFields: []string{},
			err:               "CRD for GVR must have a structural schema",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			got, gotInvalidFields, err := o.fieldsExistInGVR(tc.gvr, tc.fieldPaths...)
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
				t.Errorf("fieldExistsInGVR for fieldpath %q: got %t, want %t", tc.fieldPaths, got, tc.want)
			}
			if diff := stringcmp.SliceDiffIgnoreOrder(gotInvalidFields, tc.wantInvalidFields); diff != "" {
				t.Errorf("got invalid fields %v, want invalid fields %v", gotInvalidFields, tc.wantInvalidFields)
			}
		})
	}
}
