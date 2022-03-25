// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package osimage provides helper functions to work with OSImage data.
package osimage

import (
	"fmt"

	"k8s.io/apimachinery/pkg/labels"
)

// SetRefLabels sets labels from the data in OSImage.spec.image.ref field.
func SetRefLabels(ls labels.Set, prefix string, ref map[string]interface{}) {
	for name, value := range ref {
		prefixedName := prefix + "-" + name
		if value, ok := value.(map[string]interface{}); ok {
			SetRefLabels(ls, prefixedName, value)
			continue
		}
		ls[prefixedName] = fmt.Sprint(value)
	}
}
