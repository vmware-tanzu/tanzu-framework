// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var testData1 = `---
apiVersion: cli.tanzu.vmware.com/v1alpha1
kind: CLIPlugin
metadata:
  name: foo
spec:
  artifacts:
    v0.0.1:
    - arch: amd64
      image: tanzu-cli-plugins/foo-darwin-amd64:latest
      os: darwin
      type: oci
    - arch: amd64
      image: tanzu-cli-plugins/foo-linux-amd64:latest
      os: linux
      type: oci
    - arch: amd64
      image: tanzu-cli-plugins/foo-windows-amd64:latest
      os: windows
      type: foo
  description: Foo description
  optional: false
  recommendedVersion: v0.0.1
`

var testData2 = `---
apiVersion: cli.tanzu.vmware.com/v1alpha1
kind: CLIPlugin
metadata:
  name: foo
spec:
  artifacts:
    v0.0.1:
    - arch: amd64
      image: tanzu-cli-plugins/foo-darwin-amd64:latest
      os: darwin
      type: oci
    - arch: amd64
      image: tanzu-cli-plugins/foo-linux-amd64:latest
      os: linux
      type: oci
    - arch: amd64
      image: tanzu-cli-plugins/foo-windows-amd64:latest
      os: windows
      type: oci
  description: Foo description
  optional: false
  recommendedVersion: v0.0.1
---
apiVersion: cli.tanzu.vmware.com/v1alpha1
kind: CLIPlugin
metadata:
  name: bar
spec:
  artifacts:
    v0.0.1:
    - arch: amd64
      image: tanzu-cli-plugins/foo-darwin-amd64:latest
      os: darwin
      type: oci
    - arch: amd64
      image: tanzu-cli-plugins/foo-linux-amd64:latest
      os: linux
      type: oci
    - arch: amd64
      image: tanzu-cli-plugins/foo-windows-amd64:latest
      os: windows
      type: oci
    v0.0.2:
    - arch: amd64
      image: tanzu-cli-plugins/foo-darwin-amd64:latest
      os: darwin
      type: oci
    - arch: amd64
      image: tanzu-cli-plugins/foo-linux-amd64:latest
      os: linux
      type: oci
    - arch: amd64
      image: tanzu-cli-plugins/foo-windows-amd64:latest
      os: windows
      type: oci
  description: Bar description
  optional: false
  recommendedVersion: v0.0.2
`

func Test_ProcessOCIPluginManifest(t *testing.T) {
	assert := assert.New(t)

	plugins, err := processDiscoveryManifestData([]byte(testData1), "test-discovery")
	assert.Nil(err)
	assert.NotNil(plugins)
	assert.Equal(1, len(plugins))
	assert.Equal("foo", plugins[0].Name)
	assert.Equal("v0.0.1", plugins[0].RecommendedVersion)
	assert.Equal("Foo description", plugins[0].Description)
	assert.Equal("test-discovery", plugins[0].Source)
	assert.EqualValues([]string{"v0.0.1"}, plugins[0].SupportedVersions)

	plugins, err = processDiscoveryManifestData([]byte(testData2), "test-discovery")
	assert.Nil(err)
	assert.NotNil(plugins)
	assert.Equal(2, len(plugins))

	assert.Equal("foo", plugins[0].Name)
	assert.Equal("v0.0.1", plugins[0].RecommendedVersion)
	assert.Equal("Foo description", plugins[0].Description)
	assert.Equal("test-discovery", plugins[0].Source)
	assert.Equal(1, len(plugins[0].SupportedVersions))
	assert.EqualValues([]string{"v0.0.1"}, plugins[0].SupportedVersions)

	assert.Equal("bar", plugins[1].Name)
	assert.Equal("v0.0.2", plugins[1].RecommendedVersion)
	assert.Equal("Bar description", plugins[1].Description)
	assert.Equal("test-discovery", plugins[1].Source)
	assert.Equal(2, len(plugins[1].SupportedVersions))
	assert.Contains(plugins[1].SupportedVersions, "v0.0.1")
	assert.Contains(plugins[1].SupportedVersions, "v0.0.2")
}
