package config

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config/nodeutils"
)

func GetCLIDiscoverySources() ([]configapi.PluginDiscovery, error) {
	node, err := GetClientConfigNode()
	if err != nil {
		return nil, err
	}
	return getCLIDiscoverySources(node)
}

func GetCLIDiscoverySource(name string) (*configapi.PluginDiscovery, error) {
	node, err := GetClientConfigNode()
	if err != nil {
		return nil, err
	}
	return getCLIDiscoverySource(node, name)
}

func SetCLIDiscoverySource(discoverySource configapi.PluginDiscovery) (persist bool, err error) {
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()
	node, err := GetClientConfigNodeNoLock()
	if err != nil {
		return persist, err
	}
	persist, err = setCLIDiscoverySource(node, discoverySource)
	if err != nil {
		return persist, err
	}
	if persist {
		return persist, PersistNode(node)
	}
	return persist, err
}

func DeleteCLIDiscoverySource(name string) error {
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()
	node, err := GetClientConfigNodeNoLock()
	if err != nil {
		return err
	}
	err = deleteCLIDiscoverySource(node, name)
	if err != nil {
		return err
	}
	return PersistNode(node)
}

func getCLIDiscoverySources(node *yaml.Node) ([]configapi.PluginDiscovery, error) {
	cfg, err := nodeutils.ConvertFromNode[configapi.ClientConfig](node)
	if err != nil {
		return nil, err
	}

	if cfg.ClientOptions != nil && cfg.ClientOptions.CLI != nil && cfg.ClientOptions.CLI.DiscoverySources != nil {
		return cfg.ClientOptions.CLI.DiscoverySources, nil
	}

	return nil, errors.New("cli discovery sources not found")

}

func getCLIDiscoverySource(node *yaml.Node, name string) (*configapi.PluginDiscovery, error) {
	cfg, err := nodeutils.ConvertFromNode[configapi.ClientConfig](node)
	if err != nil {
		return nil, err
	}
	if cfg.ClientOptions != nil && cfg.ClientOptions.CLI != nil && cfg.ClientOptions.CLI.DiscoverySources != nil {
		for _, discoverySource := range cfg.ClientOptions.CLI.DiscoverySources {
			_, discoverySourceName := getDiscoverySourceTypeAndName(discoverySource)
			if discoverySourceName == name {
				return &discoverySource, nil
			}
		}
	}
	return nil, errors.New("cli discovery source not found")
}
func setCLIDiscoverySources(node *yaml.Node, discoverySources []configapi.PluginDiscovery) (err error) {
	for _, discoverySource := range discoverySources {
		_, err = setCLIDiscoverySource(node, discoverySource)
		if err != nil {
			return err
		}
	}
	return err
}

func setCLIDiscoverySource(node *yaml.Node, discoverySource configapi.PluginDiscovery) (persist bool, err error) {
	configOptions := func(c *nodeutils.Config) {
		c.ForceCreate = true
		c.Keys = []nodeutils.Key{
			{Name: KeyClientOptions, Type: yaml.MappingNode},
			{Name: KeyCLI, Type: yaml.MappingNode},
			{Name: KeyDiscoverySources, Type: yaml.SequenceNode},
		}
	}

	discoverySourcesNode, err := nodeutils.FindNode(node.Content[0], configOptions)
	if err != nil {
		return persist, err
	}

	persist, err = setDiscoverySource(discoverySourcesNode, discoverySource)
	if err != nil {
		return persist, err
	}
	return persist, err
}

func deleteCLIDiscoverySource(node *yaml.Node, name string) error {

	configOptions := func(c *nodeutils.Config) {
		c.ForceCreate = false
		c.Keys = []nodeutils.Key{
			{Name: KeyClientOptions},
			{Name: KeyCLI},
			{Name: KeyDiscoverySources},
		}
	}

	cliDiscoverySourcesNode, err := nodeutils.FindNode(node.Content[0], configOptions)
	if err != nil {
		return err
	}

	if cliDiscoverySourcesNode == nil {
		return nil
	}

	discoverySource, err := getCLIDiscoverySource(node, name)
	if err != nil {
		return nil
	}

	var result []*yaml.Node
	for _, discoverySourceNode := range cliDiscoverySourcesNode.Content {
		discoverySourceType, discoverySourceName := getDiscoverySourceTypeAndName(*discoverySource)
		if discoverySourceIndex := nodeutils.GetNodeIndex(discoverySourceNode.Content, discoverySourceType); discoverySourceIndex != -1 {
			if discoverySourceFieldIndex := nodeutils.GetNodeIndex(discoverySourceNode.Content[discoverySourceIndex].Content, "name"); discoverySourceFieldIndex != -1 && discoverySourceNode.Content[discoverySourceIndex].Content[discoverySourceFieldIndex].Value == discoverySourceName {
				continue
			}
		} else {
			result = append(result, discoverySourceNode)
		}
	}

	cliDiscoverySourcesNode.Style = 0
	cliDiscoverySourcesNode.Content = result

	return nil

}
