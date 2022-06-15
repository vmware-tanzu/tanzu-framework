// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package exec implements command execution functionality.
package exec

import (
	"context"
	"io"
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

// Command wraps exec.Command
type Command struct {
	Cmd    string
	Args   []string
	Stdin  io.Reader
	Stdout io.Writer
	Env    []string
}

// Option is a functional option type that modifies a Command.
type Option func(*Command)

// NewCommand returns a Command.
func NewCommand(opts ...Option) *Command {
	cmd := &Command{
		Stdin: nil,
	}
	for _, option := range opts {
		option(cmd)
	}
	return cmd
}

// WithCommand defines the command
func WithCommand(command string) Option {
	return func(cmd *Command) {
		cmd.Cmd = command
	}
}

// WithArgs sets the arguments for the command
func WithArgs(args ...string) Option {
	return func(cmd *Command) {
		cmd.Args = args
	}
}

// WithStdin sets up the command to read from this io.Reader.
func WithStdin(stdin io.Reader) Option {
	return func(cmd *Command) {
		cmd.Stdin = stdin
	}
}

// WithStdout sets up the command to write to io.Writer.
func WithStdout(stdout io.Writer) Option {
	return func(cmd *Command) {
		cmd.Stdout = stdout
	}
}

// WithEnv sets env variables
func WithEnv(args ...string) Option {
	return func(cmd *Command) {
		cmd.Args = args
	}
}

// Run executes the command and returns stdout, stderr and the error if there is any.
func (c *Command) Run(ctx context.Context) ([]byte, []byte, error) {
	cmd := exec.CommandContext(ctx, c.Cmd, c.Args...) // nolint:gosec
	if len(c.Env) != 0 {
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, c.Env...)
	}

	if c.Stdin != nil {
		cmd.Stdin = c.Stdin
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	if err := cmd.Start(); err != nil {
		return nil, nil, errors.WithStack(err)
	}
	output, err := io.ReadAll(stdout)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	errout, err := io.ReadAll(stderr)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	if err := cmd.Wait(); err != nil {
		return output, errout, errors.WithStack(err)
	}
	return output, errout, nil
}

// RunAndRedirectOutput executes command and redirects output
func (c *Command) RunAndRedirectOutput(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, c.Cmd, c.Args...) // nolint:gosec

	if len(c.Env) != 0 {
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, c.Env...)
	}

	if c.Stdin != nil {
		cmd.Stdin = c.Stdin
	}

	cmd.Stdout = c.Stdout
	cmd.Stderr = c.Stdout

	err := cmd.Run()
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
