// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgconfigupdater

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	ctlimg "github.com/k14s/imgpkg/pkg/imgpkg/registry"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v3"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/clientconfighelpers"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfigpaths"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/registry"
)

type provider struct {
	Name         string `yaml:"name"`
	URL          string `yaml:"url"`
	ProviderType string `yaml:"type"`
}

type certManager struct {
	URL     string `yaml:"url"`
	Version string `yaml:"version"`
	Timeout string `yaml:"timeout"`
}

type providers struct {
	Providers   []provider  `yaml:"providers"`
	CertManager certManager `yaml:"cert-manager"`
}

func (c *client) defaultProviders() (providers, error) {
	tkgDir, _, providersDir, err := c.tkgConfigPathsClient.GetTKGConfigDirectories()
	if err != nil {
		return providers{}, err
	}

	providerConfigBytes, err := os.ReadFile(filepath.Join(providersDir, constants.LocalProvidersConfigFileName))
	if err != nil {
		return providers{}, errors.Wrap(err, "cannot get provider config")
	}

	providersConfig := providers{}
	err = yaml.Unmarshal(providerConfigBytes, &providersConfig)
	if err != nil {
		return providers{}, errors.Wrapf(err, "Unable to unmarshall provider config")
	}
	for i := range providersConfig.Providers {
		providersConfig.Providers[i].URL = getFilePath(tkgDir, providersConfig.Providers[i].URL)
	}
	providersConfig.CertManager.URL = getFilePath(tkgDir, providersConfig.CertManager.URL)
	providersConfig.CertManager.Timeout = "30m" // TODO: Make this configurable and increase the default timeout value (https://github.com/vmware-tanzu/tanzu-framework/issues/2360)

	return providersConfig, nil
}

func getFilePath(tkgDir, url string) string {
	path := filepath.Join(tkgDir, url)
	if strings.Contains(path, "\\") {
		// convert windows backslash style paths 'c:\foo\....' to file:// urls
		path = "file:///" + filepath.ToSlash(path)
	}
	return path
}

func ensureCertManagerInConfig(defaultProviders providers, tkgConfigNode *yaml.Node) error {
	certManagerIndex := GetNodeIndex(tkgConfigNode.Content[0].Content, constants.CertManagerConfigKey)
	if certManagerIndex == -1 {
		tkgConfigNode.Content[0].Content = append(tkgConfigNode.Content[0].Content, createSequenceNode(constants.CertManagerConfigKey)...)
		certManagerIndex = GetNodeIndex(tkgConfigNode.Content[0].Content, constants.CertManagerConfigKey)
	}

	certManagerNode := yaml.Node{}
	if err := copyData(defaultProviders.CertManager, &certManagerNode); err != nil {
		return errors.Wrap(err, "unable to get cert-manager")
	}

	tkgConfigNode.Content[0].Content[certManagerIndex] = certManagerNode.Content[0]
	return nil
}

func copyData(from, to interface{}) error {
	bytes, err := yaml.Marshal(from)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(bytes, to); err != nil {
		return err
	}
	return nil
}

func ensureProvidersWhenKeyAbsent(defaultProviders providers, tkgConfigNode *yaml.Node) error {
	tkgConfigNode.Content[0].Content = append(tkgConfigNode.Content[0].Content, createSequenceNode(constants.ProvidersConfigKey)...)
	providerIndex := GetNodeIndex(tkgConfigNode.Content[0].Content, constants.ProvidersConfigKey)

	providerListNode := yaml.Node{}
	if err := copyData(defaultProviders.Providers, &providerListNode); err != nil {
		return errors.Wrap(err, "unable to get a list of default providers")
	}
	tkgConfigNode.Content[0].Content[providerIndex] = providerListNode.Content[0]
	return nil
}

