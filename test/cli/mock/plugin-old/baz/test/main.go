// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/plugin"
	clitest "github.com/vmware-tanzu/tanzu-framework/cli/runtime/test"
)

var descriptor = clitest.NewTestFor("baz")

func main() {
	defer Cleanup()
	p, err := plugin.NewPlugin(descriptor)
	if err != nil {
		log.Fatal(err)
	}
	p.Cmd.RunE = test
	if err := p.Execute(); err != nil {
		os.Exit(1)
	}
}

func test(c *cobra.Command, _ []string) error {
	return nil
}

// Cleanup the test.
func Cleanup() {}
