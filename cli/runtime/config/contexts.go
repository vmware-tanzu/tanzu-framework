// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"

	"gopkg.in/yaml.v3"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config/nodeutils"
)

func GetContext(name string) (*configapi.Context, error) {
	node, err := GetClientConfigNode()
	if err != nil {
		return nil, err
	}

	return getContext(node, name)
}

// Deprecated:- Use SetContext
func AddContext(c *configapi.Context, setCurrent bool) error {
	return SetContext(c, setCurrent)
}

func SetContext(c *configapi.Context, setCurrent bool) error {
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()
	node, err := GetClientConfigNodeNoLock()
	if err != nil {
		return err
	}

	persist, err := setContext(node, c)
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

	s := convertContextToServer(c)

	persist, err = setServer(node, s)
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

	return err

}

func RemoveContext(name string) error {
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()
	node, err := GetClientConfigNodeNoLock()
	if err != nil {
		return err
	}

	ctx, err := getContext(node, name)
	if err != nil {
		return err
	}

	_, err = removeCurrentContext(node, ctx.Type)
	if err != nil {
		return err
	}
	_, err = removeContext(node, name)
	if err != nil {
		return err
	}

	_, err = removeServer(node, name)
	if err != nil {
		return err
	}

	err = removeCurrentServer(node, name)
	if err != nil {
		return err
	}

	return PersistNode(node)
}

func ContextExists(name string) (bool, error) {
	_, err := GetContext(name)
	return err == nil, nil
}

func GetCurrentContext(ctxType configapi.ContextType) (c *configapi.Context, err error) {
	node, err := GetClientConfigNode()
	if err != nil {
		return nil, err
	}

	return getCurrentContext(node, ctxType)
}

func SetCurrentContext(name string) error {
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()
	node, err := GetClientConfigNodeNoLock()
	if err != nil {
		return err
	}

	ctx, err := getContext(node, name)
	if err != nil {
		return err
	}

	persist, err := setCurrentContext(node, ctx)
	if err != nil {
		return err
	}

	if persist {
		err = PersistNode(node)
		if err != nil {
			return err
		}
	}

	persist, err = setCurrentServer(node, name)
	if err != nil {
		return err
	}

	if persist {
		err = PersistNode(node)
		if err != nil {
			return err
		}
	}
	return err
}

func RemoveCurrentContext(ctxType configapi.ContextType) error {
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()
	node, err := GetClientConfigNodeNoLock()
	if err != nil {
		return err
	}

	_, err = removeCurrentContext(node, ctxType)
	if err != nil {
		return err
	}

	c, err := getCurrentContext(node, ctxType)
	if err != nil {
		return err
	}

	err = removeCurrentServer(node, c.Name)
	if err != nil {
		return err
	}

	return PersistNode(node)

}

func EndpointFromContext(s *configapi.Context) (endpoint string, err error) {
	switch s.Type {
	case configapi.CtxTypeK8s:
		return s.ClusterOpts.Endpoint, nil
	case configapi.CtxTypeTMC:
		return s.GlobalOpts.Endpoint, nil
	default:
		return endpoint, fmt.Errorf("unknown server type %q", s.Type)
	}
}

func getContext(node *yaml.Node, name string) (*configapi.Context, error) {

	cfg, err := nodeutils.ConvertFromNode[configapi.ClientConfig](node)
	if err != nil {
		return nil, err
	}

	for _, ctx := range cfg.KnownContexts {
		if ctx.Name == name {
			return ctx, nil
		}
	}
	return nil, fmt.Errorf("context %v not found", name)

}

func getCurrentContext(node *yaml.Node, ctxType configapi.ContextType) (*configapi.Context, error) {
	cfg, err := nodeutils.ConvertFromNode[configapi.ClientConfig](node)
	if err != nil {
		return nil, err
	}
	return cfg.GetCurrentContext(ctxType)
}

func setContexts(node *yaml.Node, contexts []*configapi.Context) (err error) {
	for _, c := range contexts {
		_, err = setContext(node, c)
		if err != nil {
			return err
		}
	}
	return err
}

