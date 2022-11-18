// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"fmt"

	"encoding/json"

	"github.com/pkg/errors"
)

// PluginBasicOps helps to perform the plugin command operations
type PluginBasicOps interface {
	// ListPlugins lists all plugins by running 'tanzu plugin list' command
	ListPlugins() ([]PluginListInfo, error)
}

// PluginSourceOps helps 'plugin source' commands
type PluginSourceOps interface {
	// AddPluginDiscoverySource adds plugin discovery source
	AddPluginDiscoverySource(discoveryOpts *DiscoveryOptions) error
}

// PluginLifecycleOps helps to perform the plugin and its sub-commands lifecycle operations
type PluginLifecycleOps interface {
	PluginBasicOps
	PluginSourceOps
}

type DiscoveryOptions struct {
	Name       string
	SourceType string
	URI        string
}

type pluginLifecycleOps struct {
	cmdExe CmdOps
	PluginLifecycleOps
}

func NewPluginLifecycleOps() PluginLifecycleOps {
	return &pluginLifecycleOps{
		cmdExe: NewCmdOps(),
	}
}

func (po *pluginLifecycleOps) AddPluginDiscoverySource(discoveryOpts *DiscoveryOptions) error {
	addCmd := fmt.Sprintf(AddPluginSource, discoveryOpts.Name, discoveryOpts.SourceType, discoveryOpts.URI)
	_, _, err := po.cmdExe.Exec(addCmd)
	return err
}

func (po *pluginLifecycleOps) ListPlugins() ([]PluginListInfo, error) {
	out, _, err := po.cmdExe.Exec(ListPluginsCmd)
	if err != nil {
		return nil, err
	}
	jsonStr := out.String()
	var list []PluginListInfo
	err = json.Unmarshal([]byte(jsonStr), &list)
	if err != nil {
		return nil, errors.Wrap(err, "failed to construct json node from config get output")
	}
	return list, nil
}
