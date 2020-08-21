package core

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/commands/plugin"
)

func init() {
	versionCmd.SetUsageFunc(plugin.UsageFunc)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Version information",
	Annotations: map[string]string{
		"group": string(cli.SystemCmdGroup),
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("version: %s\nbuildDate: %s\nsha: %s\n", cli.BuildVersion, cli.BuildDate, cli.BuildSHA)
		return nil
	},
}
