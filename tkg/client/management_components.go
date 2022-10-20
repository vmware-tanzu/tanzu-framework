// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/packageclients/pkg/packageclient"
	"github.com/vmware-tanzu/tanzu-framework/packageclients/pkg/packagedatamodel"
	"github.com/vmware-tanzu/tanzu-framework/tkg/carvelhelpers"
	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/tkg/managementcomponents"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigreaderwriter"
	"github.com/vmware-tanzu/tanzu-framework/tkg/utils"
)

func (c *TkgClient) InstallOrUpgradeKappController(clusterClient clusterclient.Client, operationType constants.OperationType) error {
	// Get kapp-controller configuration file
	kappControllerConfigFile, err := c.getKappControllerConfigFile()
	if err != nil {
		return err
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

// RemoveObsoleteManagementComponents lists and removes management cluster components that are obsoleted by new
// management packages, e.g. the tkr-controller-manager deployment in tkr-system namespace.
func RemoveObsoleteManagementComponents(clusterClient clusterclient.Client) error {
	objectsToDelete, err := listObjectsToDelete(clusterClient)
	if err != nil {
		return err
	}

	for _, object := range objectsToDelete {
		err := clusterClient.DeleteResource(object)
		if kerrors.FilterOut(err, apierrors.IsNotFound) != nil {
			return errors.Wrapf(err, "unable to delete resource %s: '%s/%s'",
				object.GetObjectKind().GroupVersionKind(), object.GetNamespace(), object.GetName())
		}
	}
	return nil
}

func listObjectsToDelete(clusterClient clusterclient.Client) ([]client.Object, error) {
	var objectsToDelete []client.Object

	kindsOfObjectsToDelete := map[schema.GroupVersionKind][]client.ListOption{
		{
			Group:   "addons.cluster.x-k8s.io",
			Version: "v1beta1",
			Kind:    "ClusterResourceSet",
		}: {client.InNamespace(constants.TkrNamespace)},
		{
			Group:   "apps",
			Version: "v1",
			Kind:    "Deployment",
		}: {client.InNamespace(constants.TkrNamespace)},
	}

	for gvk, listOptions := range kindsOfObjectsToDelete {
		objectList := &unstructured.UnstructuredList{}
		objectList.SetGroupVersionKind(gvk)

		if err := clusterClient.ListResources(objectList, listOptions...); err != nil {
			return nil, errors.Wrapf(err, "unable to list resources: %s", gvk.String())
		}

		for i := range objectList.Items {
			objectsToDelete = append(objectsToDelete, &objectList.Items[i])
		}
	}

	return objectsToDelete, nil
}

// InstallOrUpgradeManagementComponents install management components to the cluster
func (c *TkgClient) InstallOrUpgradeManagementComponents(mcClient clusterclient.Client, pkgClient packageclient.PackageClient, kubecontext string, upgrade bool) error {
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

	var addonsManagerPackageVersion string
	if upgrade {
		addonsManagerPackageVersion, err = c.GetAddonsManagerPackageversion(managementPackageVersion)
		if err != nil {
			return err
		}
	} else {
		addonsManagerPackageVersion = managementPackageVersion
		envDefinedVersion := os.Getenv("_ADDONS_MANAGER_PACKAGE_VERSION")
		if envDefinedVersion != "" {
			addonsManagerPackageVersion = strings.TrimLeft(envDefinedVersion, "v")
		}
	}

	// Get TKG package's values file
	tkgPackageValuesFile, err := c.getTKGPackageConfigValuesFile(mcClient, managementPackageVersion, addonsManagerPackageVersion, upgrade)
	if err != nil {
		return err
	}

	managementcomponentsInstallOptions := managementcomponents.ManagementComponentsInstallOptions{
		ClusterOptions: managementcomponents.ClusterOptions{
			Kubecontext: kubecontext,
		},
		ManagementPackageRepositoryOptions: managementcomponents.ManagementPackageRepositoryOptions{
			ManagementPackageRepoImage: managementPackageRepoImage,
			TKGPackageValuesFile:       tkgPackageValuesFile,
			PackageVersion:             managementPackageVersion,
			PackageInstallTimeout:      c.getPackageInstallTimeoutFromConfig(),
		},
	}

	err = managementcomponents.InstallManagementComponents(mcClient, pkgClient, &managementcomponentsInstallOptions)

	// Remove intermediate config files if err is empty
	if err == nil {
		os.Remove(tkgPackageValuesFile)
	}

	return err
}

// InstallAKO install AKO to the cluster
func (c *TkgClient) InstallAKO(mcClient clusterclient.Client) error {
	// Get AKO file
	akoPackageInstallFile, err := c.getAKOPackageInstallFile()
	if err != nil {
		return err
	}

	// Apply ako packageinstall configuration
	if err := mcClient.ApplyFile(akoPackageInstallFile); err != nil {
		return errors.Wrapf(err, "error installing %s", constants.AKODeploymentName)
	}
	// Remove intermediate config files if err is empty
	if err == nil {
		os.Remove(akoPackageInstallFile)
	}
	// no need to wait for AKO packageInstall to be ready. It will be ready once the AKOO
	// creates the secret for it when a cluster is created.
	// this is to workaround a bug that AKO might allocate control plane HA IP to pinniped service
	return err
}

// GetAddonsManagerPackageversion returns a addons manager package version
func (c *TkgClient) GetAddonsManagerPackageversion(managementPackageVersion string) (string, error) {
	envDefinedVersion := os.Getenv("_ADDONS_MANAGER_PACKAGE_VERSION")
	if envDefinedVersion != "" {
		return strings.TrimLeft(envDefinedVersion, "v"), nil
	}
	packageVersion := managementPackageVersion
	var err error
	if packageVersion == "" {
		packageVersion, err = c.tkgBomClient.GetManagementPackagesVersion()
		if err != nil {
			return "", err
		}
	}
	packageVersion = strings.TrimLeft(packageVersion, "v")
	// the following is done to address https://github.com/vmware-tanzu/tanzu-framework/issues/3894
	match, _ := regexp.MatchString("\\+vmware.\\d+$", packageVersion)
	if !match {
		packageVersion = packageVersion + "+vmware.1"
	}
	return packageVersion, nil
}

func (c *TkgClient) getTKGPackageConfigValuesFile(mcClient clusterclient.Client, managementPackageVersion, addonsManagerPackageVersion string, upgrade bool) (string, error) {
	var userProviderConfigValues map[string]interface{}
	var err error

	if upgrade {
		userProviderConfigValues, err = c.getUserConfigVariableValueMapFromSecret(mcClient)
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

	valuesFile, err := managementcomponents.GetTKGPackageConfigValuesFileFromUserConfig(managementPackageVersion, addonsManagerPackageVersion, userProviderConfigValues, tkgBomConfig, c.TKGConfigReaderWriter())
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

func (c *TkgClient) getUserConfigVariableValueMapFromSecret(clusterClient clusterclient.Client) (map[string]interface{}, error) {
	pollOptions := &clusterclient.PollOptions{Interval: clusterclient.CheckResourceInterval, Timeout: 3 * clusterclient.CheckResourceInterval}
	configValues := make(map[string]interface{})
	// In ClusterClass based cluster, the user config variables can be retrieved from the tkg-pkg package data values secret directly
	bytes, err := clusterClient.GetSecretValue(fmt.Sprintf(packagedatamodel.SecretName, constants.TKGManagementPackageInstallName, constants.TkgNamespace), constants.TKGPackageValuesFile, constants.TkgNamespace, pollOptions)
	if err == nil {
		var tkgPackageConfig managementcomponents.TKGPackageConfig

		err = yaml.Unmarshal(bytes, &tkgPackageConfig)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to unmarshal configuration from secret:  %v-%v-values, namespace: %v", constants.TKGManagementPackageInstallName, constants.TkgNamespace, constants.TkgNamespace)
		}
		configValues = tkgPackageConfig.ConfigValues

	} else if err != nil && apierrors.IsNotFound(err) {
		// Handle the upgrade from legacy (non-package-based-lcm) management cluster as
		// legacy (non-package-based-lcm) management cluster will not have the secret tkg-pkg-tkg-system-values
		// defined on the cluster. Github issue: https://github.com/vmware-tanzu/tanzu-framework/issues/2147
		clusterName, _, err := c.getRegionalClusterNameAndNamespace(clusterClient)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get management cluster name and namespace")
		}
		// So we retrieve the user config variables from the <cluster-name>-config-values secret
		// which was managed in legacy ytt template providers/ytt/09_miscellaneous
		bytes, err := clusterClient.GetSecretValue(fmt.Sprintf("%s-config-values", clusterName), "value", constants.TkgNamespace, pollOptions)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to get the %s-config-values secret", clusterName)
		}

		err = yaml.Unmarshal(bytes, &configValues)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to yaml unmashal the data.value of %s-config-values secret", clusterName)
		}
	} else {
		return nil, errors.Wrapf(err, "unable to get the secret %v-%v-values, namespace: %v", constants.TKGManagementPackageInstallName, constants.TkgNamespace, constants.TkgNamespace)
	}

	err = c.mutateUserConfigVariableValueMap(configValues)
	if err != nil {
		return nil, errors.Wrap(err, "unable to mapping the current configuration variables to the cluster's existing configuration")
	}

	return configValues, nil
}

