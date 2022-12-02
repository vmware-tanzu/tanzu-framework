// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

// Keys used to parse the yaml node to retrieve specific stanza of the config file
const (
	KeyServers                 = "servers"
	KeyContexts                = "contexts"
	KeyCurrentServer           = "current"
	KeyCurrentContext          = "currentContext"
	KeyClientOptions           = "clientOptions"
	KeyCLI                     = "cli"
	KeyFeatures                = "features"
	KeyEnv                     = "env"
	KeyDiscoverySources        = "discoverySources"
	KeyRepositories            = "repositories"
	KeyUnstableVersionSelector = "unstableVersionSelector"
	KeyEdition                 = "edition"
	KeyKind                    = "kind"
	KeyMetadata                = "metadata"
	KeyAPIVersion              = "apiVersion"
	KeyBomRepo                 = "bomRepo"
	KeyCompatibilityFilePath   = "compatibilityFilePath"
)
