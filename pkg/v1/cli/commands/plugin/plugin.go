package plugin

import (
	"github.com/spf13/cobra"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
)

// Plugin is a Tanzu CLI plugin.
type Plugin struct {
	Cmd *cobra.Command
}

// NewPlugin creates an instance of Plugin.
func NewPlugin(descriptor cli.PluginDescriptor) (*Plugin, error) {
	err := descriptor.Validate()
	if err != nil {
		return nil, err
	}
	p := &Plugin{
		Cmd: newRootCmd(descriptor),
	}
	return p, nil
}

// NewPluginFromFile create a new instance of Plugin from a file descriptor.
func NewPluginFromFile(path string) (*Plugin, error) {
	descriptor, err := cli.ParsePluginDescriptor(path)
	if err != nil {
		return nil, err
	}
	plugin, err := NewPlugin(descriptor)
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
