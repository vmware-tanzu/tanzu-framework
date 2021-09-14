// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"golang.org/x/mod/semver"

	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo"
)

// CoreName is the name of the core binary.
const CoreName = "core"

const coreDescription = "The core Tanzu CLI"

// CoreDescriptor is the core descriptor.
var CoreDescriptor = cliv1alpha1.PluginDescriptor{
	Name:        CoreName,
	Description: coreDescription,
	Version:     buildinfo.Version,
	BuildSHA:    buildinfo.SHA,
}

// CorePlugin is the core plugin.
var CorePlugin = Plugin{
	Name:        CoreName,
	Description: coreDescription,
}

// HasUpdate tells whether the core plugin has an update.
func HasUpdate(repo Repository) (update bool, version string, err error) {
	plugin, err := repo.Describe(CoreName)
	if err != nil {
		return false, version, err
	}
	versionSelector := repo.VersionSelector()

	version = plugin.FindVersion(versionSelector)
	compared := semver.Compare(version, buildinfo.Version)
	if compared == 1 {
		return true, version, nil
	}
	return false, version, nil
}

// Update the core CLI.
func Update(repo Repository) error {
	var executable string
	update, version, err := HasUpdate(repo)
	if err != nil {
		return err
	}
	if !update {
		return nil
	}
	b, err := repo.Fetch(CoreName, version, BuildArch())
	if err != nil {
		return err
	}

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return errors.Wrap(err, "could not locate filepath of running binary")
	}

	newCliFile := filepath.Join(dir, "tanzu_new")
	outFile, err := os.OpenFile(newCliFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return errors.Wrap(err, "could not create new binary file")
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, bytes.NewReader(b))
	if err != nil {
		return errors.Wrap(err, "could not copy new binary file")
	}
	outFile.Close()

	if BuildArch().IsWindows() {
		executable = outFile.Name() + ".exe"
		err = os.Rename(outFile.Name(), executable)
		if err != nil {
			return errors.Wrap(err, "could not rename binary file")
		}
	} else {
		executable, err = os.Executable()
		if err != nil {
			return errors.Wrap(err, "could not locate current executable")
		}
		err = os.Rename(outFile.Name(), executable)
		if err != nil {
			return errors.Wrap(err, "could not rename binary file")
		}
	}
	return nil
}
