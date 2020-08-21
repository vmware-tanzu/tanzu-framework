package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"
	"unicode"

	"github.com/spf13/cobra"

	"github.com/logrusorgru/aurora"
)

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

	d := struct {
		*cobra.Command
		CmdMap CmdMap
	}{
		c,
		cmdMap,
	}

	t := template.Must(template.New("usage").Funcs(TemplateFuncs).Parse(u.Template()))
	err := t.Execute(w, d)
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
{{.LocalFlags.FlagUsages}}

Use "{{.CommandPath}} [command] --help" for more information about a command.
`
}

// TemplateFuncs are the template usage funcs.
var TemplateFuncs = template.FuncMap{
	"rpad":                    rpad,
	"bold":                    bold,
	"trimTrailingWhitespaces": trimRightSpace,
}

// rpad adds padding to the right of a string.
// from https://github.com/spf13/cobra/blob/993cc5372a05240dfd59e3ba952748b36b2cd117/cobra.go#L29
func rpad(s string, padding int) string {
	template := fmt.Sprintf("%%-%ds", padding)
	return fmt.Sprintf(template, s)
}

func bold(s string) string {
	return aurora.Bold(s).String()
}

func trimRightSpace(s string) string {
	return strings.TrimRightFunc(s, unicode.IsSpace)
}
