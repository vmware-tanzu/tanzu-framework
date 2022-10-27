package config

import (
	"fmt"
	"strings"

	"github.com/aunum/log"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config/nodeutils"

	"gopkg.in/yaml.v3"
)

func GetServer(name string) (*configapi.Server, error) {
	node, err := GetClientConfigNode()
	if err != nil {
		return nil, err
	}
	return getServer(node, name)
}

func ServerExists(name string) (bool, error) {
	node, err := GetClientConfigNode()
	if err != nil {
		return false, err
	}
	exists, err := getServer(node, name)
	return exists != nil, err
}

func GetCurrentServer() (*configapi.Server, error) {
	node, err := GetClientConfigNode()
	if err != nil {
		return nil, err
	}
	return getCurrentServer(node)
}

func SetCurrentServer(name string) error {
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()
	node, err := GetClientConfigNodeNoLock()
	if err != nil {
		return err
	}

	s, err := getServer(node, name)
	if err != nil {
		return err
	}

	persist, err := setCurrentServer(node, name)
	if err != nil {
		return err
	}
	if persist {
		err = PersistNode(node)
		if err != nil {
			return err
		}
	}

	// Front fill CurrentContext
	c := convertServerToContext(s)
	persist, err = setCurrentContext(node, c)
	if err != nil {
		return err
	}
	if persist {
		err = PersistNode(node)
		if err != nil {
			return err
		}
	}

	return nil

}

func RemoveCurrentServer(name string) error {
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()
	node, err := GetClientConfigNodeNoLock()
	if err != nil {
		return err
	}

	_, err = getServer(node, name)
	if err != nil {
		return err
	}

	err = removeCurrentServer(node, name)
	if err != nil {
		return err
	}

	// Front fill Context and CurrentContext
	c, err := getContext(node, name)
	if err != nil {
		return err
	}

	_, err = removeCurrentContext(node, c.Type)
	if err != nil {
		return err
	}

	return PersistNode(node)

}

// Deprecated:- Use SetServer
func PutServer(s *configapi.Server, setCurrent bool) error {
	return SetServer(s, setCurrent)
}

// Deprecated:- Use SetServer
func AddServer(s *configapi.Server, setCurrent bool) error {
	return SetServer(s, setCurrent)

}

func SetServer(s *configapi.Server, setCurrent bool) error {
	// Acquire tanzu config lock
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()

	node, err := GetClientConfigNodeNoLock()
	if err != nil {
		return err
	}

	persist, err := setServer(node, s)
	if err != nil {
		return err
	}

	if persist {
		err = PersistNode(node)
		if err != nil {
			return err
		}
	}

	if setCurrent {
		persist, err = setCurrentServer(node, s.Name)
		if err != nil {
			return err
		}
		if persist {
			err = PersistNode(node)
			if err != nil {
				return err
			}
		}
	}

	// Front fill Context and CurrentContext
	c := convertServerToContext(s)

	persist, err = setContext(node, c)
	if err != nil {
		return err
	}

	if persist {
		err = PersistNode(node)
		if err != nil {
			return err
		}
	}

	if setCurrent {
		persist, err = setCurrentContext(node, c)
		if err != nil {
			return err
		}
		if persist {
			err = PersistNode(node)
			if err != nil {
				return err
			}
		}
	}

	return nil

}

func RemoveServer(name string) error {
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()
	node, err := GetClientConfigNodeNoLock()
	if err != nil {
		return err
	}

	_, err = getServer(node, name)
	if err != nil {
		return err

	}

	err = removeCurrentServer(node, name)
	if err != nil {
		return err

	}

	_, err = removeServer(node, name)
	if err != nil {
		return err

	}

	// Front fill Context and CurrentContext
	c, err := getContext(node, name)
	if err != nil {
		return err

	}

	_, err = removeCurrentContext(node, c.Type)
	if err != nil {
		return err

	}

	_, err = removeContext(node, name)
	if err != nil {
		return err

	}

	return PersistNode(node)
}

// GetDiscoverySources returns all discovery sources
// Includes standalone discovery sources and if server is available
// it also includes context based discovery sources as well
func GetDiscoverySources(serverName string) []configapi.PluginDiscovery {
	server, err := GetServer(serverName)
	if err != nil {
		log.Warningf("unknown server '%s', Unable to get server based discovery sources: %s", serverName, err.Error())
		return []configapi.PluginDiscovery{}
	}

	discoverySources := server.DiscoverySources
	// If current server type is management-cluster, then add
	// the default kubernetes discovery endpoint pointing to the
	// management-cluster kubeconfig
	if server.Type == configapi.ManagementClusterServerType {
		defaultClusterK8sDiscovery := configapi.PluginDiscovery{
			Kubernetes: &configapi.KubernetesDiscovery{
				Name:    fmt.Sprintf("default-%s", serverName),
				Path:    server.ManagementClusterOpts.Path,
				Context: server.ManagementClusterOpts.Context,
			},
		}
		discoverySources = append(discoverySources, defaultClusterK8sDiscovery)
	}

	// If the current server type is global, then add the default REST endpoint
	// for the discovery service
	if server.Type == configapi.GlobalServerType && server.GlobalOpts != nil {
		defaultRestDiscovery := configapi.PluginDiscovery{
			REST: &configapi.GenericRESTDiscovery{
				Name:     fmt.Sprintf("default-%s", serverName),
				Endpoint: appendURLScheme(server.GlobalOpts.Endpoint),
				BasePath: "v1alpha1/system/binaries/plugins",
			},
		}
		discoverySources = append(discoverySources, defaultRestDiscovery)
	}

	return discoverySources
}

