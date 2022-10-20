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
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigreaderwriter"
	"github.com/vmware-tanzu/tanzu-framework/tkg/utils"
)

const (
	packagePollInterval = 5 * time.Second
	packagePollTimeout  = 10 * time.Minute
)

// GetTKGPackageConfigValuesFileFromUserConfig returns values file from user configuration
func GetTKGPackageConfigValuesFileFromUserConfig(managementPackageVersion, addonsManagerPackageVersion string, userProviderConfigValues map[string]interface{}, tkgBomConfig *tkgconfigbom.BOMConfiguration, readerWriter tkgconfigreaderwriter.TKGConfigReaderWriter) (string, error) {
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
	default:
		return "", errors.Errorf("unknown provider type %q", providerType)
	}

	// get cacert value from user input for tkr-controller-config cm
	caCerts, imageRepo := getCaCertAndImageRepoFromUserProviderConfigValues(userProviderConfigValues, tkgBomConfig, readerWriter)

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
				VersionConstraints: addonsManagerPackageVersion,
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
				CaCerts:              caCerts,
				ImageRepo:            imageRepo,
			},
		},
		CoreManagementPluginsPackage: CoreManagementPluginsPackage{
			VersionConstraints: managementPackageVersion,
		},
		AkoOperatorPackage: AkoOperatorPackage{
			AkoOperatorPackageValues: AkoOperatorPackageValues{
				AviEnable:   convertToBool(userProviderConfigValues[constants.ConfigVariableAviEnable]),
				ClusterName: convertToString(userProviderConfigValues[constants.ConfigVariableClusterName]),
				AkoOperatorConfig: AkoOperatorConfig{
					AviControllerAddress:                           convertToString(userProviderConfigValues[constants.ConfigVariableAviControllerAddress]),
					AviControllerUsername:                          convertToString(userProviderConfigValues[constants.ConfigVariableAviControllerUsername]),
					AviControllerPassword:                          convertToString(userProviderConfigValues[constants.ConfigVariableAviControllerPassword]),
					AviControllerCA:                                convertToString(userProviderConfigValues[constants.ConfigVariableAviControllerCA]),
					AviCloudName:                                   convertToString(userProviderConfigValues[constants.ConfigVariableAviCloudName]),
					AviServiceEngineGroup:                          convertToString(userProviderConfigValues[constants.ConfigVariableAviServiceEngineGroup]),
					AviManagementClusterServiceEngineGroup:         convertToString(userProviderConfigValues[constants.ConfigVariableAviManagementClusterServiceEngineGroup]),
					AviDataPlaneNetworkName:                        convertToString(userProviderConfigValues[constants.ConfigVariableAviDataPlaneNetworkName]),
					AviDataPlaneNetworkCIDR:                        convertToString(userProviderConfigValues[constants.ConfigVariableAviDataPlaneNetworkCIDR]),
					AviControlPlaneNetworkName:                     convertToString(userProviderConfigValues[constants.ConfigVariableAviControlPlaneNetworkName]),
					AviControlPlaneNetworkCIDR:                     convertToString(userProviderConfigValues[constants.ConfigVariableAviControlPlaneNetworkCIDR]),
					AviManagementClusterDataPlaneNetworkName:       convertToString(userProviderConfigValues[constants.ConfigVariableAviManagementClusterDataPlaneNetworkName]),
					AviManagementClusterDataPlaneNetworkCIDR:       convertToString(userProviderConfigValues[constants.ConfigVariableAviManagementClusterDataPlaneNetworkCIDR]),
					AviManagementClusterControlPlaneVipNetworkName: convertToString(userProviderConfigValues[constants.ConfigVariableAviManagementClusterControlPlaneVipNetworkName]),
					AviManagementClusterControlPlaneVipNetworkCIDR: convertToString(userProviderConfigValues[constants.ConfigVariableAviManagementClusterControlPlaneVipNetworkCIDR]),
					AviControlPlaneHaProvider:                      convertToBool(userProviderConfigValues[constants.ConfigVariableVsphereHaProvider]),
				},
			},
		},
	}

	//Auto fill empty fields in AkoOperatorConfig
	autofillAkoOperatorConfig(&tkgPackageConfig.AkoOperatorPackage.AkoOperatorPackageValues.AkoOperatorConfig)
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

//convertToString converts config into string type, and return "" if config is not set
func convertToString(config interface{}) string {
	switch config.(type) {
	case string:
		return config.(string)
	default:
		return ""
	}
}

