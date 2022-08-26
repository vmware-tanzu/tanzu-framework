// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"fmt"
	"os"
	"strings"
	"text/template"
	"unicode"

	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"
)

// UsageFunc is the usage func for a plugin.
var UsageFunc = func(c *cobra.Command) error {
	t, err := template.New("usage").Funcs(TemplateFuncs).Parse(CmdTemplate)
	if err != nil {
		return err
	}
	return t.Execute(os.Stdout, c)
}

// CmdTemplate is the template for plugin commands.
const CmdTemplate = `{{ bold "Usage:" }}{{if .Runnable}}
  tanzu {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  tanzu {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

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

Use "{{if beginsWith .CommandPath "tanzu "}}{{.CommandPath}}{{else}}tanzu {{.CommandPath}}{{end}} [command] --help" for more information about a command.{{end}}
`

// TemplateFuncs are the template usage funcs.
var TemplateFuncs = template.FuncMap{
	"rpad":                    rpad,
	"bold":                    bold,
	"underline":               underline,
	"trimTrailingWhitespaces": trimRightSpace,
	"beginsWith":              beginsWith,
}

// rpad adds padding to the right of a string.
// from https://github.com/spf13/cobra/blob/993cc5372a05240dfd59e3ba952748b36b2cd117/cobra.go#L29
func rpad(s string, padding int) string {
	tmpl := fmt.Sprintf("%%-%ds", padding)
	return fmt.Sprintf(tmpl, s)
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

func beginsWith(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}
