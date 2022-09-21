// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package command provides functions to manipulate tanzu cli commands
package command

import (
	"fmt"

	"github.com/spf13/cobra"
)

// DeprecateCommand marks the command as deprecated and adds deprecation message.
func DeprecateCommand(cmd *cobra.Command, removalVersion string) {
	msg := fmt.Sprintf("will be removed in version %q.", removalVersion)
	cmd.Deprecated = msg
}

// DeprecateCommandWithAlternative marks the commands as deprecated and adds deprecation message with an alternative.
func DeprecateCommandWithAlternative(cmd *cobra.Command, removalVersion, alternative string) {
	msg := fmt.Sprintf("will be removed in version %q. Use %q instead.", removalVersion, alternative)
	cmd.Deprecated = msg
}

// DeprecateFlag marks the flag as deprecated and hidden with a deprecation message.
func DeprecateFlag(cmd *cobra.Command, flag, removalVersion string) {
	f := cmd.Flags().Lookup(flag)
	if f != nil {
		msg := fmt.Sprintf("will be removed in version %q.", removalVersion)
		f.Deprecated = msg
		f.Hidden = true
	}
}

// DeprecateFlagWithAlternative marks the flag as deprecated and hidden with deprecation message recommending an
// alternative flag to use.
func DeprecateFlagWithAlternative(cmd *cobra.Command, flag, removalVersion, alternative string) {
	f := cmd.Flags().Lookup(flag)
	if f != nil {
		msg := fmt.Sprintf("will be removed in version %q. Use %q instead.", removalVersion, alternative)
		f.Deprecated = msg
		f.Hidden = true
	}
}
