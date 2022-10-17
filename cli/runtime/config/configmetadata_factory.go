// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"

	"github.com/pkg/errors"
	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config/nodeutils"
	"gopkg.in/yaml.v3"
)

// GetClientConfigNodeNoLock retrieves the config from the local directory without acquiring the lock

func GetConfigMetadataNode() (*yaml.Node, error) {
	// Acquire tanzu config lock
	AcquireTanzuConfigMetadataLock()
	defer ReleaseTanzuConfigMetadataLock()
	return GetConfigMetadataNodeNoLock()
}

// GetConfigMetadataNodeNoLock retrieves the config from the local directory without acquiring the lock
func GetConfigMetadataNodeNoLock() (*yaml.Node, error) {
	cfgPath, err := MetadataFilePath()
	if err != nil {
		return nil, errors.Wrap(err, "failed getting config metadata path")
	}

	bytes, err := os.ReadFile(cfgPath)
	if err != nil || len(bytes) == 0 {
		node, err := NewConfigMetadataNode()
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

func NewConfigMetadataNode() (*yaml.Node, error) {
	c := newConfigMetadataNode()
	node, err := nodeutils.ConvertToNode[configapi.ConfigMetadata](c)
	node.Content[0].Style = 0
	if err != nil {
		return nil, err
	}
	return node, nil
}

func newConfigMetadataNode() *configapi.ConfigMetadata {
	c := &configapi.ConfigMetadata{}

	// Check if the lock is acquired by the current process or not
	// If not try to acquire the lock before Storing the client config
	// and release the lock after updating the config
	if !IsTanzuConfigMetadataLockAcquired() {
		AcquireTanzuConfigMetadataLock()
		defer ReleaseTanzuConfigMetadataLock()
	}

	return c
}
