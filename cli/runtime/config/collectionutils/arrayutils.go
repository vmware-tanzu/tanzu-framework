// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package collectionutils provide helper functions to operate on collections like array, map
package collectionutils

// SomeBool method tests whether at least one element in the array passes the test implemented by the provided function. It returns true if, in the array, it finds an element for which the provided function returns true; otherwise it returns false. It doesn't modify the array.
func SomeBool(arr []bool, condition func(t bool) bool) bool {
	for _, val := range arr {
		if condition(val) {
			return true
		}
	}
	return false
}
