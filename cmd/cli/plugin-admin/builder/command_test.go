// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"os"
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
	args := []string{"testrepo", "--dry-run", "--repo-type", "github"}
	cmd := NewInitCmd()
	cmd.SetArgs(args)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	err = cmd.Execute()
	assert.Nil(err)

	// Assert repo creation
	args = []string{"testrepo", "--repo-type", "github"}
	cmd = NewInitCmd()
	cmd.SetArgs(args)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	err = cmd.Execute()
	assert.Nil(err)

	err = os.Chdir(filepath.Join(dir, "testrepo"))
	assert.Nil(err)

	// Assert plugin creation
	stdout = bytes.Buffer{}
	stderr = bytes.Buffer{}
	cmd = newAddPluginCmd()
	args = []string{"testplugin", "--description", "something"}
	cmd.SetArgs(args)
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	err = cmd.Execute()
	assert.Nil(err)
}
