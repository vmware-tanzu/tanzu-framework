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

var descriptor = clitest.NewTestFor("builder")

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
	m := clitest.NewMain("cluster", c, Cleanup)
	defer m.Finish()

	// TODO (pbarker): use virtual filesystem
	err := m.RunTest(
		"create a repo",
		"builder init myrepo --dry-run",
		func(t *clitest.Test) error {
			err := t.ExecContainsString("myrepo")
			if err != nil {
				return err
			}
			return nil
		},
	)
	if err != nil {
		return err
	}
	return nil
}

// Cleanup the test.
func Cleanup() error {
	return nil
}
