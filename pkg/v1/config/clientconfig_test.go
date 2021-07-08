// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"

	configv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
)

func cleanupDir(dir string) {
	p, _ := localDirPath(dir)
	_ = os.RemoveAll(p)
}

func randString() string {
	return uuid.NewV4().String()[:5]
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

	err := StoreClientConfig(testCtx)
	require.NoError(t, err)

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

	err = RemoveServer("test")
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

	err = StoreClientConfig(testCtx)
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
