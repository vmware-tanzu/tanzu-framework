// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/kube-openapi/pkg/util/proto"
	kubectl "k8s.io/kubectl/pkg/explain"
	"k8s.io/kubectl/pkg/util/openapi"
)

// openAPIParser is an interface extracted from openapi.CachedOpenAPIParser methods for testing.
type openAPIParser interface {
	Parse() (openapi.Resources, error)
}

// openAPISchemaHelper provides methods to work with a server's OpenAPI schema.
type openAPISchemaHelper struct {
	// openAPIParser fetches OpenAPI data once and caches the parsed data in memory.
	openAPIParser openAPIParser
	restMapper    meta.RESTMapper
}

// fieldsExistInGVR checks if a list of fields exist in a GVR's schema.
func (o *openAPISchemaHelper) fieldsExistInGVR(gvr schema.GroupVersionResource, fieldPaths ...string) (exists bool, unmatched []string, err error) {
	gvk, err := o.restMapper.KindFor(gvr)
	if err != nil {
		return false, unmatched, err
	}

	openapischema, err := o.openAPIParser.Parse()
	if err != nil {
		return false, unmatched, fmt.Errorf("failed to get OpenAPI schema: %w", err)
	}
	s := openapischema.LookupResource(gvk)
	if _, ok := s.(*proto.Kind); !ok {
		return false, unmatched, fmt.Errorf("failed to lookup fields for GVR %q: CRD for GVR must have a structural schema", stringifyGVR(gvr))
	}

	for _, fp := range fieldPaths {
		_, err = kubectl.LookupSchemaForField(s, splitFields(fp))
		if err != nil {
			// Lookup does not return custom error, so we need to compare error string :(
			if strings.Contains(err.Error(), "does not exist") {
				unmatched = append(unmatched, fp)
			} else {
				return false, unmatched, err
			}
		}
	}
	return len(unmatched) == 0, unmatched, nil
}

// newOpenAPISchemaHelper returns an instance of OpenAPISchemaHelper.
func newOpenAPISchemaHelper(openAPIClient discovery.OpenAPISchemaInterface, restMapper meta.RESTMapper) *openAPISchemaHelper {
	return &openAPISchemaHelper{openAPIParser: openapi.NewOpenAPIParser(openapi.NewOpenAPIGetter(openAPIClient)), restMapper: restMapper}
}

// splitFields splits a dot-separated field name into a string slice.
func splitFields(fieldPath string) []string {
	fieldPath = strings.TrimPrefix(fieldPath, ".")
	fieldPath = strings.TrimSuffix(fieldPath, ".")
	return strings.Split(fieldPath, ".")
}
