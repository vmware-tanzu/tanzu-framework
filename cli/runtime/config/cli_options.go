package config

import (
	"gopkg.in/yaml.v3"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config/nodeutils"
)

func GetEdition() (string, error) {
	node, err := GetClientConfigNode()
	if err != nil {
		return "", err
	}
	return getEdition(node)

}

func SetEdition(val string) (persist bool, err error) {
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()
	node, err := GetClientConfigNodeNoLock()
	if err != nil {
		return persist, err
	}
	persist, err = setEdition(node, val)
	if err != nil {
		return persist, err
	}

	if persist {
		return persist, PersistNode(node)
	}

	return persist, err

}

func setEdition(node *yaml.Node, val string) (persist bool, err error) {
	configOptions := func(c *nodeutils.Config) {
		c.ForceCreate = true
		c.Keys = []nodeutils.Key{
			{Name: KeyClientOptions, Type: yaml.MappingNode},
			{Name: KeyCLI, Type: yaml.MappingNode},
			{Name: KeyEdition, Type: yaml.ScalarNode, Value: ""},
		}
	}
	editionNode, err := nodeutils.FindNode(node.Content[0], configOptions)
	if err != nil {
		return false, err
	}

	if editionNode.Value != val {
		editionNode.Value = val
		persist = true

	}

	return persist, err
}

func getEdition(node *yaml.Node) (string, error) {
	cfg, err := nodeutils.ConvertFromNode[configapi.ClientConfig](node)
	if err != nil {
		return "", err
	}

	if cfg != nil && cfg.ClientOptions != nil && cfg.ClientOptions.CLI != nil {
		return string(cfg.ClientOptions.CLI.Edition), nil
	}
	return "", nil
}

func setUnstableVersionSelector(node *yaml.Node, name string) (persist bool, err error) {

	configOptions := func(c *nodeutils.Config) {
		c.ForceCreate = true
		c.Keys = []nodeutils.Key{
			{Name: KeyClientOptions, Type: yaml.MappingNode},
			{Name: KeyCLI, Type: yaml.MappingNode},
			{Name: KeyUnstableVersionSelector, Type: yaml.ScalarNode, Value: ""},
		}
	}

	unstableVersionSelectorNode, err := nodeutils.FindNode(node.Content[0], configOptions)
	if err != nil {
		return persist, err
	}

	if unstableVersionSelectorNode.Value != name {
		unstableVersionSelectorNode.Value = name
		persist = true
	}

	return persist, err

}

func setBomRepo(node *yaml.Node, repo string) (persist bool, err error) {

	configOptions := func(c *nodeutils.Config) {
		c.ForceCreate = true
		c.Keys = []nodeutils.Key{
			{Name: KeyClientOptions, Type: yaml.MappingNode},
			{Name: KeyCLI, Type: yaml.MappingNode},
			{Name: KeyBomRepo, Type: yaml.ScalarNode, Value: repo},
		}
	}

	bomRepoNode, err := nodeutils.FindNode(node.Content[0], configOptions)
	if err != nil {
		return persist, err
	}

	if bomRepoNode.Value != repo {
		bomRepoNode.Value = repo
		persist = true
	}

	return persist, err

}

func setCompatibilityFilePath(node *yaml.Node, filepath string) (persist bool, err error) {

	configOptions := func(c *nodeutils.Config) {
		c.ForceCreate = true
		c.Keys = []nodeutils.Key{
			{Name: KeyClientOptions, Type: yaml.MappingNode},
			{Name: KeyCLI, Type: yaml.MappingNode},
			{Name: KeyCompatibilityFilePath, Type: yaml.ScalarNode, Value: ""},
		}
	}

	compatibilityFilePathNode, err := nodeutils.FindNode(node.Content[0], configOptions)
	if err != nil {
		return persist, err
	}
	if compatibilityFilePathNode.Value != filepath {
		compatibilityFilePathNode.Value = filepath
		persist = true
	}

	return persist, err

}
