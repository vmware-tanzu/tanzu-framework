// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Skip duplicate that matching with metadata config file

//nolint:dupl
package config

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

// getMetadataNode retrieves the config from the local directory without acquiring the lock
func getMetadataNode() (*yaml.Node, error) {
	// Retrieve config metadata node
	AcquireTanzuMetadataLock()
	defer ReleaseTanzuMetadataLock()
	return getMetadataNodeNoLock()
}

// getMetadataNodeNoLock retrieves the config from the local directory without acquiring the lock
func getMetadataNodeNoLock() (*yaml.Node, error) {
	cfgPath, err := CfgMetadataFilePath()
	if err != nil {
		return nil, errors.Wrap(err, "failed getting config metadata path")
	}

	bytes, err := os.ReadFile(cfgPath)
	if err != nil || len(bytes) == 0 {
		node, err := newMetadataNode()
		if err != nil {
			return nil, errors.Wrap(err, "failed to create new config metadata")
		}
		return node, nil
	}
	var node yaml.Node

	err = yaml.Unmarshal(bytes, &node)
	if err != nil {
		return nil, errors.Wrap(err, "failed to construct struct from config metadata data")
	}
	node.Content[0].Style = 0

	return &node, nil
}

func newMetadataNode() (*yaml.Node, error) {
	c := &configapi.Metadata{}
	node, err := convertMetadataToNode(c)
	node.Content[0].Style = 0
	if err != nil {
		return nil, err
	}
	return node, nil
}

func persistConfigMetadata(node *yaml.Node) error {
	path, err := CfgMetadataFilePath()
	if err != nil {
		return errors.Wrap(err, "could not find config metadata path")
	}
	return persistNode(node, WithCfgPath(path))
}
