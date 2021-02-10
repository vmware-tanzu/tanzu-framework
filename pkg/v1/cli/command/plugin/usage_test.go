// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestUsageFunc(t *testing.T) {
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

	cmd := &cobra.Command{
		Use:   "Sub1",
		Short: "Sub1 description",
	}
	err = UsageFunc(cmd)
	w.Close()
	assert.Nil(err)

	got := <-c
	assert.Contains(string(got), "Usage:")
}
