// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package publish

import (
	"path/filepath"

	"github.com/otiai10/copy"
)

// LocalPublisher defines local publisher configuration
type LocalPublisher struct {
	LocalDistributionPath string
}

// NewLocalPublisher create new local publisher
func NewLocalPublisher(localDistributionPath string) Publisher {
	return &LocalPublisher{
		LocalDistributionPath: localDistributionPath,
	}
}

// PublishPlugin publishes plugin binaries to local distribution directory
func (l *LocalPublisher) PublishPlugin(sourcePath, version, os, arch, plugin string) (string, error) {
	destPath := filepath.Join(l.LocalDistributionPath, os, arch, "cli", plugin, version, "tanzu-"+plugin+"-"+os+"_"+arch)
	if os == osTypeWindows {
		destPath += fileExtensionWindows
	}

	_ = ensureResourceDir(filepath.Dir(destPath), false)
	err := copy.Copy(sourcePath, destPath)
	if err != nil {
		return "", err
	}
	return destPath, nil
}

// PublishDiscovery publishes the CLIPlugin resources YAML to a local discovery directory
func (l *LocalPublisher) PublishDiscovery() error {
	return nil
}
