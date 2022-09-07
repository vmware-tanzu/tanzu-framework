// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/component"
	"github.com/vmware-tanzu/tanzu-framework/plugin-admin/builder/pkg/command"
)

const desc = `Initialize a new plugin repository with scaffolding for:

* Tanzu Framework CLI integration
* GolangCI linting config
* GitHub or GitLab CI config
* A Makefile`

var (
	repoType string
)

// NewInitCmd initializes a repository.
func NewInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init PLUGIN_NAME",
		Short: "Initialize a new plugin repository",
		Long:  desc,
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			if repoType == "" {
				repoType, err = selectCIProvider()
				if err != nil {
					return err
				}
			}
			return command.Initialize(args[0], repoType, dryRun)
		},
		Args: cobra.ExactArgs(1),
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print generated files to stdout")
	cmd.Flags().StringVar(&repoType, "repo-type", "", "Type of repository: github or gitlab")

	return cmd
}

func selectCIProvider() (selection string, err error) {
	cfg := &component.SelectConfig{
		Message: "choose a repository type",
		Options: []string{
			"GitHub",
			"GitLab",
		},
	}
	err = component.Select(cfg, &selection)
	if err != nil {
		return
	}
	return
}