// mutateUserConfigVariableValueMap get user config variables to overwrite the existing config variables that
// retrieved from the cluster. This is mainly for mutating during cluster upgrading.
func (c *TkgClient) mutateUserConfigVariableValueMap(configValues map[string]interface{}) error {
	userProvidedConfigValues, err := c.getUserConfigVariableValueMap()
	if err != nil {
		return err
	}
	for k, v := range userProvidedConfigValues {
		configValues[k] = v
	}
	return nil
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

func (c *TkgClient) getAKOPackageInstallFile() (string, error) {
	akoPackageInstallTemplateFile, err := utils.CreateTempFile("", "*.yaml")
	if err != nil {
		return "", err
	}
	err = utils.WriteToFile(akoPackageInstallTemplateFile, []byte(constants.AKOPackageInstall))
	if err != nil {
		return "", err
	}

	userConfigValuesFile, err := c.getUserConfigVariableValueMapFile()
	if err != nil {
		return "", err
	}

	akoPackageInstallFile, err := ProcessAKOPackageInstallFile(akoPackageInstallTemplateFile, userConfigValuesFile)
	if err != nil {
		return "", err
	}

	return akoPackageInstallFile, nil
}

func ProcessAKOPackageInstallFile(akoPackageInstallTemplateFile, userConfigValuesFile string) (string, error) {
	akoPackageInstallContent, err := carvelhelpers.ProcessYTTPackage(akoPackageInstallTemplateFile, userConfigValuesFile)
	if err != nil {
		return "", err
	}

	akoPackageInstallFile, err := utils.CreateTempFile("", "*.yaml")
	if err != nil {
		return "", err
	}

	if err := utils.WriteToFile(akoPackageInstallFile, akoPackageInstallContent); err != nil {
		return "", err
	}

	return akoPackageInstallFile, nil
}
