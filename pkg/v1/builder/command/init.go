// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/builder/template"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/component"
)

const (
	github = "github"
)

var (
	repoType string
)

func init() {
	NewInitCmd()
}

// NewInitCmd initializes a repository.
func NewInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [name]",
		Short: "Initialize a repository",
		RunE:  initialize,
		Args:  cobra.ExactArgs(1),
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print generated files to stdout")
	cmd.Flags().StringVar(&repoType, "repo-type", "", "Type of repository: github or gitlab")

	return cmd
}

func initialize(cmd *cobra.Command, args []string) error {
	name := args[0]
	var err error

	if repoType == "" {
		repoType, err = selectCIProvider()
		if err != nil {
			return err
		}
	}
	data := struct {
		RepositoryName string
	}{
		RepositoryName: name,
	}
	targets := template.DefaultInitTargets
	if strings.EqualFold(repoType, github) {
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

	if dryRun {
		return nil
	}

	c := exec.Command("git", "init", name)
	b, err := c.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s -- %s", err, string(b))
	}
	cmd.Print("successfully created repository")
	return nil
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
