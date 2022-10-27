// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package kctrl provides function to add package command tree from kctrl
package kctrl

import (
	"github.com/cppforlife/go-cli-ui/ui"

	kctrlcmd "github.com/vmware-tanzu/carvel-kapp-controller/cli/pkg/kctrl/cmd"
	kctrlcmdcore "github.com/vmware-tanzu/carvel-kapp-controller/cli/pkg/kctrl/cmd/core"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/plugin"
)

func Invoke(p *plugin.Plugin) {
	writerUI := ui.NewWriterUI(p.Cmd.OutOrStdout(), p.Cmd.ErrOrStderr(), ui.NewNoopLogger())
	adapterUI := &AdapterUI{WriterUI: *writerUI, outWriter: p.Cmd.OutOrStdout()}
	confUI := ui.NewWrappingConfUI(ui.NewPaddingUI(adapterUI), ui.NewNoopLogger())
	defer confUI.Flush()

	kctrlcmd.AttachKctrlPackageCommandTree(p.Cmd, confUI, kctrlcmdcore.PackageCommandTreeOpts{BinaryName: "tanzu", PositionalArgs: true,
		Color: false, JSON: false})
	setOutputFormatFlag(p.Cmd, adapterUI)
}
