package command

import (
	"github.com/spf13/cobra"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/component"
)

// InitCmd initializes a repository.
var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a repository",
	RunE:  initialize,
}

func initialize(cmd *cobra.Command, args []string) error {
	cfg := &component.SelectConfig{
		Message: "choose a repository type",
		Options: []string{
			"Github",
			"Gitlab",
		},
	}
	var selection string
	err := component.Select(cfg, &selection)
	if err != nil {
		return err
	}
	return nil
}
