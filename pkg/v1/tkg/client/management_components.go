// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/managementcomponents"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfigreaderwriter"
)

// InstallManagementComponents install management components to the cluster
func (c *TkgClient) InstallManagementComponents(kubeconfig, kubecontext string) error {
	managementPackageRepoImage, err := c.tkgBomClient.GetManagementPackageRepositoryImage()
	if err != nil {
		return errors.Wrap(err, "unable to get management package repository image")
	}

	// Override management package repository image if specified as part of below environment variable
	// NOTE: this override is only for testing purpose and we don't expect this to be used in production scenario
	mprImage := os.Getenv("MANAGEMENT_PACKAGE_REPO_IMAGE")
	if mprImage != "" {
		managementPackageRepoImage = mprImage
	}

	tkgPackageValuesFile, err := c.getTKGPackageConfigValuesFile()
	if err != nil {
		return err
	}

	managementcomponentsInstallOptions := managementcomponents.ManagementComponentsInstallOptions{
		ClusterOptions: managementcomponents.ClusterOptions{
			Kubeconfig:  kubeconfig,
			Kubecontext: kubecontext,
		},
		// TODO: Install kapp-controller using from the tanzu managed template manifest (https://github.com/vmware-tanzu/tanzu-framework/issues/1672)
		KappControllerOptions: managementcomponents.KappControllerOptions{
			KappControllerConfigFile:       "https://github.com/vmware-tanzu/carvel-kapp-controller/releases/download/v0.31.0/release.yml",
			KappControllerInstallNamespace: "kapp-controller",
		},
		ManagementPackageRepositoryOptions: managementcomponents.ManagementPackageRepositoryOptions{
			ManagementPackageRepoImage: managementPackageRepoImage,
			TKGPackageValuesFile:       tkgPackageValuesFile,
		},
	}

	return managementcomponents.InstallManagementComponents(&managementcomponentsInstallOptions)
}

func (c *TkgClient) getTKGPackageConfigValuesFile() (string, error) {
	path, err := c.tkgConfigPathsClient.GetConfigDefaultsFilePath()
	if err != nil {
		return "", err
	}

	userProviderConfigValues, err := c.GetUserConfigVariableValueMap(path, c.TKGConfigReaderWriter())
	if err != nil {
		return "", err
	}

	valuesFile, err := managementcomponents.GetTKGPackageConfigValuesFileFromUserConfig(userProviderConfigValues)
	if err != nil {
		return "", err
	}

	return valuesFile, nil
}

// GetUserConfigVariableValueMap is a specific implementation expecting to use a flat key-value
// file to provide a source of keys to filter for the valid user provided values.
// For example, this function uses config_default.yaml filepath to find relevant config variables
// and returns the config map of user provided variable among all applicable config variables
func (c *TkgClient) GetUserConfigVariableValueMap(configDefaultFilePath string, rw tkgconfigreaderwriter.TKGConfigReaderWriter) (map[string]string, error) {
	bytes, err := os.ReadFile(configDefaultFilePath)
	if err != nil {
		return nil, err
	}

	variables, err := GetConfigVariableListFromYamlData(bytes)
	if err != nil {
		return nil, err
	}

	userProvidedConfigValues := map[string]string{}
	for _, k := range variables {
		if v, e := rw.Get(k); e == nil {
			userProvidedConfigValues[k] = v
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
