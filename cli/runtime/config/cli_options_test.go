// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

func TestGetEdition(t *testing.T) {
	// Setup config data
	f1, err := os.CreateTemp("", "tanzu_config")
	assert.Nil(t, err)
	err = os.WriteFile(f1.Name(), []byte(""), 0644)
	assert.Nil(t, err)

	err = os.Setenv(EnvConfigKey, f1.Name())
	assert.NoError(t, err)

	f2, err := os.CreateTemp("", "tanzu_config_ng")
	assert.Nil(t, err)
	err = os.WriteFile(f2.Name(), []byte(""), 0644)
	assert.Nil(t, err)

	err = os.Setenv(EnvConfigNextGenKey, f2.Name())
	assert.NoError(t, err)

	//Setup metadata
	fMeta, err := os.CreateTemp("", "tanzu_config_metadata")
	assert.Nil(t, err)
	err = os.WriteFile(fMeta.Name(), []byte(""), 0644)
	assert.Nil(t, err)

	err = os.Setenv(EnvConfigMetadataKey, fMeta.Name())
	assert.NoError(t, err)

	// Cleanup
	defer func(name string) {
		err = os.Remove(name)
		assert.NoError(t, err)
	}(f1.Name())

	defer func(name string) {
		err = os.Remove(name)
		assert.NoError(t, err)
	}(f2.Name())

	defer func(name string) {
		err = os.Remove(name)
		assert.NoError(t, err)
	}(fMeta.Name())

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

func TestSetEdition(t *testing.T) {
	// Setup config data
	f1, err := os.CreateTemp("", "tanzu_config")
	assert.Nil(t, err)
	err = os.WriteFile(f1.Name(), []byte(""), 0644)
	assert.Nil(t, err)

	err = os.Setenv(EnvConfigKey, f1.Name())
	assert.NoError(t, err)

	f2, err := os.CreateTemp("", "tanzu_config_ng")
	assert.Nil(t, err)
	err = os.WriteFile(f2.Name(), []byte(""), 0644)
	assert.Nil(t, err)

	err = os.Setenv(EnvConfigNextGenKey, f2.Name())
	assert.NoError(t, err)

	//Setup metadata
	fMeta, err := os.CreateTemp("", "tanzu_config_metadata")
	assert.Nil(t, err)
	err = os.WriteFile(fMeta.Name(), []byte(""), 0644)
	assert.Nil(t, err)

	err = os.Setenv(EnvConfigMetadataKey, fMeta.Name())
	assert.NoError(t, err)

	// Cleanup
	defer func(name string) {
		err = os.Remove(name)
		assert.NoError(t, err)
	}(f1.Name())

	defer func(name string) {
		err = os.Remove(name)
		assert.NoError(t, err)
	}(f2.Name())

	defer func(name string) {
		err = os.Remove(name)
		assert.NoError(t, err)
	}(fMeta.Name())

	tests := []struct {
		name  string
		value string
	}{
		{
			name:  "should persist tanzu when empty client config",
			value: "tanzu",
		},
		{
			name:  "should update and persist update-tanzu",
			value: "update-tanzu",
		},
		{
			name:  "should not persist same value update-tanzu",
			value: "update-tanzu",
		},
	}

	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			err := SetEdition(spec.value)
			assert.NoError(t, err)
			c, err := GetEdition()
			assert.Equal(t, spec.value, c)
			assert.NoError(t, err)
		})
	}
}
