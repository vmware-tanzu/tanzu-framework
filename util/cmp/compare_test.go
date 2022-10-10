// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmp

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type TestStruct struct {
	S1 string `json:"s1,omitempty"`
	S2 string `json:"s2,omitempty"`
	I1 int    `json:"i1,omitempty"`
	A  *A     `json:"a,omitempty"`
}

type A struct {
	AA string `json:"aa,omitempty"`
}

func TestDefined(t *testing.T) {
	cc := DefinedComparer{}
	for _, test := range []struct {
		name      string
		a         interface{}
		b         interface{}
		shouldErr bool
	}{
		{
			name: "basic test",
			a: TestStruct{
				S1: "test",
			},
			b: TestStruct{
				S1: "test",
				S2: "test2",
			},
			shouldErr: false,
		},
		{
			name: "nested missing",
			a: TestStruct{
				S1: "test",
				A:  &A{AA: "aa"},
			},
			b: TestStruct{
				S1: "test",
				S2: "test2",
			},
			shouldErr: true,
		},
		{
			name: "nested present",
			a: TestStruct{
				S1: "test",
				A:  &A{AA: "aa"},
			},
			b: TestStruct{
				S1: "test",
				S2: "test2",
				A:  &A{AA: "aa"},
			},
			shouldErr: false,
		},
		{
			name: "nested value wrong",
			a: TestStruct{
				S1: "test",
				A:  &A{AA: "aa"},
			},
			b: TestStruct{
				S1: "test",
				S2: "test2",
				I1: 5,
				A:  &A{AA: "b"},
			},
			shouldErr: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := cc.Eq(test.a, test.b)
			if test.shouldErr {
				require.NotNil(t, err, "test should error")
			} else {
				require.Nil(t, err, "test should not error")
			}
		})
	}
}
