// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package lint

import (
	_ "embed" // required to embed file
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
)

const (
	noLint   = "no-lint"
	lintName = "lint"
	help     = "help"
)

//go:embed cli-wordlist.yml
var wordList []byte

type tanzuTerms struct {
	Nouns       []string `yaml:"nouns"`
	Verbs       []string `yaml:"verbs"`
	GlobalFlags []string `yaml:"global-flags"`
	CmdFlags    []string `yaml:"command-flags"`
}

// TKGTerms analyzes plugin command nouns and verbs
type TKGTerms struct {
	cmd   *cobra.Command
	nouns []string
	verbs []string
}

// Init initializes TKGTerms using a config.
func (l *TKGTerms) Init(c *cobraLintConfig) {
	l.cmd = c.cmd.Parent()
	l.nouns = c.cliTerms.Nouns
	l.verbs = c.cliTerms.Verbs
}

// Execute runs the analysis and reports results.
func (l *TKGTerms) Execute() *Results {
	results := make(Results)

	if l.cmd != nil {
		if _, ok := l.cmd.Annotations[noLint]; ok {
			return nil
		}
	}
	// Top level plugin nouns
	if !contains(l.nouns, rawUse(l.cmd.Use)) {
		results[l.cmd.Use] = append(results[l.cmd.Use],
			fmt.Sprintf("unknown top-level term %s, expected standard noun",
				rawUse(l.cmd.Use)))
	}

	// Subcommands can be either a valid noun or verb, depending on the commands format
	if l.cmd.HasSubCommands() {
		for _, subCmd := range l.cmd.Commands() {
			if rawUse(subCmd.Use) == lintName || rawUse(subCmd.Use) == help {
				continue
			}
			if contains(l.nouns, rawUse(subCmd.Use)) || contains(l.verbs, rawUse(subCmd.Use)) {
				continue
			}
			results[l.cmd.Use] = append(results[l.cmd.Use],
				fmt.Sprintf("unknown subcommand term %s, expected standard term",
					rawUse(subCmd.Use)))
		}
	}
	return &results
}

// TKGFlags analyzes local and persistent commands.
type TKGFlags struct {
	cmd         *cobra.Command
	cmdFlags    []string
	globalFlags []string
	results     *Results
}

// Init initializes TKGFlags analyzer.
func (l *TKGFlags) Init(c *cobraLintConfig) {
	l.cmd = c.cmd.Parent()
	l.cmdFlags = c.cliTerms.CmdFlags
	l.globalFlags = c.cliTerms.GlobalFlags
	r := make(Results)
	l.results = &r
}

// Execute runs the analysis and reports results.
func (l *TKGFlags) Execute() *Results {
	l.lint(l.cmd)

	return l.results
}

// lint recursively process subcommand tree flags
func (l *TKGFlags) lint(cmd *cobra.Command) {
	if cmd.Use == lintName {
		return
	}

	cmd.Flags().VisitAll(func(f *flag.Flag) {
		if cmd.PersistentFlags().Lookup(f.Name) != nil {
			return
		}
		if _, ok := f.Annotations[noLint]; ok {
			return
		}
		if !contains(l.cmdFlags, f.Name) {
			(*l.results)[cmd.Use] = append((*l.results)[cmd.Use],
				fmt.Sprintf("unexpected flag %s, expected standard flag", f.Name))
		}
	})

	if cmd.HasPersistentFlags() {
		cmd.PersistentFlags().VisitAll(func(f *flag.Flag) {
			if !contains(l.globalFlags, f.Name) {
				(*l.results)[cmd.Use] = append((*l.results)[cmd.Use],
					fmt.Sprintf("unexpected global flag %s, expected standard flag", f.Name))
			}
		})
	}

	for _, subCmd := range cmd.Commands() {
		if subCmd.HasFlags() {
			subCmd.Flags().VisitAll(func(f *flag.Flag) {
				l.lint(subCmd)
			})
		}
		if subCmd.HasPersistentFlags() {
			subCmd.PersistentFlags().VisitAll(func(f *flag.Flag) {
				l.lint(subCmd)
			})
		}
	}
}

func loadPluginWords(cmd *cobra.Command) (*tanzuTerms, error) {
	t := new(tanzuTerms)
	err := yaml.Unmarshal(wordList, t)
	if err != nil {
		cmd.Printf("Unmarshal: %v", err)
		return nil, err
	}
	return t, nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func rawUse(s string) string {
	return strings.Fields(s)[0]
}
