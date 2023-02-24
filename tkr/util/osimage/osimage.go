// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package osimage provides helper functions to work with OSImage data.
package osimage

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/labels"
)

const (
	maxLabelValueLength = 63
)

// SetRefLabels sets labels from the data in OSImage.spec.image.ref field.
func SetRefLabels(ls labels.Set, prefix string, ref map[string]interface{}) {
	for name, value := range ref {
		prefixedName := prefix + "-" + name
		if value, ok := value.(map[string]interface{}); ok {
			SetRefLabels(ls, prefixedName, value)
			continue
		}
		ls[prefixedName] = labelFormat(value)
	}
}

func labelFormat(value interface{}) string {
	s := fmt.Sprint(value)
	s = strings.ReplaceAll(s, "+", "---")

	// truncate label value if length is more than maxLabelValueLength
	if len(s) > maxLabelValueLength {
		s = s[:maxLabelValueLength]
	}

	return s
}
