// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

func TestGetMetadata(t *testing.T) {
	// setup
	func() {
		LocalDirName = TestLocalDirName
	}()

	defer func() {
		cleanupDir(LocalDirName)
	}()

	tests := []struct {
		name   string
		in     string
		out    *configapi.Metadata
		errStr string
	}{
		{
			name: "success empty metadata",
			in:   ``,
			out:  &configapi.Metadata{},
		},
		{
			name: "success with patch strategies",
			in: `configMetadata:
  patchStrategy:
    contexts.group: replace
    contexts.clusterOpts.annotation: replace
    contexts.discoverySources.gcp.annotation: replace`,
			out: &configapi.Metadata{
				ConfigMetadata: &configapi.ConfigMetadata{
					PatchStrategy: map[string]string{
						"contexts.group": "replace",
						"contexts.discoverySources.gcp.annotation": "replace",
						"contexts.clusterOpts.annotation":          "replace",
					},
				},
			},
		},
	}
	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			// Setup Input Data
			f, err := os.CreateTemp("", "tanzu_config")
			assert.Nil(t, err)
			err = os.WriteFile(f.Name(), []byte(spec.in), 0644)
			assert.Nil(t, err)
			defer func(name string) {
				err = os.Remove(name)
				assert.NoError(t, err)
			}(f.Name())
			err = os.Setenv(EnvConfigMetadataKey, f.Name())
			assert.NoError(t, err)

			//Test case
			c, err := GetMetadata()
			assert.Equal(t, c, spec.out)
			assert.NoError(t, err)
		})
	}
}

func TestGetConfigMetadata(t *testing.T) {
	// setup
	func() {
		LocalDirName = TestLocalDirName
	}()

	defer func() {
		cleanupDir(LocalDirName)
	}()

	tests := []struct {
		name   string
		in     string
		out    *configapi.ConfigMetadata
		errStr string
	}{
		{
			name:   "success empty metadata",
			in:     ``,
			out:    &configapi.ConfigMetadata{},
			errStr: "config metadata not found",
		},
		{
			name: "success with patch strategies",
			in: `configMetadata:
  patchStrategy:
    contexts.group: replace
    contexts.clusterOpts.annotation: replace
    contexts.discoverySources.gcp.annotation: replace`,
			out: &configapi.ConfigMetadata{
				PatchStrategy: map[string]string{
					"contexts.group": "replace",
					"contexts.discoverySources.gcp.annotation": "replace",
					"contexts.clusterOpts.annotation":          "replace",
				},
			},
		},
	}
	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			// Setup Input Data
			f, err := os.CreateTemp("", "tanzu_config")
			assert.Nil(t, err)
			err = os.WriteFile(f.Name(), []byte(spec.in), 0644)
			assert.Nil(t, err)
			defer func(name string) {
				err = os.Remove(name)
				assert.NoError(t, err)
			}(f.Name())
			err = os.Setenv(EnvConfigMetadataKey, f.Name())
			assert.NoError(t, err)

			//Test case
			c, err := GetConfigMetadata()
			if spec.errStr != "" {
				assert.Equal(t, spec.errStr, err.Error())
			} else {
				assert.Equal(t, c, spec.out)
				assert.NoError(t, err)
			}
		})
	}
}

func TestSetConfigMetadataPatchStrategy(t *testing.T) {
	// setup
	func() {
		LocalDirName = TestLocalDirName
	}()

	defer func() {
		cleanupDir(LocalDirName)
	}()

	tests := []struct {
		name   string
		key    string
		value  string
		errStr string
	}{
		{
			name:  "success add new patch strategy",
			key:   "contexts.group",
			value: "replace",
		},
		{
			name:  "success update existing patch strategy",
			key:   "contexts.group",
			value: "merge",
		},
		{
			name:  "success add existing patch strategy",
			key:   "contexts.group2",
			value: "replace",
		},
		{
			name:   "failed add new patch strategy invalid value",
			key:    "contexts.clusterOpts.annotation",
			value:  "add",
			errStr: "allowed values are replace or merge",
		},
	}
	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			err := SetConfigMetadataPatchStrategy(spec.key, spec.value)
			if spec.errStr != "" {
				assert.Equal(t, err.Error(), spec.errStr)
				c, err := GetConfigMetadataPatchStrategy()
				assert.NoError(t, err)
				assert.Equal(t, "", c[spec.key])
			} else {
				assert.NoError(t, err)
				c, err := GetConfigMetadataPatchStrategy()
				assert.NoError(t, err)
				assert.Equal(t, c[spec.key], spec.value)
			}
		})
	}
}
