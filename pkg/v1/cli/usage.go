package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"
	"unicode"

	clientv1alpha1 "github.com/vmware-tanzu-private/core/apis/client/v1alpha1"
	"github.com/vmware-tanzu-private/core/pkg/v1/client"

	"github.com/spf13/cobra"

	"github.com/logrusorgru/aurora"
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

	s, err := client.GetCurrentServer()
	if err != nil {
		return err
	}
	d := struct {
		*cobra.Command
		CmdMap CmdMap
		Server clientv1alpha1.Server
	}{
		c,
		cmdMap,
		s,
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

Use "{{.CommandPath}} [command] --help" for more information about a command.

Logged in to {{ underline .Server.Name }}
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
	"rpad":                    rpad,
	"bold":                    bold,
	"underline":               underline,
	"trimTrailingWhitespaces": trimRightSpace,
}

// rpad adds padding to the right of a string.
// from https://github.com/spf13/cobra/blob/993cc5372a05240dfd59e3ba952748b36b2cd117/cobra.go#L29
func rpad(s string, padding int) string {
	template := fmt.Sprintf("%%-%ds", padding)
	return fmt.Sprintf(template, s)
}

func underline(s string) string {
	return aurora.Underline(s).String()
}

func bold(s string) string {
	return aurora.Bold(s).String()
}

func trimRightSpace(s string) string {
	return strings.TrimRightFunc(s, unicode.IsSpace)
}
