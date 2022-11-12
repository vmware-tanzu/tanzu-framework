// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

// Metadata struct type to store config related metadata
type Metadata struct {
	// ConfigMetadata to store any config related metadata or settings
	ConfigMetadata *ConfigMetadata `json:"configMetadata,omitempty" yaml:"configMetadata,omitempty" mapstructure:"configMetadata,omitempty"`
}

// ConfigMetadata to store any config related metadata or settings
type ConfigMetadata struct {
	// PatchStrategy patch strategy to determine merge of nodes in config file. Two ways of patch strategies are merge and replace
	PatchStrategy map[string]string `json:"patchStrategy,omitempty" yaml:"patchStrategy,omitempty" mapstructure:"patchStrategy,omitempty"`
}
