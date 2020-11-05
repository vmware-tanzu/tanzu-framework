package plugin

import (
	"encoding/json"
	"fmt"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli"

	"github.com/spf13/cobra"
)

func newInfoCmd(desc *cli.PluginDescriptor) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "info",
		Short:  "Plugin info",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := json.Marshal(desc)
			if err != nil {
				return err
			}
			fmt.Println(string(b))
			return nil
		},
	}

	return cmd
}
