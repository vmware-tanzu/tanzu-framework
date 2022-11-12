// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Skip duplicate that matching with metadata config file

// Package config Provide API methods to Read/Write specific stanza of config file
//
//nolint:dupl
package config

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

// getClientConfigNode retrieves the config from the local directory with file lock
func getClientConfigNode() (*yaml.Node, error) {
	// Acquire tanzu config lock
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()
	return getClientConfigNodeNoLock()
}

// getClientConfigNodeNoLock retrieves the config from the local directory without acquiring the lock
func getClientConfigNodeNoLock() (*yaml.Node, error) {
	cfgPath, err := ClientConfigPath()
	if err != nil {
		return nil, errors.Wrap(err, "getClientConfigNodeNoLock: failed getting client config path")
	}
	bytes, err := os.ReadFile(cfgPath)
	if err != nil || len(bytes) == 0 {
		node, err := newClientConfigNode()
		if err != nil {
			return nil, errors.Wrap(err, "getClientConfigNodeNoLock: failed to create new client config")
		}
		return node, nil
	}
	var node yaml.Node
	err = yaml.Unmarshal(bytes, &node)
	if err != nil {
		return nil, errors.Wrap(err, "getClientConfigNodeNoLock: failed to construct struct from config data")
	}
	node.Content[0].Style = 0
	return &node, nil
}

// newClientConfigNode create and return new client config node
func newClientConfigNode() (*yaml.Node, error) {
	c := &configapi.ClientConfig{}
	node, err := convertClientConfigToNode(c)
	node.Content[0].Style = 0
	if err != nil {
		return nil, err
	}
	return node, nil
}

func persistConfig(node *yaml.Node) error {
	path, err := ClientConfigPath()
	if err != nil {
		return errors.Wrap(err, "could not find config path")
	}
	return persistNode(node, WithCfgPath(path))
}
