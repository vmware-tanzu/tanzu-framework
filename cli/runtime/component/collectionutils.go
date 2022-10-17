// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package component provides various components and utility helper functions
package component

// Some method tests whether at least one element in the array passes the test implemented by the provided function. It returns true if, in the array, it finds an element for which the provided function returns true; otherwise it returns false. It doesn't modify the array.
func Some[T any](arr []T, condition func(t T) bool) bool {
	for _, val := range arr {
		if condition(val) {
			return true
		}
	}
	return false
}
