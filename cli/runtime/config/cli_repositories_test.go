// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

func TestSetGetRepository(t *testing.T) {
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
		name string
		cfg  *configapi.ClientConfig
		in   configapi.PluginRepository
		out  configapi.PluginRepository
	}{
		{
			name: "should persist repository",
			cfg:  &configapi.ClientConfig{},
			in: configapi.PluginRepository{
				GCPPluginRepository: &configapi.GCPPluginRepository{
					Name:       "test",
					BucketName: "bucket",
					RootPath:   "root-path",
				},
			},
			out: configapi.PluginRepository{
				GCPPluginRepository: &configapi.GCPPluginRepository{
					Name:       "test",
					BucketName: "bucket",
					RootPath:   "root-path",
				},
			},
		},
		{
			name: "should not persist same repo",
			cfg: &configapi.ClientConfig{
				ClientOptions: &configapi.ClientOptions{
					CLI: &configapi.CLIOptions{
						Repositories: []configapi.PluginRepository{
							{
								GCPPluginRepository: &configapi.GCPPluginRepository{
									Name:       "test",
									BucketName: "bucket",
									RootPath:   "root-path",
								},
							},
						},
					},
				},
			},
			in: configapi.PluginRepository{
				GCPPluginRepository: &configapi.GCPPluginRepository{
					Name:       "test",
					BucketName: "bucket",
					RootPath:   "root-path",
				},
			},
			out: configapi.PluginRepository{
				GCPPluginRepository: &configapi.GCPPluginRepository{
					Name:       "test",
					BucketName: "bucket",
					RootPath:   "root-path",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := StoreClientConfig(tc.cfg)
			assert.NoError(t, err)
			err = SetCLIRepository(tc.in)
			assert.NoError(t, err)
			r, err := GetCLIRepository(tc.out.GCPPluginRepository.Name)
			assert.NoError(t, err)
			assert.Equal(t, tc.out.GCPPluginRepository.Name, r.GCPPluginRepository.Name)
		})
	}
}
