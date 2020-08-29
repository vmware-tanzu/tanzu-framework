package client

import (
	"testing"

	"github.com/stretchr/testify/require"
	clientv1alpha1 "github.com/vmware-tanzu-private/core/apis/client/v1alpha1"
)

func TestContext(t *testing.T) {
	testCtx := &clientv1alpha1.Context{
		Spec:   clientv1alpha1.ContextSpec{},
		Status: clientv1alpha1.ContextStatus{},
	}
	err := StoreContext(testCtx)
	require.NoError(t, err)

	ctx, err := GetContext()
	require.NoError(t, err)

	require.Equal(t, testCtx, ctx)
}
