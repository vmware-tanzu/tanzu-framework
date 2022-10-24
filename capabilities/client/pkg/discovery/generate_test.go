// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"fmt"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	testapigroup "k8s.io/apimachinery/pkg/apis/testapigroup/v1"

	runv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
)

// Create an unknownQueryType that implements QueryTarget interface to ensure
// the correct error is triggered in one of table test cases.
type unknownQueryType string

func (unk unknownQueryType) Name() string {
	return ""
}
func (unk unknownQueryType) Run(conf *clusterQueryClientConfig) (bool, error) {
	return false, nil
}
func (unk unknownQueryType) Reason() string {
	return ""
}

func TestQueryTargetsToCapabilityResource(t *testing.T) {
	clusterGVR := Group("pondQuery", testapigroup.SchemeGroupVersion.Group).
		WithVersions(testapigroup.SchemeGroupVersion.Version).
		WithResource("carps")

	koi := corev1.ObjectReference{
		Kind:       "Carp",
		Name:       "koi",
		Namespace:  "koi-pond",
		APIVersion: testapigroup.SchemeGroupVersion.String(),
	}
	annotations := map[string]string{
		"cluster.x-k8s.io/provider": "infrastructure-fake",
	}
	clusterObject := Object("pondQuery", &koi).WithAnnotations(annotations)

	clusterSchema := Schema("pondQuery", "partial schema")

	testCases := []struct {
		description  string
		queryTargets []QueryTarget
		want         runv1alpha1.Capability
		err          error
	}{
		{
			description:  "one of each query target type (gvr, object, and partial schema)",
			queryTargets: []QueryTarget{clusterGVR, clusterObject, clusterSchema},
			want: runv1alpha1.Capability{
				Spec: runv1alpha1.CapabilitySpec{
					Queries: []runv1alpha1.Query{
						{
							GroupVersionResources: []runv1alpha1.QueryGVR{
								{
									Group:    testapigroup.SchemeGroupVersion.Group,
									Resource: "carps",
									Versions: []string{testapigroup.SchemeGroupVersion.Version},
								},
							},
							Objects: []runv1alpha1.QueryObject{
								{
									ObjectReference:    koi,
									WithAnnotations:    annotations,
									WithoutAnnotations: map[string]string{},
								},
							},
							PartialSchemas: []runv1alpha1.QueryPartialSchema{
								{
									PartialSchema: "partial schema",
								},
							},
						},
					},
				},
			},
			err: nil,
		},
		{
			description:  "unknown query type",
			queryTargets: []QueryTarget{unknownQueryType("shark")},
			err:          fmt.Errorf("unknown QueryTarget type: %T", unknownQueryType("shark")),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			got, err := QueryTargetsToCapabilityResource(tc.queryTargets)

			if tc.err == nil {
				// Check QueryGVR fields.
				if len(got.Spec.Queries[0].GroupVersionResources[0].Name) == 0 {
					t.Errorf("QueryGVR name: got: %s, but want: a non-empty string", got.Spec.Queries[0].GroupVersionResources[0].Name)
				}
				if got.Spec.Queries[0].GroupVersionResources[0].Group != tc.want.Spec.Queries[0].GroupVersionResources[0].Group {
					t.Errorf("QueryGVR resource group: got: %s, but want: %s", got.Spec.Queries[0].GroupVersionResources[0].Group, tc.want.Spec.Queries[0].GroupVersionResources[0].Group)
				}
				if !reflect.DeepEqual(got.Spec.Queries[0].GroupVersionResources[0].Versions, tc.want.Spec.Queries[0].GroupVersionResources[0].Versions) {
					t.Errorf("QueryGVR versions: got: %+v, but want: %+v", got.Spec.Queries[0].GroupVersionResources[0].Versions, tc.want.Spec.Queries[0].GroupVersionResources[0].Versions)
				}

				// Check QueryObject fields.
				if len(got.Spec.Queries[0].Objects[0].Name) == 0 {
					t.Errorf("QueryObject name: got: %s, but want: a non-empty string", got.Spec.Queries[0].Objects[0].Name)
				}
				if got.Spec.Queries[0].Objects[0].ObjectReference != tc.want.Spec.Queries[0].Objects[0].ObjectReference {
					t.Errorf("QueryObject object reference: got: %+v, but want: %+v", got.Spec.Queries[0].Objects[0].ObjectReference, tc.want.Spec.Queries[0].Objects[0].ObjectReference)
				}
				if !reflect.DeepEqual(got.Spec.Queries[0].Objects[0].WithAnnotations, tc.want.Spec.Queries[0].Objects[0].WithAnnotations) {
					t.Errorf("QueryGVR versions: got: %+v, but want: %+v", got.Spec.Queries[0].Objects[0].WithAnnotations, tc.want.Spec.Queries[0].Objects[0].WithAnnotations)
				}
				if !reflect.DeepEqual(got.Spec.Queries[0].Objects[0].WithoutAnnotations, tc.want.Spec.Queries[0].Objects[0].WithoutAnnotations) {
					t.Errorf("QueryGVR versions: got: %+v, but want: %+v", got.Spec.Queries[0].Objects[0].WithoutAnnotations, tc.want.Spec.Queries[0].Objects[0].WithoutAnnotations)
				}

				// Check QueryPartialSchema fields.
				if len(got.Spec.Queries[0].PartialSchemas[0].Name) == 0 {
					t.Errorf("QueryPartialSchema name: got: %s, but want: a non-empty string", got.Spec.Queries[0].PartialSchemas[0].Name)
				}
				if got.Spec.Queries[0].PartialSchemas[0].PartialSchema != tc.want.Spec.Queries[0].PartialSchemas[0].PartialSchema {
					t.Errorf("QueryPartialSchema partial schema: got: %s, but want: %s", got.Spec.Queries[0].PartialSchemas[0].PartialSchema, tc.want.Spec.Queries[0].PartialSchemas[0].PartialSchema)
				}
			}

			if tc.err != nil {
				if err.Error() != tc.err.Error() {
					t.Errorf("want error: %s but got: %s", tc.err, err)
				}
				if got != nil {
					t.Errorf("expected nil Capability object, but got: %+v", got)
				}
			}
		})
	}
}
