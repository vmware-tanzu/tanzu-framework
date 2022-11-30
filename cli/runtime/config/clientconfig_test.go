// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
	"golang.org/x/sync/errgroup"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

func cleanupDir(dir string) {
	p, _ := localDirPath(dir)
	_ = os.RemoveAll(p)
}

func randString() string {
	return uuid.NewString()[:5]
}

func TestClientConfig(t *testing.T) {
	LocalDirName = fmt.Sprintf(".tanzu-test-%s", randString())
	server0 := &configapi.Server{
		Name: "test",
		Type: configapi.ManagementClusterServerType,
		ManagementClusterOpts: &configapi.ManagementClusterServer{
			Path: "test",
		},
	}
	testCtx := &configapi.ClientConfig{
		KnownServers: []*configapi.Server{
			server0,
		},
		CurrentServer: "test",
	}
	AcquireTanzuConfigLock()
	err := StoreClientConfig(testCtx)
	require.NoError(t, err)
	ReleaseTanzuConfigLock()
	defer cleanupDir(LocalDirName)
	_, err = GetClientConfig()
	require.NoError(t, err)
	s, err := GetServer("test")
	require.NoError(t, err)
	require.Equal(t, s, server0)
	e, err := ServerExists("test")
	require.NoError(t, err)
	require.True(t, e)
	server1 := &configapi.Server{
		Name: "test1",
		Type: configapi.ManagementClusterServerType,
		ManagementClusterOpts: &configapi.ManagementClusterServer{
			Path: "test1",
		},
	}
	err = SetServer(server1, true)
	require.NoError(t, err)
	c, err := GetClientConfig()
	require.NoError(t, err)
	require.Len(t, c.KnownServers, 2)
	require.Equal(t, c.CurrentServer, "test1")
	err = SetCurrentServer("test")
	require.NoError(t, err)
	c, err = GetClientConfig()
	require.NoError(t, err)
	require.Len(t, c.KnownServers, 2)
	require.Equal(t, "test", c.CurrentServer)
	err = RemoveServer("test")
	require.NoError(t, err)
	c, err = GetClientConfig()
	require.NoError(t, err)
	require.Len(t, c.KnownServers, 1)
	require.Equal(t, "", c.CurrentServer)
	err = SetServer(server1, true)
	require.NoError(t, err)
	c, err = GetClientConfig()
	require.NoError(t, err)
	require.Len(t, c.KnownServers, 1)
	err = RemoveServer("test1")
	require.NoError(t, err)
	c, err = GetClientConfig()
	require.NoError(t, err)
	require.Len(t, c.KnownServers, 0)
	require.Equal(t, c.CurrentServer, "")
	err = DeleteClientConfig()
	require.NoError(t, err)
}

func TestConfigLegacyDir(t *testing.T) {
	r := randString()
	LocalDirName = fmt.Sprintf(".tanzu-test-%s", r)
	// Setup legacy config dir.
	legacyLocalDirName = fmt.Sprintf(".tanzu-test-legacy-%s", r)
	legacyLocalDir, err := legacyLocalDir()
	require.NoError(t, err)
	err = os.MkdirAll(legacyLocalDir, 0755)
	require.NoError(t, err)
	legacyCfgPath, err := legacyConfigPath()
	require.NoError(t, err)
	server0 := &configapi.Server{
		Name: "test",
		Type: configapi.ManagementClusterServerType,
		ManagementClusterOpts: &configapi.ManagementClusterServer{
			Path: "test",
		},
	}
	testCtx := &configapi.ClientConfig{
		KnownServers: []*configapi.Server{
			server0,
		},
		CurrentServer: "test",
	}
	AcquireTanzuConfigLock()
	err = StoreClientConfig(testCtx)
	ReleaseTanzuConfigLock()
	require.NoError(t, err)
	require.FileExists(t, legacyCfgPath)
	defer cleanupDir(LocalDirName)
	defer cleanupDir(legacyLocalDirName)
	_, err = GetClientConfig()
	require.NoError(t, err)
	server1 := &configapi.Server{
		Name: "test1",
		Type: configapi.ManagementClusterServerType,
		ManagementClusterOpts: &configapi.ManagementClusterServer{
			Path: "test1",
		},
	}
	err = SetServer(server1, true)
	require.NoError(t, err)
	c, err := GetClientConfig()
	require.NoError(t, err)
	require.Len(t, c.KnownServers, 2)
	require.Equal(t, c.CurrentServer, "test1")
	err = RemoveServer("test")
	require.NoError(t, err)
	c, err = GetClientConfig()
	require.NoError(t, err)
	require.Len(t, c.KnownServers, 1)
	tmp := LocalDirName
	LocalDirName = legacyLocalDirName
	configCopy, err := GetClientConfig()
	require.NoError(t, err)
	if diff := cmp.Diff(c, configCopy); diff != "" {
		t.Errorf("ClientConfig object mismatch between legacy and new config location (-want +got): \n%s", diff)
	}
	LocalDirName = tmp
	err = DeleteClientConfig()
	require.NoError(t, err)
}

