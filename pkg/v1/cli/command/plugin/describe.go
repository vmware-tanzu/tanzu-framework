// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newDescribeCmd(description string) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "describe",
		Short:  "Describes the plugin",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("This command only describes a plugin's purpose... it doesn't accept any inputs.")
			}
			fmt.Println("Plugin description:", description)
			return nil
		},
	}

	return cmd
}
