// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

type Proxy struct {
	// HTTP proxy setting
	// +kubebuilder:validation:Optional
	HTTPProxy *string `json:"http_proxy,omitempty"`

	// HTTPS proxy setting
	// +kubebuilder:validation:Optional
	HTTPSProxy *string `json:"https_proxy,omitempty"`

	// No-proxy setting
	// +kubebuilder:validation:Optional
	NoProxy *string `json:"no_proxy,omitempty"`
}
