// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

//go:build gofuzz
// +build gofuzz

package ini

import (
	"bytes"
)

// Fuzz defines fuzz
func Fuzz(data []byte) int { // nolint:docStub
	b := bytes.NewReader(data)

	if _, err := Parse(b); err != nil {
		return 0
	}

	return 1
}
