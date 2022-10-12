// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigpaths"
)

const (
	// DefaultTCEBomRepo is OCI repository containing the BOM for TCE
	DefaultTCEBomRepo = "projects.registry.vmware.com/tce"
	// DefaultCompatibilityPath the path (project) of the compatibility file
	DefaultCompatibilityPath = "tkg-compatibility"
)

// SetCompatibilityFileBasedOnEdition changes the compatibility file for the edition.
//
//nolint:staticcheck
func SetCompatibilityFileBasedOnEdition() error {
	// Acquire tanzu config lock
	config.AcquireTanzuConfigLock()
	defer config.ReleaseTanzuConfigLock()

	clientConfig, err := config.GetClientConfigNoLock()
	if err != nil {
		return err
	}

	if clientConfig.ClientOptions == nil || clientConfig.ClientOptions.CLI == nil {
		return nil
	}

	switch clientConfig.ClientOptions.CLI.Edition {
	case configapi.EditionCommunity:
		clientConfig.ClientOptions.CLI.BOMRepo = DefaultTCEBomRepo
		clientConfig.ClientOptions.CLI.CompatibilityFilePath = DefaultCompatibilityPath
	case configapi.EditionStandard:
		clientConfig.ClientOptions.CLI.BOMRepo = tkgconfigpaths.TKGDefaultImageRepo
		clientConfig.ClientOptions.CLI.CompatibilityFilePath = tkgconfigpaths.TKGDefaultCompatibilityImagePath
	}
	return config.StoreClientConfig(clientConfig)
}
