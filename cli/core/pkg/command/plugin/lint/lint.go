// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package lint

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/component"
)

var cobraLints = []cobraLint{
	&TKGTerms{},
	&TKGFlags{},
}

// NewCobraLinter returns an instance of CobraLintRunner.
func NewCobraLinter(cmd *cobra.Command) (*CobraLintRunner, error) {
	terms, err := loadPluginWords(cmd)
	if err != nil {
		return nil, err
	}

	r := make(Results)

	cfg := &cobraLintConfig{
		cliTerms: terms,
		cmd:      cmd,
	}
	return &CobraLintRunner{
		results: &r,
		config:  cfg,
	}, nil
}

// Results is a map of commands and lint errors associated with them.
type Results map[string][]string

type cobraLint interface {
	Init(*cobraLintConfig)
	Execute() *Results
}

// CobraLintRunner lints cobra commands and reports results.
type CobraLintRunner struct {
	results *Results
	config  *cobraLintConfig
}

type cobraLintConfig struct {
	cliTerms *tanzuTerms
	cmd      *cobra.Command
}

// Run runs the linter and reports success or failure.
func (c *CobraLintRunner) Run() bool {
	success := true
	for _, l := range cobraLints {
		l.Init(c.config)
		for key, value := range *l.Execute() {
			if success {
				success = false
			}
			(*c.results)[key] = append((*c.results)[key], value...)
		}
	}
	return success
}

// Output writes the results of linting in a table form.
func (c *CobraLintRunner) Output() {
	t := component.NewOutputWriter(c.config.cmd.OutOrStdout(), "table", "command", "lint")
	for k, vs := range *c.results {
		for _, v := range vs {
			t.AddRow(k, v)
		}
	}
	t.Render()
	fmt.Println("---")
}
