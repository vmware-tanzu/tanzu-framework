// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package region provides region context functionalities
package region

import (
	"os"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v3"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
)

const yamlSeqTag = "!!seq"

type manager struct {
	tkgConfigPath string
}

// New creates regional manager
func New(tkgConfigPath string) (Manager, error) {
	if _, err := os.Stat(tkgConfigPath); err != nil {
		return nil, err
	}

	return &manager{
		tkgConfigPath: tkgConfigPath,
	}, nil
}

// ListRegionContexts lists all the regions in tkg config file
func (m *manager) ListRegionContexts() ([]RegionContext, error) {
	var regions []RegionContext

	// load tkg config
	tkgConfigNode, err := m.loadTkgConfig()
	if err != nil {
		return regions, errors.Wrapf(err, "unable to unmarshal tkg configuration file %s", m.tkgConfigPath)
	}

	regionsNode := m.findNodeUnderTkg(&tkgConfigNode, constants.KeyRegions)
	if regionsNode == nil {
		return regions, nil
	}

	regionsByte, err := yaml.Marshal(regionsNode.Content)
	if err != nil {
		return regions, errors.Wrap(err, "unable to parse regions")
	}

	err = yaml.Unmarshal(regionsByte, &regions)
	if err != nil {
		return regions, err
	}

	currentContextName, _ := m.getCurrentContextName()
	for i := range regions {
		if regions[i].ContextName == currentContextName {
			regions[i].IsCurrentContext = true
		}

		if regions[i].Status == "" {
			regions[i].Status = Success
		}
	}

	return regions, err
}

// SaveRegionContext saves a new region object into the tkg config file
func (m *manager) SaveRegionContext(region RegionContext) error {
	if _, err := m.getRegionContext(region.ClusterName, region.ContextName); err == nil {
		return errors.Errorf("management cluster %s with context %s already exists", region.ClusterName, region.ContextName)
	}

	tkgConfigNode, err := m.loadTkgConfig()
	if err != nil {
		return errors.Wrapf(err, "unable to unmarshal tkg configuration file %s", m.tkgConfigPath)
	}

	// find the regions node under tkg node, create one if not found
	regionsNode := m.findNodeUnderTkg(&tkgConfigNode, constants.KeyRegions)
	if regionsNode == nil {
		regionsNode, err = m.createNodeUnderTkg(&tkgConfigNode, constants.KeyRegions, "", yaml.SequenceNode)
		if err != nil {
			return err
		}
	}

	regionsNode.Kind = yaml.SequenceNode
	regionsNode.Tag = yamlSeqTag

	// convert region into yaml node
	regionByte, err := yaml.Marshal(region)
	if err != nil {
		return errors.Wrap(err, "unable to marshal the new management cluster object into yaml")
	}

	node := yaml.Node{}
	err = yaml.Unmarshal(regionByte, &node)
	if err != nil {
		return errors.Wrap(err, "unable to marshal the new management cluster object into yaml")
	}
	// append new region in regions section
	regionsNode.Content = append(regionsNode.Content, node.Content[0])

	out, err := yaml.Marshal(&tkgConfigNode)
	if err != nil {
		return err
	}
	err = os.WriteFile(m.tkgConfigPath, out, constants.ConfigFilePermissions)
	if err != nil {
		return err
	}

	return nil
}

// UpsertRegionContext saves/updates a region object in the tkg config file
func (m *manager) UpsertRegionContext(region RegionContext) error {
	tkgConfigNode, err := m.loadTkgConfig()
	if err != nil {
		return errors.Wrapf(err, "unable to unmarshal tkg configuration file %s", m.tkgConfigPath)
	}

	// find the regions node under tkg node, create one if not found
	regionsNode := m.findNodeUnderTkg(&tkgConfigNode, constants.KeyRegions)
	if regionsNode == nil {
		regionsNode, err = m.createNodeUnderTkg(&tkgConfigNode, constants.KeyRegions, "", yaml.SequenceNode)
		if err != nil {
			return err
		}
	}

	// convert region into yaml node
	regionByte, err := yaml.Marshal(region)
	if err != nil {
		return errors.Wrap(err, "unable to marshal the management cluster objects into yaml")
	}

	updatedNode := yaml.Node{}
	err = yaml.Unmarshal(regionByte, &updatedNode)
	if err != nil {
		return errors.Wrap(err, "unable to unmarshal the management cluster object into yaml")
	}

	// update the region info with given clusterName and contextName if already exists
	result := []*yaml.Node{}
	found := false
	for _, node := range regionsNode.Content {
		nameIndex := getNodeIndex(node.Content, constants.KeyRegionName)
		contextIndex := getNodeIndex(node.Content, constants.KeyRegionContext)
		if nameIndex != -1 && node.Content[nameIndex].Value == region.ClusterName &&
			contextIndex != -1 && node.Content[contextIndex].Value == region.ContextName {
			result = append(result, updatedNode.Content[0])
			found = true
			continue
		}
		result = append(result, node)
	}
	// add the region info if it doesn't exist
	if !found {
		result = append(result, updatedNode.Content[0])
	}

	regionsNode.Content = result
	regionsNode.Kind = yaml.SequenceNode
	regionsNode.Tag = yamlSeqTag

	out, err := yaml.Marshal(&tkgConfigNode)
	if err != nil {
		return err
	}
	err = os.WriteFile(m.tkgConfigPath, out, constants.ConfigFilePermissions)
	if err != nil {
		return err
	}

	return nil
}

