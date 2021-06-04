// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgconfigupdater

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/constants"
)

type provider struct {
	Name         string `yaml:"name"`
	URL          string `yaml:"url"`
	ProviderType string `yaml:"type"`
}

type providers struct {
	Providers []provider `yaml:"providers"`
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
		path := filepath.Join(tkgDir, providersConfig.Providers[i].URL)
		if strings.Contains(path, "\\") {
			// convert windows backslash style paths 'c:\foo\....' to file:// urls
			path = "file:///" + filepath.ToSlash(path)
		}

		providersConfig.Providers[i].URL = path
	}
	return providersConfig, nil
}

// EnsureProviders ensures the providers section in tkgconfig exisits and it is synchronized with the latest providers
func (c *client) EnsureProviders(needUpdate bool, tkgConfigNode *yaml.Node) error { //nolint:gocyclo
	providerIndex := GetNodeIndex(tkgConfigNode.Content[0].Content, constants.ProvidersConfigKey)
	if providerIndex != -1 && !needUpdate {
		return nil
	}
	defaultProviders, err := c.defaultProviders()
	if err != nil {
		return errors.Wrap(err, "unable to get a list of default providers")
	}

	if providerIndex == -1 {
		tkgConfigNode.Content[0].Content = append(tkgConfigNode.Content[0].Content, createSequenceNode(constants.ProvidersConfigKey)...)
		providerIndex = GetNodeIndex(tkgConfigNode.Content[0].Content, constants.ProvidersConfigKey)

		defaultProvidersBytes, err := yaml.Marshal(defaultProviders.Providers)
		if err != nil {
			return errors.Wrap(err, "unable to get a list of default providers")
		}
		providerListNode := yaml.Node{}
		err = yaml.Unmarshal(defaultProvidersBytes, &providerListNode)
		if err != nil {
			return errors.Wrap(err, "unable to get a list of default providers")
		}

		tkgConfigNode.Content[0].Content[providerIndex] = providerListNode.Content[0]
		return nil
	}

	userTKGConfigBytes, err := yaml.Marshal(tkgConfigNode)
	if err != nil {
		return err
	}

	userProviders := providers{}
	err = yaml.Unmarshal(userTKGConfigBytes, &userProviders)
	if err != nil {
		return err
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

	updatedProviderListBytes, err := yaml.Marshal(userProviders.Providers)
	if err != nil {
		return err
	}
	updatedproviderListNode := yaml.Node{}
	err = yaml.Unmarshal(updatedProviderListBytes, &updatedproviderListNode)
	if err != nil {
		return err
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

// CheckTkgConfigNeedUpdate checks if the providers section in tkg configuration file is synchronized with $HOME/.tkg/providers/config.yaml
func (c *client) CheckTkgConfigNeedUpdate() (bool, string, error) { //nolint:gocyclo
	tkgConfigPath, err := c.tkgConfigPathsClient.GetTKGConfigPath()
	if err != nil {
		return false, "", err
	}
	if _, err := os.Stat(tkgConfigPath); os.IsNotExist(err) {
		return false, tkgConfigPath, nil
	} else if err != nil {
		return false, "", err
	}

	providerConfigPath, err := c.tkgConfigPathsClient.GetProvidersConfigFilePath()
	if err != nil {
		return false, "", err
	}
	if _, err := os.Stat(providerConfigPath); os.IsNotExist(err) {
		return false, tkgConfigPath, nil
	} else if err != nil {
		return false, "", err
	}

	defaultProviders, err := c.defaultProviders()
	if err != nil {
		return false, "", err
	}

	tkgConfigData, err := os.ReadFile(tkgConfigPath)
	if err != nil {
		return false, "", errors.Wrapf(err, "unable to read tkg configuration from: %s", tkgConfigPath)
	}

	var tkgConfigProviders providers
	err = yaml.Unmarshal(tkgConfigData, &tkgConfigProviders)
	if err != nil {
		return false, "", err
	}

	if len(tkgConfigProviders.Providers) == 0 {
		return false, tkgConfigPath, nil
	}

	for _, dp := range defaultProviders.Providers {
		found := false
		for _, p := range tkgConfigProviders.Providers {
			if p.Name == dp.Name && p.ProviderType == dp.ProviderType {
				if p.URL != dp.URL {
					return true, tkgConfigPath, nil
				}
				found = true
				break
			}
		}
		if !found {
			return true, tkgConfigPath, nil
		}
	}

	return false, tkgConfigPath, nil
}
