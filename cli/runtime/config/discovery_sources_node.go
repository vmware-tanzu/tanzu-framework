// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"gopkg.in/yaml.v3"

	"github.com/pkg/errors"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config/nodeutils"
)

// DiscoveryType constants
const (
	DiscoveryTypeOCI        = "oci"
	DiscoveryTypeLocal      = "local"
	DiscoveryTypeGCP        = "gcp"
	DiscoveryTypeKubernetes = "kubernetes"
	DiscoveryTypeREST       = "rest"
)

// setDiscoverySources adds or updates the node discoverySources
func setDiscoverySources(node *yaml.Node, discoverySources []configapi.PluginDiscovery, patchStrategyOptions *nodeutils.PatchStrategyOptions) (persist bool, err error) {
	var anyPersists []bool
	isTrue := func(item bool) bool { return item }
	opts := func(c *nodeutils.Config) {
		c.ForceCreate = true
		c.Keys = []nodeutils.Key{
			{Name: KeyDiscoverySources, Type: yaml.SequenceNode},
		}
	}
	discoverySourcesNode := nodeutils.FindNode(node, opts)
	if discoverySourcesNode == nil {
		return persist, err
	}
	for _, discoverySource := range discoverySources {
		persist, err = setDiscoverySource(discoverySourcesNode, discoverySource, patchStrategyOptions)
		anyPersists = append(anyPersists, persist)
		if err != nil {
			return persist, err
		}
	}
	persist = Some[bool](anyPersists, isTrue)
	return persist, err
}

func setDiscoverySource(discoverySourcesNode *yaml.Node, discoverySource configapi.PluginDiscovery, patchStrategyOptions *nodeutils.PatchStrategyOptions) (persist bool, err error) {
	// convert the discoverySource change obj to yaml node
	newNode, err := nodeutils.ConvertToNode[configapi.PluginDiscovery](&discoverySource)
	if err != nil {
		return persist, err
	}

	exists := false
	var result []*yaml.Node

	discoverySourceType, discoverySourceName := getDiscoverySourceTypeAndName(discoverySource)
	if discoverySourceType == "" || discoverySourceName == "" {
		return persist, errors.New("not found")
	}

	for _, discoverySourceNode := range discoverySourcesNode.Content {
		if discoverySourceIndex := nodeutils.GetNodeIndex(discoverySourceNode.Content, discoverySourceType); discoverySourceIndex != -1 {
			if discoverySourceFieldIndex := nodeutils.GetNodeIndex(discoverySourceNode.Content[discoverySourceIndex].Content, "name"); discoverySourceFieldIndex != -1 &&
				discoverySourceNode.Content[discoverySourceIndex].Content[discoverySourceFieldIndex].Value == discoverySourceName {
				exists = true
				// only merge and persist if the change is not equal to existing
				persist, err = nodeutils.NotEqual(newNode.Content[0], discoverySourceNode)
				if err != nil {
					return persist, err
				}
				if persist {
					err = nodeutils.ReplaceNodes(newNode.Content[0], discoverySourceNode, patchStrategyOptions)
					if err != nil {
						return persist, err
					}
					err = nodeutils.MergeNodes(newNode.Content[0], discoverySourceNode)
					if err != nil {
						return persist, err
					}
				}
				result = append(result, discoverySourceNode)
				continue
			}
		}
		result = append(result, discoverySourceNode)
	}
	if !exists {
		result = append(result, newNode.Content[0])
		persist = true
	}
	discoverySourcesNode.Style = 0
	discoverySourcesNode.Content = result
	return persist, err
}

func getDiscoverySourceTypeAndName(discoverySource configapi.PluginDiscovery) (string, string) {
	if discoverySource.GCP != nil && discoverySource.GCP.Name != "" {
		return DiscoveryTypeGCP, discoverySource.GCP.Name
	} else if discoverySource.OCI != nil && discoverySource.OCI.Name != "" {
		return DiscoveryTypeOCI, discoverySource.OCI.Name
	} else if discoverySource.Local != nil && discoverySource.Local.Name != "" {
		return DiscoveryTypeLocal, discoverySource.Local.Name
	} else if discoverySource.Kubernetes != nil && discoverySource.Kubernetes.Name != "" {
		return DiscoveryTypeKubernetes, discoverySource.Kubernetes.Name
	} else if discoverySource.REST != nil && discoverySource.REST.Name != "" {
		return DiscoveryTypeREST, discoverySource.REST.Name
	}
	return "", ""
}
