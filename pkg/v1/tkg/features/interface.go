// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package features provides TKG related feature enablement functionalities
package features

// Client defines methods to access feature flags
type Client interface {
	GetFeatureFlags() (map[string]string, error)
	IsFeatureFlagEnabled(string) (bool, error)
	WriteFeatureFlags(map[string]string) error
	GetFeatureFlag(string) (string, error)
}
