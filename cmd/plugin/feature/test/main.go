// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-plugin-runtime/plugin"
	clitest "github.com/vmware-tanzu/tanzu-plugin-runtime/test/framework"
)

var descriptor = clitest.NewTestFor("feature")

func main() {
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
