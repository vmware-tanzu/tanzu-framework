// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"

	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/command/plugin"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"
)

var pluginDesc = cliv1alpha1.PluginDescriptor{
	Name:        "diagnostics",
	Description: "Cluster diagnostics",
	Group:       cliv1alpha1.RunCmdGroup,
	Aliases:     []string{"diag", "diags", "diagnostic"},
	Version:     "v0.0.1",
}

var (
	//go:embed scripts
	scriptFS embed.FS

)

func main() {

	workDir, err  := getDefaultWorkdir()
	if err != nil {
		fmt.Printf("main: workdir: %s\n", err)
		os.Exit(1)
	}

	outputDir := getDefaultOutputDir()

	p, err := plugin.NewPlugin(&pluginDesc)
	if err != nil {
		log.Fatal(err)
	}
	p.Cmd.PersistentFlags().StringVar(&workDir, "work-dir", workDir, "a working location while collecting diagnostics information")
	p.Cmd.PersistentFlags().StringVar(&outputDir, "output-dir", outputDir, "the output location for collected diagnostics information")
	p.AddCommands(collectCmd())

	if err := p.Execute(); err != nil {
		os.Exit(1)
	}
}

func getDefaultWorkdir() (string, error) {
	tanzuConfigDir, err := config.LocalDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(tanzuConfigDir, "crashd"), nil
}

func getDefaultOutputDir() string {
	return "./"
}

func getDefaultKubeconfig() string {
	kcfg := os.Getenv("KUBECONFIG")
	if kcfg == ""{
		kcfg = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	}
	return kcfg
}

func getDefaultClusterContext(clusterName string) string {
	return fmt.Sprintf("%s-admin@%s", clusterName, clusterName)
}

// getCurrentManagementSvr returns the current management server
//func getCurrentManagementSvr() (*v1alpha1.Server, error) {
//	svr, err := config.GetCurrentServer()
//	if err != nil {
//		return nil, err
//	}
//	if !svr.IsManagementCluster() {
//		return nil, fmt.Errorf("current server is management server")
//	}
//	return svr, nil
//}
//
//func getClusterClient()(clusterclient.Client, error){
//	svr, err := getCurrentManagementSvr()
//	if err != nil {
//		return nil, err
//	}
//	return clusterclient.NewClient(svr.ManagementClusterOpts.Path, svr.ManagementClusterOpts.Context, clusterclient.Options{})
//}