// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
	clientv1alpha1 "github.com/vmware-tanzu-private/core/apis/client/v1alpha1"
	"github.com/vmware-tanzu-private/core/pkg/v1/client"
)

func TestGCPRepository(t *testing.T) {
	r := clientv1alpha1.PluginRepository{GCPPluginRepository: &client.CoreGCPBucketRepository}
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

func TestPlugin(t *testing.T) {
	for _, test := range []struct {
		name     string
		versions []string
		max      string
	}{
		{
			name:     "basic patch",
			versions: []string{"v0.0.1", "v0.0.2"},
			max:      "v0.0.2",
		},
		{
			name:     "release candidates",
			versions: []string{"v0.0.1", "v1.3.0-rc.1", "v1.3.0-pre-alpha-1"},
			max:      "v0.0.1",
		},
		{
			name:     "release candidates same",
			versions: []string{"v0.0.1", "v1.3.0", "v1.3.0-rc.1", "v1.3.0-pre-alpha-1"},
			max:      "v1.3.0",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			p := Plugin{
				Name:     "foo",
				Versions: test.versions,
			}
			v := p.VersionLatest()
			require.Equal(t, test.max, v)
		})
	}
}
