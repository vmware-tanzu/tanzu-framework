// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package publish

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/aunum/log"
	"github.com/spf13/afero"
	"github.com/tj/assert"
)

func Test_DetectAvailablePluginInfo(t *testing.T) {
	assert := assert.New(t)
	artifactDir := "artifacts"
	// Set `fs` as in memory file system for testing
	fs = afero.NewMemMapFs()

	// Setup artifacts directory for testing
	createDummyArtifactDir(filepath.Join(artifactDir, "windows", "amd64", "cli"), "fake-plugin-foo", "v1.1.0", "fake description", []string{"v1.0.0"})
	createDummyArtifactDir(filepath.Join(artifactDir, "windows", "arm32", "cli"), "fake-plugin-foo", "v1.1.0", "fake description", []string{"v1.0.0"})
	createDummyArtifactDir(filepath.Join(artifactDir, "darwin", "amd64", "cli"), "fake-plugin-foo", "v1.1.0", "fake description", []string{"v1.0.0", "v1.1.0"})
	createDummyArtifactDir(filepath.Join(artifactDir, "darwin", "arm32", "cli"), "fake-plugin-foo", "v1.1.0", "fake description", []string{"v1.0.0", "v1.1.0"})

	plugins := []string{"fake-plugin-foo"}
	osArch := []string{"darwin-amd64", "darwin-arm32", "windows-amd64", "windows-arm32"}
	pluginInfo, err := detectAvailablePluginInfo(artifactDir, plugins, osArch, "v1.1.0")
	assert.Nil(err)
	assert.NotNil(pluginInfo["fake-plugin-foo"])

	// should include 2 versions. "v1.0.0" and "v1.1.0" and no other versions
	assert.Equal(2, len(pluginInfo["fake-plugin-foo"].versions))
	assert.NotNil(pluginInfo["fake-plugin-foo"].versions["v1.0.0"])
	assert.NotNil(pluginInfo["fake-plugin-foo"].versions["v1.1.0"])
	assert.Nil(pluginInfo["fake-plugin-foo"].versions["v1.2.0"])

	// should include 4 os-arch for version v1.0.0. "darwin-amd64", "darwin-arm32", "windows-amd64", "windows-arm32"
	assert.Equal(4, len(pluginInfo["fake-plugin-foo"].versions["v1.0.0"]))
	arrOsArch := convertVersionToStringArray(pluginInfo["fake-plugin-foo"].versions["v1.0.0"])
	assert.Contains(arrOsArch, "darwin-amd64")
	assert.Contains(arrOsArch, "darwin-arm32")
	assert.Contains(arrOsArch, "windows-amd64")
	assert.Contains(arrOsArch, "windows-arm32")

	// should include 2 os-arch for version v1.1.0. "darwin-amd64", "darwin-amd64"
	assert.Equal(2, len(pluginInfo["fake-plugin-foo"].versions["v1.1.0"]))
	arrOsArch = convertVersionToStringArray(pluginInfo["fake-plugin-foo"].versions["v1.1.0"])
	assert.Contains(arrOsArch, "darwin-amd64")
	assert.Contains(arrOsArch, "darwin-arm32")
	assert.NotContains(arrOsArch, "windows-amd64")
	assert.NotContains(arrOsArch, "windows-arm32")

	assert.Equal("v1.1.0", pluginInfo["fake-plugin-foo"].recommendedVersion)
	assert.Equal("fake description", pluginInfo["fake-plugin-foo"].description)
}

func convertVersionToStringArray(arrOsArchInfo []osArch) []string {
	oa := []string{}
	for idx := range arrOsArchInfo {
		oa = append(oa, arrOsArchInfo[idx].os+"-"+arrOsArchInfo[idx].arch)
	}
	return oa
}

func createDummyArtifactDir(directoryBasePath, pluginName, recommendedVersion, description string, versions []string) { // `pluginName` always receives `"fake-plugin-foo"
	var err error

	for _, v := range versions {
		err = fs.MkdirAll(filepath.Join(directoryBasePath, pluginName, v), 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	data := `name: %s
description: %s
version: %s`

	err = afero.WriteFile(fs, filepath.Join(directoryBasePath, pluginName, "plugin.yaml"), []byte(fmt.Sprintf(data, pluginName, description, recommendedVersion)), 0644)
	if err != nil {
		log.Fatal(err)
	}
}
