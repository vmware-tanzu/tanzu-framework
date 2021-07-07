package discovery

import (
	"fmt"
	"math/rand"

	runv1alpha1 "github.com/vmware-tanzu-private/core/apis/run/v1alpha1"
)

// QueryTargetsToCapabilityResource is a helper function to generate a Capability v1alpha1 resource from a slice of QueryTarget.
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
				Name:     fmt.Sprintf("gvr-%d", rand.Int31()),
				Group:    query.group,
				Versions: query.versions,
				Resource: query.resource,
			}
			gvrQueries = append(gvrQueries, q)
		case *QueryObject:
			q := runv1alpha1.QueryObject{
				Name:               fmt.Sprintf("object-%d", rand.Int31()),
				ObjectReference:    *query.object,
				WithAnnotations:    query.annotationsMap(true),
				WithoutAnnotations: query.annotationsMap(false),
			}
			objectQueries = append(objectQueries, q)
		case *QueryPartialSchema:
			q := runv1alpha1.QueryPartialSchema{
				Name:          fmt.Sprintf("partialSchema-%d", rand.Int31()),
				PartialSchema: query.schema,
			}
			partialSchemaQueries = append(partialSchemaQueries, q)
		default:
			return nil, fmt.Errorf("unknown QueryTarget type: %T", qt)
		}
	}

	capability := &runv1alpha1.Capability{
		Spec: runv1alpha1.CapabilitySpec{
			Query: runv1alpha1.Query{
				GroupVersionResources: gvrQueries,
				Objects:               objectQueries,
				PartialSchemas:        partialSchemaQueries,
			},
		},
	}

	return capability, nil
}
