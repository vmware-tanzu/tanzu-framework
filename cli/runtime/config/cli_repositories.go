// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config/nodeutils"
)

// GetCLIRepositories retrieves cli repositories
func GetCLIRepositories() ([]configapi.PluginRepository, error) {
	node, err := getClientConfigNode()
	if err != nil {
		return nil, err
	}
	return getCLIRepositories(node)
}

// GetCLIRepository retrieves cli repository by name
func GetCLIRepository(name string) (*configapi.PluginRepository, error) {
	node, err := getClientConfigNode()
	if err != nil {
		return nil, err
	}
	return getCLIRepository(node, name)
}

// SetCLIRepository add or update a repository
func SetCLIRepository(repository configapi.PluginRepository) (err error) {
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()
	node, err := getClientConfigNodeNoLock()
	if err != nil {
		return err
	}
	persist, err := setCLIRepository(node, repository)
	if err != nil {
		return err
	}
	if persist {
		err = persistNode(node)
		if err != nil {
			return err
		}
	}
	return err
}

// DeleteCLIRepository delete a cli repository by name
func DeleteCLIRepository(name string) error {
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()
	node, err := getClientConfigNodeNoLock()
	if err != nil {
		return err
	}
	err = deleteCLIRepository(node, name)
	if err != nil {
		return err
	}
	return persistNode(node)
}

func getCLIRepositories(node *yaml.Node) ([]configapi.PluginRepository, error) {
	cfg, err := nodeutils.ConvertFromNode[configapi.ClientConfig](node)
	if err != nil {
		return nil, err
	}
	if cfg.ClientOptions != nil && cfg.ClientOptions.CLI != nil && cfg.ClientOptions.CLI.Repositories != nil {
		return cfg.ClientOptions.CLI.Repositories, nil
	}
	return nil, errors.New("cli repositories not found")
}

//nolint: dupl
func getCLIRepository(node *yaml.Node, name string) (*configapi.PluginRepository, error) {
	cfg, err := nodeutils.ConvertFromNode[configapi.ClientConfig](node)
	if err != nil {
		return nil, err
	}
	if cfg.ClientOptions != nil && cfg.ClientOptions.CLI != nil && cfg.ClientOptions.CLI.Repositories != nil {
		for _, repository := range cfg.ClientOptions.CLI.Repositories {
			_, repositoryName := getRepositoryTypeAndName(repository)
			if repositoryName == name {
				return &repository, nil
			}
		}
	}
	return nil, errors.New("cli repository not found")
}

func setCLIRepositories(node *yaml.Node, repos []configapi.PluginRepository) (err error) {
	for _, repository := range repos {
		_, err = setCLIRepository(node, repository)
		if err != nil {
			return err
		}
	}
	return err
}

func setCLIRepository(node *yaml.Node, repository configapi.PluginRepository) (persist bool, err error) {
	configOptions := func(c *nodeutils.Config) {
		c.ForceCreate = true
		c.Keys = []nodeutils.Key{
			{Name: KeyClientOptions, Type: yaml.MappingNode},
			{Name: KeyCLI, Type: yaml.MappingNode},
			{Name: KeyRepositories, Type: yaml.SequenceNode},
		}
	}
	repositoriesNode := nodeutils.FindNode(node.Content[0], configOptions)
	if repositoriesNode == nil {
		return persist, err
	}
	return setRepository(repositoriesNode, repository)
}

//nolint: dupl
func deleteCLIRepository(node *yaml.Node, name string) error {
	configOptions := func(c *nodeutils.Config) {
		c.ForceCreate = false
		c.Keys = []nodeutils.Key{
			{Name: KeyClientOptions},
			{Name: KeyCLI},
			{Name: KeyRepositories},
		}
	}
	cliRepositoriesNode := nodeutils.FindNode(node.Content[0], configOptions)
	if cliRepositoriesNode == nil {
		return nil
	}
	repository, err := getCLIRepository(node, name)
	if err != nil {
		return err
	}
	repositoryType, repositoryName := getRepositoryTypeAndName(*repository)
	var result []*yaml.Node
	for _, repositoryNode := range cliRepositoriesNode.Content {
		if repositoryIndex := nodeutils.GetNodeIndex(repositoryNode.Content, repositoryType); repositoryIndex != -1 {
			if repositoryFieldIndex := nodeutils.GetNodeIndex(repositoryNode.Content[repositoryIndex].Content, "name"); repositoryFieldIndex != -1 && repositoryNode.Content[repositoryIndex].Content[repositoryFieldIndex].Value == repositoryName {
				continue
			}
		}
		result = append(result, repositoryNode)
	}
	cliRepositoriesNode.Style = 0
	cliRepositoriesNode.Content = result
	return nil
}

func setRepository(repositoriesNode *yaml.Node, repository configapi.PluginRepository) (persist bool, err error) {
	newNode, err := nodeutils.ConvertToNode[configapi.PluginRepository](&repository)
	if err != nil {
		return persist, err
	}
	exists := false
	var result []*yaml.Node
	repositoryType, repositoryName := getRepositoryTypeAndName(repository)
	if repositoryType == "" || repositoryName == "" {
		return persist, errors.New("not found")
	}
	for _, repositoryNode := range repositoriesNode.Content {
		if repositoryIndex := nodeutils.GetNodeIndex(repositoryNode.Content, repositoryType); repositoryIndex != -1 {
			if repositoryFieldIndex := nodeutils.GetNodeIndex(repositoryNode.Content[repositoryIndex].Content, "name"); repositoryFieldIndex != -1 &&
				repositoryNode.Content[repositoryIndex].Content[repositoryFieldIndex].Value == repositoryName {
				exists = true
				persist, err = nodeutils.NotEqual(newNode.Content[0], repositoryNode)
				if persist {
					err = nodeutils.MergeNodes(newNode.Content[0], repositoryNode)
					if err != nil {
						return persist, err
					}
				}
				result = append(result, repositoryNode)
				continue
			}
		}
		result = append(result, repositoryNode)
	}
	if !exists {
		result = append(result, newNode.Content[0])
		persist = true
	}
	repositoriesNode.Style = 0
	repositoriesNode.Content = result
	return persist, err
}

func getRepositoryTypeAndName(repository configapi.PluginRepository) (string, string) {
	if repository.GCPPluginRepository != nil && repository.GCPPluginRepository.Name != "" {
		return "gcpPluginRepository", repository.GCPPluginRepository.Name
	}
	return "", ""
}