func appendURLScheme(endpoint string) string {
	e := strings.Split(endpoint, ":")[0]
	if !strings.Contains(e, "https") {
		return fmt.Sprintf("https://%s", e)
	}
	return e
}

func setCurrentServer(node *yaml.Node, name string) (persist bool, err error) {

	configOptions := func(c *nodeutils.Config) {
		c.ForceCreate = true
		c.Keys = []nodeutils.Key{
			{Name: KeyCurrentServer, Type: yaml.ScalarNode, Value: ""},
		}
	}

	currentServerNode, err := nodeutils.FindNode(node.Content[0], configOptions)
	if err != nil {
		return persist, err
	}
	if currentServerNode.Value != name {
		currentServerNode.Value = name
		persist = true
	}

	return persist, err
}

func getServer(node *yaml.Node, name string) (*configapi.Server, error) {
	cfg, err := nodeutils.ConvertFromNode[configapi.ClientConfig](node)

	if err != nil {
		return nil, err
	}

	for _, server := range cfg.KnownServers {
		if server.Name == name {
			return server, nil
		}
	}

	return nil, fmt.Errorf("could not find server %q", name)

}

func getCurrentServer(node *yaml.Node) (s *configapi.Server, err error) {
	cfg, err := nodeutils.ConvertFromNode[configapi.ClientConfig](node)
	if err != nil {
		return nil, err
	}
	for _, server := range cfg.KnownServers {
		if server.Name == cfg.CurrentServer {
			return server, nil
		}
	}
	return s, fmt.Errorf("current server %q not found in tanzu config", cfg.CurrentServer)
}

func removeCurrentServer(node *yaml.Node, name string) error {

	configOptions := func(c *nodeutils.Config) {
		c.Keys = []nodeutils.Key{
			{Name: KeyCurrentServer},
		}
	}

	currentServerNode, err := nodeutils.FindNode(node.Content[0], configOptions)
	if err != nil {
		return err
	}

	if currentServerNode == nil {
		return nil
	}

	if currentServerNode.Value == name {
		currentServerNode.Value = ""
	}

	return nil
}

func removeServer(node *yaml.Node, name string) (ok bool, err error) {

	configOptions := func(c *nodeutils.Config) {
		c.Keys = []nodeutils.Key{
			{Name: KeyServers},
		}
	}

	serversNode, err := nodeutils.FindNode(node.Content[0], configOptions)
	if err != nil {
		return true, err
	}

	var servers []*yaml.Node
	for _, serverNode := range serversNode.Content {
		if index := nodeutils.GetNodeIndex(serverNode.Content, "name"); index != -1 && serverNode.Content[index].Value == name {
			continue
		}
		servers = append(servers, serverNode)
	}

	serversNode.Content = servers

	return true, nil
}

func setServers(node *yaml.Node, servers []*configapi.Server) error {
	for _, server := range servers {
		_, err := setServer(node, server)
		if err != nil {
			return err
		}
	}
	return nil
}

func setServer(node *yaml.Node, s *configapi.Server) (persist bool, err error) {

	var persistDiscoverySources bool

	// convert server to node
	newServerNode, err := nodeutils.ConvertToNode[configapi.Server](s)
	if err != nil {
		return persist, err
	}

	configOptions := func(c *nodeutils.Config) {
		c.ForceCreate = true
		c.Keys = []nodeutils.Key{
			{Name: KeyServers, Type: yaml.SequenceNode},
		}
	}

	serversNode, err := nodeutils.FindNode(node.Content[0], configOptions)
	if err != nil {
		return persist, err
	}

	exists := false
	var result []*yaml.Node
	for _, serverNode := range serversNode.Content {
		if index := nodeutils.GetNodeIndex(serverNode.Content, "name"); index != -1 &&
			serverNode.Content[index].Value == s.Name {
			exists = true

			persist, err = nodeutils.NotEqual(newServerNode.Content[0], serverNode)
			if err != nil {
				return persist, err
			}
			if persist {
				err = nodeutils.MergeNodes(newServerNode.Content[0], serverNode)
				if err != nil {
					return false, err
				}
			}

			persistDiscoverySources, err = setDiscoverySources(serverNode, s.DiscoverySources)
			if err != nil {
				return persistDiscoverySources, err
			}

			if persistDiscoverySources {
				err = nodeutils.MergeNodes(newServerNode.Content[0], serverNode)
				if err != nil {
					return persistDiscoverySources, err
				}
			}

			result = append(result, serverNode)
			continue
		}
		result = append(result, serverNode)
	}

	if !exists {
		result = append(result, newServerNode.Content[0])
		persist = true
	}

	serversNode.Content = result

	return persistDiscoverySources || persist, err

}

// EndpointFromServer returns the endpoint from server.
func EndpointFromServer(s *configapi.Server) (endpoint string, err error) {
	switch s.Type {
	case configapi.ManagementClusterServerType:
		return s.ManagementClusterOpts.Endpoint, nil
	case configapi.GlobalServerType:
		return s.GlobalOpts.Endpoint, nil
	default:
		return endpoint, fmt.Errorf("unknown server type %q", s.Type)
	}
}
