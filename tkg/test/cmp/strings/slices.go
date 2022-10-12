// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package strings

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// SliceDiffIgnoreOrder returns a human-readable diff of two string slices.
// Two slices are considered equal when they have the same length and same elements. The order of the elements is
// ignored while comparing. Nil and empty slices are considered equal.
//
// This function is intended to be used in tests for comparing expected and actual values, and printing the diff for
// users to debug:
//
//	if diff := SliceDiffIgnoreOrder(got, want); diff != "" {
//	    t.Errorf("got: %v, want: %v, diff: %s", got, want, diff)
//	}
func SliceDiffIgnoreOrder(a, b []string) string {
	return cmp.Diff(a, b, cmpopts.EquateEmpty(), cmpopts.SortSlices(func(x, y string) bool { return x < y }))
}
