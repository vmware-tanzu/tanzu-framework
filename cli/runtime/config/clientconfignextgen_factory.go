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
)

// getClientConfigNextGenNode retrieves the config from the local directory with file lock
func getClientConfigNextGenNode() (*yaml.Node, error) {
	// Acquire tanzu config v2 lock
	AcquireTanzuConfigNextGenLock()
	defer ReleaseTanzuConfigNextGenLock()
	return getClientConfigNextGenNodeNoLock()
}

// getClientConfigNextGenNodeNoLock retrieves the config from the local directory without acquiring the lock
func getClientConfigNextGenNodeNoLock() (*yaml.Node, error) {
	cfgPath, err := ClientConfigNextGenPath()
	if err != nil {
		return nil, errors.Wrap(err, "failed getting client config path")
	}
	bytes, err := os.ReadFile(cfgPath)
	if err != nil || len(bytes) == 0 {
		node, err := newClientConfigNode()
		if err != nil {
			return nil, errors.Wrap(err, "failed to create new client config ng")
		}
		return node, nil
	}
	var node yaml.Node
	err = yaml.Unmarshal(bytes, &node)
	if err != nil {
		return nil, errors.Wrap(err, "failed to construct struct from config ng data")
	}
	node.Content[0].Style = 0
	return &node, nil
}

func persistClientConfigNextGen(node *yaml.Node) error {
	path, err := ClientConfigNextGenPath()
	if err != nil {
		return errors.Wrap(err, "could not find config ng path")
	}
	return persistNode(node, WithCfgPath(path))
}
