// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package command creates and initializes the tanzu CLI.
package command

import (
	"context"
	"fmt"
	"os"
	"text/template"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/cli"
	coreTemplates "github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/command/templates"
	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/pluginmanager"
	cliapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/cli/v1alpha1"
)

// DefaultDocsDir is the base docs directory
const DefaultDocsDir = "docs/cli/commands"
const ErrorDocsOutputFolderNotExists = "error reading docs output directory '%v', make sure directory exists or provide docs output directory as input value to '--docs-dir' flag"

var (
	docsDir string
)

func init() {
	genAllDocsCmd.Flags().StringVarP(&docsDir, "docs-dir", "d", DefaultDocsDir, "destination for docs output")
}

var genAllDocsCmd = &cobra.Command{
	Use:    "generate-all-docs",
	Short:  "Generate Cobra CLI docs for all plugins installed",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {

		if docsDir == "" {
			docsDir = DefaultDocsDir
		}
		if dir, err := os.Stat(docsDir); err != nil || !dir.IsDir() {
			return errors.Wrap(err, fmt.Sprintf(ErrorDocsOutputFolderNotExists, docsDir))
		}
		// Generate standard tanzu.md command file
		if err := genCoreCMD(cmd); err != nil {
			return fmt.Errorf("error generate core tanzu cmd markdown %q", err)
		}

		var pluginDescriptions []*cliapi.PluginDescriptor
		var err error

		serverPlugins, standalonePlugins, err := pluginmanager.InstalledPlugins()
		if err != nil {
			return fmt.Errorf("error while getting installed plugins descriptors: %q", err)
		}

		combinedPds := append(serverPlugins, standalonePlugins...)
		for i := range combinedPds {
			pluginDescriptions = append(pluginDescriptions, &combinedPds[i])
		}

		if err := genREADME(pluginDescriptions); err != nil {
			return fmt.Errorf("error generate core tanzu README markdown %q", err)
		}

		if err := genMarkdownTreePlugins(pluginDescriptions); err != nil {
			return fmt.Errorf("error generating plugin docs %q", err)
		}

		return nil
	},
}

func genCoreCMD(cmd *cobra.Command) error {
	tanzuMD := fmt.Sprintf("%s/%s", docsDir, "tanzu.md")
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

func genREADME(plugins []*cliapi.PluginDescriptor) error {
	readmeFilename := fmt.Sprintf("%s/%s", docsDir, "README.md")
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

func genMarkdownTreePlugins(plugins []*cliapi.PluginDescriptor) error {
	args := []string{"generate-docs", "--docs-dir", docsDir}
	for _, p := range plugins {
		runner := cli.NewRunner(p.Name, p.InstallationPath, args)
		ctx := context.Background()
		if err := runner.Run(ctx); err != nil {
			return err
		}
	}
	return nil
}
