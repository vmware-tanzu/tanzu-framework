// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package template

import "github.com/vmware-tanzu/tanzu-framework/plugin-admin/builder/pkg/template/plugintemplates"

const gitignore = `artifacts
tools/bin`

// GoMod target
var GoMod = Target{
	Filepath: "go.mod",
	Template: plugintemplates.Gomod,
}

// BuildVersion target
var BuildVersion = Target{
	Filepath: "BUILD_VERSION",
	Template: `v0.0.1`,
}

// GitIgnore target
var GitIgnore = Target{
	Filepath: ".gitignore",
	Template: gitignore,
}

// GitLabCI target
var GitLabCI = Target{
	Filepath: ".gitlab-ci.yaml",
	Template: plugintemplates.GitlabCI,
}

// GitHubCI target
// TODO (pbarker): should we push everything to a single repository, or at least make that possible?
// TODO (pbarker): should report stats
var GitHubCI = Target{
	Filepath: ".github/workflows/release.yaml",
	Template: plugintemplates.GithubWorkflowRelease,
}

// CommonMK MK4 target
var CommonMK = Target{
	Filepath: "common.mk",
	Template: plugintemplates.CommonMK,
}

// Makefile target
var Makefile = Target{
	Filepath: "Makefile",
	Template: plugintemplates.Makefile,
}

// Codeowners target
// TODO (pbarker): replace with the CLI reviewers group
var Codeowners = Target{
	Filepath: "CODEOWNERS",
	Template: `* @vuil`,
}

// Tools target.
var Tools = Target{
	Filepath: "tools/tools.go",
	Template: plugintemplates.ToolsGo,
}

// MainReadMe target
var MainReadMe = Target{
	Filepath: "README.md",
	Template: plugintemplates.PluginReadme,
}

// GolangCIConfig target.
var GolangCIConfig = Target{
	Filepath: ".golangci.yaml",
	Template: plugintemplates.GolangCIConfig,
}
