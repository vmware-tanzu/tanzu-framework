// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"os"

	"github.com/spf13/cobra"
	crashdexec "github.com/vmware-tanzu/crash-diagnostics/exec"
)

type collectCmdArgs struct {
	infra string
	clusterName string
	kubeconfig string
	contextName string
}

var(
	collectArgs = collectCmdArgs{
		infra: "docker",
		kubeconfig: getDefaultKubeconfig(),
	}
)
func collectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collect",
		Short: "Collect cluster diagnostics for the specified cluster",
		Long:  `Collect cluster diagnostics for the specified cluster`,
	}
	cmd.Flags().StringVar(&collectArgs.infra, "infra", collectArgs.infra, "The cluster infrastructure (i.e. aws, azure, docker, vsphere, etc)")
	cmd.Flags().StringVar(&collectArgs.clusterName, "cluster-name", collectArgs.clusterName, "The name of cluster to diagnose")
	cmd.Flags().StringVar(&collectArgs.kubeconfig, "kubeconfig", collectArgs.kubeconfig, "The kubeconfig file to use")
	cmd.RunE = collectFunc
	return cmd
}

func collectFunc(cmd *cobra.Command, args []string) error {
	scriptData, err := scriptFS.ReadFile("scripts/standalone_cluster.crsh")
	if err != nil {
		return err
	}

	// clean up workdir
	defer  os.RemoveAll(workDir)


	argsMap :=crashdexec.ArgMap{
		"workdir": workDir,
		"infra": collectArgs.infra,
		"kubeconfig": collectArgs.kubeconfig,
		"cluster_context": collectArgs.contextName,
		"cluster_name": collectArgs.clusterName,
	}

	return crashdexec.Execute("standalone_cluster.crsh", bytes.NewReader(scriptData), argsMap)
}
