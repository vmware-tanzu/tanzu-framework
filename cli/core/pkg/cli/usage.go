// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"os"
	"text/template"

	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/component"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
)

// CmdMap is the map of command groups to plugins
type CmdMap map[string][]*cobra.Command

// MainUsage create the main usage display for tanzu cli.
type MainUsage struct{}

// NewMainUsage creates an instance of Usage.
func NewMainUsage() *MainUsage {
	return &MainUsage{}
}

// Func generates a usage func for cobra.
func (u *MainUsage) Func() func(*cobra.Command) error {
	return func(c *cobra.Command) error {
		return u.GenerateDescriptor(c, os.Stdout)
	}
}

// GenerateDescriptor generates a descriptor
func (u *MainUsage) GenerateDescriptor(c *cobra.Command, w io.Writer) error {
	cmdMap := CmdMap{}
	for _, cmd := range c.Commands() {
		if cmd.Hidden {
			continue
		}
		group := cmd.Annotations["group"]
		if group == "" {
			continue
		}
		g, ok := cmdMap[group]
		if !ok {
			g = []*cobra.Command{}
		}
		g = append(g, cmd)
		cmdMap[group] = g
	}

	var serverString string

	s, err := config.GetCurrentServer()
	if err != nil {
		serverString = "Not logged in"
	} else {
		serverString = fmt.Sprintf("Logged in to %s", component.Underline(s.Name))
	}
	d := struct {
		*cobra.Command
		CmdMap       CmdMap
		ServerString string
	}{
		c,
		cmdMap,
		serverString,
	}

	t := template.Must(template.New("usage").Funcs(TemplateFuncs).Parse(u.Template()))
	err = t.Execute(w, d)
	if err != nil {
		return err
	}
	return nil
}

// Template returns the template for the main usage.
func (u *MainUsage) Template() string {
	return `{{ bold "Usage:" }}
  {{.Command.CommandPath}} [command]{{if .HasExample}}

{{ bold "Examples:" }}
  {{.Example}}{{end}}

{{ bold "Available command groups:" }}
{{ range $group, $cmds := .CmdMap}}
  {{ bold $group }}{{ range $cmd := $cmds }}
    {{rpad $cmd.Name 24}}{{$cmd.Short}} {{end}}
	{{end}}

{{ bold "Flags:" }}
{{.LocalFlags.FlagUsages  | trimTrailingWhitespaces}}

Use "{{.CommandPath}} [command] --help" for more information about a command. {{ if ne .ServerString "" }}

{{ .ServerString }}{{end}}
`
}

// SubCmdUsageFunc is the usage func for a plugin.
var SubCmdUsageFunc = func(c *cobra.Command) error {
	t, err := template.New("usage").Funcs(TemplateFuncs).Parse(SubCmdTemplate)
	if err != nil {
		return err
	}
	return t.Execute(os.Stdout, c)
}

// SubCmdTemplate is the template for plugin commands.
const SubCmdTemplate = `{{ bold "Usage:" }}{{if .Runnable}}
{{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

{{ bold "Aliases:" }}
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

{{ bold "Examples:" }}
  {{.Example}}{{end}}{{if .HasAvailableSubCommands}}

{{ bold "Available Commands:" }}{{range .Commands}}{{if .IsAvailableCommand }}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

{{ bold "Flags:" }}
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

{{ bold "Global Flags:" }}
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

{{ bold "Additional help topics:" }}{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

// TemplateFuncs are the template usage funcs.
var TemplateFuncs = template.FuncMap{
	"rpad":                    component.Rpad,
	"bold":                    component.Bold,
	"underline":               component.Underline,
	"trimTrailingWhitespaces": component.TrimRightSpace,
	"beginsWith":              component.BeginsWith,
}
