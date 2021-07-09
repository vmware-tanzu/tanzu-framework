// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPluginNameFromBin(t *testing.T) {
	for _, test := range []struct {
		binName    string
		pluginName string
	}{
		{
			binName:    fmt.Sprintf("%stest", BinNamePrefix),
			pluginName: "test",
		},
		{
			binName:    fmt.Sprintf("%stest-charlie_1", BinNamePrefix),
			pluginName: "test-charlie_1",
		},
	} {
		t.Run(fmt.Sprintf("%s name", test.binName), func(t *testing.T) {
			p := PluginNameFromBin(test.binName)
			require.Equal(t, test.pluginName, p)
		})
	}
}

func TestBinFromPluginName(t *testing.T) {
	for _, test := range []struct {
		binName    string
		pluginName string
	}{
		{
			binName:    fmt.Sprintf("%stest", BinNamePrefix),
			pluginName: "test",
		},
		{
			binName:    fmt.Sprintf("%stest-charlie_1", BinNamePrefix),
			pluginName: "test-charlie_1",
		},
	} {
		t.Run(fmt.Sprintf("%s name", test.pluginName), func(t *testing.T) {
			b := BinFromPluginName(test.pluginName)
			require.Equal(t, test.binName, b)
		})
	}
}

func TestMakeArtifactName(t *testing.T) {
	for _, test := range []struct {
		arch         Arch
		pluginName   string
		artifactName string
	}{
		{
			arch:         LinuxAMD64,
			pluginName:   "test",
			artifactName: fmt.Sprintf("%s-test-%s", ArtifactNamePrefix, LinuxAMD64),
		},
		{
			arch:         WinAMD64,
			pluginName:   "test-charlie_1",
			artifactName: fmt.Sprintf("%s-test-charlie_1-%s.exe", ArtifactNamePrefix, WinAMD64),
		},
	} {
		t.Run(fmt.Sprintf("%s test", test.arch), func(t *testing.T) {
			b := MakeArtifactName(test.pluginName, test.arch)
			require.Equal(t, test.artifactName, b)
		})
	}
}
