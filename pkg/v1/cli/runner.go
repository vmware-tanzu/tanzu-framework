// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/aunum/log"
)

// Runner is a plugin runner.
type Runner struct {
	name       string
	args       []string
	pluginRoot string
}

// NewRunner creates an instance of Runner.
func NewRunner(name string, args []string, options ...Option) *Runner {
	opts := makeDefaultOptions(options...)

	r := &Runner{
		name:       name,
		args:       args,
		pluginRoot: opts.pluginRoot,
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
	stdout, stderr, err := r.run(ctx, pluginPath)

	// We want to write any available output, regardless of whether an error occurred.
	if stdout != "" {
		os.Stdout.WriteString(stdout)
	}
	if stderr != "" {
		os.Stderr.WriteString(stderr)
	}

	if err != nil {
		return err
	}
	return nil
}

// RunOutput runs a plugin and returns the output.
func (r *Runner) RunOutput(ctx context.Context) (string, string, error) {
	return r.run(ctx, r.pluginPath())
}

func (r *Runner) run(ctx context.Context, pluginPath string) (string, string, error) {
	if BuildArch().IsWindows() {
		pluginPath += ".exe"
	}

	info, err := os.Stat(pluginPath)
	if err != nil {
		// TODO (pbarker): should check if the plugin exists in the repository using fuzzy search and display how to install.
		return "", "", fmt.Errorf("plugin %q does not exist, try using `tanzu plugin install %s` to install or `tanzu plugin list` to find plugins", r.name, r.name)
	}

	if info.IsDir() {
		return "", "", fmt.Errorf("%q is a directory", pluginPath)
	}

	stateFile, err := ioutil.TempFile("", "tanzu-cli-state")
	if err != nil {
		return "", "", fmt.Errorf("create state file: %w", err)
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
		return "", "", fmt.Errorf("encode state: %w", err)
	}

	if err := stateFile.Close(); err != nil {
		return "", "", fmt.Errorf("close state file: %w", err)
	}

	env := append(os.Environ(), fmt.Sprintf("%s=%s", EnvPluginStateKey, stateFile.Name()))

	log.Debugf("running command path %s args: %+v", pluginPath, r.args)
	cmd := exec.CommandContext(ctx, pluginPath, r.args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdin = os.Stdin
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	cmd.Env = env

	err = cmd.Run()
	return stdout.String(), stderr.String(), err
}

func (r *Runner) pluginName() string {
	return BinFromPluginName(r.name)
}

func (r *Runner) pluginPath() string {
	return filepath.Join(r.pluginRoot, r.pluginName())
}

func (r *Runner) testPluginPath() string {
	return filepath.Join(r.pluginRoot, "test", BinTestFromPluginName(r.name))
}
