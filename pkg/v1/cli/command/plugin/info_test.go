// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	cliv1alpha1 "github.com/vmware-tanzu-private/core/apis/cli/v1alpha1"
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

	descriptor := cliv1alpha1.PluginDescriptor{
		Name:        "Test Plugin",
		Description: "Description of the plugin",
		Version:     "1.2.3",
		BuildSHA:    "cafecafe",
		Group:       "TestGroup",
		DocURL:      "https://docs.example.com",
		Hidden:      false,
	}
	expected, err := json.Marshal(descriptor)
	if err != nil {
		t.Error(err)
	}

	infoCmd := newInfoCmd(&descriptor)
	err = infoCmd.Execute()
	w.Close()
	assert.Nil(err)

	got := <-c
	assert.Equal(fmt.Sprintf("%s\n", expected), string(got))
}
