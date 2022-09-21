// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"fmt"

	"github.com/aunum/log"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/cli"
	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/component"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
)

var (
	gcpBucketName, gcpRootPath, name string
)

var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Manage plugin repositories",
}

func init() {
	repoCmd.SetUsageFunc(cli.SubCmdUsageFunc)
	repoCmd.AddCommand(
		listRepoCmd,
		addRepoCmd,
		updateRepoCmd,
		deleteRepoCmd,
	)
	addRepoCmd.Flags().StringVarP(&name, "name", "n", "", "name of repository")
	addRepoCmd.Flags().StringVarP(&gcpBucketName, "gcp-bucket-name", "b", "", "name of gcp bucket")
	addRepoCmd.Flags().StringVarP(&gcpRootPath, "gcp-root-path", "p", "", "root path in gcp bucket")
	cobra.MarkFlagRequired(addRepoCmd.Flags(), "name")            //nolint
	cobra.MarkFlagRequired(addRepoCmd.Flags(), "gcp-bucket-name") //nolint
	cobra.MarkFlagRequired(addRepoCmd.Flags(), "gcp-root-path")   //nolint

	updateRepoCmd.Flags().StringVarP(&gcpBucketName, "gcp-bucket-name", "b", "", "name of gcp bucket")
	updateRepoCmd.Flags().StringVarP(&gcpRootPath, "gcp-root-path", "p", "", "root path in gcp bucket")

	listRepoCmd.Flags().StringVarP(&outputFormat, "output", "o", "", "Output format (yaml|json|table)")
}

var listRepoCmd = &cobra.Command{
	Use:   "list",
	Short: "List available repositories",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.GetClientConfig()
		if err != nil {
			return err
		}

		repos := cli.LoadRepositories(cfg)
		output := component.NewOutputWriter(cmd.OutOrStdout(), outputFormat, "name")
		for index := range repos {
			output.AddRow(repos[index].Name())
		}
		output.Render()

		return nil
	},
}

var addRepoCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a repository",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Acquire tanzu config lock
		config.AcquireTanzuConfigLock()
		defer config.ReleaseTanzuConfigLock()

		cfg, err := config.GetClientConfigNoLock()
		if err != nil {
			return err
		}
		if cfg.ClientOptions == nil {
			cfg.ClientOptions = &configapi.ClientOptions{}
		}
		if cfg.ClientOptions.CLI == nil {
			cfg.ClientOptions.CLI = &configapi.CLIOptions{}
		}
		repos := cfg.ClientOptions.CLI.Repositories
		gcpRepo := &configapi.GCPPluginRepository{
			Name:       name,
			BucketName: gcpBucketName,
			RootPath:   gcpRootPath,
		}
		pluginRepo := configapi.PluginRepository{
			GCPPluginRepository: gcpRepo,
		}
		for _, repo := range repos {
			if repo.GCPPluginRepository != nil {
				if repo.GCPPluginRepository.Name == name {
					return fmt.Errorf("repo name %q already exists", name)
				}
			}
		}
		repos = append(repos, pluginRepo)
		cfg.ClientOptions.CLI.Repositories = repos
		err = config.StoreClientConfig(cfg)
		if err != nil {
			return err
		}
		log.Successf("successfully added repository %s", pluginRepo.GCPPluginRepository.Name)
		return nil
	},
}

var updateRepoCmd = &cobra.Command{
	Use:   "update [name]",
	Short: "Update a repository configuration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoName := args[0]

		// Acquire tanzu config lock
		config.AcquireTanzuConfigLock()
		defer config.ReleaseTanzuConfigLock()

		cfg, err := config.GetClientConfigNoLock()
		if err != nil {
			return err
		}

		repoNoExistError := fmt.Errorf("repo %q does not exist", repoName)
		if cfg.ClientOptions == nil {
			return repoNoExistError
		}
		if cfg.ClientOptions.CLI == nil {
			return repoNoExistError
		}
		repos := cfg.ClientOptions.CLI.Repositories

		newRepos := []configapi.PluginRepository{}
		for _, repo := range repos {
			if repo.GCPPluginRepository != nil {
				if repo.GCPPluginRepository.Name == repoName {
					if gcpBucketName != "" {
						repo.GCPPluginRepository.BucketName = gcpBucketName
					}
					if gcpRootPath != "" {
						repo.GCPPluginRepository.RootPath = gcpRootPath
					}
				}
				newRepos = append(newRepos, repo)
			}
		}
		cfg.ClientOptions.CLI.Repositories = newRepos
		err = config.StoreClientConfig(cfg)
		if err != nil {
			return err
		}
		log.Successf("successfully updated repository configuration for %s", repoName)
		return nil
	},
}

var deleteRepoCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a repository",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		repoName := args[0]

		// Acquire tanzu config lock
		config.AcquireTanzuConfigLock()
		defer config.ReleaseTanzuConfigLock()

		cfg, err := config.GetClientConfigNoLock()
		if err != nil {
			return err
		}
		if cfg.ClientOptions == nil || cfg.ClientOptions.CLI == nil {
			return fmt.Errorf("repository %q unknown", repoName)
		}

		r := cfg.ClientOptions.CLI.Repositories
		newRepos := []configapi.PluginRepository{}
		for _, repo := range r {
			if repo.GCPPluginRepository == nil {
				continue
			}
			if repo.GCPPluginRepository.Name == repoName {
				continue
			}
			newRepos = append(newRepos, repo)
		}
		cfg.ClientOptions.CLI.Repositories = newRepos
		err = config.StoreClientConfig(cfg)
		if err != nil {
			return err
		}
		log.Successf("successfully deleted repository %s", repoName)
		return nil
	},
}
