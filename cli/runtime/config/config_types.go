// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package config Provide API methods to Read/Write specific stanza of config file
package config

type CfgOptions struct {
	CfgPath string // file path to the config file
}

type CfgOpts func(config *CfgOptions)

func WithCfgPath(path string) CfgOpts {
	return func(config *CfgOptions) {
		config.CfgPath = path
	}
}
