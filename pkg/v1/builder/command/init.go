package command

import (
	"github.com/aunum/log"
	"github.com/spf13/cobra"
	"github.com/vmware-tanzu-private/core/pkg/v1/builder/template"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/component"
)

// InitCmd initializes a repository.
var InitCmd = &cobra.Command{
	Use:   "init [name]",
	Short: "Initialize a repository",
	RunE:  initialize,
	Args:  cobra.ExactArgs(1),
}

func initialize(cmd *cobra.Command, args []string) error {
	name := args[0]
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

	data := struct {
		RepositoryName string
	}{
		RepositoryName: name,
	}
	targets := template.DefaultInitTargets
	if selection == "Github" {
		targets = append(targets, template.GithubCI)
	} else {
		targets = append(targets, template.GitlabCI)
	}
	for _, target := range targets {
		err = target.Run(name, data)
		if err != nil {
			return err
		}
	}
	log.Success("succesfully created repository")
	return nil
}
