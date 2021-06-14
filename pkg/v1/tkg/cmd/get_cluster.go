// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli/component"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgctl"
)

type getClustersOptions struct {
	namespace string
	includeMC bool
}

var gco = &getClustersOptions{}

var getClustersCmd = &cobra.Command{
	Use:     "cluster",
	Aliases: []string{"clusters"},
	Short:   "Get Tanzu Kubernetes clusters",
	Long:    `Get Tanzu Kubernetes clusters managed by the active management cluster context`,
	Run: func(cmd *cobra.Command, args []string) {
		err := runGetClusters(cmd)
		verifyCommandError(err)
	},
}

func init() {
	getClustersCmd.Flags().StringVarP(&gco.namespace, "namespace", "n", "", "The namespace from which to get workload clusters. If not provided clusters from all namespaces will be returned")
	getClustersCmd.Flags().BoolVarP(&gco.includeMC, "include-management-cluster", "", false, "Show active management cluster information as well")
	getClustersCmd.Flags().StringVarP(&outputFormat, "output", "o", "", "Output format. Supported formats: json|yaml|table")

	getCmd.AddCommand(getClustersCmd)
}

func runGetClusters(cmd *cobra.Command) error {
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

	var t component.OutputWriter
	if outputFormat == string(component.JSONOutputType) || outputFormat == string(component.YAMLOutputType) {
		t = component.NewObjectWriter(cmd.OutOrStdout(), outputFormat, clusters)
	} else {
		t = component.NewOutputWriter(cmd.OutOrStdout(), outputFormat, "NAME", "NAMESPACE", "STATUS", "CONTROLPLANE", "WORKERS", "KUBERNETES", "ROLES")
		for i := range clusters {
			clusterRoles := ClusterRoleNone
			if len(clusters[i].Roles) != 0 {
				clusterRoles = strings.Join(clusters[i].Roles, ",")
			}
			t.AddRow(clusters[i].Name, clusters[i].Namespace, clusters[i].Status, clusters[i].ControlPlaneCount, clusters[i].WorkerCount, clusters[i].K8sVersion, clusterRoles)
		}
	}
	t.Render()

	return nil
}
