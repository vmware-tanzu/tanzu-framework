// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package managementcomponents implements management component installation helpers
package managementcomponents

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
func GetTKGPackageConfigFromUserConfig(managementPackageVersion, addonsManagerPackageVersion string, userProviderConfigValues map[string]interface{}, tkgBomConfig *tkgconfigbom.BOMConfiguration, readerWriter tkgconfigreaderwriter.TKGConfigReaderWriter, onBootstrapCluster bool) (*TKGPackageConfig, error) {
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
		tkrRepoImagePath = fmt.Sprintf("%s/%s", tkgBomConfig.ImageConfig.ImageRepository, tkgBomConfig.TKRPackageRepo.Oracle)
	default:
		return nil, errors.Errorf("unknown provider type %q", providerType)
	}

	skipVerifyCert := getSkipVerify(userProviderConfigValues, readerWriter)

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
				DeployCLIPluginCRD: false,
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
				SkipVerifyCert:       skipVerifyCert,
				ImageRepo:            imageRepo,
			},
		},
		CoreManagementPluginsPackage: CoreManagementPluginsPackage{
			VersionConstraints: managementPackageVersion,
		},
	}

	// fill in nsx advanced load balancer(a.k.a avi) config
	if err := setAkoOperatorConfig(&tkgPackageConfig, userProviderConfigValues, onBootstrapCluster); err != nil {
		return nil, err
	}
	setProxyConfiguration(&tkgPackageConfig, userProviderConfigValues)

	return &tkgPackageConfig, nil
}

// GetTKGPackageConfigValuesFileFromUserConfig returns values file from user configuration
func GetTKGPackageConfigValuesFileFromUserConfig(managementPackageVersion, addonsManagerPackageVersion string, userProviderConfigValues map[string]interface{}, tkgBomConfig *tkgconfigbom.BOMConfiguration, readerWriter tkgconfigreaderwriter.TKGConfigReaderWriter, onBootstrapCluster bool) (string, error) {

	tkgPackageConfig, err := GetTKGPackageConfigFromUserConfig(managementPackageVersion, addonsManagerPackageVersion, userProviderConfigValues, tkgBomConfig, readerWriter, onBootstrapCluster)
	if err != nil {
		return "", err
	}

	configBytes, err := yaml.Marshal(tkgPackageConfig)
	if err != nil {
		return "", err
	}

	valuesFile := filepath.Join(os.TempDir(), constants.TKGPackageValuesFile)
	if err = utils.SaveFile(valuesFile, configBytes); err != nil {
		return "", err
	}

	return valuesFile, nil
}

// convertToString converts config into string type, and return "" if config is not set
func convertToString(config interface{}) string {
	if config != nil {
		return config.(string)
	}
	return ""
}

// convertNodeNetworkList converts config into avi node network list type, and return default value if config is not set
func convertNodeNetworkList(userProviderConfigValues map[string]interface{}) (string, error) {
	var nodeNetworkList []NodeNetwork
	config := userProviderConfigValues[constants.ConfigVariableAviIngressNodeNetworkList]
	if config == nil || config.(string) == "" || config.(string) == `""` {
		// return vsphere network if node network list is not set
		network_pathes := strings.Split(convertToString(userProviderConfigValues[constants.ConfigVariableVsphereNetwork]), "/")
		network := network_pathes[len(network_pathes)-1]
		nodeNetworkList = []NodeNetwork{
			{
				NetworkName: network,
			},
		}
	} else {
		if err := yaml.Unmarshal([]byte(config.(string)), &nodeNetworkList); err != nil {
			return "", errors.Errorf("Invalid node network list %s", config.(string))
		}
	}
	// convert nodeNetworkList to json string
	jsonBytes, err := json.Marshal(nodeNetworkList)
	if err != nil {
		return "", errors.Errorf("Cannot convert nodeNetworkList to json string")
	}

	return string(jsonBytes), nil
}

// convertToBool converts config into bool type, and return false if config is not set
func convertToBool(config interface{}) bool {
	if config != nil {
		return config.(bool)
	}
	return false
}

// convertAVILabels converts config into string type, and return empty string if config is not set
func convertAVILabels(config interface{}) (string, error) {
	if config != nil {
		switch config.(type) {
		case string:
			return config.(string), nil
		default:
			jsonBytes, err := json.Marshal(config)
			if err != nil {
				return "", err
			}
			return string(jsonBytes), nil
		}
	}
	return "", nil
}

