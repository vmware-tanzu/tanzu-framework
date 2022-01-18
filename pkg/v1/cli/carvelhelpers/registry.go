// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package carvelhelpers

import (
	"os"
	"runtime"
	"strings"

	ctlimg "github.com/k14s/imgpkg/pkg/imgpkg/registry"
	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/clientconfighelpers"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfigpaths"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/registry"
)

// GetFilesMapFromImage returns map of files metadata
// It takes os environment variables for custom repository and proxy
// configuration into account while downloading image from repository
func GetFilesMapFromImage(imageWithTag string) (map[string][]byte, error) {
	reg, err := newRegistry()
	if err != nil {
		return nil, errors.Wrapf(err, "unable to initialize registry")
	}
	return reg.GetFiles(imageWithTag)
}

// DownloadImageBundleAndSaveFilesToTempDir reads OCI image and saves file to temp dir
// returns temp configuration dir with downloaded imgpkg bundle
func DownloadImageBundleAndSaveFilesToTempDir(imageWithTag string) (string, error) {
	reg, err := newRegistry()
	if err != nil {
		return "", errors.Wrapf(err, "unable to initialize registry")
	}
	tmpDir, err := os.MkdirTemp("", "oci_image")
	if err != nil {
		return "", errors.Wrap(err, "error creating temporary directory")
	}
	err = reg.DownloadBundle(imageWithTag, tmpDir)
	return tmpDir, err
}

// newRegistry returns a new registry object by also
// taking into account for any custom registry or proxy
// environment variable provided by the user
func newRegistry() (registry.Registry, error) {
	verifyCerts := true
	skipVerifyCerts := os.Getenv(constants.ConfigVariableCustomImageRepositorySkipTLSVerify)
	if strings.EqualFold(skipVerifyCerts, "true") {
		verifyCerts = false
	}

	registryOpts := &ctlimg.Opts{
		VerifyCerts: verifyCerts,
		Anon:        true,
	}

	if runtime.GOOS == "windows" {
		err := clientconfighelpers.AddRegistryTrustedRootCertsFileForWindows(registryOpts)
		if err != nil {
			return nil, err
		}
	}

	caCertBytes, err := clientconfighelpers.GetCustomRepositoryCaCertificateForClient(nil)
	if err == nil && len(caCertBytes) != 0 {
		filePath, err := tkgconfigpaths.GetRegistryCertFile()
		if err != nil {
			return nil, err
		}
		err = os.WriteFile(filePath, caCertBytes, 0o644)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to write the custom image registry CA cert to file '%s'", filePath)
		}
		registryOpts.CACertPaths = append(registryOpts.CACertPaths, filePath)
	}

	return registry.New(registryOpts)
}
