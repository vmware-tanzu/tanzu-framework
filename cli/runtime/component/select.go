// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package component

import (
	"github.com/AlecAivazis/survey/v2"
)

// SelectConfig is the configuration for a selection.
type SelectConfig struct {
	// Message to display to user.
	Message string

	// Default option.
	Default interface{}

	// Options to select frorm.
	Options []string

	// Sensitive information.
	Sensitive bool

	// Help for the prompt.
	Help string

	// PageSize defines how many options per page.
	PageSize int
}

// Run the selection.
func (p *SelectConfig) Run(response interface{}) error {
	return Select(p, response)
}

// Select an option.
func Select(p *SelectConfig, response interface{}) error {
	prompt := translateSelectConfig(p)
	return survey.AskOne(prompt, response)
}

func translateSelectConfig(p *SelectConfig) survey.Prompt {
	return &survey.Select{
		Message:  p.Message,
		Options:  p.Options,
		Default:  p.Default,
		Help:     p.Help,
		PageSize: p.PageSize,
	}
}
