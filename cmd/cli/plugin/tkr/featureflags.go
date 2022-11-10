// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

const (
	// TKR version v1alpha3 feature flag determines whether to use Tanzu Kubernetes Release API version v1alpha3. Setting
	// feature flag to true will allow to use the TKR version v1alpha3; false allows to use legacy TKR version v1alpha1
	FeatureFlagTKRVersionV1Alpha3 = "features.global.tkr-version-v1alpha3-beta"
)

// DefaultFeatureFlagsForTKRPlugin is used to populate default feature-flags for the tkr plugin
var (
	DefaultFeatureFlagsForTKRPlugin = map[string]bool{
		FeatureFlagTKRVersionV1Alpha3: false,
	}
)
