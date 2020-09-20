package core

import (
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
	"github.com/vmware-tanzu-private/core/pkg/v1/client"
)

func init() {
	useCmd.SetUsageFunc(cli.SubCmdUsageFunc)
}

var useCmd = &cobra.Command{
	Use:   "use [name]",
	Short: "Use a server",
	Annotations: map[string]string{
		"group": string(cli.SystemCmdGroup),
	},
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		serverName := args[0]
		return client.SetCurrentServer(serverName)
	},
}
