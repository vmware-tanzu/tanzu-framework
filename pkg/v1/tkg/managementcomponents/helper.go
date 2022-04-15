// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package managementcomponents implements management component installation helpers
package managementcomponents

import (
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/utils"
)

const (
	packagePollInterval = 5 * time.Second
	packagePollTimeout  = 10 * time.Minute
)

// TKGPackageConfigurationOptions defines configuration needed to define TKG package values
type TKGPackageConfigurationOptions struct {
	ManagementPackageRepoImage string
	ManagementPackageVersion   string
	UserProviderConfigValues   map[string]string
	TKRPackageRepository       TKRPackageRepository
}

// GetTKGPackageConfigValuesFile returns values file from user configuration
func GetTKGPackageConfigValuesFile(options TKGPackageConfigurationOptions) (string, error) {
	tkgPackageConfig := TKGPackageConfig{
		Metadata: Metadata{
			InfraProvider: options.UserProviderConfigValues[constants.ConfigVariableProviderType],
		},
		ConfigValues:         options.UserProviderConfigValues,
		TKRPackageRepository: options.TKRPackageRepository,
		FrameworkPackage: FrameworkPackage{
			VersionConstraints: options.ManagementPackageVersion,
			FeaturegatePackageValues: FeaturegatePackageValues{
				VersionConstraints: options.ManagementPackageVersion,
			},
			TKRServicePackageValues: TKRServicePackageValues{
				VersionConstraints: options.ManagementPackageVersion,
			},
			CLIPluginsPackageValues: CLIPluginsPackageValues{
				VersionConstraints: options.ManagementPackageVersion,
			},
			AddonsManagerPackageValues: AddonsManagerPackageValues{
				VersionConstraints: options.ManagementPackageVersion,
			},
		},
		ClusterClassPackage: ClusterClassPackage{
			VersionConstraints: options.ManagementPackageVersion,
			ClusterClassInfraPackageValues: ClusterClassInfraPackageValues{
				VersionConstraints: options.ManagementPackageVersion,
			},
		},
	}

	configBytes, err := yaml.Marshal(tkgPackageConfig)
	if err != nil {
		return "", err
	}

	valuesFile := filepath.Join(os.TempDir(), constants.TKGPackageValuesFile)
	err = utils.SaveFile(valuesFile, configBytes)
	if err != nil {
		return "", err
	}
	return valuesFile, nil
}
