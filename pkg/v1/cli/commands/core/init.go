package core

import (
	"github.com/aunum/log"
	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/commands/plugin"
)

func init() {
	initCmd.SetUsageFunc(plugin.UsageFunc)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the CLI",
	Annotations: map[string]string{
		"group": string(cli.SystemCmdGroup),
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		s := spin.New("%s   initializing")
		s.Start()
		catalog, err := cli.NewCatalog()
		if err != nil {
			return err
		}
		repo := cli.NewDefaultRepository()

		err = catalog.EnsureDistro(repo)
		if err != nil {
			return err
		}
		s.Stop()
		log.Success("successfully initialized CLI")
		return nil
	},
}
