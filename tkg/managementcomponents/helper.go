// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package managementcomponents implements management component installation helpers
package managementcomponents

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigbom"
	"github.com/vmware-tanzu/tanzu-framework/tkg/utils"
)

const (
	packagePollInterval           = 5 * time.Second
	packagePollTimeout            = 10 * time.Minute
	aviNamespace                  = "avi-system"
	defaultServiceEngineGroupName = "Default-Group"
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
	default:
		return "", errors.Errorf("unknown provider type %q", providerType)
	}
	clusterNamespace := convertToString(userProviderConfigValues[constants.ConfigVariableNamespace])
	if clusterNamespace == "" {
		clusterNamespace = "default"
	}
	aviCertificate, err := base64.StdEncoding.DecodeString(convertToString(userProviderConfigValues[constants.ConfigVariableAviControllerCA]))
	if err != nil {
		return "", errors.Errorf("fail to get avi certificate")
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
		AkoOperatorPackage: AkoOperatorPackage{
			AkoOperatorPackageValues: AkoOperatorPackageValues{
				AviEnable:   convertToBool(userProviderConfigValues[constants.ConfigVariableAviEnable]),
				ClusterName: convertToString(userProviderConfigValues[constants.ConfigVariableClusterName]),
				AviOperatorConfig: AviOperatorConfig{
					AviControllerAddress:                           convertToString(userProviderConfigValues[constants.ConfigVariableAviControllerAddress]),
					AviControllerVersion:                           convertToString(userProviderConfigValues[constants.ConfigVariableAviControllerVersion]),
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
		LoadBalancerAndIngressServicePackage: LoadBalancerAndIngressServicePackage{
			LoadBalancerAndIngressServicePackageValues: LoadBalancerAndIngressServicePackageValues{
				Name:      "ako-" + clusterNamespace + "-" + convertToString(userProviderConfigValues[constants.ConfigVariableClusterName]),
				Namespace: aviNamespace,
				LoadBalancerAndIngressServiceConfig: LoadBalancerAndIngressServiceConfig{
					AkoSettings: AkoSettings{
						DisableStaticRouteSync: "true",
						ClusterName:            convertToString(userProviderConfigValues[constants.ConfigVariableClusterName]),
						CniPlugin:              convertToString(userProviderConfigValues[constants.ConfigVariableCNI]),
					},
					NetworkSettings: NetworkSettings{
						VipNetworkList: "[]",
					},
					ControllerSettings: ControllerSettings{
						ServiceEngineGroupName: defaultServiceEngineGroupName,
						CloudName:              convertToString(userProviderConfigValues[constants.ConfigVariableAviCloudName]),
						ControllerIp:           convertToString(userProviderConfigValues[constants.ConfigVariableAviControllerAddress]),
					},
					AviCredentials: AviCredentials{
						Username:                 convertToString(userProviderConfigValues[constants.ConfigVariableAviControllerUsername]),
						Password:                 convertToString(userProviderConfigValues[constants.ConfigVariableAviControllerPassword]),
						CertificateAuthorityData: string(aviCertificate),
					},
				},
			},
		},
	}

	aviOperatorValidator(&tkgPackageConfig.AkoOperatorPackage.AkoOperatorPackageValues.AviOperatorConfig)
	err = akoValidator(&tkgPackageConfig)
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

func convertToString(aviOperatorConfig interface{}) string {
	switch aviOperatorConfig.(type) {
	case string:
		return aviOperatorConfig.(string)
	default:
		return ""
	}
	if aviOperatorConfig == nil || reflect.ValueOf(aviOperatorConfig).IsNil() {
		return ""
	}
	return ""
}

func convertToBool(aviOperatorConfig interface{}) bool {
	switch aviOperatorConfig.(type) {
	case bool:
		return aviOperatorConfig.(bool)
	default:
		return false
	}
	if aviOperatorConfig == nil || reflect.ValueOf(aviOperatorConfig).IsNil() {
		return false
	}
	return false
}

func aviOperatorValidator(aviOperatorConfig *AviOperatorConfig) {
	if aviOperatorConfig.AviManagementClusterServiceEngineGroup == "" {
		aviOperatorConfig.AviManagementClusterServiceEngineGroup = aviOperatorConfig.AviServiceEngineGroup
	}

	if aviOperatorConfig.AviManagementClusterDataPlaneNetworkName != "" && aviOperatorConfig.AviManagementClusterDataPlaneNetworkCIDR != "" {
		aviOperatorConfig.AviControlPlaneNetworkName = aviOperatorConfig.AviManagementClusterDataPlaneNetworkName
		aviOperatorConfig.AviControlPlaneNetworkCIDR = aviOperatorConfig.AviManagementClusterDataPlaneNetworkCIDR
	} else if aviOperatorConfig.AviControlPlaneNetworkName == "" || aviOperatorConfig.AviControlPlaneNetworkCIDR == "" {
		aviOperatorConfig.AviControlPlaneNetworkName = aviOperatorConfig.AviDataPlaneNetworkName
		aviOperatorConfig.AviControlPlaneNetworkCIDR = aviOperatorConfig.AviDataPlaneNetworkCIDR
	}

	if aviOperatorConfig.AviManagementClusterDataPlaneNetworkName != "" && aviOperatorConfig.AviManagementClusterDataPlaneNetworkCIDR != "" {
		aviOperatorConfig.AviManagementClusterControlPlaneVipNetworkName = aviOperatorConfig.AviManagementClusterDataPlaneNetworkName
		aviOperatorConfig.AviManagementClusterControlPlaneVipNetworkCIDR = aviOperatorConfig.AviManagementClusterDataPlaneNetworkCIDR
	} else if aviOperatorConfig.AviManagementClusterControlPlaneVipNetworkName == "" || aviOperatorConfig.AviManagementClusterControlPlaneVipNetworkCIDR == "" {
		aviOperatorConfig.AviManagementClusterControlPlaneVipNetworkName = aviOperatorConfig.AviDataPlaneNetworkName
		aviOperatorConfig.AviManagementClusterControlPlaneVipNetworkCIDR = aviOperatorConfig.AviDataPlaneNetworkCIDR
	}

	if aviOperatorConfig.AviManagementClusterDataPlaneNetworkName == "" || aviOperatorConfig.AviManagementClusterDataPlaneNetworkCIDR == "" {
		aviOperatorConfig.AviManagementClusterDataPlaneNetworkName = aviOperatorConfig.AviDataPlaneNetworkName
		aviOperatorConfig.AviManagementClusterDataPlaneNetworkCIDR = aviOperatorConfig.AviDataPlaneNetworkCIDR
	}
}

// VipNetwork
type VipNetwork struct {
	NetworkName string `json:"networkName,omitempty"`
	Cidr        string `json:"cidr,omitempty"`
}

func akoValidator(tkgPackageConfig *TKGPackageConfig) error {
	aviOperatorConfig := tkgPackageConfig.AkoOperatorPackage.AkoOperatorPackageValues.AviOperatorConfig
	tkgPackageConfig.LoadBalancerAndIngressServicePackage.LoadBalancerAndIngressServicePackageValues.LoadBalancerAndIngressServiceConfig.ControllerSettings.ServiceEngineGroupName = aviOperatorConfig.AviManagementClusterServiceEngineGroup

	vipNetwork := []VipNetwork{
		VipNetwork{
			NetworkName: aviOperatorConfig.AviControlPlaneNetworkName,
			Cidr:        aviOperatorConfig.AviControlPlaneNetworkCIDR,
		},
	}
	data, err := json.Marshal(vipNetwork)
	if err != nil {
		return err
	}
	tkgPackageConfig.LoadBalancerAndIngressServicePackage.LoadBalancerAndIngressServicePackageValues.LoadBalancerAndIngressServiceConfig.NetworkSettings.VipNetworkList = string(data)
	return nil
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
