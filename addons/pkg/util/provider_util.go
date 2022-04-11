// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"reflect"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ExtractTypedLocalObjectRef extracts the TypedLocalObjectReference from the unstructuredContent by looking at the fields
// which has the specified fieldSuffix.
// Returns a map with groupKind as the key and object names as the value
func ExtractTypedLocalObjectRef(unstructuredContent map[string]interface{}, fieldSuffix string) map[schema.GroupKind][]string {
	if unstructuredContent == nil || fieldSuffix == "" {
		return nil
	}
	var extractedGKs = make(map[schema.GroupKind][]string)
	for k, v := range unstructuredContent {
		if v != nil && reflect.TypeOf(v).Kind() == reflect.Map {
			if strings.HasSuffix(k, fieldSuffix) {
				localObjRef := v.(map[string]interface{})

				if !isValidLocalObjectRef(localObjRef) {
					return nil
				}

				apiGroup := ""
				if localObjRef["apiGroup"] != nil {
					apiGroup = localObjRef["apiGroup"].(string)
				}
				kind := localObjRef["kind"].(string)
				name := localObjRef["name"].(string)
				groupKind := schema.GroupKind{Group: apiGroup, Kind: kind}
				extractedGKs[groupKind] = append(extractedGKs[groupKind], name)
			} else {
				extractedGVRsFromNested := ExtractTypedLocalObjectRef(v.(map[string]interface{}), fieldSuffix)
				// Combine the result from nested fields into the extractedGKs
				for k, v := range extractedGVRsFromNested {
					extractedGKs[k] = append(extractedGKs[k], v...)
				}
			}
		}
	}
	return extractedGKs
}

func isValidLocalObjectRef(localObjRef map[string]interface{}) bool {
	if localObjRef == nil {
		return false
	}
	if _, exist := localObjRef["kind"]; !exist || reflect.TypeOf(localObjRef["kind"]).Kind() != reflect.String {
		return false
	}
	if _, exist := localObjRef["name"]; !exist || reflect.TypeOf(localObjRef["name"]).Kind() != reflect.String {
		return false
	}
	if localObjRef["apiGroup"] != nil && reflect.TypeOf(localObjRef["apiGroup"]).Kind() != reflect.String {
		return false
	}
	return true
}
