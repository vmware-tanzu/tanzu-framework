package client

import (
	"testing"

	"github.com/stretchr/testify/require"
	clientv1alpha1 "github.com/vmware-tanzu-private/core/apis/client/v1alpha1"
)

func TestConfig(t *testing.T) {
	testCtx := &clientv1alpha1.Config{
		KnownServers:  []*clientv1alpha1.Server{},
		CurrentServer: "test",
	}
	err := StoreConfig(testCtx)
	require.NoError(t, err)

	ctx, err := GetConfig()
	require.NoError(t, err)

	require.Equal(t, testCtx, ctx)
}
