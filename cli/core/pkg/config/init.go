// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package config contains useful functionality for config updates
package config

import (
	"github.com/aunum/log"

	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
)

func init() {
	// Acquire tanzu config lock
	config.AcquireTanzuConfigLock()
	defer config.ReleaseTanzuConfigLock()

	c, err := config.GetClientConfigNoLock()
	if err != nil {
		log.Warningf("unable to get client config: %v", err)
	}

	addedDefaultDiscovery := populateDefaultStandaloneDiscovery(c)
	addedFeatureFlags := AddDefaultFeatureFlagsIfMissing(c, DefaultCliFeatureFlags)
	addedEdition := addDefaultEditionIfMissing(c)
	addedBomRepo := AddBomRepoIfMissing(c)
	addedCompatabilityFile := AddCompatibilityFileIfMissing(c)
	// contexts could be lost when older plugins edit the config, so populate them from servers
	addedContexts := config.PopulateContexts(c)

	if addedFeatureFlags || addedDefaultDiscovery || addedEdition || addedCompatabilityFile || addedBomRepo || addedContexts {
		_ = config.StoreClientConfig(c)
	}
}
