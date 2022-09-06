// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/cli"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/component"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/builder/command"
)

var (
	dryRun      bool
	description string
)

var compileArgs = &command.PluginCompileArgs{
	Match:        "*",
	TargetArch:   []string{"all"},
	SourcePath:   "./cmd/cli/plugin",
	ArtifactsDir: cli.DefaultArtifactsDirectory,
}

// NewCLICmd creates the CLI builder commands.
func NewCLICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cli",
		Short: "Build CLIs",
	}

	cmd.AddCommand(newCompileCmd())
	cmd.AddCommand(newAddPluginCmd())
	return cmd
}

// newCompileCmd compiles CLI plugins.
func newCompileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compile",
		Short: "Compile a repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			return command.Compile(compileArgs)
		},
	}

	cmd.Flags().StringVar(&compileArgs.Version, "version", "", "Version of the root cli (required)")
	cmd.Flags().StringVar(&compileArgs.LDFlags, "ldflags", "", "ldflags to set on build")
	cmd.Flags().StringVar(&compileArgs.Tags, "tags", "", "Tags to set on build")
	cmd.Flags().StringVar(&compileArgs.Match, "match", compileArgs.Match, "Match a plugin name to build, supports globbing")
	cmd.Flags().StringArrayVar(&compileArgs.TargetArch, "target", compileArgs.TargetArch, "Only compile for specific target(s), use 'local' to compile for host os")
	cmd.Flags().StringVar(&compileArgs.SourcePath, "path", compileArgs.SourcePath, "Path of the plugins source directory")
	cmd.Flags().StringVar(&compileArgs.ArtifactsDir, "artifacts", compileArgs.ArtifactsDir, "Path to output artifacts")
	cmd.Flags().StringVar(&compileArgs.CorePath, "corepath", "", "Path for core binary")
	cmd.Flags().StringVar(&compileArgs.GoPrivate, "goprivate", "", "Comma-separated list of glob patterns of module path prefixes to set as GOPRIVATE on build")

	_ = cmd.MarkFlagRequired("version")

	return cmd
}

// newAddPluginCmd adds a cli plugin to the repository.
func newAddPluginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-plugin NAME",
		Short: "Add a plugin to a repository",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error

			name := args[0]
			if description == "" {
				description, err = askDescription()
				if err != nil {
					return err
				}
			}

			return command.AddPlugin(name, description, dryRun)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print generated files to stdout")
	cmd.Flags().StringVar(&description, "description", "", "Required plugin description")

	return cmd
}

func askDescription() (answer string, err error) {
	questioncfg := &component.QuestionConfig{
		Message: "provide a description",
	}
	err = component.Ask(questioncfg, &answer)
	if err != nil {
		return
	}
	return
}
