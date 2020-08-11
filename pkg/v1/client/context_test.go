package client

import (
	"fmt"
	"testing"

	clientv1alpha1 "github.com/vmware-tanzu-private/core/apis/client.tanzu.cloud.vmware.com/v1alpha1"
)

func TestContext(t *testing.T) {
	testCtx := clientv1alpha1.Context{
		Spec:   clientv1alpha1.ContextSpec{},
		Status: clientv1alpha1.ContextStatus{},
	}
	fmt.Println(testCtx)
}
