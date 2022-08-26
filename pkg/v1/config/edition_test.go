// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	"github.com/stretchr/testify/require"

	configv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
)

func TestConfigFeaturesDefaultEditionAdded(t *testing.T) {
	cfg := &configv1alpha1.ClientConfig{
		ClientOptions: &configv1alpha1.ClientOptions{
			CLI: &configv1alpha1.CLIOptions{
				Repositories:            DefaultRepositories,
				UnstableVersionSelector: DefaultVersionSelector,
			},
		},
	}

	added := addDefaultEditionIfMissing(cfg)
	require.True(t, added, "addDefaultEditionIfMissing should have returned true (having added missing default edition value)")
	errMsg := "addDefaultEditionIfMissing should have added default edition (" + configv1alpha1.EditionStandard + ") instead of " + cfg.ClientOptions.CLI.Edition
	require.Equal(t, cfg.ClientOptions.CLI.Edition, configv1alpha1.EditionSelector(configv1alpha1.EditionStandard), errMsg)
}

func TestConfigFeaturesDefaultEditionNotAdded(t *testing.T) {
	cfg := &configv1alpha1.ClientConfig{
		ClientOptions: &configv1alpha1.ClientOptions{
			CLI: &configv1alpha1.CLIOptions{
				Repositories:            DefaultRepositories,
				UnstableVersionSelector: DefaultVersionSelector,
				Edition:                 "tce",
			},
		},
	}

	added := addDefaultEditionIfMissing(cfg)
	require.False(t, added, "addDefaultEditionIfMissing should have returned false (without adding default edition value)")
	errMsg := "addDefaultEditionIfMissing should have left existing edition value intact instead of replacing with [" + cfg.ClientOptions.CLI.Edition + "]"
	require.Equal(t, cfg.ClientOptions.CLI.Edition, configv1alpha1.EditionSelector(configv1alpha1.EditionCommunity), errMsg)
}
