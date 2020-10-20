package core

import (
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
)

func init() {
	pluginCmd.SetUsageFunc(cli.SubCmdUsageFunc)
	pluginCmd.AddCommand(
		listRepoCmd,
	)
	pluginCmd.PersistentFlags().StringVarP(&local, "local", "l", "", "path to local repository")
}

var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Manage plugin repositories",
	Annotations: map[string]string{
		"group": string(cli.SystemCmdGroup),
	},
}

var listRepoCmd = &cobra.Command{
	Use:   "list",
	Short: "List available repositories",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

var addRepoCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a repository",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

var deleteRepoCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a repository",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		return nil
	},
}