// DeleteRegionContext deletes a management cluster data from tkg config file
func (m *manager) DeleteRegionContext(clusterName string) error {
	tkgConfigNode, err := m.loadTkgConfig()
	if err != nil {
		return errors.Wrapf(err, "unable to unmarshal tkg configuration file %s", m.tkgConfigPath)
	}
	regionsNode := m.findNodeUnderTkg(&tkgConfigNode, constants.KeyRegions)
	if regionsNode == nil {
		return nil
	}

	currentContext := ""
	currentContextNode := m.findNodeUnderTkg(&tkgConfigNode, constants.KeyCurrentRegionContext)

	if currentContextNode != nil {
		currentContext = currentContextNode.Value
	}

	// removed the all region info with given clusterName, regardless the context
	// unset current context if the current context is deleted
	result := []*yaml.Node{}
	for _, node := range regionsNode.Content {
		if index := getNodeIndex(node.Content, constants.KeyRegionName); index != -1 && node.Content[index].Value == clusterName {
			if index := getNodeIndex(node.Content, constants.KeyRegionContext); index != -1 && node.Content[index].Value == currentContext {
				currentContextNode.Value = ""
			}
			continue
		}
		result = append(result, node)
	}

	if len(result) == 0 {
		regionsNode.Kind = yaml.ScalarNode
		regionsNode.Tag = yamlSeqTag
	} else {
		regionsNode.Content = result
	}

	out, err := yaml.Marshal(&tkgConfigNode)
	if err != nil {
		return err
	}
	err = os.WriteFile(m.tkgConfigPath, out, constants.ConfigFilePermissions)
	if err != nil {
		return err
	}

	return nil
}

// GetRegionContext gets a management cluster info by cluster name, contextName is an optional parameter
func (m *manager) getRegionContext(clusterName, contextName string) (RegionContext, error) {
	regions, err := m.ListRegionContexts()
	if err != nil {
		return RegionContext{}, err
	}
	var result []RegionContext

	for _, r := range regions {
		if r.ClusterName == clusterName {
			result = append(result, r)
		}
	}

	if len(result) == 0 {
		return RegionContext{}, errors.Errorf("cannot find management cluster %s", clusterName)
	}

	if contextName != "" {
		for _, r := range result {
			if r.ContextName == contextName {
				return r, nil
			}
		}
		return RegionContext{}, errors.Errorf("cannot find management cluster %s with context %s", clusterName, contextName)
	}

	if len(result) == 1 {
		return result[0], nil
	}

	return RegionContext{}, errors.Errorf("multiple contexts are found for cluster %s, please specify a context name", clusterName)
}

// SetCurrentContext sets current management cluster context into tkg config file
func (m *manager) SetCurrentContext(clusterName, contextName string) error {
	region, err := m.getRegionContext(clusterName, contextName)
	if err != nil {
		return errors.Wrapf(err, "unable to set current management cluster to %s", clusterName)
	}

	tkgConfigNode, err := m.loadTkgConfig()
	if err != nil {
		return errors.Wrapf(err, "unable to unmarshal tkg configuration file %s", m.tkgConfigPath)
	}

	currentContextNode := m.findNodeUnderTkg(&tkgConfigNode, constants.KeyCurrentRegionContext)

	if currentContextNode == nil {
		_, err = m.createNodeUnderTkg(&tkgConfigNode, constants.KeyCurrentRegionContext, region.ContextName, yaml.ScalarNode)
		if err != nil {
			return errors.Wrap(err, "unable to set current context in tkg config file")
		}
	} else {
		currentContextNode.Tag = yamlSeqTag
		currentContextNode.Value = region.ContextName
	}

	out, err := yaml.Marshal(&tkgConfigNode)
	if err != nil {
		return err
	}
	err = os.WriteFile(m.tkgConfigPath, out, constants.ConfigFilePermissions)
	if err != nil {
		return err
	}

	return nil
}

