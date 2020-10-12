package core

import (
	"github.com/aunum/log"
	"github.com/caarlos0/spin"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
)

func init() {
	initCmd.SetUsageFunc(cli.SubCmdUsageFunc)
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
		err = catalog.EnsureDistro(cli.DefaultMultiRepo)
		if err != nil {
			return err
		}
		s.Stop()
		log.Success("successfully initialized CLI")
		return nil
	},
}
