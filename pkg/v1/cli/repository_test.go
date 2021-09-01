// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"testing"

	"github.com/stretchr/testify/require"

	configv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"
)

func TestGCPRepository(t *testing.T) {
	modes := []configv1alpha1.VersionSelectorLevel{
		configv1alpha1.AllUnstableVersions,
		configv1alpha1.AlphaUnstableVersions,
		configv1alpha1.ExperimentalUnstableVersions,
		configv1alpha1.NoUnstableVersions}

	for _, v := range modes {
		testRepository(t, v)
	}
}

func testRepository(t *testing.T, versionSelectorName configv1alpha1.VersionSelectorLevel) {
	vs := LoadVersionSelector(versionSelectorName)

	r := configv1alpha1.PluginRepository{GCPPluginRepository: &config.CoreGCPBucketRepository}
	repo := loadRepository(r, vs)
	list, err := repo.List()
	require.NoError(t, err)

	require.GreaterOrEqual(t, len(list), 2)

	_, err = repo.Describe("cluster")
	require.NoError(t, err)

	// TODO(vuil): restore once repository is seeded with legitimate binaries
	// bin, err := repo.Fetch("cluster", VersionLatest, LinuxAMD64)
	// require.NoError(t, err)
	// require.GreaterOrEqual(t, len(bin), 10)
}
