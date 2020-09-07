package csp

import (
	"testing"

	"github.com/stretchr/testify/require"
	authv1alpha1 "github.com/vmware-tanzu-private/core/apis/auth/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestConfig(t *testing.T) {
	testCfg := &authv1alpha1.CSPConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test1",
		},
		Spec:   authv1alpha1.CSPConfigSpec{},
		Status: authv1alpha1.CSPConfigStatus{},
	}
	err := StoreConfig(testCfg)
	require.NoError(t, err)

	cfg, err := GetConfig("test1")
	require.NoError(t, err)

	require.Equal(t, testCfg, cfg)

	err = DeleteConfig("test1")
	require.NoError(t, err)
}
