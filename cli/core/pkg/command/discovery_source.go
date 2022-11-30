// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"fmt"
	"strings"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/config"

	"github.com/aunum/log"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/cli"
	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/common"
	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/discovery"
	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/component"
	configlib "github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
)

var (
	discoverySourceType, discoverySourceName, uri string
)

var discoverySourceCmd = &cobra.Command{
	Use:   "source",
	Short: "Manage plugin discovery sources",
	Long:  "Manage plugin discovery sources. Discovery source provides metadata about the list of available plugins, their supported versions and how to download them.",
}

func init() {
	discoverySourceCmd.SetUsageFunc(cli.SubCmdUsageFunc)
	discoverySourceCmd.AddCommand(
		listDiscoverySourceCmd,
		addDiscoverySourceCmd,
		updateDiscoverySourceCmd,
		deleteDiscoverySourceCmd,
	)
	addDiscoverySourceCmd.Flags().StringVarP(&discoverySourceName, "name", "n", "", "name of discovery source")
	addDiscoverySourceCmd.Flags().StringVarP(&discoverySourceType, "type", "t", "", "type of discovery source")
	addDiscoverySourceCmd.Flags().StringVarP(&uri, "uri", "u", "", "URI for discovery source. URI format might be different based on the type of discovery source")

	// Not handling errors below because cobra handles the error when flag user doesn't provide these required flags
	_ = cobra.MarkFlagRequired(addDiscoverySourceCmd.Flags(), "name")
	_ = cobra.MarkFlagRequired(addDiscoverySourceCmd.Flags(), "type")
	_ = cobra.MarkFlagRequired(addDiscoverySourceCmd.Flags(), "uri")

	updateDiscoverySourceCmd.Flags().StringVarP(&discoverySourceType, "type", "t", "", "type of discovery source")
	updateDiscoverySourceCmd.Flags().StringVarP(&uri, "uri", "u", "", "URI for discovery source. URI format might be different based on the type of discovery source")

	listDiscoverySourceCmd.Flags().StringVarP(&outputFormat, "output", "o", "", "Output format (yaml|json|table)")
}

var listDiscoverySourceCmd = &cobra.Command{
	Use:   "list",
	Short: "List available discovery sources",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := configlib.GetClientConfig()
		if err != nil {
			return err
		}

		output := component.NewOutputWriter(cmd.OutOrStdout(), outputFormat, "name", "type", "scope")

		// Get standalone scoped discoveries
		if cfg.ClientOptions != nil && cfg.ClientOptions.CLI != nil && cfg.ClientOptions.CLI.DiscoverySources != nil {
			outputFromDiscoverySources(cfg.ClientOptions.CLI.DiscoverySources, common.PluginScopeStandalone, output)
		}

		// If context-target feature is activated, get discovery sources from all active context
		// else get discovery sources from current server
		if configlib.IsFeatureActivated(config.FeatureContextCommand) {
			mapContexts, err := configlib.GetAllCurrentContextsMap()
			if err == nil {
				for _, context := range mapContexts {
					outputFromDiscoverySources(context.DiscoverySources, common.PluginScopeContext, output)
				}
			}
		} else {
			server, err := configlib.GetCurrentServer()
			if err == nil && server != nil {
				outputFromDiscoverySources(server.DiscoverySources, common.PluginScopeContext, output)
			}
		}

		output.Render()

		return nil
	},
}

func outputFromDiscoverySources(discoverySources []configapi.PluginDiscovery, scope string, output component.OutputWriter) {
	for _, ds := range discoverySources {
		dsName, dsType := discoverySourceNameAndType(ds)
		output.AddRow(dsName, dsType, scope)
	}
}

var addDiscoverySourceCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a discovery source",
	Long:  "Add a discovery source. Supported discovery types are: oci, local",
	Example: `
    # Add a local discovery source. If URI is relative path,
    # $HOME/.config/tanzu-plugins will be considered based path
    tanzu plugin source add --name standalone-local --type local --uri path/to/local/discovery

    # Add an OCI discovery source. URI should be an OCI image.
    tanzu plugin source add --name standalone-oci --type oci --uri projects.registry.vmware.com/tkg/tanzu-plugins/standalone:latest`,

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

		discoverySources, err := addDiscoverySource(cfg.ClientOptions.CLI.DiscoverySources, discoverySourceName, discoverySourceType, uri)
		if err != nil {
			return err
		}

		cfg.ClientOptions.CLI.DiscoverySources = discoverySources
		err = configlib.StoreClientConfig(cfg)
		if err != nil {
			return err
		}
		log.Successf("successfully added discovery source %s", discoverySourceName)
		return nil
	},
}

var updateDiscoverySourceCmd = &cobra.Command{
	Use:   "update [name]",
	Short: "Update a discovery source configuration",
	Args:  cobra.ExactArgs(1),
	Example: `
    # Update a local discovery source. If URI is relative path, 
    # $HOME/.config/tanzu-plugins will be considered base path
    tanzu plugin source update standalone-local --type local --uri new/path/to/local/discovery

    # Update an OCI discovery source. URI should be an OCI image.
    tanzu plugin source update standalone-oci --type oci --uri projects.registry.vmware.com/tkg/tanzu-plugins/standalone:v1.0`,

	RunE: func(cmd *cobra.Command, args []string) error {
		discoveryName := args[0]

		// Acquire tanzu config lock
		configlib.AcquireTanzuConfigLock()
		defer configlib.ReleaseTanzuConfigLock()

		cfg, err := configlib.GetClientConfigNoLock()
		if err != nil {
			return err
		}

		discoveryNoExistError := fmt.Errorf("discovery %q does not exist", discoveryName)
		if cfg.ClientOptions == nil {
			return discoveryNoExistError
		}
		if cfg.ClientOptions.CLI == nil {
			return discoveryNoExistError
		}

		newDiscoverySources, err := updateDiscoverySources(cfg.ClientOptions.CLI.DiscoverySources, discoveryName, discoverySourceType, uri)
		if err != nil {
			return err
		}

		cfg.ClientOptions.CLI.DiscoverySources = newDiscoverySources
		err = configlib.StoreClientConfig(cfg)
		if err != nil {
			return err
		}
		log.Successf("updated discovery source %s", discoveryName)
		return nil
	},
}

var deleteDiscoverySourceCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a discovery source",
	Args:  cobra.ExactArgs(1),
	Example: `
    # Delete a discovery source
    tanzu plugin discovery delete standalone-oci`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		discoveryName := args[0]

		// Acquire tanzu config lock
		configlib.AcquireTanzuConfigLock()
		defer configlib.ReleaseTanzuConfigLock()

		cfg, err := configlib.GetClientConfigNoLock()
		if err != nil {
			return err
		}
		if cfg.ClientOptions == nil || cfg.ClientOptions.CLI == nil {
			return fmt.Errorf("discovery %q unknown", discoveryName)
		}

		newDiscoverySources, err := deleteDiscoverySource(cfg.ClientOptions.CLI.DiscoverySources, discoveryName)
		if err != nil {
			return err
		}

		cfg.ClientOptions.CLI.DiscoverySources = newDiscoverySources
		err = configlib.StoreClientConfig(cfg)
		if err != nil {
			return err
		}
		log.Successf("deleted discovery source %s", discoveryName)
		return nil
	},
}

