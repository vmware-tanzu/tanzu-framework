// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package plugintemplates

import _ "embed" // Import files for plugin templates

// CommandReadme contains the stock readme template
//
//go:embed command_readme.md.tmpl
var CommandReadme string

// PluginReadme contains the stock readme template
//
//go:embed plugin_readme.md.tmpl
var PluginReadme string

// GolangCIConfig contains the stock golangci configuration
//
//go:embed dotgolangci.txt.tmpl
var GolangCIConfig string

// Gomod contains the stock Gomod file for the scaffolding
//
//go:embed gomod.tmpl
var Gomod string

// Makefile contains the stock Makefile
//
//go:embed Makefile.tmpl
var Makefile string

// CommonMK contains the stock Common.mk
//
//go:embed common.mk.tmpl
var CommonMK string

// MainGo contains the plugin main.go template
//
//go:embed main.go.tmpl
var MainGo string

// MainTestGo contains the plugin main test template
//
//go:embed main_test.go.tmpl
var MainTestGo string

// ToolsGo contains the plugin tools.go template
//
//go:embed tools.go.tmpl
var ToolsGo string

// GithubWorkflowRelease contains the Github release workflow template
//
//go:embed github_workflow_release.yaml.tmpl
var GithubWorkflowRelease string

// GitlabCI contains the Gitlab CI template
//
//go:embed gitlab-ci.yaml.tmpl
var GitlabCI string