//convertToBool converts config into bool type, and return false if config is not set
func convertToBool(config interface{}) bool {
	switch config.(type) {
	case bool:
		return config.(bool)
	default:
		return false
	}
}

//autofillAkoOperatorConfig autofills empty fields in AkoOperatorConfig
func autofillAkoOperatorConfig(akoOperatorConfig *AkoOperatorConfig) {
	if akoOperatorConfig.AviManagementClusterServiceEngineGroup == "" {
		akoOperatorConfig.AviManagementClusterServiceEngineGroup = akoOperatorConfig.AviServiceEngineGroup
	}

	if akoOperatorConfig.AviManagementClusterDataPlaneNetworkName != "" && akoOperatorConfig.AviManagementClusterDataPlaneNetworkCIDR != "" {
		akoOperatorConfig.AviControlPlaneNetworkName = akoOperatorConfig.AviManagementClusterDataPlaneNetworkName
		akoOperatorConfig.AviControlPlaneNetworkCIDR = akoOperatorConfig.AviManagementClusterDataPlaneNetworkCIDR
	} else if akoOperatorConfig.AviControlPlaneNetworkName == "" || akoOperatorConfig.AviControlPlaneNetworkCIDR == "" {
		akoOperatorConfig.AviControlPlaneNetworkName = akoOperatorConfig.AviDataPlaneNetworkName
		akoOperatorConfig.AviControlPlaneNetworkCIDR = akoOperatorConfig.AviDataPlaneNetworkCIDR
	}

	if akoOperatorConfig.AviManagementClusterDataPlaneNetworkName != "" && akoOperatorConfig.AviManagementClusterDataPlaneNetworkCIDR != "" {
		akoOperatorConfig.AviManagementClusterControlPlaneVipNetworkName = akoOperatorConfig.AviManagementClusterDataPlaneNetworkName
		akoOperatorConfig.AviManagementClusterControlPlaneVipNetworkCIDR = akoOperatorConfig.AviManagementClusterDataPlaneNetworkCIDR
	} else if akoOperatorConfig.AviManagementClusterControlPlaneVipNetworkName == "" || akoOperatorConfig.AviManagementClusterControlPlaneVipNetworkCIDR == "" {
		akoOperatorConfig.AviManagementClusterControlPlaneVipNetworkName = akoOperatorConfig.AviDataPlaneNetworkName
		akoOperatorConfig.AviManagementClusterControlPlaneVipNetworkCIDR = akoOperatorConfig.AviDataPlaneNetworkCIDR
	}

	if akoOperatorConfig.AviManagementClusterDataPlaneNetworkName == "" || akoOperatorConfig.AviManagementClusterDataPlaneNetworkCIDR == "" {
		akoOperatorConfig.AviManagementClusterDataPlaneNetworkName = akoOperatorConfig.AviDataPlaneNetworkName
		akoOperatorConfig.AviManagementClusterDataPlaneNetworkCIDR = akoOperatorConfig.AviDataPlaneNetworkCIDR
	}
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

func getCaCertAndImageRepoFromUserProviderConfigValues(userProviderConfigValues map[string]interface{}, bomConfig *tkgconfigbom.BOMConfiguration,
	readerWriter tkgconfigreaderwriter.TKGConfigReaderWriter) (string, string) {
	caCert := ""
	imageRepo := bomConfig.ImageConfig.ImageRepository
	// implement the same logic as legacy func tkg_image_repo_ca_cert() in providers/ytt/lib/helpers.star
	if val, ok := userProviderConfigValues[constants.TKGProxyCACert]; ok {
		caCert = val.(string)
	} else if val, ok := userProviderConfigValues[constants.ConfigVariableCustomImageRepositoryCaCertificate]; ok {
		caCert = val.(string)
	}

	// implement the same logic as legacy func tkg_image_repo() in providers/ytt/lib/helpers.star
	if val, ok := userProviderConfigValues[constants.ConfigVariableCustomImageRepository]; ok {
		imageRepo = val.(string)
	}

	// override the values with tkgconfig readerwriter if exists
	if readerWriter != nil {
		repo, err := readerWriter.Get(constants.ConfigVariableCustomImageRepository)
		if err == nil && repo != "" {
			imageRepo = repo
		}
		if ca, err := readerWriter.Get(constants.TKGProxyCACert); err == nil && ca != "" {
			caCert = ca
		} else if ca, err := readerWriter.Get(constants.ConfigVariableCustomImageRepositoryCaCertificate); err == nil && ca != "" {
			caCert = ca
		}
	}

	return caCert, imageRepo
}