func (m *manager) getCurrentContextName() (string, error) {
	tkgConfigNode, err := m.loadTkgConfig()
	if err != nil {
		return "", errors.Wrapf(err, "unable to unmarshal tkg configuration file %s", m.tkgConfigPath)
	}

	currentContextNode := m.findNodeUnderTkg(&tkgConfigNode, constants.KeyCurrentRegionContext)
	if currentContextNode == nil || currentContextNode.Value == "" {
		return "", errors.New("unable to find current context node")
	}
	return currentContextNode.Value, nil
}

func (m *manager) GetCurrentContext() (RegionContext, error) {
	tkgConfigNode, err := m.loadTkgConfig()
	if err != nil {
		return RegionContext{}, errors.Wrapf(err, "unable to unmarshal tkg configuration file %s", m.tkgConfigPath)
	}

	currentContextNode := m.findNodeUnderTkg(&tkgConfigNode, constants.KeyCurrentRegionContext)

	if currentContextNode != nil && currentContextNode.Value != "" {
		regions, err := m.ListRegionContexts()
		if err != nil {
			return RegionContext{}, errors.Wrap(err, "unable to get current management cluster context info")
		}

		for _, r := range regions {
			if r.ContextName == currentContextNode.Value {
				return r, nil
			}
		}
	}
	return RegionContext{}, errors.New("current management cluster context is not set")
}

func (m *manager) loadTkgConfig() (yaml.Node, error) {
	// load the whole tkg config into yaml Node
	tkgConfigNode := yaml.Node{}

	fileData, err := os.ReadFile(m.tkgConfigPath)
	if err != nil {
		return tkgConfigNode, errors.Errorf("unable to read tkg configuration from: %s", m.tkgConfigPath)
	}

	err = yaml.Unmarshal(fileData, &tkgConfigNode)

	return tkgConfigNode, err
}

func (m *manager) createNodeUnderTkg(tkgConfigNode *yaml.Node, key, value string, kind yaml.Kind) (*yaml.Node, error) {
	tkgNodeIndex := getNodeIndex(tkgConfigNode.Content[0].Content, constants.KeyTkg)
	// create tkg section if it does not exist
	if tkgNodeIndex == -1 {
		tkgConfigNode.Content[0].Content = append(tkgConfigNode.Content[0].Content, m.createMappingNode(constants.KeyTkg)...)
	}

	tkgNodeIndex = getNodeIndex(tkgConfigNode.Content[0].Content, constants.KeyTkg)

	// if tkg node presents, but nothing under it
	if len(tkgConfigNode.Content[0].Content[tkgNodeIndex].Content) == 0 {
		tkgConfigNode.Content[0].Content[tkgNodeIndex] = &yaml.Node{
			Kind: yaml.MappingNode,
		}
	}

	tkgNode := tkgConfigNode.Content[0].Content[tkgNodeIndex]

	switch kind {
	case yaml.SequenceNode:
		tkgNode.Content = append(tkgNode.Content, m.createSequenceNode(key)...)
	case yaml.MappingNode:
		tkgNode.Content = append(tkgNode.Content, m.createMappingNode(key)...)
	case yaml.ScalarNode:
		tkgNode.Content = append(tkgNode.Content, m.createScalarNode(key, value)...)
	default:
		return nil, errors.New("node kind is not supported under tkg node")
	}

	index := getNodeIndex(tkgNode.Content, key)
	return tkgNode.Content[index], nil
}

func (m *manager) createSequenceNode(key string) []*yaml.Node {
	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: key,
	}
	valueNode := &yaml.Node{
		Kind: yaml.SequenceNode,
	}

	return []*yaml.Node{keyNode, valueNode}
}

func (m *manager) createScalarNode(key, value string) []*yaml.Node {
	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: key,
	}
	valueNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: value,
	}

	return []*yaml.Node{keyNode, valueNode}
}

func (m *manager) createMappingNode(key string) []*yaml.Node {
	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: key,
	}
	valueNode := &yaml.Node{
		Kind: yaml.MappingNode,
	}

	return []*yaml.Node{keyNode, valueNode}
}

// find direct children under tkg node
func (m *manager) findNodeUnderTkg(tkgConfigNode *yaml.Node, key string) *yaml.Node {
	// find the tkg node
	tkgNodeIndex := getNodeIndex(tkgConfigNode.Content[0].Content, constants.KeyTkg)
	if tkgNodeIndex == -1 {
		return nil
	}
	tkgNode := tkgConfigNode.Content[0].Content[tkgNodeIndex]

	// find the tkg region node
	targetNodeIndex := getNodeIndex(tkgNode.Content, key)
	if targetNodeIndex == -1 {
		return nil
	}

	return tkgNode.Content[targetNodeIndex]
}

func getNodeIndex(node []*yaml.Node, key string) int {
	appIdx := -1
	for i, k := range node {
		if i%2 == 0 && k.Value == key {
			appIdx = i + 1
			break
		}
	}
	return appIdx
}
