// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package plugin provides functions to create new CLI plugins.
package plugin

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"

	core "github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/command/core"
)

var (
	docsDir string
)

func init() {
	genDocsCmd.Flags().StringVarP(&docsDir, "docs-dir", "d", core.DefaultDocsDir, "destination for docss output")
}

var genDocsCmd = &cobra.Command{
	Use:    "generate-docs",
	Short:  "Generate Cobra CLI docs for all subcommands",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		identity := func(s string) string {
			if !strings.HasPrefix(s, "tanzu") {
				return fmt.Sprintf("tanzu_%s", s)
			}
			return s
		}
		emptyStr := func(s string) string { return "" }

		// Necessary to generate correct output
		core.RootCmd.AddCommand(cmd.Parent())
		if err := doc.GenMarkdownTreeCustom(cmd.Parent(), docsDir, emptyStr, identity); err != nil {
			return fmt.Errorf("error generating docs %q", err)
		}

		return nil
	},
}
