// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config/collectionutils"

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

const (
	Default      = "default"
	DefaultLocal = "default-local"
)

// setDiscoverySources adds or updates the node discoverySources
func setDiscoverySources(node *yaml.Node, discoverySources []configapi.PluginDiscovery, patchStrategyOpts ...nodeutils.PatchStrategyOpts) (persist bool, err error) {
	var anyPersists []bool
	isTrue := func(item bool) bool { return item }
	// Find the discovery sources node in the specific yaml node
	keys := []nodeutils.Key{
		{Name: KeyDiscoverySources, Type: yaml.SequenceNode},
	}
	discoverySourcesNode := nodeutils.FindNode(node, nodeutils.WithForceCreate(), nodeutils.WithKeys(keys))
	if discoverySourcesNode == nil {
		return persist, err
	}
	// Add or update discovery sources in the discovery sources node
	for _, discoverySource := range discoverySources {
		persist, err = setDiscoverySource(discoverySourcesNode, discoverySource, patchStrategyOpts...)
		anyPersists = append(anyPersists, persist)
		if err != nil {
			return persist, err
		}
	}
	persist = collectionutils.SomeBool(anyPersists, isTrue)
	return persist, err
}

//nolint:funlen,gocyclo
func setDiscoverySource(discoverySourcesNode *yaml.Node, discoverySource configapi.PluginDiscovery, patchStrategyOpts ...nodeutils.PatchStrategyOpts) (persist bool, err error) {
	// Convert discoverySource change obj to yaml node
	newNode, err := convertPluginDiscoveryToNode(&discoverySource)
	if err != nil {
		return persist, err
	}

	exists := false
	var result []*yaml.Node

	// Get discovery source type and name
	newOrUpdatedDiscoverySourceType, newOrUpdatedDiscoverySourceName := getDiscoverySourceTypeAndName(discoverySource)
	if newOrUpdatedDiscoverySourceType == "" || newOrUpdatedDiscoverySourceName == "" {
		return persist, errors.New("not found")
	}

	// Find the matching discovery source type and index from possible source values
	findDiscoverySourceIndexOfAnyType := func(currentNodes []*yaml.Node) (string, int) {
		possibleValues := []string{DiscoveryTypeOCI, DiscoveryTypeLocal, DiscoveryTypeGCP, DiscoveryTypeKubernetes, DiscoveryTypeREST}
		for _, discoverySourceType := range possibleValues {
			idx := nodeutils.GetNodeIndex(currentNodes, discoverySourceType)
			if idx != -1 {
				return discoverySourceType, idx
			}
		}
		return "", -1
	}
	// Loop through each discovery source node
	for _, discoverySourceNode := range discoverySourcesNode.Content {
		// Find discovery source by any type
		discoverySourceTypeOfAnyType, discoverySourceIndexOfAnyType := findDiscoverySourceIndexOfAnyType(discoverySourceNode.Content)

		// Find discovery source by exact type
		discoverySourceIndexOfExactType := nodeutils.GetNodeIndex(discoverySourceNode.Content, newOrUpdatedDiscoverySourceType)

		// If strict match i.e. change discovery source type and current discovery source type is of same type proceed with regular merge
		if discoverySourceIndexOfAnyType != -1 && discoverySourceIndexOfExactType != -1 {
			// find name field node
			nameIdx := nodeutils.GetNodeIndex(discoverySourceNode.Content[discoverySourceIndexOfAnyType].Content, "name")
			if nameIdx != -1 {
				// check if same name already exists
				isSameNameAlreadyExists := discoverySourceNode.Content[discoverySourceIndexOfAnyType].Content[nameIdx].Value == newOrUpdatedDiscoverySourceName
				// check if name is default or default-local
				isNameDefaultOrDefaultLocal := checkFoDiscoverySourceDefaultAndDefaultLocalName(discoverySourceNode.Content[discoverySourceIndexOfAnyType].Content[nameIdx].Value, newOrUpdatedDiscoverySourceName)
				// check if matching discovery source found based on name
				if isSameNameAlreadyExists || isNameDefaultOrDefaultLocal {
					// match found proceed with regular merge
					exists = true
					// only replace, merge and persist if the change is not equal to existing
					persist, err = nodeutils.NotEqual(newNode.Content[0], discoverySourceNode)
					if err != nil {
						return persist, err
					}
					if persist {
						// Replace nodes as per patch strategy defined in config-metadata.yaml
						err = nodeutils.ReplaceNodes(newNode.Content[0], discoverySourceNode, patchStrategyOpts...)
						if err != nil {
							return false, err
						}
						// Merge the new node into discovery source node
						err = nodeutils.MergeNodes(newNode.Content[0], discoverySourceNode)
						if err != nil {
							return false, err
						}
					}
				}
			}
			// If not a strict match i.e. change discovery source type is of different current discovery type
		} else if discoverySourceIndexOfAnyType != -1 || discoverySourceIndexOfExactType != -1 {
			// find name field node
			nameIdx := nodeutils.GetNodeIndex(discoverySourceNode.Content[discoverySourceIndexOfAnyType].Content, "name")
			if nameIdx != -1 {
				// check if newOrUpdatedDiscoverySource ds name and current ds name are same
				isSameNameAlreadyExists := discoverySourceNode.Content[discoverySourceIndexOfAnyType].Content[nameIdx].Value == newOrUpdatedDiscoverySourceName
				// check if newOrUpdatedDiscoverySource name and current discovery source name is default or default-local
				isNameDefaultOrDefaultLocal := checkFoDiscoverySourceDefaultAndDefaultLocalName(discoverySourceNode.Content[discoverySourceIndexOfAnyType].Content[nameIdx].Value, newOrUpdatedDiscoverySourceName)
				if isSameNameAlreadyExists || isNameDefaultOrDefaultLocal {
					exists = true
					// only replace, merge and persist if the change is not equal to existing
					persist, err = nodeutils.NotEqual(newNode.Content[0], discoverySourceNode)
					if err != nil {
						return persist, err
					}
					if persist {
						// Since merging discovery sources of different discovery source types we need to replace the nodes of different discovery type
						options := &nodeutils.PatchStrategyOptions{}
						for _, opt := range patchStrategyOpts {
							opt(options)
						}
						replaceDiscoverySourceTypeKey := fmt.Sprintf("%v.%v", options.Key, discoverySourceTypeOfAnyType)
						replaceDiscoverySourceContextTypeKey := fmt.Sprintf("%v.%v", options.Key, "contextType")
						options.PatchStrategies[replaceDiscoverySourceTypeKey] = nodeutils.PatchStrategyReplace
						options.PatchStrategies[replaceDiscoverySourceContextTypeKey] = nodeutils.PatchStrategyReplace

						// Replace nodes as per patch strategy defined in config-metadata.yaml
						err = nodeutils.ReplaceNodes(newNode.Content[0], discoverySourceNode, patchStrategyOpts...)
						if err != nil {
							return false, err
						}
						// Merge the new node into discovery source node
						err = nodeutils.MergeNodes(newNode.Content[0], discoverySourceNode)
						if err != nil {
							return false, err
						}
					}
				}
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

func checkFoDiscoverySourceDefaultAndDefaultLocalName(existingDiscoverySource, newOrUpdatedDiscoverySource string) bool {
	return (existingDiscoverySource == Default || existingDiscoverySource == DefaultLocal) && (newOrUpdatedDiscoverySource == Default || newOrUpdatedDiscoverySource == DefaultLocal)
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
