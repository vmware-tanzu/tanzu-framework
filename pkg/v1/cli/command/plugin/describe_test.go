// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func readOutput(t *testing.T, r io.Reader, c chan<- []byte) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		t.Error(err)
	}
	c <- data
}

func TestDescribe(t *testing.T) {
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

	expected := "Example plugin description"
	describeCmd := newDescribeCmd(expected)
	err = describeCmd.Execute()
	assert.Nil(err)
	w.Close()

	got := <-c
	assert.Equal(fmt.Sprintf("%s\n", expected), string(got))
}