func setContext(node *yaml.Node, ctx *configapi.Context) (persist bool, err error) {

	var persistDiscoverySources bool

	// convert context to node
	newContextNode, err := nodeutils.ConvertToNode[configapi.Context](ctx)
	if err != nil {
		return persist, err
	}

	configOptions := func(c *nodeutils.Config) {
		c.ForceCreate = true
		c.Keys = []nodeutils.Key{
			{Name: KeyContexts, Type: yaml.SequenceNode},
		}
	}

	contextsNode, err := nodeutils.FindNode(node.Content[0], configOptions)
	if err != nil {
		return persist, err
	}

	exists := false
	var result []*yaml.Node
	for _, contextNode := range contextsNode.Content {
		if index := nodeutils.GetNodeIndex(contextNode.Content, "name"); index != -1 &&
			contextNode.Content[index].Value == ctx.Name {
			exists = true

			persist, err = nodeutils.NotEqual(newContextNode.Content[0], contextNode)
			if err != nil {
				return persist, err
			}
			if persist {
				err = nodeutils.MergeNodes(newContextNode.Content[0], contextNode)
				if err != nil {
					return persist, err
				}
			}

			persistDiscoverySources, err = setDiscoverySources(contextNode, ctx.DiscoverySources)
			if err != nil {
				return persistDiscoverySources, err
			}
			if persistDiscoverySources {
				err = nodeutils.MergeNodes(newContextNode.Content[0], contextNode)
				if err != nil {
					return persistDiscoverySources, err
				}
			}

			result = append(result, contextNode)
			continue
		}
		result = append(result, contextNode)
	}

	if !exists {
		result = append(result, newContextNode.Content[0])
		persist = true
	}

	contextsNode.Content = result

	return persist, err

}

func setCurrentContext(node *yaml.Node, ctx *configapi.Context) (persist bool, err error) {

	configOptions := func(c *nodeutils.Config) {
		c.ForceCreate = true
		c.Keys = []nodeutils.Key{
			{Name: KeyCurrentContext, Type: yaml.MappingNode},
		}
	}

	currentContextNode, err := nodeutils.FindNode(node.Content[0], configOptions)
	if err != nil {
		return persist, err
	}

	if index := nodeutils.GetNodeIndex(currentContextNode.Content, string(ctx.Type)); index != -1 {
		if currentContextNode.Content[index].Value != ctx.Name {
			currentContextNode.Content[index].Value = ctx.Name
			persist = true
		}
	} else {
		currentContextNode.Content = append(currentContextNode.Content, nodeutils.CreateScalarNode(string(ctx.Type), ctx.Name)...)
		persist = true
	}

	return persist, err

}

func removeCurrentContext(node *yaml.Node, ctxType configapi.ContextType) (ok bool, err error) {

	configOptions := func(c *nodeutils.Config) {
		c.Keys = []nodeutils.Key{
			{Name: KeyCurrentContext},
			{Name: string(ctxType)},
		}
	}

	currentContextNode, err := nodeutils.FindNode(node.Content[0], configOptions)
	if err != nil {
		return false, err
	}

	if currentContextNode == nil {
		return true, nil
	}

	currentContextNode.Value = ""

	return true, nil
}

func removeContext(node *yaml.Node, name string) (ok bool, err error) {

	configOptions := func(c *nodeutils.Config) {
		c.Keys = []nodeutils.Key{
			{Name: KeyContexts},
		}
	}

	contextsNode, err := nodeutils.FindNode(node.Content[0], configOptions)

	if err != nil {
		return false, err
	}

	if contextsNode == nil {
		return true, nil
	}

	var contexts []*yaml.Node
	for _, contextNode := range contextsNode.Content {
		if index := nodeutils.GetNodeIndex(contextNode.Content, "name"); index != -1 && contextNode.Content[index].Value == name {
			continue
		}
		contexts = append(contexts, contextNode)
	}

	contextsNode.Content = contexts

	return true, nil
}
