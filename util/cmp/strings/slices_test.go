// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package strings

import "testing"

func TestSliceDiffIgnoreOrder(t *testing.T) {
	testCases := []struct {
		description string
		a           []string
		b           []string
		diffEmpty   bool
	}{
		{
			"Non empty diff - elements different",
			[]string{"abc", "def"},
			[]string{"abc", "ghi"},
			false,
		},
		{
			"Empty diff - same elements",
			[]string{"abc", "def"},
			[]string{"abc", "def"},
			true,
		},
		{
			"Empty diff - same elements unordered",
			[]string{"abc", "def", "ghi"},
			[]string{"def", "ghi", "abc"},
			true,
		},
		{
			"Empty diff - nil slices",
			nil,
			nil,
			true,
		},
		{
			"Empty diff - non nil empty slices",
			[]string{},
			[]string{},
			true,
		},
		{
			"Empty diff - nil and non-nil slices",
			nil,
			[]string{},
			true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			diff := SliceDiffIgnoreOrder(tc.a, tc.b)
			if (len(diff) == 0) != tc.diffEmpty {
				t.Errorf("expected diffEmpty to be %t, got diff: %s", tc.diffEmpty, diff)
			}
		})
	}
}
