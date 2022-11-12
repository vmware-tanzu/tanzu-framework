// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config/nodeutils"
)

// GetCLIRepositories retrieves cli repositories
func GetCLIRepositories() ([]configapi.PluginRepository, error) {
	// Retrieve client config node
	node, err := getClientConfigNode()
	if err != nil {
		return nil, err
	}

	return getCLIRepositories(node)
}

// GetCLIRepository retrieves cli repository by name
func GetCLIRepository(name string) (*configapi.PluginRepository, error) {
	// Retrieve client config node
	node, err := getClientConfigNode()
	if err != nil {
		return nil, err
	}

	return getCLIRepository(node, name)
}

// SetCLIRepository add or update a repository
func SetCLIRepository(repository configapi.PluginRepository) (err error) {
	// Retrieve client config node
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()
	node, err := getClientConfigNodeNoLock()
	if err != nil {
		return err
	}

	// Add or update cli repository in the yaml node
	persist, err := setCLIRepository(node, repository)
	if err != nil {
		return err
	}

	// Persist the config node to the file
	if persist {
		err = persistConfig(node)
		if err != nil {
			return err
		}
	}

	return err
}

// DeleteCLIRepository delete a cli repository by name
func DeleteCLIRepository(name string) error {
	// Retrieve client config node
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()
	node, err := getClientConfigNodeNoLock()
	if err != nil {
		return err
	}

	// Delete the matching cli repository from the yaml node
	err = deleteCLIRepository(node, name)
	if err != nil {
		return err
	}

	// Persist the config node to the file
	return persistConfig(node)
}

func getCLIRepositories(node *yaml.Node) ([]configapi.PluginRepository, error) {
	cfg, err := convertNodeToClientConfig(node)
	if err != nil {
		return nil, err
	}
	if cfg.ClientOptions != nil && cfg.ClientOptions.CLI != nil && cfg.ClientOptions.CLI.Repositories != nil {
		return cfg.ClientOptions.CLI.Repositories, nil
	}
	return nil, errors.New("cli repositories not found")
}

func getCLIRepository(node *yaml.Node, name string) (*configapi.PluginRepository, error) {
	cfg, err := convertNodeToClientConfig(node)
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
	// Retrieve the patch strategies from config metadata
	patchStrategies, err := GetConfigMetadataPatchStrategy()
	if err != nil {
		patchStrategies = make(map[string]string)
	}

	// Find the cli repositories node in the yaml node
	keys := []nodeutils.Key{
		{Name: KeyClientOptions, Type: yaml.MappingNode},
		{Name: KeyCLI, Type: yaml.MappingNode},
		{Name: KeyRepositories, Type: yaml.SequenceNode},
	}
	cliRepositoriesNode := nodeutils.FindNode(node.Content[0], nodeutils.WithForceCreate(), nodeutils.WithKeys(keys))
	if cliRepositoriesNode == nil {
		return persist, err
	}

	// Add or Update cli repository to cli repositories node based on patch strategy
	return setRepository(cliRepositoriesNode, repository, nodeutils.WithPatchStrategies(patchStrategies), nodeutils.WithPatchStrategyKey(fmt.Sprintf("%v.%v.%v", KeyClientOptions, KeyCLI, KeyRepositories)))
}

func deleteCLIRepository(node *yaml.Node, name string) error {
	// Find the cli repositories node in the yaml node
	keys := []nodeutils.Key{
		{Name: KeyClientOptions, Type: yaml.MappingNode},
		{Name: KeyCLI, Type: yaml.MappingNode},
		{Name: KeyRepositories, Type: yaml.SequenceNode},
	}
	cliRepositoriesNode := nodeutils.FindNode(node.Content[0], nodeutils.WithKeys(keys))
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

func setRepository(repositoriesNode *yaml.Node, repository configapi.PluginRepository, patchStrategyOpts ...nodeutils.PatchStrategyOpts) (persist bool, err error) {
	newNode, err := convertPluginRepositoryToNode(&repository)
	if err != nil {
		return persist, err
	}

	exists := false
	var result []*yaml.Node

	repositoryType, repositoryName := getRepositoryTypeAndName(repository)
	if repositoryType == "" || repositoryName == "" {
		return persist, errors.New("not found")
	}

	// loop through the repositories seqence node
	for _, repositoryNode := range repositoriesNode.Content {
		// find the repository matching by repository type
		if repositoryIndex := nodeutils.GetNodeIndex(repositoryNode.Content, repositoryType); repositoryIndex != -1 {
			// find the repository matching by name
			if repositoryFieldIndex := nodeutils.GetNodeIndex(repositoryNode.Content[repositoryIndex].Content, "name"); repositoryFieldIndex != -1 &&
				repositoryNode.Content[repositoryIndex].Content[repositoryFieldIndex].Value == repositoryName {
				exists = true
				// persist change only if it's not the same as existing node
				persist, err = nodeutils.NotEqual(newNode.Content[0], repositoryNode)
				if persist {
					// replace nodes specified in the patch strategy
					err = nodeutils.ReplaceNodes(newNode.Content[0], repositoryNode, patchStrategyOpts...)
					if err != nil {
						return false, err
					}
					// merge the new node into repository node
					err = nodeutils.MergeNodes(newNode.Content[0], repositoryNode)
					if err != nil {
						return false, err
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
