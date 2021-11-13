// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package clientconfighelpers implements helper function
// related to client configurations
package clientconfighelpers

import (
	"encoding/base64"
	"os"

	ctlimg "github.com/k14s/imgpkg/pkg/imgpkg/registry"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfigpaths"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfigreaderwriter"
)

// GetCustomRepositoryCaCertificateForClient returns CA certificate to use with cli client
// This function reads the CA certificate from following variables in decreasing order of precedence:
// 1. PROXY_CA_CERT
// 2. TKG_PROXY_CA_CERT
// 3. TKG_CUSTOM_IMAGE_REPOSITORY_CA_CERTIFICATE
func GetCustomRepositoryCaCertificateForClient(tkgconfigReaderWriter tkgconfigreaderwriter.TKGConfigReaderWriter) ([]byte, error) {
	caCert := ""
	var errProxyCACert, errTkgProxyCACertValue, errCustomImageRepoCACert error
	var proxyCACertValue, tkgProxyCACertValue, customImageRepoCACert string

	// Get the proxy configuration from tkgconfigreaderwriter if not nil
	// otherwise get the same proxy configuration from os environment variable
	if tkgconfigReaderWriter != nil {
		proxyCACertValue, errProxyCACert = tkgconfigReaderWriter.Get(constants.ProxyCACert)
		tkgProxyCACertValue, errTkgProxyCACertValue = tkgconfigReaderWriter.Get(constants.TKGProxyCACert)
		customImageRepoCACert, errCustomImageRepoCACert = tkgconfigReaderWriter.Get(constants.ConfigVariableCustomImageRepositoryCaCertificate)
	} else {
		proxyCACertValue = os.Getenv(constants.ProxyCACert)
		tkgProxyCACertValue = os.Getenv(constants.TKGProxyCACert)
		customImageRepoCACert = os.Getenv(constants.ConfigVariableCustomImageRepositoryCaCertificate)
	}

	if errProxyCACert == nil && proxyCACertValue != "" {
		caCert = proxyCACertValue
	} else if errTkgProxyCACertValue == nil && tkgProxyCACertValue != "" {
		caCert = tkgProxyCACertValue
	} else if errCustomImageRepoCACert == nil && customImageRepoCACert != "" {
		caCert = customImageRepoCACert
	} else {
		// return empty content when none is specified
		return []byte{}, nil
	}

	decoded, err := base64.StdEncoding.DecodeString(caCert)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode the base64-encoded custom registry CA certificate string")
	}
	return decoded, nil
}

// AddRegistryTrustedRootCertsFileForWindows adds CA certificate to registry options for windows environments
func AddRegistryTrustedRootCertsFileForWindows(registryOpts *ctlimg.Opts) error {
	filePath, err := tkgconfigpaths.GetRegistryTrustedCACertFileForWindows()
	if err != nil {
		return err
	}
	err = os.WriteFile(filePath, projectsRegistryCA, constants.ConfigFilePermissions)
	if err != nil {
		return errors.Wrapf(err, "failed to write the registry trusted CA cert to file '%s'", filePath)
	}
	registryOpts.CACertPaths = append(registryOpts.CACertPaths, filePath)
	return nil
}
