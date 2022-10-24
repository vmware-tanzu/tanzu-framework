// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"fmt"
	"math/rand"

	runv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
)

// QueryTargetsToCapabilityResource is a helper function to generate a
// Capability v1alpha1 resource from a slice of QueryTarget.
func QueryTargetsToCapabilityResource(queryTargets []QueryTarget) (*runv1alpha1.Capability, error) {
	var (
		gvrQueries           []runv1alpha1.QueryGVR
		objectQueries        []runv1alpha1.QueryObject
		partialSchemaQueries []runv1alpha1.QueryPartialSchema
	)

	for _, qt := range queryTargets {
		switch query := qt.(type) {
		case *QueryGVR:
			q := runv1alpha1.QueryGVR{
				Name:     fmt.Sprintf("gvr-%d", rand.Int31()), //nolint:gosec
				Group:    query.group,
				Versions: query.versions,
				Resource: query.resource.String,
			}
			gvrQueries = append(gvrQueries, q)
		case *QueryObject:
			q := runv1alpha1.QueryObject{
				Name:               fmt.Sprintf("object-%d", rand.Int31()), //nolint:gosec
				ObjectReference:    *query.object,
				WithAnnotations:    query.annotationsMap(true),
				WithoutAnnotations: query.annotationsMap(false),
			}
			objectQueries = append(objectQueries, q)
		case *QueryPartialSchema:
			q := runv1alpha1.QueryPartialSchema{
				Name:          fmt.Sprintf("partialSchema-%d", rand.Int31()), //nolint:gosec
				PartialSchema: query.schema,
			}
			partialSchemaQueries = append(partialSchemaQueries, q)
		default:
			return nil, fmt.Errorf("unknown QueryTarget type: %T", qt)
		}
	}

	capability := &runv1alpha1.Capability{
		Spec: runv1alpha1.CapabilitySpec{
			Queries: []runv1alpha1.Query{
				{
					Name:                  fmt.Sprintf("query-%d", rand.Int31()), //nolint:gosec
					GroupVersionResources: gvrQueries,
					Objects:               objectQueries,
					PartialSchemas:        partialSchemaQueries,
				},
			},
		},
	}

	return capability, nil
}
