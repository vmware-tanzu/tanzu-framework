// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package collectionutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSomeWithBoolType(t *testing.T) {
	boolTests := []struct {
		name      string
		input     []bool
		condition func(b bool) bool
		out       bool
	}{
		{
			name:  "should return true",
			input: []bool{true, false, false, false},
			condition: func(b bool) bool {
				return b
			},
			out: true,
		},
		{
			name:  "should return true",
			input: []bool{false, false, false, false},
			condition: func(b bool) bool {
				return b
			},
			out: false,
		},
	}

	for _, tc := range boolTests {
		t.Run(tc.name, func(t *testing.T) {
			ans := SomeBool(tc.input, tc.condition)
			assert.Equal(t, ans, tc.out)
		})
	}
}
