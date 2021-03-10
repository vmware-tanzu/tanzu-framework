// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"fmt"

	"github.com/pkg/errors"
)

var (
	okResponsesMap = map[string]struct{}{
		"y": {},
		"Y": {},
	}
)

// AskForConfirmation is used to prompt the user to confirm or deny a choice
func AskForConfirmation(message string) error {
	var response string
	msg := message + " [y/N]: "
	fmt.Print(msg)
	_, err := fmt.Scanln(&response)
	if err != nil {
		return errors.New("aborted")
	}
	if _, exit := okResponsesMap[response]; !exit {
		return errors.New("aborted")
	}
	return nil
}
