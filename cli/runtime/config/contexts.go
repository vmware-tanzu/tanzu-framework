// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"

	"gopkg.in/yaml.v3"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config/nodeutils"
)

// GetContext retrieves the context by name
func GetContext(name string) (*configapi.Context, error) {
	node, err := getClientConfigNode()
	if err != nil {
		return nil, err
	}
	return getContext(node, name)
}

// AddContext add or update context and currentContext
func AddContext(c *configapi.Context, setCurrent bool) error {
	return SetContext(c, setCurrent)
}

// SetContext add or update context and currentContext
//
//nolint:gocyclo
func SetContext(c *configapi.Context, setCurrent bool) error {
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()
	node, err := getClientConfigNodeNoLock()
	if err != nil {
		return err
	}
	persist, err := setContext(node, c)
	if err != nil {
		return err
	}
	if persist {
		err = persistNode(node)
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
			err = persistNode(node)
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
		err = persistNode(node)
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
			err = persistNode(node)
			if err != nil {
				return err
			}
		}
	}
	return err
}

// DeleteContext delete a context by name
func DeleteContext(name string) error {
	return RemoveContext(name)
}

// RemoveContext delete a context by name
func RemoveContext(name string) error {
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()
	node, err := getClientConfigNodeNoLock()
	if err != nil {
		return err
	}
	ctx, err := getContext(node, name)
	if err != nil {
		return err
	}
	err = removeCurrentContext(node, ctx)
	if err != nil {
		return err
	}
	err = removeContext(node, name)
	if err != nil {
		return err
	}
	err = removeServer(node, name)
	if err != nil {
		return err
	}
	err = removeCurrentServer(node, name)
	if err != nil {
		return err
	}
	return persistNode(node)
}

// ContextExists checks if context by name already exists
func ContextExists(name string) (bool, error) {
	exists, _ := GetContext(name)
	return exists != nil, nil
}

// GetCurrentContext retrieves the current context for the specified context type
func GetCurrentContext(ctxType configapi.ContextType) (c *configapi.Context, err error) {
	node, err := getClientConfigNode()
	if err != nil {
		return nil, err
	}
	return getCurrentContext(node, ctxType)
}

// SetCurrentContext sets the current context to the specified name if context is present
func SetCurrentContext(name string) error {
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()
	node, err := getClientConfigNodeNoLock()
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
		err = persistNode(node)
		if err != nil {
			return err
		}
	}
	persist, err = setCurrentServer(node, name)
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

// RemoveCurrentContext removed the current context of specified context type
func RemoveCurrentContext(ctxType configapi.ContextType) error {
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()
	node, err := getClientConfigNodeNoLock()
	if err != nil {
		return err
	}
	ctx := &configapi.Context{Type: ctxType}
	err = removeCurrentContext(node, ctx)
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
	return persistNode(node)
}

// EndpointFromContext retrieved the endpoint from the specified context
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
	cfg, err := convertNodeToClientConfig(node)
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
	cfg, err := convertNodeToClientConfig(node)
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

//nolint:dupl
func setContext(node *yaml.Node, ctx *configapi.Context) (persist bool, err error) {
	var persistDiscoverySources bool
	// convert context to node
	newContextNode, err := convertContextToNode(ctx)
	if err != nil {
		return persist, err
	}

	// config options to retrieve the contexts stanza from config file
	configOptions := func(c *nodeutils.Config) {
		c.ForceCreate = true
		c.Keys = []nodeutils.Key{
			{Name: KeyContexts, Type: yaml.SequenceNode},
		}
	}
	// find the contexts node from the root node
	contextsNode := nodeutils.FindNode(node.Content[0], configOptions)
	if contextsNode == nil {
		return persist, err
	}
	exists := false
	var result []*yaml.Node
	for _, contextNode := range contextsNode.Content {
		if index := nodeutils.GetNodeIndex(contextNode.Content, "name"); index != -1 &&
			contextNode.Content[index].Value == ctx.Name {
			exists = true
			// check if the updated context is not same as exisiting context
			persist, err = nodeutils.NotEqual(newContextNode.Content[0], contextNode)
			if err != nil {
				return persist, err
			}
			// merge the nodes only if the nodes are not equal
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
			// merge the discovery sources to context
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
	return persistDiscoverySources || persist, err
}

func setCurrentContext(node *yaml.Node, ctx *configapi.Context) (persist bool, err error) {
	configOptions := func(c *nodeutils.Config) {
		c.ForceCreate = true
		c.Keys = []nodeutils.Key{
			{Name: KeyCurrentContext, Type: yaml.MappingNode},
		}
	}
	currentContextNode := nodeutils.FindNode(node.Content[0], configOptions)
	if currentContextNode == nil {
		return persist, nodeutils.ErrNodeNotFound
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

func removeCurrentContext(node *yaml.Node, ctx *configapi.Context) error {
	configOptions := func(c *nodeutils.Config) {
		c.Keys = []nodeutils.Key{
			{Name: KeyCurrentContext},
			{Name: string(ctx.Type)},
		}
	}
	currentContextNode := nodeutils.FindNode(node.Content[0], configOptions)
	if currentContextNode == nil {
		return nil
	}
	if currentContextNode.Value == ctx.Name || ctx.Name == "" {
		currentContextNode.Value = ""
	}
	return nil
}

func removeContext(node *yaml.Node, name string) error {
	configOptions := func(c *nodeutils.Config) {
		c.Keys = []nodeutils.Key{
			{Name: KeyContexts},
		}
	}
	contextsNode := nodeutils.FindNode(node.Content[0], configOptions)
	if contextsNode == nil {
		return nil
	}
	var contexts []*yaml.Node
	for _, contextNode := range contextsNode.Content {
		if index := nodeutils.GetNodeIndex(contextNode.Content, "name"); index != -1 && contextNode.Content[index].Value == name {
			continue
		}
		contexts = append(contexts, contextNode)
	}
	contextsNode.Content = contexts
	return nil
}
