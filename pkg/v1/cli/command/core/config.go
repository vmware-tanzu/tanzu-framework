// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"

	clientv1alpha1 "github.com/vmware-tanzu-private/core/apis/client/v1alpha1"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
	"github.com/vmware-tanzu-private/core/pkg/v1/client"
)

func init() {
	configCmd.SetUsageFunc(cli.SubCmdUsageFunc)
	configCmd.AddCommand(
		showConfigCmd,
		initConfigCmd,
	)
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
