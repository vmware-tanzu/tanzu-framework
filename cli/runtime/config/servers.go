// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"strings"

	"github.com/aunum/log"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config/nodeutils"

	"gopkg.in/yaml.v3"
)

// GetServer retrieves server by name
func GetServer(name string) (*configapi.Server, error) {
	// Retrieve client config node
	node, err := getClientConfigNode()
	if err != nil {
		return nil, err
	}
	return getServer(node, name)
}

// ServerExists checks if server by specified name is present in config
func ServerExists(name string) (bool, error) {
	exists, _ := GetServer(name)
	return exists != nil, nil
}

// GetCurrentServer retrieves the current server
func GetCurrentServer() (*configapi.Server, error) {
	// Retrieve client config node
	node, err := getClientConfigNode()
	if err != nil {
		return nil, err
	}
	return getCurrentServer(node)
}

// SetCurrentServer add or update current server
func SetCurrentServer(name string) error {
	// Retrieve client config node
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()
	node, err := getClientConfigNodeNoLock()
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
		err = persistConfig(node)
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
		err = persistConfig(node)
		if err != nil {
			return err
		}
	}
	return nil
}

// RemoveCurrentServer removes the current server if server exists by specified name
func RemoveCurrentServer(name string) error {
	// Retrieve client config node
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()
	node, err := getClientConfigNodeNoLock()
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
	err = removeCurrentContext(node, c)
	if err != nil {
		return err
	}
	return persistConfig(node)
}

// PutServer add or update server and currentServer
func PutServer(s *configapi.Server, setCurrent bool) error {
	return SetServer(s, setCurrent)
}

// AddServer add or update server and currentServer
func AddServer(s *configapi.Server, setCurrent bool) error {
	return SetServer(s, setCurrent)
}

// SetServer add or update server and currentServer
func SetServer(s *configapi.Server, setCurrent bool) error {
	// Acquire tanzu config lock
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()
	node, err := getClientConfigNodeNoLock()
	if err != nil {
		return err
	}
	persist, err := setServer(node, s)
	if err != nil {
		return err
	}
	if persist {
		err = persistConfig(node)
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
			err = persistConfig(node)
			if err != nil {
				return err
			}
		}
	}

	err = frontFillContexts(s, setCurrent, node)
	if err != nil {
		return err
	}

	return nil
}

func frontFillContexts(s *configapi.Server, setCurrent bool, node *yaml.Node) error {
	// Front fill Context and CurrentContext
	c := convertServerToContext(s)
	persist, err := setContext(node, c)
	if err != nil {
		return err
	}
	if persist {
		err = persistConfig(node)
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
			err = persistConfig(node)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// DeleteServer deletes the server specified by name
func DeleteServer(name string) error {
	return RemoveServer(name)
}

// RemoveServer removed the server by name
func RemoveServer(name string) error {
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()
	node, err := getClientConfigNodeNoLock()
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
	err = removeServer(node, name)
	if err != nil {
		return err
	}
	// Front fill Context and CurrentContext
	c, err := getContext(node, name)
	if err != nil {
		return err
	}
	err = removeCurrentContext(node, c)
	if err != nil {
		return err
	}
	err = removeContext(node, name)
	if err != nil {
		return err
	}
	return persistConfig(node)
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
	// find current server node
	keys := []nodeutils.Key{
		{Name: KeyCurrentServer, Type: yaml.ScalarNode, Value: ""},
	}
	currentServerNode := nodeutils.FindNode(node.Content[0], nodeutils.WithForceCreate(), nodeutils.WithKeys(keys))
	if currentServerNode == nil {
		return persist, nodeutils.ErrNodeNotFound
	}
	if currentServerNode.Value != name {
		currentServerNode.Value = name
		persist = true
	}
	return persist, err
}

func getServer(node *yaml.Node, name string) (*configapi.Server, error) {
	cfg, err := convertNodeToClientConfig(node)
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
	cfg, err := convertNodeToClientConfig(node)
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
	// find current server node
	keys := []nodeutils.Key{
		{Name: KeyCurrentServer},
	}
	currentServerNode := nodeutils.FindNode(node.Content[0], nodeutils.WithKeys(keys))
	if currentServerNode == nil {
		return nil
	}
	if currentServerNode.Value == name {
		currentServerNode.Value = ""
	}
	return nil
}

func removeServer(node *yaml.Node, name string) error {
	// find servers node
	keys := []nodeutils.Key{
		{Name: KeyServers},
	}
	serversNode := nodeutils.FindNode(node.Content[0], nodeutils.WithKeys(keys))
	if serversNode == nil {
		return nodeutils.ErrNodeNotFound
	}
	var servers []*yaml.Node
	for _, serverNode := range serversNode.Content {
		if index := nodeutils.GetNodeIndex(serverNode.Content, "name"); index != -1 && serverNode.Content[index].Value == name {
			continue
		}
		servers = append(servers, serverNode)
	}
	serversNode.Content = servers
	return nil
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
	// Get Patch Strategies
	patchStrategies, err := GetConfigMetadataPatchStrategy()
	if err != nil {
		patchStrategies = make(map[string]string)
	}
	var persistDiscoverySources bool

	// convert server to node
	newServerNode, err := convertServerToNode(s)
	if err != nil {
		return persist, err
	}

	// find servers node
	keys := []nodeutils.Key{
		{Name: KeyServers, Type: yaml.SequenceNode},
	}
	serversNode := nodeutils.FindNode(node.Content[0], nodeutils.WithForceCreate(), nodeutils.WithKeys(keys))
	if serversNode == nil {
		return persist, nodeutils.ErrNodeNotFound
	}
	exists := false
	var result []*yaml.Node
	//nolint: dupl
	for _, serverNode := range serversNode.Content {
		if index := nodeutils.GetNodeIndex(serverNode.Content, "name"); index != -1 &&
			serverNode.Content[index].Value == s.Name {
			exists = true
			_, err = nodeutils.ReplaceNodes(newServerNode.Content[0], serverNode, nodeutils.WithPatchStrategyKey(KeyServers), nodeutils.WithPatchStrategies(patchStrategies))
			if err != nil {
				return false, err
			}
			persist, err = nodeutils.MergeNodes(newServerNode.Content[0], serverNode)
			if err != nil {
				return false, err
			}
			// add or update discovery sources of server
			persistDiscoverySources, err = setDiscoverySources(serverNode, s.DiscoverySources, nodeutils.WithPatchStrategyKey(fmt.Sprintf("%v.%v", KeyServers, KeyDiscoverySources)), nodeutils.WithPatchStrategies(patchStrategies))
			if err != nil {
				return false, err
			}
			if persistDiscoverySources {
				_, err = nodeutils.MergeNodes(newServerNode.Content[0], serverNode)
				if err != nil {
					return false, err
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
