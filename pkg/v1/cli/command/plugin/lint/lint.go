// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package lint

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli/component"
)

var cobraLints = []cobraLint{
	&TKGTerms{},
	&TKGFlags{},
}

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

type Results map[string][]string

type cobraLint interface {
	Init(*cobraLintConfig)
	Execute() *Results
}

type CobraLintRunner struct {
	results *Results
	config  *cobraLintConfig
}

type cobraLintConfig struct {
	cliTerms *tanzuTerms
	cmd      *cobra.Command
}

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

func (c *CobraLintRunner) Output() {
	t := component.NewTableWriter()
	t.SetHeader([]string{"Command", "Lint"})

	for k, vs := range *c.results {
		for _, v := range vs {
			t.Append([]string{k, v})
		}
	}
	t.Render()
	fmt.Println("---")
}
