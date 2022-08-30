// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/tj/assert"

	configv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigpaths"
)

func TestClientConfig(t *testing.T) {
	assert := assert.New(t)

	config.LocalDirName = fmt.Sprintf(".tanzu-test-%s", uuid.NewString()[:5])
	defer os.RemoveAll(config.LocalDirName)

	testConfig := &configv1alpha1.ClientConfig{
		ClientOptions: &configv1alpha1.ClientOptions{
			CLI: &configv1alpha1.CLIOptions{},
		},
	}

	// Test SetCompatibilityFileBasedOnEdition when the community edition is configure
	config.AcquireTanzuConfigLock()
	testConfig.ClientOptions.CLI.Edition = configv1alpha1.EditionCommunity
	err := config.StoreClientConfig(testConfig)
	assert.Nil(err)
	config.ReleaseTanzuConfigLock()

	err = SetCompatibilityFileBasedOnEdition()
	assert.Nil(err)
	// verify the result with GetClientConfig
	testConfig, err = config.GetClientConfig()
	assert.Nil(err)
	assert.Equal(DefaultTCEBomRepo, testConfig.ClientOptions.CLI.BOMRepo)
	assert.Equal(DefaultCompatibilityPath, testConfig.ClientOptions.CLI.CompatibilityFilePath)

	// Test SetCompatibilityFileBasedOnEdition when the standard edition is configure
	config.AcquireTanzuConfigLock()
	testConfig.ClientOptions.CLI.Edition = configv1alpha1.EditionStandard
	err = config.StoreClientConfig(testConfig)
	assert.Nil(err)
	config.ReleaseTanzuConfigLock()

	err = SetCompatibilityFileBasedOnEdition()
	assert.Nil(err)
	// verify the result with GetClientConfig
	testConfig, err = config.GetClientConfig()
	assert.Nil(err)
	assert.Equal(tkgconfigpaths.TKGDefaultImageRepo, testConfig.ClientOptions.CLI.BOMRepo)
	assert.Equal(tkgconfigpaths.TKGDefaultCompatibilityImagePath, testConfig.ClientOptions.CLI.CompatibilityFilePath)
}
