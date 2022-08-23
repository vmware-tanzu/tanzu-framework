// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package component

import (
	"testing"

	"github.com/AlecAivazis/survey/v2"
	"github.com/stretchr/testify/assert"
)

func Test_translatePromptConfig_Sensitive(t *testing.T) {
	assert := assert.New(t)

	promptConfig := PromptConfig{
		Message:   "Pick a card, any card",
		Options:   []string{"one", "two", "three"},
		Default:   "one",
		Sensitive: true,
		Help:      "Help will be given to those who need it",
	}

	prompt := translatePromptConfig(&promptConfig)
	assert.NotNil(prompt)

	// Secure should return a password prompt
	_, ok := prompt.(*survey.Password)
	assert.True(ok)
}

func Test_translatePromptConfig_OptionsSelect(t *testing.T) {
	assert := assert.New(t)

	promptConfig := PromptConfig{
		Message:   "Pick a card, any card",
		Options:   []string{"one", "two", "three"},
		Default:   "one",
		Sensitive: false,
		Help:      "Help will be given to those who need it",
	}

	prompt := translatePromptConfig(&promptConfig)
	assert.NotNil(prompt)

	// Prompt with options should return a select prompt
	selectPrompt, ok := prompt.(*survey.Select)
	assert.True(ok)
	assert.Equal(len(promptConfig.Options), len(selectPrompt.Options))
}

func Test_translatePromptConfig_Input(t *testing.T) {
	assert := assert.New(t)

	promptConfig := PromptConfig{
		Message:   "Pick a card, any card",
		Default:   "one",
		Sensitive: false,
		Help:      "Help will be given to those who need it",
	}

	prompt := translatePromptConfig(&promptConfig)
	assert.NotNil(prompt)

	// Prompt without options should return an input prompt
	_, ok := prompt.(*survey.Input)
	assert.True(ok)
}
