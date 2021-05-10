// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	cliv1alpha1 "github.com/vmware-tanzu-private/core/apis/cli/v1alpha1"
)

func TestNewPlugin(t *testing.T) {
	assert := assert.New(t)

	descriptor := cliv1alpha1.PluginDescriptor{
		Name:        "Test Plugin",
		Description: "Description of the plugin",
		Version:     "v1.2.3",
		BuildSHA:    "cafecafe",
		Group:       "TestGroup",
		DocURL:      "https://docs.example.com",
		Hidden:      false,
	}

	cmd, err := NewPlugin(&descriptor)
	if err != nil {
		t.Error(err)
	}
	assert.Equal("Test Plugin", cmd.Cmd.Use)
	assert.Equal(("Description of the plugin"), cmd.Cmd.Short)
}

func TestNewPluginFromFile(t *testing.T) {
	assert := assert.New(t)

	tmpFile, err := ioutil.TempFile("", "plugin-test-*.json")
	if err != nil {
		t.Error(err)
	}
	pluginFile := tmpFile.Name()
	defer os.Remove(pluginFile)

	descriptor := cliv1alpha1.PluginDescriptor{
		Name:        "Test Plugin",
		Description: "Description of the plugin",
		Version:     "v1.2.3",
		BuildSHA:    "cafecafe",
		Group:       "TestGroup",
		DocURL:      "https://docs.example.com",
		Hidden:      false,
	}
	pluginJSON, err := json.Marshal(descriptor)
	if err != nil {
		t.Error(err)
	}
	err = ioutil.WriteFile(pluginFile, pluginJSON, 0644)
	if err != nil {
		t.Error(err)
	}

	cmd, err := NewPluginFromFile(pluginFile)
	if err != nil {
		t.Error(err)
	}
	assert.Equal("Test Plugin", cmd.Cmd.Use)
	assert.Equal(("Description of the plugin"), cmd.Cmd.Short)
}

func TestNewPluginFromFile_Invalid(t *testing.T) {
	assert := assert.New(t)

	cmd, err := NewPluginFromFile("/tmp/does/not/exist.json")
	assert.NotNil(err)
	assert.Nil(cmd)
	assert.Contains(err.Error(), "could not read")
}

func TestAddCommands(t *testing.T) {
	assert := assert.New(t)

	descriptor := cliv1alpha1.PluginDescriptor{
		Name:        "Test Plugin",
		Description: "Description of the plugin",
		Version:     "v1.2.3",
		BuildSHA:    "cafecafe",
		Group:       "TestGroup",
		DocURL:      "https://docs.example.com",
		Hidden:      false,
	}

	cmd, err := NewPlugin(&descriptor)
	if err != nil {
		t.Error(err)
	}

	subCmd := &cobra.Command{
		Use:   "Sub1",
		Short: "Sub1 description",
	}
	cmd.AddCommands(subCmd)

	// Plugin gets four commands by default (describe, info, version, lint), ours should make five.
	assert.Equal(6, len(cmd.Cmd.Commands()))
}

func TestExecute(t *testing.T) {
	assert := assert.New(t)

	descriptor := cliv1alpha1.PluginDescriptor{
		Name:        "Test Plugin",
		Description: "Description of the plugin",
		Version:     "v1.2.3",
		BuildSHA:    "cafecafe",
		Group:       "TestGroup",
		DocURL:      "https://docs.example.com",
		Hidden:      false,
	}

	cmd, err := NewPlugin(&descriptor)
	if err != nil {
		t.Error(err)
	}

	assert.Nil(cmd.Execute())
}
