package core

import (
	"io/ioutil"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/commands/plugin"
	"github.com/vmware-tanzu-private/core/pkg/v1/client"
)

func init() {
	configCmd.SetUsageFunc(plugin.UsageFunc)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show current config",
	Annotations: map[string]string{
		"group": string(cli.SystemCmdGroup),
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath, err := client.ConfigPath()
		if err != nil {
			return err
		}
		b, err := ioutil.ReadFile(cfgPath)
		if err != nil {
			return err
		}
		println(string(b))
		return nil
	},
}
