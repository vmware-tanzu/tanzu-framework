// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package types defines type constants.
package types

const (
	// HTTPProxyConfigAnnotation is the cluster HTTP proxy annotation key
	HTTPProxyConfigAnnotation = "TKG_HTTP_PROXY"

	// HTTPSProxyConfigAnnotation is the cluster HTTPS proxy annotation key
	HTTPSProxyConfigAnnotation = "TKG_HTTPS_PROXY"

	// NoProxyConfigAnnotation is the cluster no-proxy annotation key
	NoProxyConfigAnnotation = "TKG_NO_PROXY"

	// ProxyCACertConfigAnnotation is the cluster proxy CA certificate annotation key
	ProxyCACertConfigAnnotation = "TKG_PROXY_CA_CERT"

	// IPFamilyConfigAnnotation is the cluster IP family annotation key
	IPFamilyConfigAnnotation = "TKG_IP_FAMILY"
)
