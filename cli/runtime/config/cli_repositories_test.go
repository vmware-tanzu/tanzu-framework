package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

func TestSetRepository(t *testing.T) {
	func() {
		LocalDirName = fmt.Sprintf(".tanzu-test")
	}()

	defer func() {
		cleanupDir(LocalDirName)
	}()

	tests := []struct {
		name    string
		cfg     *configapi.ClientConfig
		in      configapi.PluginRepository
		out     configapi.PluginRepository
		persist bool
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
			persist: true,
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
			persist: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			err := StoreClientConfig(tc.cfg)
			assert.NoError(t, err)

			persist, err := SetCLIRepository(tc.in)
			assert.Equal(t, tc.persist, persist)
			assert.NoError(t, err)
			r, err := GetCLIRepository(tc.out.GCPPluginRepository.Name)
			assert.Equal(t, tc.out.GCPPluginRepository.Name, r.GCPPluginRepository.Name)
		})
	}

}
