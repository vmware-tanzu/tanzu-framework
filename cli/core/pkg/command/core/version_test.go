// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo"
)

func readOutput(t *testing.T, r io.Reader, c chan<- []byte) {
	data, err := io.ReadAll(r)
	if err != nil {
		t.Error(err)
	}
	c <- data
}

func TestVersion(t *testing.T) {
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

	buildinfo.Version = "1.2.3"
	buildinfo.Date = "today"
	buildinfo.SHA = "cafecafe"
	defer func() {
		buildinfo.Version = ""
		buildinfo.Date = ""
		buildinfo.SHA = ""
	}()

	err = versionCmd.Execute()
	assert.Nil(err)
	w.Close()

	got := <-c
	expected := "version: 1.2.3\nbuildDate: today\nsha: cafecafe\n"
	assert.Equal(expected, string(got))
}
