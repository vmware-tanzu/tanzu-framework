// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package component

import (
	"github.com/AlecAivazis/survey/v2"
)

// QuestionConfig stores config for prompting a CLI question.
type QuestionConfig struct {
	Message string
}

// Run asks a question.
func (q *QuestionConfig) Run(response interface{}) error {
	return Ask(q, response)
}

// Ask asks a questions and lets the user select an option.
func Ask(q *QuestionConfig, response interface{}) error {
	return survey.AskOne(&survey.Input{Message: q.Message}, response)
}
