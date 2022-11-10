// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"fmt"
	"os"
	"strings"

	"github.com/aunum/log"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/cli"
	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/config"
	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/pluginmanager"

	cliapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/cli/v1alpha1"
	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/component"
	configlib "github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
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
		"group": string(cliapi.SystemCmdGroup),
	},
}

var getConfigCmd = &cobra.Command{
	Use:   "get",
	Short: "Get the current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath, err := configlib.ClientConfigPath()
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
		configlib.AcquireTanzuConfigLock()
		defer configlib.ReleaseTanzuConfigLock()

		cfg, err := configlib.GetClientConfigNoLock()
		if err != nil {
			return err
		}

		err = setConfiguration(cfg, args[0], args[1])
		if err != nil {
			return err
		}

		return configlib.StoreClientConfig(cfg)
	},
}

// setConfiguration sets the key-value pair for the given path
func setConfiguration(cfg *configapi.ClientConfig, pathParam, value string) error {
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

func setFeatures(cfg *configapi.ClientConfig, paramArray []string, value string) error {
	if len(paramArray) != 3 {
		return errors.New("unable to parse config path parameter into three parts [" + strings.Join(paramArray, ".") + "]  (was expecting 'features.<plugin>.<feature>'")
	}
	plugin := paramArray[1]
	featureName := paramArray[2]

	if cfg.ClientOptions == nil {
		cfg.ClientOptions = &configapi.ClientOptions{}
	}
	if cfg.ClientOptions.Features == nil {
		cfg.ClientOptions.Features = make(map[string]configapi.FeatureMap)
	}
	if cfg.ClientOptions.Features[plugin] == nil {
		cfg.ClientOptions.Features[plugin] = configapi.FeatureMap{}
	}
	cfg.ClientOptions.Features[plugin][featureName] = value
	return nil
}

func setEnvs(cfg *configapi.ClientConfig, paramArray []string, value string) error {
	if len(paramArray) != 2 {
		return errors.New("unable to parse config path parameter into two parts [" + strings.Join(paramArray, ".") + "]  (was expecting 'env.<variable>'")
	}
	envVariable := paramArray[1]

	if cfg.ClientOptions == nil {
		cfg.ClientOptions = &configapi.ClientOptions{}
	}
	if cfg.ClientOptions.Env == nil {
		cfg.ClientOptions.Env = make(map[string]string)
	}

	cfg.ClientOptions.Env[envVariable] = value
	return nil
}

func setUnstableVersions(cfg *configapi.ClientConfig, value string) error {
	optionKey := configapi.VersionSelectorLevel(value)

	switch optionKey {
	case configapi.AllUnstableVersions,
		configapi.AlphaUnstableVersions,
		configapi.ExperimentalUnstableVersions,
		configapi.NoUnstableVersions:
		cfg.SetUnstableVersionSelector(optionKey)
	default:
		return fmt.Errorf("unknown unstable-versions setting: %s; should be one of [all, none, alpha, experimental]", optionKey)
	}
	return nil
}

func setEdition(cfg *configapi.ClientConfig, edition string) error {
	editionOption := configapi.EditionSelector(edition)

	switch editionOption {
	case configapi.EditionCommunity, configapi.EditionStandard:
		cfg.SetEditionSelector(editionOption)
	default:
		return fmt.Errorf("unknown edition: %s; should be one of [%s, %s]", editionOption, configapi.EditionStandard, configapi.EditionCommunity)
	}
	return nil
}

var initConfigCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize config with defaults",
	Long:  "Initialize config with defaults including plugin specific defaults for all active and installed plugins",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Acquire tanzu config lock
		configlib.AcquireTanzuConfigLock()
		defer configlib.ReleaseTanzuConfigLock()

		cfg, err := configlib.GetClientConfigNoLock()
		if err != nil {
			return err
		}
		if cfg.ClientOptions == nil {
			cfg.ClientOptions = &configapi.ClientOptions{}
		}
		if cfg.ClientOptions.CLI == nil {
			cfg.ClientOptions.CLI = &configapi.CLIOptions{}
		}

		serverPluginDescriptors, standalonePluginDescriptors, err := pluginmanager.InstalledPlugins()
		if err != nil {
			return err
		}

		// Add the default featureflags for active plugins based on the currentContext
		// Plugins that are installed but are not active plugin will not be processed here
		// and defaultFeatureFlags will not be configured for those plugins
		for _, desc := range append(serverPluginDescriptors, standalonePluginDescriptors...) {
			config.AddDefaultFeatureFlagsIfMissing(cfg, desc.DefaultFeatureFlags)
		}

		err = configlib.StoreClientConfig(cfg)
		if err != nil {
			return err
		}

		log.Success("successfully initialized the config")
		return nil
	},
}

// Note: Shall be deprecated in a future version. Superseded by 'tanzu context' command.
var serversCmd = &cobra.Command{
	Use:   "server",
	Short: "Configured servers",
}

// Note: Shall be deprecated in a future version. Superseded by 'tanzu context list' command.
var listServersCmd = &cobra.Command{
	Use:   "list",
	Short: "List servers",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := configlib.GetClientConfig()
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

// Note: Shall be deprecated in a future version. Superseded by 'tanzu context delete' command.
var deleteServersCmd = &cobra.Command{
	Use:   "delete SERVER_NAME",
	Short: "Delete a server from the config",

	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.Errorf("Server name required. Usage: tanzu config server delete server_name")
		}

		var isAborted error
		if !unattended {
			isAborted = component.AskForConfirmation("Deleting the server entry from the config will remove it from the list of tracked servers. " +
				"You will need to use tanzu login to track this server again. Are you sure you want to continue?")
		}

		if isAborted == nil {
			log.Infof("Deleting entry for cluster %s", args[0])
			serverExists, err := configlib.ServerExists(args[0])
			if err != nil {
				return err
			}

			if serverExists {
				err := configlib.RemoveServer(args[0])
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

		return unsetConfiguration(args[0])

	},
}

// unsetConfiguration unsets the key-value pair for the given path and removes it
func unsetConfiguration(pathParam string) error {
	// parse the param
	paramArray := strings.Split(pathParam, ".")
	if len(paramArray) < 2 {
		return errors.New("unable to parse config path parameter into parts [" + pathParam + "]  (was expecting 'features.<plugin>.<feature>' or 'env.<env_variable>')")
	}

	configLiteral := paramArray[0]

	switch configLiteral {
	case ConfigLiteralFeatures:
		return unsetFeatures(paramArray)
	case ConfigLiteralEnv:
		return unsetEnvs(paramArray)
	default:
		return errors.New("unsupported config path parameter [" + configLiteral + "] (was expecting 'features.<plugin>.<feature>' or 'env.<env_variable>')")
	}
}

func unsetFeatures(paramArray []string) error {
	if len(paramArray) != 3 {
		return errors.New("unable to parse config path parameter into three parts [" + strings.Join(paramArray, ".") + "]  (was expecting 'features.<plugin>.<feature>'")
	}
	plugin := paramArray[1]
	featureName := paramArray[2]

	return configlib.DeleteFeature(plugin, featureName)
}

func unsetEnvs(paramArray []string) error {
	if len(paramArray) != 2 {
		return errors.New("unable to parse config path parameter into two parts [" + strings.Join(paramArray, ".") + "]  (was expecting 'env.<env_variable>'")
	}

	envVariable := paramArray[1]
	return configlib.DeleteEnv(envVariable)
}
