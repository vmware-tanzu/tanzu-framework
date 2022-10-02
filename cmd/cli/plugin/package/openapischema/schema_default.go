// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package openapischema provides utilities for working with openapi schema
package openapischema

import (
	"bytes"
	"reflect"

	"gopkg.in/yaml.v3"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	structuralschema "k8s.io/apiextensions-apiserver/pkg/apiserver/schema"
	"k8s.io/apimachinery/pkg/runtime"
)

const Object = "object"

// SchemaDefault returns a yaml byte array with values populated according to schema
func SchemaDefault(schema []byte) ([]byte, error) {
	jsonSchemaProps := &apiextensions.JSONSchemaProps{}

	if err := yaml.Unmarshal(schema, jsonSchemaProps); err != nil {
		return nil, err
	}
	s, err := structuralschema.NewStructural(jsonSchemaProps)
	if err != nil {
		return nil, err
	}
	unstructured := make(map[string]interface{})

	schemaDefault(unstructured, s)
	var b bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&b)
	yamlEncoder.SetIndent(2)
	if err := yamlEncoder.Encode(unstructured); err != nil {
		return nil, err
	}
	if err := yamlEncoder.Close(); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

// schemaDefault does defaulting of x depending on default values in s.
// This is adopted from k8s.io/apiextensions-apiserver/pkg/apiserver/schema/defaulting with 2 changes
// 1. Prevent deep copy of int as it panics
// 2. For type object depth first search to see if there is any property with default
//
//gocyclo:ignore
func schemaDefault(x interface{}, s *structuralschema.Structural) {
	if s == nil {
		return
	}

	switch x := x.(type) {
	case map[string]interface{}:
		for k, prop := range s.Properties { //nolint
			if prop.Default.Object == nil {
				shouldCreateDefault := false
				var b []bool
				b = createDefault(&prop, b) //nolint:gosec
				for _, x := range b {
					if x {
						shouldCreateDefault = x
						break
					}
				}
				if shouldCreateDefault {
					prop.Default.Object = make(map[string]interface{})
				} else {
					continue
				}
			}
			if _, found := x[k]; !found || isNonNullableNull(x[k], &prop) { //nolint:gosec
				if isKindInt(prop.Default.Object) {
					x[k] = prop.Default.Object
				} else {
					x[k] = runtime.DeepCopyJSONValue(prop.Default.Object)
				}
			}
		}
		for k := range x {
			if prop, found := s.Properties[k]; found {
				schemaDefault(x[k], &prop)
			} else if s.AdditionalProperties != nil {
				if isNonNullableNull(x[k], s.AdditionalProperties.Structural) {
					if isKindInt(s.AdditionalProperties.Structural.Default.Object) {
						x[k] = s.AdditionalProperties.Structural.Default.Object
					} else {
						x[k] = runtime.DeepCopyJSONValue(s.AdditionalProperties.Structural.Default.Object)
					}
				}
				schemaDefault(x[k], s.AdditionalProperties.Structural)
			}
		}
	case []interface{}:
		for i := range x {
			if isNonNullableNull(x[i], s.Items) {
				if isKindInt(s.Items.Default.Object) {
					x[i] = s.Items.Default.Object
				} else {
					x[i] = runtime.DeepCopyJSONValue(s.Items.Default.Object)
				}
			}
			schemaDefault(x[i], s.Items)
		}
	default:
		// scalars, do nothing
	}
}

func isNonNullableNull(x interface{}, s *structuralschema.Structural) bool {
	return x == nil && s != nil && !s.Generic.Nullable
}

func isKindInt(src interface{}) bool {
	if src != nil && reflect.TypeOf(src).Kind() == reflect.Int {
		return true
	}
	return false
}

func createDefault(structural *structuralschema.Structural, b []bool) []bool {
	for _, v := range structural.Properties { //nolint:gocritic
		// return true if there is a non-nested(not object) with a default value
		if v.Type != Object && v.Default.Object != nil {
			b = append(b, true)
			return b
		}
		if v.Type == Object && v.Default.Object == nil && v.Properties != nil {
			b = append(b, createDefault(&v, b)...) //nolint:gosec
		}
	}
	b = append(b, false)
	return b
}
