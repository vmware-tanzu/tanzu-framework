// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package sets provides sets of different types, e.g. StringSet.
package sets

type StringSet map[string]struct{}

func Strings(ss ...string) StringSet {
	r := make(StringSet, len(ss))
	return r.Add(ss...)
}

func (set StringSet) Add(ss ...string) StringSet {
	for _, s := range ss {
		set[s] = struct{}{}
	}
	return set
}

func (set StringSet) Remove(ss ...string) StringSet {
	for _, s := range ss {
		delete(set, s)
	}
	return set
}

func (set StringSet) Has(s string) bool {
	_, has := set[s]
	return has
}

func (set StringSet) Intersect(other StringSet) StringSet {
	return set.Filter(func(s string) bool {
		return other.Has(s)
	})
}

func (set StringSet) Map(f func(s string) string) StringSet {
	r := make(StringSet, len(set))
	for s := range set {
		r[f(s)] = struct{}{}
	}
	return r
}

func (set StringSet) Filter(f func(s string) bool) StringSet {
	r := make(StringSet, len(set))
	for s := range set {
		if f(s) {
			r[s] = struct{}{}
		}
	}
	return r
}
