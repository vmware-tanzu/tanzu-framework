// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	crashdexec "github.com/vmware-tanzu/crash-diagnostics/exec"
)

type collectBootsrapArgs struct {
	workDir     string
	skip        bool
	clusterName string
}

type collectWorkloadArgs struct {
	workDir          string
	standalone       bool
	infra            string
	kubeconfig       string
	contextName      string
	clusterName      string
	clusterNamespace string
	sshUser          string
	sskPkPath        string
}

var (
	bootstrapArgs = collectBootsrapArgs{
		skip: false,
	}

	workloadArgs = collectWorkloadArgs{
		standalone: false,
		infra:      "docker",
	}
)

func collectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collect",
		Short: "Collect cluster diagnostics for the specified cluster",
		Long:  `Collect cluster diagnostics for the specified cluster`,
	}
	// bootstrap args
	cmd.Flags().BoolVar(&bootstrapArgs.skip, "skip-bootstrap-cluster", bootstrapArgs.skip, "If true, skips diagnostics collection from the bootstrap cluster")
	cmd.Flags().StringVar(&bootstrapArgs.clusterName, "bootstrap-cluster-name", bootstrapArgs.clusterName, "A specific bootstrap cluster name to diagnose")

	// workload
	cmd.Flags().BoolVar(&workloadArgs.standalone, "standalone-workload-cluster", workloadArgs.standalone, "If true, the cluster is treated as a standalone cluster")
	cmd.Flags().StringVar(&workloadArgs.infra, "workload-cluster-infra", workloadArgs.infra, "Overrides the infrastructure type for the managed cluster (i.e. aws, azure, vsphere, etc)")
	cmd.Flags().StringVar(&workloadArgs.kubeconfig, "workload-cluster-kubeconfig", workloadArgs.kubeconfig, "Overrides the kubeconfig for the managed workload cluster")
	cmd.Flags().StringVar(&workloadArgs.contextName, "workload-cluster-context", workloadArgs.contextName, "Overrides the context name of the workload cluster")
	cmd.Flags().StringVar(&workloadArgs.clusterName, "workload-cluster-name", workloadArgs.clusterName, "The name of the managed cluster for which to collect diagnostics (required)")
	cmd.Flags().StringVar(&workloadArgs.clusterNamespace, "workload-cluster-namespace", workloadArgs.clusterNamespace, "The namespace where managed workload resources are stored (required)")

	cmd.RunE = collectFunc
	return cmd
}

func collectFunc(cmd *cobra.Command, args []string) error {
	workDir, err := getDefaultWorkdir()
	if err != nil {
		return fmt.Errorf("collect: %w", err)
	}
	defer os.RemoveAll(workDir)

	bootstrapArgs.workDir = workDir
	if err := collectBoostrapDiags(bootstrapArgs); err != nil {
		log.Printf("Bootstrap diagnostics: %s", err)
	}

	workloadArgs.workDir = workDir
	if workloadArgs.standalone {
		if err := collectStandaloneDiags(workloadArgs); err != nil {
			log.Printf("Workload cluster diagnostics: %s", err)
		}
	}

	//if err := collectWorkloadDiags(); err !=  nil {
	//	log.Printf("Workload cluster diagnostics: %s", err)
	//}

	log.Println("Done!")
	return nil
}

func collectBoostrapDiags(args collectBootsrapArgs) error {
	if args.skip {
		log.Println("skip-bootstrap-cluster = true: nothing is collected")
		return nil
	}

	scriptName := "scripts/bootstrap_cluster.crsh"
	scriptData, err := scriptFS.ReadFile(scriptName)
	if err != nil {
		return err
	}

	argsMap := crashdexec.ArgMap{
		"workdir":                args.workDir,
		"infra":                  "docker",
		"bootstrap_cluster_name": args.clusterName,
	}

	return crashdexec.Execute(scriptName, bytes.NewReader(scriptData), argsMap)
}

func collectStandaloneDiags(args collectWorkloadArgs) error {
	scriptName := "scripts/standalone_cluster.crsh"
	scriptData, err := scriptFS.ReadFile(scriptName)
	if err != nil {
		return err
	}

	argsMap := crashdexec.ArgMap{
		"workdir":                 args.workDir,
		"workload_infra":                   args.infra,
		"workload_cluster_name": args.clusterName,
		"workload_context":      args.contextName,
	}

	return crashdexec.Execute(scriptName, bytes.NewReader(scriptData), argsMap)
}
