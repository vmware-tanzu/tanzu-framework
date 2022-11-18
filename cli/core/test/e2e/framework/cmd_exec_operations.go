// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	runtimeTest "github.com/vmware-tanzu/tanzu-framework/cli/runtime/test"
)

// CmdOps performs the Command line exec operations
type CmdOps interface {
	Exec(command string) (stdOut, stdErr *bytes.Buffer, err error)
	ExecContainsString(command, contains string) error
	ExecContainsAnyString(command string, contains []string) error
	ExecContainsErrorString(command, contains string) error
	ExecNotContainsStdErrorString(command, contains string) error
	ExecNotContainsString(command, contains string) error
}

// cmdOps is the implementation of CmdOps
type cmdOps struct {
	CmdOps
}

func NewCmdOps() CmdOps {
	return &cmdOps{}
}

// Exec the command, exit on error
func (co *cmdOps) Exec(command string) (stdOut, stdErr *bytes.Buffer, err error) {
	cmdInput := strings.Split(command, " ")
	cmdName := cmdInput[0]
	cmdArgs := cmdInput[1:]
	cmd := exec.Command(cmdName, cmdArgs...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return nil, nil, fmt.Errorf(fmt.Sprintf("error while running %s", command), err)
	}
	return &stdout, &stderr, nil
}

// ExecContainsString checks that the given command output contains the string.
func (co *cmdOps) ExecContainsString(command, contains string) error {
	stdOut, _, err := co.Exec(command)
	if err != nil {
		return err
	}
	return ContainsString(stdOut, contains)
}

// ExecContainsAnyString checks that the given command output contains any of the given set of strings.
func (co *cmdOps) ExecContainsAnyString(command string, contains []string) error {
	stdOut, _, err := co.Exec(command)
	if err != nil {
		return err
	}
	return ContainsAnyString(stdOut, contains)
}

// ExecContainsErrorString checks that the given command stdErr output contains the string
func (co *cmdOps) ExecContainsErrorString(command, contains string) error {
	_, stdErr, err := co.Exec(command)
	if err != nil {
		return err
	}
	return ContainsString(stdErr, contains)
}

// ExecNotContainsStdErrorString checks that the given command stdErr output contains the string
func (co *cmdOps) ExecNotContainsStdErrorString(command, contains string) error {
	_, stdErr, err := co.Exec(command)
	if err != nil && stdErr == nil {
		return err
	}
	return NotContainsString(stdErr, contains)
}

// NotContainsString checks that the given buffer not contains the string if contains then throws error.
func NotContainsString(stdOut *bytes.Buffer, contains string) error {
	so := stdOut.String()
	if strings.Contains(so, contains) {
		return fmt.Errorf("stdOut %q contains %q", so, contains)
	}
	return nil
}

// ContainsString checks that the given buffer contains the string.
func ContainsString(stdOut *bytes.Buffer, contains string) error {
	so := stdOut.String()
	if !strings.Contains(so, contains) {
		return fmt.Errorf("stdOut %q did not contain %q", so, contains)
	}
	return nil
}

// ContainsAnyString checks that the given buffer contains any of the given set of strings.
func ContainsAnyString(stdOut *bytes.Buffer, contains []string) error {
	var containsAny bool
	so := stdOut.String()

	for _, str := range contains {
		containsAny = containsAny || strings.Contains(so, str)
	}

	if !containsAny {
		return fmt.Errorf("stdOut %q did not contain of the following %q", so, contains)
	}
	return nil
}

// ExecNotContainsString checks that the given command output not contains the string.
func (co *cmdOps) ExecNotContainsString(command, contains string) error {
	stdOut, _, err := runtimeTest.Exec(command)
	if err != nil {
		return err
	}
	return co.NotContainsString(stdOut, contains)
}

// NotContainsString checks that the given buffer not contains the string.
func (co *cmdOps) NotContainsString(stdOut *bytes.Buffer, contains string) error {
	so := stdOut.String()
	if strings.Contains(so, contains) {
		return fmt.Errorf("stdOut %q does contain %q", so, contains)
	}
	return nil
}
