// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"testing"

	"github.com/stretchr/testify/require"

	configv1alpha1 "github.com/vmware-tanzu-private/core/apis/config/v1alpha1"
	"github.com/vmware-tanzu-private/core/pkg/v1/config"
)

func TestGCPRepository(t *testing.T) {
	r := configv1alpha1.PluginRepository{GCPPluginRepository: &config.CoreGCPBucketRepository}
	repo := loadRepository(r)
	list, err := repo.List()
	require.NoError(t, err)

	require.GreaterOrEqual(t, len(list), 2)

	_, err = repo.Describe("cluster")
	require.NoError(t, err)

	bin, err := repo.Fetch("cluster", VersionLatest, LinuxAMD64)
	require.NoError(t, err)

	require.GreaterOrEqual(t, len(bin), 10)
}
