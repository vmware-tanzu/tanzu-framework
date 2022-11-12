// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package config Provide API methods to Read/Write specific stanza of config file
package config

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type CfgOptions struct {
	CfgPath string // file path to the config file
}

type CfgOpts func(config *CfgOptions)

func WithCfgPath(path string) CfgOpts {
	return func(config *CfgOptions) {
		config.CfgPath = path
	}
}

// persistNode stores/writes the yaml node to config.yaml
func persistNode(node *yaml.Node, opts ...CfgOpts) error {
	configurations := &CfgOptions{}
	for _, opt := range opts {
		opt(configurations)
	}
	cfgPathExists, err := fileExists(configurations.CfgPath)
	if err != nil {
		return errors.Wrap(err, "failed to check config path existence")
	}
	if !cfgPathExists {
		localDir, err := LocalDir()
		if err != nil {
			return errors.Wrap(err, "could not find local tanzu dir for OS")
		}
		if err := os.MkdirAll(localDir, 0755); err != nil {
			return errors.Wrap(err, "could not make local tanzu directory")
		}
	}
	data, err := yaml.Marshal(node)
	if err != nil {
		return errors.Wrap(err, "failed to marshal nodeutils")
	}
	err = os.WriteFile(configurations.CfgPath, data, 0644)
	if err != nil {
		return errors.Wrap(err, "failed to write the config to file")
	}
	storeConfigToLegacyDir(data)
	return nil
}
