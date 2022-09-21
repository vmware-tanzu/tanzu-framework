// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package publish

import (
	"path/filepath"

	"github.com/otiai10/copy"
	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/common"
)

// LocalPublisher defines local publisher configuration
type LocalPublisher struct {
	LocalDistributionPath string
}

// NewLocalPublisher create new local publisher
func NewLocalPublisher(localDistributionPath string) (Publisher, error) {
	if localDistributionPath == "" {
		return nil, errors.New("local distribution path cannot be empty")
	}
	return &LocalPublisher{
		LocalDistributionPath: localDistributionPath,
	}, nil
}

// PublishPlugin publishes plugin binaries to local distribution directory
func (l *LocalPublisher) PublishPlugin(sourcePath, version, os, arch, plugin string) (string, error) {
	relativePath := filepath.Join(os, arch, "cli", plugin, version, "tanzu-"+plugin+"-"+os+"_"+arch)
	if os == osTypeWindows {
		relativePath += fileExtensionWindows
	}

	destPath := filepath.Join(l.LocalDistributionPath, relativePath)

	_ = ensureResourceDir(filepath.Dir(destPath), false)
	err := copy.Copy(sourcePath, destPath)
	if err != nil {
		return "", err
	}
	return relativePath, nil
}

// PublishDiscovery publishes the CLIPlugin resources YAML to a local discovery directory
func (l *LocalPublisher) PublishDiscovery() error {
	return nil
}

// Type returns type of publisher
func (l *LocalPublisher) Type() string {
	return common.DistributionTypeLocal
}
