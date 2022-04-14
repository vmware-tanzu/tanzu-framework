// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package types defines type constants.
package types

const (
	// HTTPProxyConfigAnnotation is the cluster HTTP proxy annotation key
	HTTPProxyConfigAnnotation = "tkg.tanzu.vmware.com/tkg-http-proxy"

	// HTTPSProxyConfigAnnotation is the cluster HTTPS proxy annotation key
	HTTPSProxyConfigAnnotation = "tkg.tanzu.vmware.com/tkg-https-proxy"

	// NoProxyConfigAnnotation is the cluster no-proxy annotation key
	NoProxyConfigAnnotation = "tkg.tanzu.vmware.com/tkg-no-proxy"

	// ProxyCACertConfigAnnotation is the cluster proxy CA certificate annotation key
	ProxyCACertConfigAnnotation = "tkg.tanzu.vmware.com/tkg-proxy-ca-cert"

	// IPFamilyConfigAnnotation is the cluster IP family annotation key
	IPFamilyConfigAnnotation = "tkg.tanzu.vmware.com/tkg-ip-family"

	// SkipTLSVeriy is the cluster skip tls verify annotation key
	SkipTLSVerifyConfigAnnotation = "tkg.tanzu.vmware.com/skip-tls-verify"
)
