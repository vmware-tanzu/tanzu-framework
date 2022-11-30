// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

func TestSetGetRepository(t *testing.T) {
	func() {
		LocalDirName = TestLocalDirName
	}()
	defer func() {
		cleanupDir(LocalDirName)
	}()
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
