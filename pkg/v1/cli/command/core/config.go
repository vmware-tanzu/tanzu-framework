// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"fmt"
	"io/ioutil"

	"github.com/aunum/log"
	"github.com/pkg/errors"

	"github.com/spf13/cobra"

	clientv1alpha1 "github.com/vmware-tanzu-private/core/apis/client/v1alpha1"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/component"
	"github.com/vmware-tanzu-private/core/pkg/v1/client"
)

func init() {
	configCmd.SetUsageFunc(cli.SubCmdUsageFunc)
	configCmd.AddCommand(
		showConfigCmd,
		initConfigCmd,
		serversCmd,
	)
	serversCmd.AddCommand(listServersCmd)
	addDeleteServersCmd()
}

var unattended bool

func addDeleteServersCmd() {
	deleteServersCmd.Flags().BoolVarP(&unattended, "yes", "y", false, "Delete the server entry without confirmation")
	serversCmd.AddCommand(deleteServersCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration for the CLI",
	Annotations: map[string]string{
		"group": string(cli.SystemCmdGroup),
	},
}

var showConfigCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath, err := client.ConfigPath()
		if err != nil {
			return err
		}
		b, err := ioutil.ReadFile(cfgPath)
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	},
}

var initConfigCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize config with defaults",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := client.GetConfig()
		if err != nil {
			return err
		}
		if cfg.ClientOptions == nil {
			cfg.ClientOptions = &clientv1alpha1.ClientOptions{}
		}
		if cfg.ClientOptions.CLI == nil {
			cfg.ClientOptions.CLI = &clientv1alpha1.CLIOptions{}
		}
		repos := cfg.ClientOptions.CLI.Repositories
		finalRepos := []clientv1alpha1.PluginRepository{}
		for _, repo := range client.DefaultRepositories {
			var exists bool
			for _, r := range repos {
				if repo.GCPPluginRepository.Name == r.GCPPluginRepository.Name {
					finalRepos = append(finalRepos, r)
					exists = true
				}
			}
			if !exists {
				finalRepos = append(finalRepos, repo)
			}
		}
		cfg.ClientOptions.CLI.Repositories = finalRepos

		err = client.StoreConfig(cfg)
		if err != nil {
			return err
		}

		return nil
	},
}

var serversCmd = &cobra.Command{
	Use:   "server",
	Short: "Configured servers",
}

var listServersCmd = &cobra.Command{
	Use:   "list",
	Short: "List servers",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := client.GetConfig()
		if err != nil {
			return err
		}

		data := [][]string{}
		for _, server := range cfg.KnownServers {
			var endpoint, path, context string
			if server.IsGlobal() {
				endpoint = server.GlobalOpts.Endpoint
			} else {
				endpoint = server.ManagementClusterOpts.Endpoint
				path = server.ManagementClusterOpts.Path
				context = server.ManagementClusterOpts.Context
			}
			data = append(data, []string{server.Name, string(server.Type), endpoint, path, context})
		}
		table := component.NewTableWriter("Name", "Type", "Endpoint", "Path", "Context")

		for _, v := range data {
			table.Append(v)
		}
		table.Render()
		return nil
	},
}

var deleteServersCmd = &cobra.Command{
	Use:   "delete SERVER_NAME",
	Short: "delete a server from the config",

	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.Errorf("Server name required. Usage: tanzu config server delete server_name")
		}

		var isAborted error
		if !unattended {
			isAborted = cli.AskForConfirmation("Deleting the sever entry from the config will remove it from the list of tracked servers. " +
				"You will need to use tanzu login to track this server again. Are you sure you want to continue?")
		}

		if isAborted == nil {
			log.Infof("Deleting entry for cluster %s", args[0])
			serverExists, err := client.ServerExists(args[0])
			if err != nil {
				return err
			}

			if serverExists {
				err := client.RemoveServer(args[0])
				if err != nil {
					return err
				}
			} else {
				return errors.New(fmt.Sprintf("Server %s not found in list of known servers", args[0]))
			}
		}

		return nil
	},
}
