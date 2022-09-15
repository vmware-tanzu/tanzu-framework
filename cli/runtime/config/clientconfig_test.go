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

	configv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
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
	server0 := &configv1alpha1.Server{
		Name: "test",
		Type: configv1alpha1.ManagementClusterServerType,
		ManagementClusterOpts: &configv1alpha1.ManagementClusterServer{
			Path: "test",
		},
	}
	testCtx := &configv1alpha1.ClientConfig{
		KnownServers: []*configv1alpha1.Server{
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

	server1 := &configv1alpha1.Server{
		Name: "test1",
		Type: configv1alpha1.ManagementClusterServerType,
		ManagementClusterOpts: &configv1alpha1.ManagementClusterServer{
			Path: "test1",
		},
	}

	err = AddServer(server1, true)
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

	err = PutServer(server1, true)
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

	server0 := &configv1alpha1.Server{
		Name: "test",
		Type: configv1alpha1.ManagementClusterServerType,
		ManagementClusterOpts: &configv1alpha1.ManagementClusterServer{
			Path: "test",
		},
	}
	testCtx := &configv1alpha1.ClientConfig{
		KnownServers: []*configv1alpha1.Server{
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

	server1 := &configv1alpha1.Server{
		Name: "test1",
		Type: configv1alpha1.ManagementClusterServerType,
		ManagementClusterOpts: &configv1alpha1.ManagementClusterServer{
			Path: "test1",
		},
	}

	err = AddServer(server1, true)
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

func TestGetDiscoverySources(t *testing.T) {
	assert := assert.New(t)

	tanzuConfigBytes := `apiVersion: config.tanzu.vmware.com/v1alpha1
clientOptions:
  cli:
    useContextAwareDiscovery: true
current: mgmt
kind: ClientConfig
metadata:
  creationTimestamp: null
servers:
- managementClusterOpts:
    context: mgmt-admin@mgmt
    path: config
  name: mgmt
  type: managementcluster
`
	f, err := os.CreateTemp("", "tanzu_config")
	assert.Nil(err)
	err = os.WriteFile(f.Name(), []byte(tanzuConfigBytes), 0644)
	assert.Nil(err)
	defer os.Remove(f.Name())
	os.Setenv("TANZU_CONFIG", f.Name())

	pds := GetDiscoverySources("mgmt")
	assert.Equal(1, len(pds))
	assert.Equal(pds[0].Kubernetes.Name, "default-mgmt")
	assert.Equal(pds[0].Kubernetes.Path, "config")
	assert.Equal(pds[0].Kubernetes.Context, "mgmt-admin@mgmt")

	// Test tmc global server
	tanzuConfigBytes = `apiVersion: config.tanzu.vmware.com/v1alpha1
clientOptions:
  cli:
    useContextAwareDiscovery: true
current: tmc-test
kind: ClientConfig
metadata:
  creationTimestamp: null
servers:
- globalOpts:
    endpoint: test.cloud.vmware.com:443
  name: tmc-test
  type: global
`
	tf, err := os.CreateTemp("", "tanzu_tmc_config")
	assert.Nil(err)
	err = os.WriteFile(tf.Name(), []byte(tanzuConfigBytes), 0644)
	assert.Nil(err)
	defer os.Remove(tf.Name())
	os.Setenv("TANZU_CONFIG", tf.Name())

	pds = GetDiscoverySources("tmc-test")
	assert.Equal(1, len(pds))
	assert.Equal(pds[0].REST.Endpoint, "https://test.cloud.vmware.com")
	assert.Equal(pds[0].REST.BasePath, "v1alpha1/system/binaries/plugins")
	assert.Equal(pds[0].REST.Name, "default-tmc-test")
}

func TestClientConfigUpdateInParallel(t *testing.T) {
	assert := assert.New(t)
	addServer := func(mcName string) error {
		_, err := GetClientConfig()
		if err != nil {
			return err
		}

		s := &configv1alpha1.Server{
			Name: mcName,
			Type: configv1alpha1.ManagementClusterServerType,
			ManagementClusterOpts: &configv1alpha1.ManagementClusterServer{
				Context: "fake-context",
				Path:    "fake-path",
			},
		}
		err = AddServer(s, true)
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
		assert.Nil(err)
		os.Setenv("TANZU_CONFIG", f.Name())

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
		assert.Nil(readErr, "Error reading config: %s", readErr)
		assert.Nil(err, "Config file contents: \n%s", rawContents)

		// Make sure that the configuration file is not corrupted
		clientconfig, err := GetClientConfig()
		assert.Nil(err)

		// Make sure all expected servers are added to the knownServers list
		assert.Equal(parallelExecutionCounter, len(clientconfig.KnownServers))
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
		ctx      *configv1alpha1.Context
		endpoint string
		errStr   string
	}{
		{
			name: "success k8s",
			ctx: &configv1alpha1.Context{
				Name: "test-mc",
				Type: configv1alpha1.CtxTypeK8s,
				ClusterOpts: &configv1alpha1.ClusterServer{
					Endpoint:            "test-endpoint",
					Path:                "test-path",
					Context:             "test-context",
					IsManagementCluster: true,
				},
			},
		},
		{
			name: "success tmc current",
			ctx: &configv1alpha1.Context{
				Name: "test-tmc",
				Type: configv1alpha1.CtxTypeTMC,
				GlobalOpts: &configv1alpha1.GlobalServer{
					Endpoint: "test-endpoint",
				},
			},
		},
		{
			name: "failure",
			ctx: &configv1alpha1.Context{
				Name: "test-dummy",
				Type: "dummy",
				ClusterOpts: &configv1alpha1.ClusterServer{
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
