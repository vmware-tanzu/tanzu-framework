package command

import "github.com/spf13/cobra"

// InitCmd initializes a repository.
var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a repository",
	RunE:  initialize,
}

func initialize(cmd *cobra.Command, args []string) error {
	return nil
}
