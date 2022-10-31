// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/vmware-tanzu/tanzu-framework/packageclients/pkg/packagedatamodel"
	"github.com/vmware-tanzu/tanzu-framework/tkg/carvelhelpers"
	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/tkg/managementcomponents"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigreaderwriter"
	"github.com/vmware-tanzu/tanzu-framework/tkg/utils"
)

func (c *TkgClient) InstallOrUpgradeKappController(kubeconfig, kubecontext string, operationType constants.OperationType) error {
	// Get kapp-controller configuration file
	kappControllerConfigFile, err := c.getKappControllerConfigFile()
	if err != nil {
		return err
	}

	clusterClient, err := clusterclient.NewClient(kubeconfig, kubecontext, clusterclient.Options{})
	if err != nil {
		return errors.Wrap(err, "unable to get cluster client")
	}

	kappControllerOptions := managementcomponents.KappControllerOptions{
		KappControllerConfigFile:       kappControllerConfigFile,
		KappControllerInstallNamespace: constants.TkgNamespace,
	}
	err = managementcomponents.InstallKappController(clusterClient, kappControllerOptions, operationType)

	// Remove intermediate config files if err is empty
	if err == nil {
		os.Remove(kappControllerConfigFile)
	}
	return err
}

// InstallOrUpgradeManagementComponents install management components to the cluster
func (c *TkgClient) InstallOrUpgradeManagementComponents(kubeconfig, kubecontext string, upgrade bool) error {
	managementPackageRepoImage, err := c.tkgBomClient.GetManagementPackageRepositoryImage()
	if err != nil {
		return errors.Wrap(err, "unable to get management package repository image")
	}

	managementPackageVersion := ""

	// Override management package repository image if specified as part of below environment variable
	// NOTE: this override is only for testing purpose and we don't expect this to be used in production scenario
	mprImage := os.Getenv("_MANAGEMENT_PACKAGE_REPO_IMAGE")
	if mprImage != "" {
		managementPackageRepoImage = mprImage
	}

	// Override the version to use for management packages if specified as part of below environment variable
	// NOTE: this override is only for testing purpose and we don't expect this to be used in production scenario
	mpVersion := os.Getenv("_MANAGEMENT_PACKAGE_VERSION")
	if mpVersion != "" {
		managementPackageVersion = mpVersion
	}

	managementPackageVersion = strings.TrimLeft(managementPackageVersion, "v")

	// Get TKG package's values file
	tkgPackageValuesFile, err := c.getTKGPackageConfigValuesFile(managementPackageVersion, kubeconfig, kubecontext, upgrade)
	if err != nil {
		return err
	}

	managementcomponentsInstallOptions := managementcomponents.ManagementComponentsInstallOptions{
		ClusterOptions: managementcomponents.ClusterOptions{
			Kubeconfig:  kubeconfig,
			Kubecontext: kubecontext,
		},
		ManagementPackageRepositoryOptions: managementcomponents.ManagementPackageRepositoryOptions{
			ManagementPackageRepoImage: managementPackageRepoImage,
			TKGPackageValuesFile:       tkgPackageValuesFile,
			PackageVersion:             managementPackageVersion,
			PackageInstallTimeout:      c.getPackageInstallTimeoutFromConfig(),
		},
	}

	err = managementcomponents.InstallManagementComponents(&managementcomponentsInstallOptions)

	// Remove intermediate config files if err is empty
	if err == nil {
		os.Remove(tkgPackageValuesFile)
	}

	return err
}

func (c *TkgClient) getTKGPackageConfigValuesFile(managementPackageVersion, kubeconfig, kubecontext string, upgrade bool) (string, error) {
	var userProviderConfigValues map[string]interface{}
	var err error

	if upgrade {
		userProviderConfigValues, err = c.getUserConfigVariableValueMapFromSecret(kubeconfig, kubecontext)
	} else {
		userProviderConfigValues, err = c.getUserConfigVariableValueMap()
	}

	if err != nil {
		return "", err
	}

	tkgBomConfig, err := c.tkgBomClient.GetDefaultTkgBOMConfiguration()
	if err != nil {
		return "", err
	}

	valuesFile, err := managementcomponents.GetTKGPackageConfigValuesFileFromUserConfig(managementPackageVersion, userProviderConfigValues, tkgBomConfig)
	if err != nil {
		return "", err
	}

	return valuesFile, nil
}

func (c *TkgClient) getUserConfigVariableValueMap() (map[string]interface{}, error) {
	path, err := c.tkgConfigPathsClient.GetConfigDefaultsFilePath()
	if err != nil {
		return nil, err
	}

	return c.GetUserConfigVariableValueMap(path, c.TKGConfigReaderWriter())
}

func (c *TkgClient) getUserConfigVariableValueMapFromSecret(kubeconfig, kubecontext string) (map[string]interface{}, error) {
	clusterClient, err := clusterclient.NewClient(kubeconfig, kubecontext, clusterclient.Options{})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get cluster client")
	}

	var tkgPackageConfig managementcomponents.TKGPackageConfig

	// Handle the upgrade from legacy (non-package-based-lcm) management cluster as
	// legacy (non-package-based-lcm) management cluster will not have this secret defined
	// on the cluster. Github issue: https://github.com/vmware-tanzu/tanzu-framework/issues/2147
	bytes, err := clusterClient.GetSecretValue(fmt.Sprintf(packagedatamodel.SecretName, constants.TKGManagementPackageInstallName, constants.TkgNamespace), constants.TKGPackageValuesFile, constants.TkgNamespace, nil)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get cluster client")
	}

	err = yaml.Unmarshal(bytes, &tkgPackageConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to unmarshal configuration from secret: %v, namespace: %v", constants.TKGPackageValues, constants.TkgNamespace)
	}

	return tkgPackageConfig.ConfigValues, nil
}

