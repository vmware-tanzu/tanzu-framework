// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	cliapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/cli/v1alpha1"
)

func TestInfo(t *testing.T) {
	assert := assert.New(t)

	r, w, err := os.Pipe()
	if err != nil {
		t.Error(err)
	}
	c := make(chan []byte)
	go readOutput(t, r, c)

	// Set up for our test
	stdout := os.Stdout
	stderr := os.Stderr
	defer func() {
		os.Stdout = stdout
		os.Stderr = stderr
	}()
	os.Stdout = w
	os.Stderr = w

	descriptor := cliapi.PluginDescriptor{
		Name:        "Test Plugin",
		Description: "Description of the plugin",
		Version:     "1.2.3",
		BuildSHA:    "cafecafe",
		Group:       "TestGroup",
		DocURL:      "https://docs.example.com",
		Hidden:      false,
	}

	infoCmd := newInfoCmd(&descriptor)
	err = infoCmd.Execute()
	w.Close()
	assert.Nil(err)

	got := <-c

	expectedInfo := pluginInfo{
		PluginDescriptor: descriptor,
	}

	gotInfo := &pluginInfo{}
	err = json.Unmarshal(got, gotInfo)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(expectedInfo.Name, gotInfo.Name)
	assert.Equal(expectedInfo.Description, gotInfo.Description)
	assert.Equal(expectedInfo.Version, gotInfo.Version)
	assert.Equal(expectedInfo.BuildSHA, gotInfo.BuildSHA)
	assert.Equal(expectedInfo.DocURL, gotInfo.DocURL)
	assert.Equal(expectedInfo.Hidden, gotInfo.Hidden)
	assert.NotEmpty(gotInfo.PluginRuntimeVersion)
}
