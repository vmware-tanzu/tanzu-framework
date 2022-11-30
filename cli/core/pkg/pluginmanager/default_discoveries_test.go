// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package pluginmanager

import (
	"os"
	"testing"

	configlib "github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"

	"github.com/stretchr/testify/assert"
)

func Test_defaultDiscoverySourceBasedOnServer(t *testing.T) {
	tanzuConfigBytes := `apiVersion: config.tanzu.vmware.com/v1alpha1
clientOptions:
  cli:
    useContextAwareDiscovery: true
current: mgmt
kind: ClientConfig
metadata:
  creationTimestamp: null
servers:
- managementClusterOpts:
    context: mgmt-admin@mgmt
    path: config
  name: mgmt
  type: managementcluster
`
	f, err := os.CreateTemp("", "tanzu_config")
	assert.Nil(t, err)
	err = os.WriteFile(f.Name(), []byte(tanzuConfigBytes), 0644)
	assert.Nil(t, err)
	defer func(name string) {
		err = os.Remove(name)
		assert.NoError(t, err)
	}(f.Name())
	err = os.Setenv("TANZU_CONFIG", f.Name())
	assert.NoError(t, err)
	server, err := configlib.GetServer("mgmt")
	assert.Nil(t, err)
	pds := append(server.DiscoverySources, defaultDiscoverySourceBasedOnServer(server)...)
	assert.Equal(t, 1, len(pds))
	assert.Equal(t, pds[0].Kubernetes.Name, "default-mgmt")
	assert.Equal(t, pds[0].Kubernetes.Path, "config")
	assert.Equal(t, pds[0].Kubernetes.Context, "mgmt-admin@mgmt")
}

func Test_defaultDiscoverySourceBasedOnContext(t *testing.T) {
	// Test tmc global server
	tanzuConfigBytes := `apiVersion: config.tanzu.vmware.com/v1alpha1
clientOptions:
  cli:
    useContextAwareDiscovery: true
current: tmc-test
kind: ClientConfig
metadata:
  creationTimestamp: null
contexts:
- globalOpts:
    endpoint: test.cloud.vmware.com:443
  name: tmc-test
  type: tmc
- clusterOpts:
    context: mgmt-admin@mgmt
    path: config
  name: mgmt
  type: k8s
`
	tf, err := os.CreateTemp("", "tanzu_tmc_config")
	assert.Nil(t, err)
	err = os.WriteFile(tf.Name(), []byte(tanzuConfigBytes), 0644)
	assert.Nil(t, err)
	defer func(name string) {
		err = os.Remove(name)
		assert.NoError(t, err)
	}(tf.Name())
	err = os.Setenv("TANZU_CONFIG", tf.Name())
	assert.Nil(t, err)

	context, err := configlib.GetContext("tmc-test")
	assert.Nil(t, err)
	pdsTMC := append(context.DiscoverySources, defaultDiscoverySourceBasedOnContext(context)...)
	assert.Equal(t, 1, len(pdsTMC))
	assert.Equal(t, pdsTMC[0].REST.Endpoint, "https://test.cloud.vmware.com")
	assert.Equal(t, pdsTMC[0].REST.BasePath, "v1alpha1/system/binaries/plugins")
	assert.Equal(t, pdsTMC[0].REST.Name, "default-tmc-test")

	context, err = configlib.GetContext("mgmt")
	assert.Nil(t, err)
	pdsK8s := append(context.DiscoverySources, defaultDiscoverySourceBasedOnContext(context)...)
	assert.Equal(t, 1, len(pdsK8s))
	assert.Equal(t, pdsK8s[0].Kubernetes.Name, "default-mgmt")
	assert.Equal(t, pdsK8s[0].Kubernetes.Path, "config")
	assert.Equal(t, pdsK8s[0].Kubernetes.Context, "mgmt-admin@mgmt")
}
