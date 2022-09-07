// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package template

import "github.com/vmware-tanzu/tanzu-framework/plugin-admin/builder/pkg/template/plugintemplates"

// PluginReadMe target
var PluginReadMe = Target{
	Filepath: "cmd/plugin/{{ .PluginName }}/README.md",
	Template: plugintemplates.CommandReadme,
}

// PluginMain target
// TODO (pbarker): proper logging
var PluginMain = Target{
	Filepath: "cmd/plugin/{{ .PluginName | ToLower }}/main.go",
	Template: plugintemplates.MainGo,
}

// PluginTest target
var PluginTest = Target{
	Filepath: "cmd/plugin/{{ .PluginName }}/test/main.go",
	Template: plugintemplates.MainTestGo,
}