func setAkoOperatorConfig(tkgPackageConfig *TKGPackageConfig, userProviderConfigValues map[string]interface{}, onBootstrapCluster bool) error {
	if !convertToBool(userProviderConfigValues[constants.ConfigVariableAviEnable]) {
		return nil
	}

	nodeNetworkList, err := convertNodeNetworkList(userProviderConfigValues)
	if err != nil {
		return errors.Wrapf(err, "Error convert node network list")
	}

	aviLabelsJsonString, err := convertAVILabels(userProviderConfigValues[constants.ConfigVariableAviLabels])
	if err != nil {
		return err
	}

	tkgPackageConfig.AkoOperatorPackage = AkoOperatorPackage{
		AkoOperatorPackageValues: AkoOperatorPackageValues{
			AviEnable:          convertToBool(userProviderConfigValues[constants.ConfigVariableAviEnable]),
			ClusterName:        convertToString(userProviderConfigValues[constants.ConfigVariableClusterName]),
			OnBootstrapCluster: onBootstrapCluster,
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
				AviNSXTT1Router:                                convertToString(userProviderConfigValues[constants.ConfigVariableAviNSXTT1Router]),
				AviLabels:                                      aviLabelsJsonString,
				AviControlPlaneHaProvider:                      convertToBool(userProviderConfigValues[constants.ConfigVariableVsphereHaProvider]),
				AviIngressNodeNetworkList:                      nodeNetworkList,
			},
		},
	}

	// auto fill in vip networks
	autofillAkoOperatorConfig(&tkgPackageConfig.AkoOperatorPackage.AkoOperatorPackageValues.AkoOperatorConfig)
	return nil
}

// autofillAkoOperatorConfig autofills empty fields in AkoOperatorConfig
func autofillAkoOperatorConfig(akoOperatorConfig *AkoOperatorConfig) {
	if akoOperatorConfig.AviManagementClusterServiceEngineGroup == "" {
		akoOperatorConfig.AviManagementClusterServiceEngineGroup = akoOperatorConfig.AviServiceEngineGroup
	}

	// fill in mgmt cluster data plane VIP networks
	if akoOperatorConfig.AviManagementClusterDataPlaneNetworkName == "" || akoOperatorConfig.AviManagementClusterDataPlaneNetworkCIDR == "" {
		akoOperatorConfig.AviManagementClusterDataPlaneNetworkName = akoOperatorConfig.AviDataPlaneNetworkName
		akoOperatorConfig.AviManagementClusterDataPlaneNetworkCIDR = akoOperatorConfig.AviDataPlaneNetworkCIDR
	}
	// fill in workload clusters' control plane VIP network
	if akoOperatorConfig.AviControlPlaneNetworkName == "" || akoOperatorConfig.AviControlPlaneNetworkCIDR == "" {
		akoOperatorConfig.AviControlPlaneNetworkName = akoOperatorConfig.AviManagementClusterDataPlaneNetworkName
		akoOperatorConfig.AviControlPlaneNetworkCIDR = akoOperatorConfig.AviManagementClusterDataPlaneNetworkCIDR
	}
	// fill in management cluster control plane VIP network
	if akoOperatorConfig.AviManagementClusterControlPlaneVipNetworkName == "" || akoOperatorConfig.AviManagementClusterControlPlaneVipNetworkCIDR == "" {
		akoOperatorConfig.AviManagementClusterControlPlaneVipNetworkName = akoOperatorConfig.AviManagementClusterDataPlaneNetworkName
		akoOperatorConfig.AviManagementClusterControlPlaneVipNetworkCIDR = akoOperatorConfig.AviManagementClusterDataPlaneNetworkCIDR
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
	setProxyInTKRServicePackage(tkgPackageConfig, httpProxy, httpsProxy, noProxy)
}

func setProxyInTKRSourceControllerPackage(tkgPackageConfig *TKGPackageConfig, httpProxy, httpsProxy, noProxy string) {
	tkgPackageConfig.TKRSourceControllerPackage.TKRSourceControllerPackageValues.Deployment =
		TKRSourceControllerPackageValuesDeployment{
			HttpProxy:  httpProxy,
			HttpsProxy: httpsProxy,
			NoProxy:    noProxy,
		}
}

func setProxyInTKRServicePackage(tkgPackageConfig *TKGPackageConfig, httpProxy, httpsProxy, noProxy string) {
	tkgPackageConfig.FrameworkPackage.TKRServicePackageValues.Deployment =
		TKRServicePackageValuesDeployment{
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

func getSkipVerify(userProviderConfigValues map[string]interface{}, readerWriter tkgconfigreaderwriter.TKGConfigReaderWriter) bool {
	defer func() {
		recover() // don't panic
	}()

	if readerWriter != nil {
		if skipVerifyStr, err := readerWriter.Get(constants.ConfigVariableCustomImageRepositorySkipTLSVerify); err == nil {
			if skipVerifyBool, err := strconv.ParseBool(skipVerifyStr); err == nil {
				return skipVerifyBool
			}
		}
	}

	if skipVerifyVal, ok := userProviderConfigValues[constants.ConfigVariableCustomImageRepositorySkipTLSVerify]; ok {
		if skipVerifyBool, err := strconv.ParseBool(fmt.Sprint(skipVerifyVal)); err == nil {
			return skipVerifyBool
		}
	}

	return false
}
