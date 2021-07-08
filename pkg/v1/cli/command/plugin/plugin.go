// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"github.com/spf13/cobra"

	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli"
)

// Plugin is a Tanzu CLI plugin.
type Plugin struct {
	Cmd *cobra.Command
}

// NewPlugin creates an instance of Plugin.
func NewPlugin(descriptor *cliv1alpha1.PluginDescriptor) (*Plugin, error) {
	cli.ApplyDefaultConfig(descriptor)
	err := cli.ValidatePlugin(descriptor)
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
	descriptor, err := cli.ParsePluginDescriptor(path)
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
