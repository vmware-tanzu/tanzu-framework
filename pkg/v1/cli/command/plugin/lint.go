// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/command/plugin/lint"
)

// LintCmd inspects the Cobra command tree to ensure all commands meet the linting standards
// todo: checks for duplicate commands and duplicate aliases
var lintCmd = &cobra.Command{
	Use:          "lint",
	Short:        "Lint on cobra command structure",
	Long:         "Lint this command's full flag and cmd tree. The cmd or flag will be skipped when annotated with 'no-lint'",
	Hidden:       true,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		linter, err := lint.NewCobraLinter(cmd)
		if err != nil {
			return err
		}

		if !linter.Run() {
			linter.Output()
			return errors.New("cobra command linting failed")
		}

		return nil
	},
}
