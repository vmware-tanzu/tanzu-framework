// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package managementcomponents

// TKGPackageConfig defines TKG package configuration
type TKGPackageConfig struct {
	Metadata     Metadata          `yaml:"metadata"`
	ConfigValues map[string]string `yaml:"configvalues"`
}

// Metadata specifies metadata as part of TKG package config
type Metadata struct {
	InfraProvider string `yaml:"infraProvider"`
}
