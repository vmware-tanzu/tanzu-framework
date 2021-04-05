// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package component

import (
	"github.com/AlecAivazis/survey/v2"
)

type QuestionConfig struct {
	Message string
}

func (q *QuestionConfig) Run(response interface{}) error {
	return Ask(q, response)
}

// Select an option.
func Ask(q *QuestionConfig, response interface{}) error {
	return survey.AskOne(&survey.Input{Message: q.Message}, response)
}
