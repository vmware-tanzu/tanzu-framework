// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
)

func readOutput(t *testing.T, r io.Reader, c chan<- []byte) {
	data, err := ioutil.ReadAll(r)
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

	cli.BuildVersion = "1.2.3"
	cli.BuildDate = "today"
	cli.BuildSHA = "cafecafe"
	defer func() {
		cli.BuildVersion = ""
		cli.BuildDate = ""
		cli.BuildSHA = ""
	}()

	err = versionCmd.Execute()
	assert.Nil(err)
	w.Close()

	got := <-c
	expected := "version: 1.2.3\nbuildDate: today\nsha: cafecafe\n"
	assert.Equal(expected, string(got))
}
