// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	kubectl "k8s.io/kubectl/pkg/explain"
	"k8s.io/kubectl/pkg/util/openapi"
)

// openAPISchemaHelper provides methods to work with a server's OpenAPI schema.
type openAPISchemaHelper struct {
	openAPIGetter openapi.Getter
	restMapper    meta.RESTMapper
}

// openAPISchema returns OpenAPI schema for the server.
// The schema is cached in the underlying getter, so repeated calls are okay.
func (o *openAPISchemaHelper) openAPISchema() (openapi.Resources, error) {
	return o.openAPIGetter.Get()
}

// fieldExistsInGVR checks if a field exists in a GVR's schema.
func (o *openAPISchemaHelper) fieldExistsInGVR(gvr schema.GroupVersionResource, fieldPath string) (bool, error) {
	gvk, err := o.restMapper.KindFor(gvr)
	if err != nil {
		return false, err
	}

	openapischema, err := o.openAPISchema()
	if err != nil {
		return false, fmt.Errorf("failed to get OpenAPI schema: %w", err)
	}
	s := openapischema.LookupResource(gvk)
	_, err = kubectl.LookupSchemaForField(s, splitFields(fieldPath))
	if err != nil {
		// Lookup does not return custom error, so we need to compare error string :(
		if strings.Contains(err.Error(), "does not exist") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// newOpenAPISchemaHelper returns an instance of OpenAPISchemaHelper.
func newOpenAPISchemaHelper(openAPIClient discovery.OpenAPISchemaInterface, restMapper meta.RESTMapper) *openAPISchemaHelper {
	return &openAPISchemaHelper{openAPIGetter: openapi.NewOpenAPIGetter(openAPIClient), restMapper: restMapper}
}

// splitFields splits a dot-separated field name into a string slice.
func splitFields(fieldPath string) []string {
	fieldPath = strings.TrimPrefix(fieldPath, ".")
	fieldPath = strings.TrimSuffix(fieldPath, ".")
	return strings.Split(fieldPath, ".")
}
