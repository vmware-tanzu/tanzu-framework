// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
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
			ans := Some(tc.input, tc.condition)
			assert.Equal(t, ans, tc.out)
		})
	}
}

func TestSomeWithContext(t *testing.T) {
	contextTests := []struct {
		name      string
		input     []configapi.Context
		condition func(c configapi.Context) bool
		out       bool
	}{
		{
			name: "should return true",
			input: []configapi.Context{
				{
					Name: "test-mc",
					Type: "k8s",
				},
				{
					Name: "test-mc1",
					Type: "k8s",
				},
				{
					Name: "test-mc2",
					Type: "k8s",
				},
			},
			condition: func(c configapi.Context) bool {
				return c.Name == "test-mc"
			},
			out: true,
		},
		{
			name: "should return true",
			input: []configapi.Context{
				{
					Name: "test-mc",
					Type: "k8s",
				},
				{
					Name: "test-mc1",
					Type: "k8s",
				},
				{
					Name: "test-mc2",
					Type: "k8s",
				},
			},
			condition: func(c configapi.Context) bool {
				return c.Name == "test-mc4"
			},
			out: false,
		},
	}

	for _, tc := range contextTests {
		t.Run(tc.name, func(t *testing.T) {
			ans := Some[configapi.Context](tc.input, tc.condition)
			assert.Equal(t, ans, tc.out)
		})
	}
}
