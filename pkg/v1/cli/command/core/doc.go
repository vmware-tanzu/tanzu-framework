// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package core creates and initializes the tanzu CLI.
package core

import (
	"context"
	"fmt"
	"os"
	"text/template"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	cli "github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli"
	coreTemplates "github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/command/core/templates"
)

// DefaultDocsDir is the base docs directory
const DefaultDocsDir = "docs/cli/commands"

var (
	docsDir string
)

func init() {
	genAllDocsCmd.Flags().StringVarP(&docsDir, "docs-dir", "d", DefaultDocsDir, "destination for docss output")
}

var genAllDocsCmd = &cobra.Command{
	Use:    "generate-all-docs",
	Short:  "Generate Cobra CLI docs for all plugins installed",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Generate standard tanzu.md command file
		if err := genCoreCMD(cmd); err != nil {
			return fmt.Errorf("error generate core tanzu cmd markdown %q", err)
		}

		plugins, err := cli.ListPlugins()
		if err != nil {
			return err
		}

		// Generate README TOC
		if err := genREADME(plugins); err != nil {
			return fmt.Errorf("error generate core tanzu README markdown %q", err)
		}

		if err := genMarkdownTreePlugins(plugins); err != nil {
			return fmt.Errorf("error generating plugin docs %q", err)
		}

		return nil
	},
}

func genCoreCMD(cmd *cobra.Command) error {
	tanzuMD := fmt.Sprintf("%s/%s", DefaultDocsDir, "tanzu.md")
	t, err := os.OpenFile(tanzuMD, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error opening tanzu.md %q", err)
	}
	defer t.Close()
	if err := doc.GenMarkdown(cmd.Root(), t); err != nil {
		return fmt.Errorf("error generating markdown %q", err)
	}
	return nil
}

func genREADME(plugins []*cliv1alpha1.PluginDescriptor) error {
	readmeFilename := fmt.Sprintf("%s/%s", DefaultDocsDir, "README.md")
	readme, err := os.OpenFile(readmeFilename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error opening readme %q", err)
	}
	defer readme.Close()

	tmpl := template.Must(template.New("readme").Parse(coreTemplates.CoreREADME))
	err = tmpl.Execute(readme, plugins)
	if err != nil {
		return err
	}
	return nil
}

func genMarkdownTreePlugins(plugins []*cliv1alpha1.PluginDescriptor) error {
	args := []string{"generate-docs"}
	for _, p := range plugins {
		runner := cli.NewRunner(p.Name, p.InstallationPath, args)
		ctx := context.Background()
		if err := runner.Run(ctx); err != nil {
			return err
		}
	}
	return nil
}
