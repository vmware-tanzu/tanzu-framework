// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"fmt"
	"os/exec"

	"github.com/aunum/log"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/builder/template"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/component"
)

func init() {
	InitCmd.Flags().BoolVar(&dryRun, "dry-run", false, "print generated files to stdout")
}

// InitCmd initializes a repository.
var InitCmd = &cobra.Command{
	Use:   "init [name]",
	Short: "Initialize a repository",
	RunE:  initialize,
	Args:  cobra.ExactArgs(1),
}

func initialize(cmd *cobra.Command, args []string) error {
	name := args[0]
	cfg := &component.SelectConfig{
		Message: "choose a repository type",
		Options: []string{
			"GitHub",
			"GitLab",
		},
	}
	var selection string
	err := component.Select(cfg, &selection)
	if err != nil {
		return err
	}

	data := struct {
		RepositoryName string
	}{
		RepositoryName: name,
	}
	targets := template.DefaultInitTargets
	if selection == "GitHub" {
		targets = append(targets, template.GitHubCI)
	} else {
		targets = append(targets, template.GitLabCI)
	}
	for _, target := range targets {
		err = target.Run(name, data, dryRun)
		if err != nil {
			return err
		}
	}

	c := exec.Command("git", "init", name)
	b, err := c.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s -- %s", err, string(b))
	}
	log.Success("successfully created repository")
	return nil
}