func (c *TkgClient) getUserConfigVariableValueMapFile() (string, error) {
	userConfigValues, err := c.getUserConfigVariableValueMap()
	if err != nil {
		return "", err
	}

	configBytes, err := yaml.Marshal(userConfigValues)
	if err != nil {
		return "", err
	}

	prefix := []byte(`#@data/values
#@overlay/match-child-defaults missing_ok=True
---
`)
	configBytes = append(prefix, configBytes...)

	configFile, err := utils.CreateTempFile("", "*.yaml")
	if err != nil {
		return "", err
	}
	err = utils.WriteToFile(configFile, configBytes)
	if err != nil {
		return "", err
	}
	return configFile, nil
}

func (c *TkgClient) getKappControllerConfigFile() (string, error) {
	kappControllerPackageImage, err := c.tkgBomClient.GetKappControllerPackageImage()
	if err != nil {
		return "", err
	}

	path, err := c.tkgConfigPathsClient.GetTKGProvidersDirectory()
	if err != nil {
		return "", err
	}
	kappControllerValuesDirPath := filepath.Join(path, "kapp-controller-values")

	userConfigValuesFile, err := c.getUserConfigVariableValueMapFile()
	if err != nil {
		return "", err
	}

	defer func() {
		// Remove intermediate config files if err is empty
		if err == nil {
			os.Remove(userConfigValuesFile)
		}
	}()

	log.V(6).Infof("User ConfigValues File: %v", userConfigValuesFile)

	kappControllerConfigFile, err := ProcessKappControllerPackage(kappControllerPackageImage, userConfigValuesFile, kappControllerValuesDirPath)
	if err != nil {
		return "", err
	}

	return kappControllerConfigFile, nil
}

func ProcessKappControllerPackage(kappControllerPackageImage, userConfigValuesFile, kappControllerValuesDirPath string) (string, error) {
	kappControllerValuesFile, err := GetKappControllerConfigValuesFile(userConfigValuesFile, kappControllerValuesDirPath)
	if err != nil {
		return "", err
	}

	defer func() {
		// Remove intermediate config files if err is empty
		if err == nil {
			os.Remove(kappControllerValuesFile)
		}
	}()

	log.V(6).Infof("Kapp-controller values-file: %v", kappControllerValuesFile)

	configBytes, err := carvelhelpers.ProcessCarvelPackage(kappControllerPackageImage, kappControllerValuesFile)
	if err != nil {
		return "", err
	}

	configFile, err := utils.CreateTempFile("", "")
	if err != nil {
		return "", err
	}
	err = utils.WriteToFile(configFile, configBytes)
	if err != nil {
		return "", err
	}

	log.V(6).Infof("Kapp-controller configuration file: %v", configFile)
	return configFile, nil
}

func GetKappControllerConfigValuesFile(userConfigValuesFile, kappControllerValuesDir string) (string, error) {
	kappControllerValuesBytes, err := carvelhelpers.ProcessYTTPackage(kappControllerValuesDir, userConfigValuesFile)
	if err != nil {
		return "", err
	}

	prefix := []byte(`#@data/values
#@overlay/match-child-defaults missing_ok=True
#@overlay/replace
---
`)
	kappControllerValuesBytes = append(prefix, kappControllerValuesBytes...)
	kappControllerValuesFile, err := utils.CreateTempFile("", "*.yaml")
	if err != nil {
		return "", err
	}
	err = utils.WriteToFile(kappControllerValuesFile, kappControllerValuesBytes)
	if err != nil {
		return "", err
	}

	return kappControllerValuesFile, nil
}

// GetUserConfigVariableValueMap is a specific implementation expecting to use a flat key-value
// file to provide a source of keys to filter for the valid user provided values.
// For example, this function uses config_default.yaml filepath to find relevant config variables
// and returns the config map of user provided variable among all applicable config variables
func (c *TkgClient) GetUserConfigVariableValueMap(configDefaultFilePath string, rw tkgconfigreaderwriter.TKGConfigReaderWriter) (map[string]interface{}, error) {
	bytes, err := os.ReadFile(configDefaultFilePath)
	if err != nil {
		return nil, err
	}

	variables, err := GetConfigVariableListFromYamlData(bytes)
	if err != nil {
		return nil, err
	}

	userProvidedConfigValues := map[string]interface{}{}
	for _, k := range variables {
		if v, e := rw.Get(k); e == nil {
			userProvidedConfigValues[k] = utils.Convert(v)
		}
	}

	return userProvidedConfigValues, nil
}

func GetConfigVariableListFromYamlData(bytes []byte) ([]string, error) {
	configValues := map[string]interface{}{}
	err := yaml.Unmarshal(bytes, &configValues)
	if err != nil {
		return nil, errors.Wrap(err, "error while unmarshaling")
	}

	keys := make([]string, 0, len(configValues))
	for k := range configValues {
		keys = append(keys, k)
	}

	return keys, nil
}
