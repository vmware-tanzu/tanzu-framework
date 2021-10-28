// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/aunum/log"
)

// Runner is a plugin runner.
type Runner struct {
	name          string
	args          []string
	pluginRoot    string
	pluginAbsPath string
}

// NewRunner creates an instance of Runner.
func NewRunner(name, pluginAbsPath string, args []string, options ...Option) *Runner {
	opts := makeDefaultOptions(options...)

	r := &Runner{
		name:          name,
		args:          args,
		pluginRoot:    opts.pluginRoot,
		pluginAbsPath: pluginAbsPath,
	}
	return r
}

// Run runs a plugin.
func (r *Runner) Run(ctx context.Context) error {
	return r.runStdOutput(ctx, r.pluginPath())
}

// RunTest runs a plugin test.
func (r *Runner) RunTest(ctx context.Context) error {
	return r.runStdOutput(ctx, r.testPluginPath())
}

// runStdOutput runs a plugin and writes any output to the standard os.Stdout and os.Stderr.
func (r *Runner) runStdOutput(ctx context.Context, pluginPath string) error {
	err := r.run(ctx, pluginPath, nil, nil)
	if err != nil {
		return err
	}
	return nil
}

// RunOutput runs a plugin and returns the output.
func (r *Runner) RunOutput(ctx context.Context) (string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := r.run(ctx, r.pluginPath(), &stdout, &stderr)
	return stdout.String(), stderr.String(), err
}

// run executes a command at pluginPath. If stdout and stderr are nil, any ouptut from command
// execution is emitted to os.Stdout and os.Stderr respectively. Otherwise any command output
// is captured in the bytes.Buffer.
func (r *Runner) run(ctx context.Context, pluginPath string, stdout, stderr *bytes.Buffer) error {
	if BuildArch().IsWindows() {
		pluginPath += ".exe"
	}

	info, err := os.Stat(pluginPath)
	if err != nil {
		// TODO (pbarker): should check if the plugin exists in the repository using fuzzy search and display how to install.
		return fmt.Errorf("plugin %q does not exist, try using `tanzu plugin install %s` to install or `tanzu plugin list` to find plugins", r.name, r.name)
	}

	if info.IsDir() {
		return fmt.Errorf("%q is a directory", pluginPath)
	}

	stateFile, err := os.CreateTemp("", "tanzu-cli-state")
	if err != nil {
		return fmt.Errorf("create state file: %w", err)
	}

	defer func() {
		if rErr := os.Remove(stateFile.Name()); rErr != nil {
			log.Errorf("unable to remove plugin state file: %v", err)
		}
	}()

	state := &PluginState{
		Auth: "auth",
	}

	if err := json.NewEncoder(stateFile).Encode(state); err != nil {
		return fmt.Errorf("encode state: %w", err)
	}

	if err := stateFile.Close(); err != nil {
		return fmt.Errorf("close state file: %w", err)
	}

	env := append(os.Environ(), fmt.Sprintf("%s=%s", EnvPluginStateKey, stateFile.Name()))

	log.Debugf("running command path %s args: %+v", pluginPath, r.args)
	cmd := exec.CommandContext(ctx, pluginPath, r.args...)

	cmd.Env = env
	cmd.Stdin = os.Stdin
	// Check if the execution output should be captured
	if stderr != nil {
		cmd.Stderr = stderr
	} else {
		cmd.Stderr = os.Stderr
	}
	if stdout != nil {
		cmd.Stdout = stdout
	} else {
		cmd.Stdout = os.Stdout
	}

	err = cmd.Run()
	return err
}

func (r *Runner) pluginName() string {
	return BinFromPluginName(r.name)
}

func (r *Runner) pluginPath() string {
	if r.pluginAbsPath != "" {
		return r.pluginAbsPath
	}
	return filepath.Join(r.pluginRoot, r.pluginName())
}

func (r *Runner) testPluginPath() string {
	return filepath.Join(r.pluginRoot, "test", BinTestFromPluginName(r.name))
}
