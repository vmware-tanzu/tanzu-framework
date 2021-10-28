// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"fmt"
	"os"
	"strings"

	"github.com/aunum/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	kerrors "k8s.io/apimachinery/pkg/util/errors"

	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	configv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/component"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"
)

func init() {
	configCmd.SetUsageFunc(cli.SubCmdUsageFunc)
	configCmd.AddCommand(
		getConfigCmd,
		initConfigCmd,
		setConfigCmd,
		serversCmd,
	)
	serversCmd.AddCommand(listServersCmd)
	addDeleteServersCmd()
}

var unattended bool

func addDeleteServersCmd() {
	listServersCmd.Flags().StringVarP(&outputFormat, "output", "o", "", "Output format (yaml|json|table)")
	deleteServersCmd.Flags().BoolVarP(&unattended, "yes", "y", false, "Delete the server entry without confirmation")
	serversCmd.AddCommand(deleteServersCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration for the CLI",
	Annotations: map[string]string{
		"group": string(cliv1alpha1.SystemCmdGroup),
	},
}

var getConfigCmd = &cobra.Command{
	Use:   "get",
	Short: "Get the current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath, err := config.ClientConfigPath()
		if err != nil {
			return err
		}
		b, err := os.ReadFile(cfgPath)
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	},
}

var setConfigCmd = &cobra.Command{
	Use:   "set <path> <value>",
	Short: "Set config values at the given path. path values: [unstable-versions, features.global.<feature>, features.<plugin>.<feature>]",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.Errorf("both path and value are required")
		}
		if len(args) > 2 {
			return errors.Errorf("only path and value are allowed")
		}
		cfg, err := config.GetClientConfig()
		if err != nil {
			return err
		}

		err = setFeature(cfg, args[0], args[1])
		if err != nil {
			return err
		}

		return config.StoreClientConfig(cfg)
	},
}

// setFeature sets the key-value pair for the given path
func setFeature(cfg *configv1alpha1.ClientConfig, pathParam, value string) error {
	// special cases:
	// backward compatibility
	if pathParam == "unstable-versions" {
		return setUnstableVersions(cfg, value)
	}

	// parse the param
	paramArray := strings.Split(pathParam, ".")
	if len(paramArray) != 3 {
		return errors.New("unable to parse config path parameter into three parts [" + pathParam + "]  (was expecting features.<plugin>.<feature>)")
	}

	featuresLiteral := paramArray[0]
	plugin := paramArray[1]
	key := paramArray[2]

	if featuresLiteral != "features" {
		return errors.New("unsupported config path parameter [" + featuresLiteral + "] (was expecting 'features.<plugin>.<feature>')")
	}

	if cfg.ClientOptions == nil {
		cfg.ClientOptions = &configv1alpha1.ClientOptions{}
	}
	if cfg.ClientOptions.Features == nil {
		cfg.ClientOptions.Features = make(map[string]configv1alpha1.FeatureMap)
	}
	if cfg.ClientOptions.Features[plugin] == nil {
		cfg.ClientOptions.Features[plugin] = configv1alpha1.FeatureMap{}
	}
	cfg.ClientOptions.Features[plugin][key] = value

	return nil
}

func setUnstableVersions(cfg *configv1alpha1.ClientConfig, value string) error {
	optionKey := configv1alpha1.VersionSelectorLevel(value)

	switch optionKey {
	case configv1alpha1.AllUnstableVersions,
		configv1alpha1.AlphaUnstableVersions,
		configv1alpha1.ExperimentalUnstableVersions,
		configv1alpha1.NoUnstableVersions:
		cfg.SetUnstableVersionSelector(optionKey)
	default:
		return fmt.Errorf("unknown unstable-versions setting: %s; should be one of [all, none, alpha, experimental]", optionKey)
	}
	return nil
}

var initConfigCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize config with defaults",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.GetClientConfig()
		if err != nil {
			return err
		}
		if cfg.ClientOptions == nil {
			cfg.ClientOptions = &configv1alpha1.ClientOptions{}
		}
		if cfg.ClientOptions.CLI == nil {
			cfg.ClientOptions.CLI = &configv1alpha1.CLIOptions{}
		}
		repos := cfg.ClientOptions.CLI.Repositories
		finalRepos := []configv1alpha1.PluginRepository{}
		for _, repo := range config.DefaultRepositories {
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

		err = config.StoreClientConfig(cfg)
		if err != nil {
			return err
		}

		descriptors, err := cli.ListPlugins()
		if err != nil {
			return err
		}

		errList := []error{}
		for _, desc := range descriptors {
			if err := cli.InitializePlugin(desc.Name); err != nil {
				errList = append(errList, err)
			}
		}
		return kerrors.NewAggregate(errList)
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
		cfg, err := config.GetClientConfig()
		if err != nil {
			return err
		}

		output := component.NewOutputWriter(cmd.OutOrStdout(), outputFormat, "Name", "Type", "Endpoint", "Path", "Context")
		for _, server := range cfg.KnownServers {
			var endpoint, path, context string
			if server.IsGlobal() {
				endpoint = server.GlobalOpts.Endpoint
			} else {
				endpoint = server.ManagementClusterOpts.Endpoint
				path = server.ManagementClusterOpts.Path
				context = server.ManagementClusterOpts.Context
			}
			output.AddRow(server.Name, server.Type, endpoint, path, context)
		}
		output.Render()
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
			isAborted = cli.AskForConfirmation("Deleting the server entry from the config will remove it from the list of tracked servers. " +
				"You will need to use tanzu login to track this server again. Are you sure you want to continue?")
		}

		if isAborted == nil {
			log.Infof("Deleting entry for cluster %s", args[0])
			serverExists, err := config.ServerExists(args[0])
			if err != nil {
				return err
			}

			if serverExists {
				err := config.RemoveServer(args[0])
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