func TestClientConfigUpdateInParallel(t *testing.T) {
	addServer := func(mcName string) error {
		_, err := GetClientConfig()
		if err != nil {
			return err
		}
		s := &configapi.Server{
			Name: mcName,
			Type: configapi.ManagementClusterServerType,
			ManagementClusterOpts: &configapi.ManagementClusterServer{
				Context: "fake-context",
				Path:    "fake-path",
			},
		}
		err = SetServer(s, true)
		if err != nil {
			return err
		}
		_, err = GetClientConfig()
		return err
	}
	// Creates temp configuration file and runs addServer in parallel
	runTestInParallel := func() {
		// Get the temp tanzu config file
		f, err := os.CreateTemp("", "tanzu_config*")
		assert.Nil(t, err)
		defer func(name string) {
			err = os.Remove(name)
			assert.NoError(t, err)
		}(f.Name())
		err = os.Setenv("TANZU_CONFIG", f.Name())
		assert.NoError(t, err)
		// run addServer in parallel
		parallelExecutionCounter := 100
		group, _ := errgroup.WithContext(context.Background())
		for i := 1; i <= parallelExecutionCounter; i++ {
			id := i
			group.Go(func() error {
				return addServer(fmt.Sprintf("mc-%v", id))
			})
		}
		err = group.Wait()
		rawContents, readErr := os.ReadFile(f.Name())
		assert.Nil(t, readErr, "Error reading config: %s", readErr)
		assert.Nil(t, err, "Config file contents: \n%s", rawContents)
		// Make sure that the configuration file is not corrupted
		clientconfig, err := GetClientConfig()
		assert.Nil(t, err)
		// Make sure all expected servers are added to the knownServers list
		assert.Equal(t, parallelExecutionCounter, len(clientconfig.KnownServers))
	}
	// Run the parallel tests of reading and updating the configuration file
	// multiple times to make sure all the attempts are successful
	for testCounter := 1; testCounter <= 5; testCounter++ {
		runTestInParallel()
	}
}

func TestEndpointFromContext(t *testing.T) {
	tcs := []struct {
		name     string
		ctx      *configapi.Context
		endpoint string
		errStr   string
	}{
		{
			name: "success k8s",
			ctx: &configapi.Context{
				Name: "test-mc",
				Type: "k8s",
				ClusterOpts: &configapi.ClusterServer{
					Endpoint:            "test-endpoint",
					Path:                "test-path",
					Context:             "test-context",
					IsManagementCluster: true,
				},
			},
		},
		{
			name: "success tmc current",
			ctx: &configapi.Context{
				Name: "test-tmc",
				Type: "tmc",
				GlobalOpts: &configapi.GlobalServer{
					Endpoint: "test-endpoint",
				},
			},
		},
		{
			name: "failure",
			ctx: &configapi.Context{
				Name: "test-dummy",
				Type: "dummy",
				ClusterOpts: &configapi.ClusterServer{
					Endpoint:            "test-endpoint",
					Path:                "test-path",
					Context:             "test-context",
					IsManagementCluster: true,
				},
			},
			errStr: "unknown server type \"dummy\"",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			endpoint, err := EndpointFromContext(tc.ctx)
			if tc.errStr == "" {
				assert.NoError(t, err)
				assert.Equal(t, "test-endpoint", endpoint)
			} else {
				assert.EqualError(t, err, tc.errStr)
			}
		})
	}
}
