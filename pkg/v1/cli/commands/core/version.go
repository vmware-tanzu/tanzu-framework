package core

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
)

func init() {
	versionCmd.SetUsageFunc(cli.SubCmdUsageFunc)
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
