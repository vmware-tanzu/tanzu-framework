// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgctl"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/utils"
)

type getClustersOptions struct {
	namespace    string
	includeMC    bool
	outputFormat string
}

var gco = &getClustersOptions{}

var getClustersCmd = &cobra.Command{
	Use:     "cluster",
	Aliases: []string{"clusters"},
	Short:   "Get Tanzu Kubernetes clusters",
	Long:    `Get Tanzu Kubernetes clusters managed by the active management cluster context`,
	Run: func(cmd *cobra.Command, args []string) {
		err := runGetClusters()
		verifyCommandError(err)
	},
}

func init() {
	getClustersCmd.Flags().StringVarP(&gco.namespace, "namespace", "n", "", "The namespace from which to get workload clusters. If not provided clusters from all namespaces will be returned")
	getClustersCmd.Flags().BoolVarP(&gco.includeMC, "include-management-cluster", "", false, "Show active management cluster information as well")
	getClustersCmd.Flags().StringVarP(&gco.outputFormat, "output", "o", "", "Output format. Supported formats: json|yaml")

	getCmd.AddCommand(getClustersCmd)
}

func runGetClusters() error {
	tkgClient, err := newTKGCtlClient()
	if err != nil {
		return err
	}

	options := tkgctl.ListTKGClustersOptions{
		ClusterName: "",
		Namespace:   gco.namespace,
		IncludeMC:   gco.includeMC,
	}

	clusters, err := tkgClient.GetClusters(options)
	if err != nil {
		return err
	}

	if gco.outputFormat != "" {
		return utils.RenderOutput(clusters, gco.outputFormat)
	}

	t := utils.CreateTableWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"NAME", "NAMESPACE", "STATUS", "CONTROLPLANE", "WORKERS", "KUBERNETES", "ROLES"})
	for i := range clusters {
		clusterRoles := ClusterRoleNone
		if len(clusters[i].Roles) != 0 {
			clusterRoles = strings.Join(clusters[i].Roles, ",")
		}
		t.AppendRow(table.Row{clusters[i].Name, clusters[i].Namespace, clusters[i].Status, clusters[i].ControlPlaneCount, clusters[i].WorkerCount, clusters[i].K8sVersion, clusterRoles})
	}
	t.Render()

	return nil
}
