// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package exec

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/pkg/errors"
)

type Command struct {
	Cmd    string
	Args   []string
	Stdin  io.Reader
	Stdout io.Writer
	Env    []string
}

type Option func(*Command)

func NewCommand(opts ...Option) *Command {
	cmd := &Command{
		Stdin: nil,
	}
	for _, option := range opts {
		option(cmd)
	}
	return cmd
}

func WithCommand(command string) Option {
	return func(cmd *Command) {
		cmd.Cmd = command
	}
}

func WithArgs(args ...string) Option {
	return func(cmd *Command) {
		cmd.Args = args
	}
}

func WithStdin(stdin io.Reader) Option {
	return func(cmd *Command) {
		cmd.Stdin = stdin
	}
}

func WithStdout(stdout io.Writer) Option {
	return func(cmd *Command) {
		cmd.Stdout = stdout
	}
}

func WithEnv(args ...string) Option {
	return func(cmd *Command) {
		cmd.Args = args
	}
}

func (c *Command) Run(ctx context.Context) ([]byte, []byte, error) {
	cmd := exec.CommandContext(ctx, c.Cmd, c.Args...)
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
	output, err := ioutil.ReadAll(stdout)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	errout, err := ioutil.ReadAll(stderr)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	if err := cmd.Wait(); err != nil {
		return output, errout, errors.WithStack(err)
	}
	return output, errout, nil
}

func (c *Command) RunAndRedirectOutput(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, c.Cmd, c.Args...)

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
