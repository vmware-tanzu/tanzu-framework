// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_BuilderInitAndAddPlugin(t *testing.T) {
	assert := assert.New(t)

	dir, err := os.MkdirTemp("", "core")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(dir)
	err = os.Chdir(dir)
	assert.Nil(err)

	var stdout, stderr bytes.Buffer

	// Assert dry-run does not create a repo
	expected := "module testrepo"
	args := []string{"testrepo", "--dry-run", "--repo-type", "github"}
	cmd := NewInitCmd()
	cmd.SetArgs(args)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	err = cmd.Execute()
	assert.Nil(err)
	assert.Contains(expected, stdout.String())

	// Assert repo creation
	expected = "successfully created repository"
	args = []string{"testrepo", "--repo-type", "github"}
	cmd = NewInitCmd()
	cmd.SetArgs(args)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	err = cmd.Execute()
	assert.Nil(err)
	assert.Equal(expected, stdout.String())

	err = os.Chdir(filepath.Join(dir, "testrepo"))
	assert.Nil(err)

	// Assert plugin creation
	stdout = bytes.Buffer{}
	stderr = bytes.Buffer{}
	expected = "successfully created plugin"
	cmd = NewAddPluginCmd()
	args = []string{"testplugin", "--description", "something"}
	cmd.SetArgs(args)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	err = cmd.Execute()
	assert.Nil(err)
	assert.Equal(expected, stdout.String())

	// Assert make init and test exit 0
	osCmd := exec.Command("make", "init")
	_, err = osCmd.CombinedOutput()
	assert.Nil(err)
	osCmd = exec.Command("make", "test")
	_, err = osCmd.CombinedOutput()
	assert.Nil(err)
}
