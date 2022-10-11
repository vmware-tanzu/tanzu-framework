// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/multierr"
	"golang.org/x/mod/semver"

	cliapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/cli/v1alpha1"
)

// Plugin is a Tanzu CLI plugin.
type Plugin struct {
	Cmd *cobra.Command
}

// NewPlugin creates an instance of Plugin.
func NewPlugin(descriptor *cliapi.PluginDescriptor) (*Plugin, error) {
	ApplyDefaultConfig(descriptor)
	err := ValidatePlugin(descriptor)
	if err != nil {
		return nil, err
	}
	p := &Plugin{
		Cmd: newRootCmd(descriptor),
	}
	p.Cmd.AddCommand(lintCmd)
	p.Cmd.AddCommand(genDocsCmd)
	p.Cmd.AddCommand(newPostInstallCmd(descriptor))
	return p, nil
}

// NewPluginFromFile create a new instance of Plugin from a file descriptor.
func NewPluginFromFile(path string) (*Plugin, error) {
	descriptor, err := parsePluginDescriptor(path)
	if err != nil {
		return nil, err
	}
	plugin, err := NewPlugin(&descriptor)
	if err != nil {
		return nil, err
	}
	return plugin, nil
}

// AddCommands adds commands to the plugin.
func (p *Plugin) AddCommands(commands ...*cobra.Command) {
	p.Cmd.AddCommand(commands...)
}

// Execute executes the plugin.
func (p *Plugin) Execute() error {
	return p.Cmd.Execute()
}

// parsePluginDescriptor parses a plugin descriptor in yaml.
func parsePluginDescriptor(path string) (desc cliapi.PluginDescriptor, err error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return desc, errors.Wrap(err, "could not read plugin descriptor")
	}

	err = json.Unmarshal(b, &desc)
	if err != nil {
		return desc, errors.Wrap(err, "could not unmarshal plugin descriptor")
	}
	return
}

// ApplyDefaultConfig applies default configurations to plugin descriptor.
func ApplyDefaultConfig(p *cliapi.PluginDescriptor) {
	if p.PostInstallHook == nil {
		p.PostInstallHook = func() error {
			return nil
		}
	}
}

// ValidatePlugin validates the plugin descriptor.
func ValidatePlugin(p *cliapi.PluginDescriptor) (err error) {
	// skip builder plugin for bootstrapping
	if p.Name == "builder" {
		return nil
	}
	if p.Name == "" {
		err = multierr.Append(err, fmt.Errorf("plugin %q name cannot be empty", p.Name))
	}
	if p.Version == "" {
		err = multierr.Append(err, fmt.Errorf("plugin %q version cannot be empty", p.Name))
	}
	if !semver.IsValid(p.Version) && p.Version != "dev" {
		err = multierr.Append(err, fmt.Errorf("version %q %q is not a valid semantic version", p.Name, p.Version))
	}
	if p.Description == "" {
		err = multierr.Append(err, fmt.Errorf("plugin %q description cannot be empty", p.Name))
	}
	if p.Group == "" {
		err = multierr.Append(err, fmt.Errorf("plugin %q group cannot be empty", p.Name))
	}
	return
}
