// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"testing"

	"github.com/stretchr/testify/require"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/config"
)

func TestGCPRepository(t *testing.T) {
	modes := []configapi.VersionSelectorLevel{
		configapi.AllUnstableVersions,
		configapi.AlphaUnstableVersions,
		configapi.ExperimentalUnstableVersions,
		configapi.NoUnstableVersions}

	for _, v := range modes {
		testRepository(t, v)
	}
}

func testRepository(t *testing.T, versionSelectorName configapi.VersionSelectorLevel) {
	vs := LoadVersionSelector(versionSelectorName)

	r := configapi.PluginRepository{GCPPluginRepository: &config.CoreGCPBucketRepository}
	repo := loadRepository(r, vs)
	list, err := repo.List()
	require.NoError(t, err)

	require.GreaterOrEqual(t, len(list), 2)

	_, err = repo.Describe("cluster")
	require.NoError(t, err)

	bin, err := repo.Fetch("cluster", VersionLatest, LinuxAMD64)
	require.NoError(t, err)

	require.GreaterOrEqual(t, len(bin), 10)
}
