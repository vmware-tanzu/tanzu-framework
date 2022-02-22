// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package types defines type constants.
package types

const (
	// HTTPProxyConfigAnnotation is the cluster HTTP proxy annotation key
	HTTPProxyConfigAnnotation = "tkg.tanzu.vmware.com/tkg_http_proxy"

	// HTTPSProxyConfigAnnotation is the cluster HTTPS proxy annotation key
	HTTPSProxyConfigAnnotation = "tkg.tanzu.vmware.com/tkg_https_proxy"

	// NoProxyConfigAnnotation is the cluster no-proxy annotation key
	NoProxyConfigAnnotation = "tkg.tanzu.vmware.com/tkg_no_proxy"

	// ProxyCACertConfigAnnotation is the cluster proxy CA certificate annotation key
	ProxyCACertConfigAnnotation = "tkg.tanzu.vmware.com/tkg_proxy_ca_cert"

	// IPFamilyConfigAnnotation is the cluster IP family annotation key
	IPFamilyConfigAnnotation = "tkg.tanzu.vmware.com/tkg_ip_family"
)
