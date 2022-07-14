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

// ConfigLiterals used with set/unset commands
const (
	ConfigLiteralFeatures = "features"
	ConfigLiteralEnv      = "env"
)

func init() {
	configCmd.SetUsageFunc(cli.SubCmdUsageFunc)
	configCmd.AddCommand(
		getConfigCmd,
		initConfigCmd,
		setConfigCmd,
		unsetConfigCmd,
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
	Short: "Set config values at the given path",
	Long:  "Set config values at the given path. path values: [unstable-versions, cli.edition, features.global.<feature>, features.<plugin>.<feature>, env.<variable>]",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.Errorf("both path and value are required")
		}
		if len(args) > 2 {
			return errors.Errorf("only path and value are allowed")
		}

		// Acquire tanzu config lock
		config.AcquireTanzuConfigLock()
		defer config.ReleaseTanzuConfigLock()

		cfg, err := config.GetClientConfig()
		if err != nil {
			return err
		}

		err = setConfiguration(cfg, args[0], args[1])
		if err != nil {
			return err
		}

		return config.StoreClientConfig(cfg)
	},
}

// setConfiguration sets the key-value pair for the given path
func setConfiguration(cfg *configv1alpha1.ClientConfig, pathParam, value string) error {
	// special cases:
	// backward compatibility
	if pathParam == "unstable-versions" || pathParam == "cli.unstable-versions" {
		return setUnstableVersions(cfg, value)
	}

	if pathParam == "cli.edition" {
		return setEdition(cfg, value)
	}

	// parse the param
	paramArray := strings.Split(pathParam, ".")
	if len(paramArray) < 2 {
		return errors.New("unable to parse config path parameter into parts [" + pathParam + "]  (was expecting 'features.<plugin>.<feature>' or 'env.<env_variable>')")
	}

	configLiteral := paramArray[0]

	switch configLiteral {
	case ConfigLiteralFeatures:
		return setFeatures(cfg, paramArray, value)
	case ConfigLiteralEnv:
		return setEnvs(cfg, paramArray, value)
	default:
		return errors.New("unsupported config path parameter [" + configLiteral + "] (was expecting 'features.<plugin>.<feature>' or 'env.<env_variable>')")
	}
}

func setFeatures(cfg *configv1alpha1.ClientConfig, paramArray []string, value string) error {
	if len(paramArray) != 3 {
		return errors.New("unable to parse config path parameter into three parts [" + strings.Join(paramArray, ".") + "]  (was expecting 'features.<plugin>.<feature>'")
	}
	plugin := paramArray[1]
	featureName := paramArray[2]

	if cfg.ClientOptions == nil {
		cfg.ClientOptions = &configv1alpha1.ClientOptions{}
	}
	if cfg.ClientOptions.Features == nil {
		cfg.ClientOptions.Features = make(map[string]configv1alpha1.FeatureMap)
	}
	if cfg.ClientOptions.Features[plugin] == nil {
		cfg.ClientOptions.Features[plugin] = configv1alpha1.FeatureMap{}
	}
	cfg.ClientOptions.Features[plugin][featureName] = value
	return nil
}

func setEnvs(cfg *configv1alpha1.ClientConfig, paramArray []string, value string) error {
	if len(paramArray) != 2 {
		return errors.New("unable to parse config path parameter into two parts [" + strings.Join(paramArray, ".") + "]  (was expecting 'env.<variable>'")
	}
	envVariable := paramArray[1]

	if cfg.ClientOptions == nil {
		cfg.ClientOptions = &configv1alpha1.ClientOptions{}
	}
	if cfg.ClientOptions.Env == nil {
		cfg.ClientOptions.Env = make(map[string]string)
	}

	cfg.ClientOptions.Env[envVariable] = value
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

func setEdition(cfg *configv1alpha1.ClientConfig, edition string) error {
	editionOption := configv1alpha1.EditionSelector(edition)

	switch editionOption {
	case configv1alpha1.EditionCommunity, configv1alpha1.EditionStandard:
		cfg.SetEditionSelector(editionOption)
		// when community edition is set, configure the compatibility file to use
		// community edition's.
		err := cfg.SetCompatibilityFile(editionOption)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown edition: %s; should be one of [%s, %s]", editionOption, configv1alpha1.EditionStandard, configv1alpha1.EditionCommunity)
	}
	return nil
}

var initConfigCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize config with defaults",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Acquire tanzu config lock
		config.AcquireTanzuConfigLock()
		defer config.ReleaseTanzuConfigLock()

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
	Short: "Delete a server from the config",

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
				return fmt.Errorf("server %s not found in list of known servers", args[0])
			}
		}

		return nil
	},
}

var unsetConfigCmd = &cobra.Command{
	Use:   "unset <path>",
	Short: "Unset config values at the given path",
	Long:  "Unset config values at the given path. path values: [features.global.<feature>, features.<plugin>.<feature>, env.global.<variable>, env.<plugin>.<variable>]",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.Errorf("path is required")
		}
		if len(args) > 1 {
			return errors.Errorf("only path is allowed")
		}

		// Acquire tanzu config lock
		config.AcquireTanzuConfigLock()
		defer config.ReleaseTanzuConfigLock()

		cfg, err := config.GetClientConfig()
		if err != nil {
			return err
		}

		err = unsetConfiguration(cfg, args[0])
		if err != nil {
			return err
		}

		return config.StoreClientConfig(cfg)
	},
}

// unsetConfiguration unsets the key-value pair for the given path and removes it
func unsetConfiguration(cfg *configv1alpha1.ClientConfig, pathParam string) error {
	// parse the param
	paramArray := strings.Split(pathParam, ".")
	if len(paramArray) < 2 {
		return errors.New("unable to parse config path parameter into parts [" + pathParam + "]  (was expecting 'features.<plugin>.<feature>' or 'env.<env_variable>')")
	}

	configLiteral := paramArray[0]

	switch configLiteral {
	case ConfigLiteralFeatures:
		return unsetFeatures(cfg, paramArray)
	case ConfigLiteralEnv:
		return unsetEnvs(cfg, paramArray)
	default:
		return errors.New("unsupported config path parameter [" + configLiteral + "] (was expecting 'features.<plugin>.<feature>' or 'env.<env_variable>')")
	}
}

func unsetFeatures(cfg *configv1alpha1.ClientConfig, paramArray []string) error {
	if len(paramArray) != 3 {
		return errors.New("unable to parse config path parameter into three parts [" + strings.Join(paramArray, ".") + "]  (was expecting 'features.<plugin>.<feature>'")
	}
	plugin := paramArray[1]
	featureName := paramArray[2]

	if cfg.ClientOptions == nil || cfg.ClientOptions.Features == nil ||
		cfg.ClientOptions.Features[plugin] == nil {
		return nil
	}
	delete(cfg.ClientOptions.Features[plugin], featureName)
	return nil
}

func unsetEnvs(cfg *configv1alpha1.ClientConfig, paramArray []string) error {
	if len(paramArray) != 2 {
		return errors.New("unable to parse config path parameter into two parts [" + strings.Join(paramArray, ".") + "]  (was expecting 'env.<env_variable>'")
	}

	envVariable := paramArray[1]
	if cfg.ClientOptions == nil || cfg.ClientOptions.Env == nil {
		return nil
	}
	delete(cfg.ClientOptions.Env, envVariable)

	return nil
}
