// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/plugin"
)

func Test_getInstalledElseAvailablePluginVersion(t *testing.T) {
	assert := assert.New(t)

	p := plugin.Discovered{
		InstalledVersion:   "v1.0.0",
		RecommendedVersion: "v2.0.0",
	}

	version := getInstalledElseAvailablePluginVersion(p)
	assert.Equal(version, p.InstalledVersion)

	p.InstalledVersion = ""
	version = getInstalledElseAvailablePluginVersion(p)
	assert.Equal(version, p.RecommendedVersion)
}
