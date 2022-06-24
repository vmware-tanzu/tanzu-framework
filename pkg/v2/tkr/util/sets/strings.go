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

func (set StringSet) Intersect(others ...StringSet) StringSet {
	return set.Filter(func(s string) bool {
		for _, other := range others {
			if !other.Has(s) {
				return false
			}
		}
		return true
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

func (set StringSet) Union(sets ...StringSet) StringSet {
	sets = append(sets, set)
	sizeBound := 0
	for _, aSet := range sets {
		sizeBound += len(aSet)
	}
	r := make(StringSet, sizeBound)
	for _, aSet := range sets {
		for s := range aSet {
			r[s] = struct{}{}
		}
	}
	return r
}

func (set StringSet) Slice() []string {
	r := make([]string, len(set))
	i := 0
	for key := range set {
		r[i] = key
		i++
	}
	return r
}
