package main

import (
	"fmt"
	"os"

	"github.com/aunum/log"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/command/plugin"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/component"
)

var descriptor = cli.PluginDescriptor{
	Name:        "cluster",
	Description: "Kubernetes cluster operations",
	Version:     "v0.0.1",
	Group:       cli.RunCmdGroup,
}

func main() {
	p, err := plugin.NewPlugin(&descriptor)
	if err != nil {
		log.Fatal(err)
	}
	p.AddCommands(
		createClusterCmd,
		getClusterCmd,
	)
	if err := p.Execute(); err != nil {
		os.Exit(1)
	}
}

var getClusterCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("in progress...")
		return nil
	},
}

var listClustersCmd = &cobra.Command{
	Use:   "list",
	Short: "List clusters",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("in progress...")
		return nil
	},
}

var createClusterCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("in progress...")
		c := component.PromptConfig{
			Message: "enter cluster name",
			Help:    "give the cluster a name",
			Default: "mycluster",
		}
		var name string
		err := c.Run(&name)
		if err != nil {
			return err
		}
		fmt.Println("name: ", name)
		return nil
	},
}

var updateClusterCmd = &cobra.Command{
	Use:   "update",
	Short: "update a cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("in progress...")
		return nil
	},
}

var deleteClusterCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete a cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("in progress...")
		return nil
	},
}
