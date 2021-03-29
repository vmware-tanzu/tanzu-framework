// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package plugintemplates

import _ "embed" // Import files for plugin templates

// PluginReadme contains the stock readme template
//go:embed plugin_readme.md.tmpl
var PluginReadme string

// GolangCIConfig contains the stock golangci configuration
//go:embed dotgolangci.txt.tmpl
var GolangCIConfig string

// Gomod contains the stock Gomod file for the scaffolding
//go:embed gomod.tmpl
var Gomod string
