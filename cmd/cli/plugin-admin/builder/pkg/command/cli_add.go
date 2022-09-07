// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package command provides handling to generate new scaffolding, compile, and
// publish CLI plugins.
package command

import (
	"errors"
	"os"

	"github.com/aunum/log"

	"github.com/vmware-tanzu/tanzu-framework/plugin-admin/builder/pkg/template"
)

// looksLikeARepo tries to verify we are running from the root of a repo or
// submodule. It looks for the presence of a go.mod and Makefile. If either is
// not found, returns an error to indicate it does not look like we are running
// from the repo root.
func looksLikeARepo() error {
	if _, err := os.Stat("go.mod"); errors.Is(err, os.ErrNotExist) {
		return errors.New("command should be run from the root of the repo, go.mod not found")
	}
	if _, err := os.Stat("Makefile"); errors.Is(err, os.ErrNotExist) {
		return errors.New("command should be run from the root of the repo, Makefile not found")
	}
	return nil
}

// AddPlugin generates the skeleton for a new plugin.
func AddPlugin(name, description string, dryRun bool) error {
	// Try to ensure we are in the root of the repo.
	if err := looksLikeARepo(); err != nil {
		return err
	}

	if description == "" {
		return errors.New("plugin description is required")
	}

	data := struct {
		PluginName  string
		Description string
	}{
		PluginName:  name,
		Description: description,
	}
	targets := template.DefaultPluginTargets
	for _, target := range targets {
		err := target.Run("", data, dryRun)
		if err != nil {
			return err
		}
	}
	log.Success("successfully created plugin")

	return nil
}
