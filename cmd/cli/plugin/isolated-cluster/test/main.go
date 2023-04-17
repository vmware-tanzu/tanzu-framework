// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// nolint: deadcode,unused,stylecheck
package test

import (
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-plugin-runtime/plugin"
	clitest "github.com/vmware-tanzu/tanzu-plugin-runtime/test/framework"
)

func main() {
	pluginName := "isolated-cluster"
	descriptor := clitest.NewTestFor(pluginName)

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
	if err := Cleanup(); err != nil {
		return err
	}
	return nil
}

// Cleanup the test.
func Cleanup() error {
	return nil
}
