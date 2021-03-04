// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
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
	pluginPath := r.pluginPath()

	if BuildArch().IsWindows() {
		pluginPath = pluginPath + ".exe"
	}
	info, err := os.Stat(pluginPath)
	if err != nil {
		// TODO (pbarker): should check if the plugin exists in the repository using fuzzy search and display how to install.
		return fmt.Errorf("plugin %q does not exist, try using `tanzu plugin install %s` to install or `tanzu plugin list` to find plugins", r.name, r.name)
	}

	if info.IsDir() {
		return fmt.Errorf("%q is a directory", pluginPath)
	}
	err = r.run(ctx, pluginPath)
	if err != nil {
		return err
	}
	return nil
}

// RunTest runs a plugin test.
func (r *Runner) RunTest(ctx context.Context) error {
	pluginPath := r.testPluginPath()
	info, err := os.Stat(pluginPath)
	if err != nil {
		// TODO (pbarker): should check if the plugin exists in the repository using fuzzy search and display how to install.
		return fmt.Errorf("plugin %q does not exist, try using `tanzu plugin install %s` to install or `tanzu plugin list` to find plugins", r.name, r.name)
	}

	if info.IsDir() {
		return fmt.Errorf("%q is a directory", pluginPath)
	}
	err = r.run(ctx, pluginPath)
	if err != nil {
		return err
	}
	return nil
}

func (r *Runner) run(ctx context.Context, pluginPath string) error {
	stateFile, err := ioutil.TempFile("", "tanzu-cli-state")
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
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Env = env

	return cmd.Run()
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
