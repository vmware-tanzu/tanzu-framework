package client

import (
	"testing"

	"github.com/stretchr/testify/require"
	clientv1alpha1 "github.com/vmware-tanzu-private/core/apis/client/v1alpha1"
)

func TestConfig(t *testing.T) {
	LocalDirName = ".tanzu-test"
	server0 := &clientv1alpha1.Server{
		Name: "test",
		Type: clientv1alpha1.ManagementClusterServerType,
		ManagementClusterOpts: &clientv1alpha1.ManagementClusterServer{
			Path: "test",
		},
	}
	testCtx := &clientv1alpha1.Config{
		KnownServers: []*clientv1alpha1.Server{
			server0,
		},
		CurrentServer: "test",
	}

	err := StoreConfig(testCtx)
	require.NoError(t, err)

	_, err = GetConfig()
	require.NoError(t, err)

	s, err := GetServer("test")
	require.NoError(t, err)

	require.Equal(t, s, server0)

	e, err := ServerExists("test")
	require.NoError(t, err)
	require.True(t, e)

	server1 := &clientv1alpha1.Server{
		Name: "test1",
		Type: clientv1alpha1.ManagementClusterServerType,
		ManagementClusterOpts: &clientv1alpha1.ManagementClusterServer{
			Path: "test1",
		},
	}

	err = AddServer(server1, true)
	require.NoError(t, err)

	c, err := GetConfig()
	require.NoError(t, err)
	require.Len(t, c.KnownServers, 2)
	require.Equal(t, c.CurrentServer, "test1")

	err = RemoveServer("test")
	require.NoError(t, err)

	c, err = GetConfig()
	require.NoError(t, err)
	require.Len(t, c.KnownServers, 1)
}