func createDiscoverySource(dsType, dsName, uri string) (configapi.PluginDiscovery, error) {
	pluginDiscoverySource := configapi.PluginDiscovery{}
	if dsType == "" {
		return pluginDiscoverySource, errors.New("discovery source type cannot be empty")
	}
	if dsName == "" {
		return pluginDiscoverySource, errors.New("discovery source name cannot be empty")
	}

	switch strings.ToLower(dsType) {
	case common.DiscoveryTypeLocal:
		pluginDiscoverySource.Local = createLocalDiscoverySource(dsName, uri)
	case common.DiscoveryTypeOCI:
		pluginDiscoverySource.OCI = createOCIDiscoverySource(dsName, uri)
	case common.DiscoveryTypeREST:
		pluginDiscoverySource.REST = createRESTDiscoverySource(dsName, uri)
	case common.DiscoveryTypeGCP, common.DiscoveryTypeKubernetes:
		return pluginDiscoverySource, errors.Errorf("discovery source type '%s' is not yet supported", dsType)
	default:
		return pluginDiscoverySource, errors.Errorf("unknown discovery source type '%s'", dsType)
	}
	return pluginDiscoverySource, nil
}

func createLocalDiscoverySource(discoveryName, uri string) *configapi.LocalDiscovery {
	return &configapi.LocalDiscovery{
		Name: discoveryName,
		Path: uri,
	}
}

func createOCIDiscoverySource(discoveryName, uri string) *configapi.OCIDiscovery {
	return &configapi.OCIDiscovery{
		Name:  discoveryName,
		Image: uri,
	}
}

func createRESTDiscoverySource(discoveryName, uri string) *configapi.GenericRESTDiscovery {
	return &configapi.GenericRESTDiscovery{
		Name:     discoveryName,
		Endpoint: uri,
	}
}

func discoverySourceNameAndType(ds configapi.PluginDiscovery) (string, string) {
	switch {
	case ds.GCP != nil:
		return ds.GCP.Name, common.DiscoveryTypeGCP
	case ds.Kubernetes != nil:
		return ds.Kubernetes.Name, common.DiscoveryTypeKubernetes
	case ds.Local != nil:
		return ds.Local.Name, common.DiscoveryTypeLocal
	case ds.OCI != nil:
		return ds.OCI.Name, common.DiscoveryTypeOCI
	case ds.REST != nil:
		return ds.REST.Name, common.DiscoveryTypeREST
	default:
		return "-", "Unknown" // Unknown discovery source found
	}
}

func addDiscoverySource(discoverySources []configapi.PluginDiscovery, dsName, dsType, uri string) ([]configapi.PluginDiscovery, error) {
	for _, ds := range discoverySources {
		if discovery.CheckDiscoveryName(ds, dsName) {
			return nil, fmt.Errorf("discovery name %q already exists", dsName)
		}
	}

	pluginDiscoverySource, err := createDiscoverySource(dsType, dsName, uri)
	if err != nil {
		return nil, err
	}

	discoverySources = append(discoverySources, pluginDiscoverySource)
	return discoverySources, nil
}

func deleteDiscoverySource(discoverySources []configapi.PluginDiscovery, discoveryName string) ([]configapi.PluginDiscovery, error) {
	newDiscoverySources := []configapi.PluginDiscovery{}
	found := false
	for _, ds := range discoverySources {
		if discovery.CheckDiscoveryName(ds, discoveryName) {
			found = true
			continue
		}
		newDiscoverySources = append(newDiscoverySources, ds)
	}
	if !found {
		return nil, fmt.Errorf("discovery source %q does not exist", discoveryName)
	}
	return newDiscoverySources, nil
}

func updateDiscoverySources(discoverySources []configapi.PluginDiscovery, dsName, dsType, uri string) ([]configapi.PluginDiscovery, error) {
	var newDiscoverySources []configapi.PluginDiscovery
	var err error

	found := false
	for _, ds := range discoverySources {
		if discovery.CheckDiscoveryName(ds, dsName) {
			found = true
			ds, err = createDiscoverySource(dsType, dsName, uri)
			if err != nil {
				return nil, err
			}
		}
		newDiscoverySources = append(newDiscoverySources, ds)
	}
	if !found {
		return nil, fmt.Errorf("discovery source %q does not exist", dsName)
	}
	return newDiscoverySources, nil
}
