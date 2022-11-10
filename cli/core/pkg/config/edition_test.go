// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	"github.com/stretchr/testify/require"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

func TestConfigFeaturesDefaultEditionAdded(t *testing.T) {
	cfg := &configapi.ClientConfig{
		ClientOptions: &configapi.ClientOptions{
			CLI: &configapi.CLIOptions{
				Repositories:            DefaultRepositories,
				UnstableVersionSelector: DefaultVersionSelector,
			},
		},
	}

	added := addDefaultEditionIfMissing(cfg)
	require.True(t, added, "addDefaultEditionIfMissing should have returned true (having added missing default edition value)")
	errMsg := "addDefaultEditionIfMissing should have added default edition (" + configapi.EditionStandard + ") instead of " + cfg.ClientOptions.CLI.Edition //nolint:staticcheck
	require.Equal(t, cfg.ClientOptions.CLI.Edition, configapi.EditionSelector(configapi.EditionStandard), errMsg)                                            //nolint:staticcheck
}

func TestConfigFeaturesDefaultEditionNotAdded(t *testing.T) {
	cfg := &configapi.ClientConfig{
		ClientOptions: &configapi.ClientOptions{
			CLI: &configapi.CLIOptions{
				Repositories:            DefaultRepositories,
				UnstableVersionSelector: DefaultVersionSelector,
				Edition:                 "tce",
			},
		},
	}

	added := addDefaultEditionIfMissing(cfg)
	require.False(t, added, "addDefaultEditionIfMissing should have returned false (without adding default edition value)")
	errMsg := "addDefaultEditionIfMissing should have left existing edition value intact instead of replacing with [" + cfg.ClientOptions.CLI.Edition + "]" //nolint:staticcheck
	require.Equal(t, cfg.ClientOptions.CLI.Edition, configapi.EditionSelector(configapi.EditionCommunity), errMsg)                                          //nolint:staticcheck
}