// EnsureProvidersInConfig ensures the providers section in tkgconfig exists and it is synchronized with the latest providers
func (c *client) EnsureProvidersInConfig(needUpdate bool, tkgConfigNode *yaml.Node) error {
	providerIndex := GetNodeIndex(tkgConfigNode.Content[0].Content, constants.ProvidersConfigKey)
	if providerIndex != -1 && !needUpdate {
		return nil
	}
	defaultProviders, err := c.defaultProviders()
	if err != nil {
		return errors.Wrap(err, "unable to get a list of default providers")
	}

	err = ensureCertManagerInConfig(defaultProviders, tkgConfigNode)
	if err != nil {
		return err
	}

	if providerIndex == -1 {
		return ensureProvidersWhenKeyAbsent(defaultProviders, tkgConfigNode)
	}

	userProviders := providers{}
	if err := copyData(tkgConfigNode, &userProviders); err != nil {
		return errors.Wrap(err, "unable to get a list of default providers")
	}

	for _, dp := range defaultProviders.Providers {
		found := false
		for i, p := range userProviders.Providers {
			if p.Name == dp.Name && p.ProviderType == dp.ProviderType {
				userProviders.Providers[i].URL = dp.URL
				found = true
				break
			}
		}
		if !found {
			userProviders.Providers = append(userProviders.Providers, provider{Name: dp.Name, ProviderType: dp.ProviderType, URL: dp.URL})
		}
	}

	updatedproviderListNode := yaml.Node{}
	if err := copyData(userProviders.Providers, &updatedproviderListNode); err != nil {
		return errors.Wrap(err, "unable to get a list of default providers")
	}
	tkgConfigNode.Content[0].Content[providerIndex] = updatedproviderListNode.Content[0]

	return nil
}

func (c *client) CheckInfrastructureVersion(providerName string) (string, error) {
	strs := strings.Split(providerName, ":")
	if len(strs) > 2 || len(strs) == 0 {
		return "", errors.New("not a valid infrastructure provider name")
	}

	if len(strs) == 1 {
		version, err := c.GetDefaultInfrastructureVersion(providerName)
		if err != nil {
			return "", errors.Wrapf(err, "not able to set default infrastructure provider version for %s", providerName)
		}

		return providerName + ":" + version, nil
	}

	match, err := regexp.MatchString("v([0-9]+).([0-9]+).([0-9]+)", strs[len(strs)-1])
	if err != nil || !match {
		return "", errors.Errorf("%s is not a valid provider version", strs[len(strs)-1])
	}

	return providerName, nil
}

func (c *client) GetDefaultInfrastructureVersion(providerName string) (string, error) {
	tkgConfigPath, err := c.tkgConfigPathsClient.GetTKGConfigPath()
	if err != nil {
		return "", err
	}

	fileData, err := os.ReadFile(tkgConfigPath)
	if err != nil {
		return "", errors.Wrapf(err, "unable to read tkg configuration from: %s", tkgConfigPath)
	}

	var providerMap providers

	count := 0
	version := ""

	if err = yaml.Unmarshal(fileData, &providerMap); err != nil {
		return "", errors.Wrapf(err, "%s does not contains valid providers info", tkgConfigPath)
	}

	for _, p := range providerMap.Providers {
		if p.Name == providerName && p.ProviderType == constants.InfrastructureProviderType {
			count++
			version, err = extractVersionFromPath(p.URL)
			if err != nil {
				return "", err
			}
		}
	}

	if count != 1 {
		return "", errors.Errorf("cannot get default infrastructure provider version for %s from config file %s, 0 or multiple versions found", providerName, tkgConfigPath)
	}
	return version, nil
}

func extractVersionFromPath(path string) (string, error) {
	// according to clusterctl provider contract, a local repository need to follow the pattern ~/local-repository/infrastructure-aws/v0.5.2/xxx.yaml
	strs := strings.Split(path, "/")

	const maxStrLen = 2
	if len(strs) < maxStrLen {
		return "", errors.Errorf("%s is not a valid local provider repository path", path)
	}

	match, err := regexp.MatchString("v([0-9]+).([0-9]+).([0-9]+)", strs[len(strs)-2])
	if err != nil || !match {
		return "", errors.Errorf("%s is not a valid local provider repository path", path)
	}
	return strs[len(strs)-2], nil
}

func (c *client) InitProvidersRegistry() (registry.Registry, error) {
	verifyCerts := true
	skipVerifyCerts, err := c.tkgConfigReaderWriter.Get(constants.ConfigVariableCustomImageRepositorySkipTLSVerify)
	if err == nil && strings.EqualFold(skipVerifyCerts, "true") {
		verifyCerts = false
	}

	registryOpts := ctlimg.Opts{
		VerifyCerts: verifyCerts,
		Anon:        true,
	}

	caCertBytes, err := clientconfighelpers.GetCustomRepositoryCaCertificateForClient(c.tkgConfigReaderWriter)
	if err == nil && len(caCertBytes) != 0 {
		filePath, err := tkgconfigpaths.GetRegistryCertFile()
		if err != nil {
			return nil, err
		}
		err = os.WriteFile(filePath, caCertBytes, 0644)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to write the custom image registry CA cert to file '%s'", filePath)
		}
		registryOpts.CACertPaths = []string{filePath}
	}

	return registry.New(&registryOpts)
}
