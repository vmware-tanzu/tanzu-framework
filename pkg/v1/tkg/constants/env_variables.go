// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package constants

// environment variables for http proxy
const (
	NoProxy     = "NO_PROXY"
	HTTPProxy   = "HTTP_PROXY"
	HTTPSProxy  = "HTTPS_PROXY"
	ProxyCACert = "PROXY_CA_CERT"
)

// environment variables for internal development use
const (
	SuppressProvidersUpdate = "SUPPRESS_PROVIDERS_UPDATE"
)

const (
	AllowedRegistries = "ALLOWED_REGISTRY"
)
