package controllers

import (
	"testing"

	"github.com/vmware-tanzu/tanzu-framework/apis/addonconfigs/cni/v1alpha2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCopyStruct(t *testing.T) {
	ccpConf := v1alpha2.CcpAdapterConf{
		EnableDebugServer: true,
		APIServerPort:     1234,
	}
	descCcpAdapterConf := ccpAdapterConf{
		EnableDebugServer: false,
		APIServerPort:     0,
	}
	err := copyStructAtoB(ccpConf, &descCcpAdapterConf)
	require.NoError(t, err, "copy CcpAdapterConf values error")
	assert.Equal(t, 1234, descCcpAdapterConf.APIServerPort)
	assert.Equal(t, true, descCcpAdapterConf.EnableDebugServer)

	mpConf := v1alpha2.MpAdapterConf{
		NSXClientAuthCertFile: "fake-cert-file",
		ConditionTimeout:      150,
	}
	descMpAdapterConf := mpAdapterConf{
		NSXClientAuthCertFile: "",
		ConditionTimeout:      0,
	}
	err = copyStructAtoB(mpConf, &descMpAdapterConf)
	require.NoError(t, err, "copy MpAdapterConf values error")
	assert.Equal(t, "fake-cert-file", descMpAdapterConf.NSXClientAuthCertFile)
	assert.Equal(t, 150, descMpAdapterConf.ConditionTimeout)
}
