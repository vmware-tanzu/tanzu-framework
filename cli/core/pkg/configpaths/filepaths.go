// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package configpaths

import (
	"os"
	"path"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/constants"
)

// GetRegistryCertFile returns the registry cert file path
func GetRegistryCertFile() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "could not locate user home dir")
	}
	return path.Join(home, constants.TKGRegistryCertFile), nil
}

// GetRegistryTrustedCACertFileForWindows returns the registry trusted root ca cert filepath for windows
func GetRegistryTrustedCACertFileForWindows() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "could not locate user home dir")
	}
	return path.Join(home, constants.TKGRegistryTrustedRootCAFileForWindows), nil
}
