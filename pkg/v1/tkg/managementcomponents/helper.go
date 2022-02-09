// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package managementcomponents implements management component installation helpers
package managementcomponents

import (
	"time"

	"gopkg.in/yaml.v3"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/utils"
)

const (
	packagePollInterval = 5 * time.Second
	packagePollTimeout  = 10 * time.Minute
)

// GetTKGPackageConfigValuesFileFromUserConfig returns values file from user configuration
func GetTKGPackageConfigValuesFileFromUserConfig(userProviderConfigValues map[string]string) (string, error) {
	tkgPackageConfig := TKGPackageConfig{
		Metadata: Metadata{
			InfraProvider: userProviderConfigValues[constants.ConfigVariableProviderType],
		},
		ConfigValues: userProviderConfigValues,
	}

	configBytes, err := yaml.Marshal(tkgPackageConfig)
	if err != nil {
		return "", err
	}

	valuesFile, err := utils.CreateTempFile("", "")
	if err != nil {
		return "", err
	}

	err = utils.WriteToFile(valuesFile, configBytes)
	if err != nil {
		return "", err
	}
	return valuesFile, nil
}
