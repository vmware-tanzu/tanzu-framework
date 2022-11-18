// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package nodeutils provides utility methods to perform operations on yaml node
package nodeutils

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// Config options to be used to find specific node based on arrays of Keys passed in hierarchical order
type Config struct {
	ForceCreate bool  // Set to True to create nodes of missing keys. Ex: True for Add/Update operations on yaml node, False for Get/Delete operations on yaml node
	Keys        []Key // keys of config nodes passed in hierarchical order. Ex: [ClientOptions, CLI, DiscoverySources] to get the DiscoverySources node from ClientOptions yaml node
}

type Key struct {
	Name  string
	Value string
	Type  yaml.Kind
}

type Options func(config *Config)

var (
	ErrNodeNotFound = errors.New("node not found")
)
