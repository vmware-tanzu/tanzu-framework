// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/plugin"
	clitest "github.com/vmware-tanzu/tanzu-framework/cli/runtime/test"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
)

var descriptor = clitest.NewTestFor("secret")

func main() {
	p, err := plugin.NewPlugin(descriptor)
	if err != nil {
		log.Fatal(err, "failed to create a new instance of the plugin")
	}
	p.Cmd.RunE = test
	if err := p.Execute(); err != nil {
		os.Exit(1)
	}
}

func test(c *cobra.Command, _ []string) error {
	return nil
}
