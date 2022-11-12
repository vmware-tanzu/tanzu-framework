// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

func TestGetAllEnvs(t *testing.T) {
	// setup
	func() {
		LocalDirName = TestLocalDirName
	}()
	defer func() {
		cleanupDir(LocalDirName)
	}()
	tests := []struct {
		name   string
		in     *configapi.ClientConfig
		out    map[string]string
		errStr string
	}{
		{
			name: "success k8s",
			in: &configapi.ClientConfig{
				ClientOptions: &configapi.ClientOptions{
					Env: map[string]string{
						"test": "test",
					},
				},
			},
			out: map[string]string{
				"test": "test",
			},
		},
	}
	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			err := StoreClientConfig(spec.in)
			assert.NoError(t, err)
			c, err := GetAllEnvs()
			assert.NoError(t, err)
			assert.Equal(t, spec.out, c)
			assert.NoError(t, err)
		})
	}
}

func TestGetEnv(t *testing.T) {
	// setup
	func() {
		LocalDirName = TestLocalDirName
	}()
	defer func() {
		cleanupDir(LocalDirName)
	}()
	tests := []struct {
		name   string
		in     *configapi.ClientConfig
		out    string
		errStr string
	}{
		{
			name: "success k8s",
			in: &configapi.ClientConfig{
				ClientOptions: &configapi.ClientOptions{
					Env: map[string]string{
						"test": "test",
					},
				},
			},
			out: "test",
		},
	}
	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			err := StoreClientConfig(spec.in)
			assert.NoError(t, err)
			c, err := GetEnv("test")
			assert.NoError(t, err)
			assert.Equal(t, spec.out, c)
			assert.NoError(t, err)
		})
	}
}

func TestSetEnv(t *testing.T) {
	// setup
	func() {
		LocalDirName = TestLocalDirName
	}()
	defer func() {
		cleanupDir(LocalDirName)
	}()
	tests := []struct {
		name    string
		key     string
		val     string
		persist bool
	}{
		{
			name: "should add new env to empty envs",
			key:  "test",
			val:  "test-test",
		},
		{
			name: "should add new env to existing envs",
			key:  "test2",
			val:  "test2",
		},
		{
			name: "should update existing env",
			key:  "test",
			val:  "updated-test",
		},
		{
			name: "should not update same env",
			key:  "test2",
			val:  "test2",
		},
	}

	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			err := SetEnv(spec.key, spec.val)
			assert.NoError(t, err)
			val, err := GetEnv(spec.key)
			assert.Equal(t, spec.val, val)
			assert.NoError(t, err)
		})
	}
}
func TestDeleteEnv(t *testing.T) {
	// setup
	func() {
		LocalDirName = TestLocalDirName
		cfg := &configapi.ClientConfig{
			ClientOptions: &configapi.ClientOptions{
				Env: map[string]string{
					"test":  "test",
					"test2": "test2",
					"test4": "test2",
				},
			},
		}
		err := StoreClientConfig(cfg)
		assert.NoError(t, err)
	}()
	defer func() {
		cleanupDir(LocalDirName)
	}()
	tests := []struct {
		name string
		in   string
		out  bool
	}{
		{
			name: "success delete test",
			in:   "test",
			out:  true,
		},
		{
			name: "success delete test2",
			in:   "test2",
			out:  true,
		},

		{
			name: "success delete test3",
			in:   "test3",
			out:  true,
		},
	}

	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			err := DeleteEnv(spec.in)
			assert.NoError(t, err)
			c, err := GetEnv(spec.in)
			assert.Equal(t, "not found", err.Error())
			assert.Equal(t, spec.out, c == "")
		})
	}
}
