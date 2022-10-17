// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package managementcomponents implements management component installation helpers
package managementcomponents

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigbom"
	"github.com/vmware-tanzu/tanzu-framework/tkg/utils"
)

const (
	packagePollInterval = 5 * time.Second
	packagePollTimeout  = 10 * time.Minute
)

// GetTKGPackageConfigValuesFileFromUserConfig returns values file from user configuration
func GetTKGPackageConfigValuesFileFromUserConfig(managementPackageVersion string, userProviderConfigValues map[string]interface{}, tkgBomConfig *tkgconfigbom.BOMConfiguration) (string, error) {
	// TODO: Temporary hack(hard coded values) to configure TKR source controller package values. This should be replaced with the logic
	// that fetches these values from tkg-bom(for bom related urls) and set the TKR source controller package values
	var tkrRepoImagePath string
	providerType := userProviderConfigValues[constants.ConfigVariableProviderType]
	switch providerType {
	case constants.InfrastructureProviderVSphere:
		tkrRepoImagePath = fmt.Sprintf("%s/%s", tkgBomConfig.ImageConfig.ImageRepository, tkgBomConfig.TKRPackageRepo.VSphereNonparavirt)
	case constants.InfrastructureProviderAWS:
		tkrRepoImagePath = fmt.Sprintf("%s/%s", tkgBomConfig.ImageConfig.ImageRepository, tkgBomConfig.TKRPackageRepo.AWS)
	case constants.InfrastructureProviderAzure:
		tkrRepoImagePath = fmt.Sprintf("%s/%s", tkgBomConfig.ImageConfig.ImageRepository, tkgBomConfig.TKRPackageRepo.Azure)
	// Using vSphere's TKR components because there are no TKR components for CAPD yet
	// The issue https://github.com/vmware-tanzu/tanzu-framework/issues/3215 has been filed to add TKR components for CAPD
	case constants.InfrastructureProviderDocker:
		tkrRepoImagePath = fmt.Sprintf("%s/%s", tkgBomConfig.ImageConfig.ImageRepository, tkgBomConfig.TKRPackageRepo.VSphereNonparavirt)
	case constants.InfrastructureProviderOCI:
		// @TODO(Fang Han): remove this hardcoded tkr-oci after its downstream version is available
		tkrRepoImagePath = "gcr.io/tkg-on-oci/tkg/tkr/tkr-oci"
	default:
		return "", errors.Errorf("unknown provider type %q", providerType)
	}

	tkgPackageConfig := TKGPackageConfig{
		Metadata: Metadata{
			InfraProvider: userProviderConfigValues[constants.ConfigVariableProviderType].(string),
		},
		ConfigValues: userProviderConfigValues,
		FrameworkPackage: FrameworkPackage{
			VersionConstraints: managementPackageVersion,
			FeaturegatePackageValues: FeaturegatePackageValues{
				VersionConstraints: managementPackageVersion,
			},
			TKRServicePackageValues: TKRServicePackageValues{
				VersionConstraints: managementPackageVersion,
			},
			CLIPluginsPackageValues: CLIPluginsPackageValues{
				VersionConstraints: managementPackageVersion,
			},
			AddonsManagerPackageValues: AddonsManagerPackageValues{
				VersionConstraints: managementPackageVersion,
				TanzuAddonsManager: TanzuAddonsManager{
					FeatureGates: AddonsFeatureGates{
						ClusterBootstrapController: true,
						PackageInstallStatus:       true,
					},
				},
			},
			TanzuAuthPackageValues: TanzuAuthPackageValues{
				VersionConstraints: managementPackageVersion,
			},
		},
		ClusterClassPackage: ClusterClassPackage{
			VersionConstraints: managementPackageVersion,
			ClusterClassInfraPackageValues: ClusterClassInfraPackageValues{
				VersionConstraints: managementPackageVersion,
			},
		},
		TKRSourceControllerPackage: TKRSourceControllerPackage{
			VersionConstraints: managementPackageVersion,
			TKRSourceControllerPackageValues: TKRSourceControllerPackageValues{
				VersionConstraints:   managementPackageVersion,
				BomImagePath:         fmt.Sprintf("%s/%s", tkgBomConfig.ImageConfig.ImageRepository, tkgBomConfig.TKRBOM.ImagePath),
				BomMetadataImagePath: fmt.Sprintf("%s/%s", tkgBomConfig.ImageConfig.ImageRepository, tkgBomConfig.TKRCompatibility.ImagePath),
				TKRRepoImagePath:     tkrRepoImagePath,
				DefaultCompatibleTKR: tkgBomConfig.Default.TKRVersion,
			},
		},
		CoreManagementPluginsPackage: CoreManagementPluginsPackage{
			VersionConstraints: managementPackageVersion,
		},
	}

	setProxyConfiguration(&tkgPackageConfig, userProviderConfigValues)

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

func setProxyConfiguration(tkgPackageConfig *TKGPackageConfig, userProviderConfigValues map[string]interface{}) {
	var (
		httpProxy  string
		httpsProxy string
		noProxy    string
	)

	if p, ok := userProviderConfigValues[constants.TKGHTTPProxy]; ok {
		httpProxy = p.(string)
	}
	if p, ok := userProviderConfigValues[constants.TKGHTTPSProxy]; ok {
		httpsProxy = p.(string)
	}
	if p, ok := userProviderConfigValues[constants.TKGNoProxy]; ok {
		noProxy = p.(string)
	}

	setProxyInTKRSourceControllerPackage(tkgPackageConfig, httpProxy, httpsProxy, noProxy)
}

func setProxyInTKRSourceControllerPackage(tkgPackageConfig *TKGPackageConfig, httpProxy, httpsProxy, noProxy string) {
	tkgPackageConfig.TKRSourceControllerPackage.TKRSourceControllerPackageValues.Deployment =
		TKRSourceControllerPackageValuesDeployment{
			HttpProxy:  httpProxy,
			HttpsProxy: httpsProxy,
			NoProxy:    noProxy,
		}
}
