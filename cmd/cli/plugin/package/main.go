// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/aunum/log"
	"github.com/cppforlife/go-cli-ui/ui"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	kctrlcmd "github.com/vmware-tanzu/carvel-kapp-controller/cli/pkg/kctrl/cmd"
	kctrlcmdcore "github.com/vmware-tanzu/carvel-kapp-controller/cli/pkg/kctrl/cmd/core"
	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/command/plugin"
)

var descriptor = cliv1alpha1.PluginDescriptor{
	Name:        "package",
	Description: "Tanzu package management",
	Group:       cliv1alpha1.RunCmdGroup,
}

func main() {
	p, err := plugin.NewPlugin(&descriptor)
	if err != nil {
		log.Fatal(err)
	}

	err = nonExitingMain(p)
	if err != nil {
		os.Exit(1)
	}
}

func nonExitingMain(p *plugin.Plugin) error {
	writerUI := ui.NewWriterUI(os.Stdout, os.Stderr, ui.NewNoopLogger())
	adapterUI := &AdapterUI{*writerUI, os.Stdout, ""}
	confUI := ui.NewWrappingConfUI(adapterUI, ui.NewNoopLogger())

	defer confUI.Flush()

	kctrlcmd.AttachKctrlPackageCommandTree(p.Cmd, confUI, kctrlcmdcore.PackageCommandTreeOpts{BinaryName: "tanzu", PositionalArgs: true,
		Color: false, JSON: false})

	setOutputFormat(p.Cmd, adapterUI)

	if err := p.Execute(); err != nil {
		return err
	}

	return nil
}
