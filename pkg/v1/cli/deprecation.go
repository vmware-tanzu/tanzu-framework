// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// DeprecateCommand marks the commands as deprecated and adds deprecation message
func DeprecateCommand(cmd *cobra.Command, removalVersion, alternative string) {
	msg := fmt.Sprintf("will be removed in version %q. Use %q instead", removalVersion, alternative)
	cmd.Deprecated = msg
}
