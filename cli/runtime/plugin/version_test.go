// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_newVersionCmd(t *testing.T) {
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

	versionStr := "v1.2.3"
	cmd := newVersionCmd(versionStr)
	err = cmd.Execute()
	w.Close()
	assert.Nil(err)

	got := <-c
	assert.Equal(fmt.Sprintf("%s\n", versionStr), string(got))
}
