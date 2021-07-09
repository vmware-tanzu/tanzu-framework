// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"
	"strings"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/component"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"
)

type listClusterOptions struct {
	namespace    string
	includeMC    bool
	outputFormat string
}

var lc = &listClusterOptions{}

var listClustersCmd = &cobra.Command{
	Use:   "list",
	Short: "List clusters",
	RunE:  list,
}

func init() {
	listClustersCmd.Flags().StringVarP(&lc.namespace, "namespace", "n", "", "The namespace from which to list workload clusters. If not provided clusters from all namespaces will be returned")
	listClustersCmd.Flags().BoolVarP(&lc.includeMC, "include-management-cluster", "", false, "Show active management cluster information as well")
	listClustersCmd.Flags().StringVarP(&lc.outputFormat, "output", "o", "", "Output format (yaml|json|table)")
}

func list(cmd *cobra.Command, args []string) error {
	server, err := config.GetCurrentServer()
	if err != nil {
		return err
	}

	if server.IsGlobal() {
		return errors.New("listing cluster with a global server is not implemented yet")
	}
	return listClusters(cmd, server)
}

//nolint:gocritic
func listClusters(cmd *cobra.Command, server *v1alpha1.Server) error {
	tkgctlClient, err := createTKGClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context)
	if err != nil {
		return err
	}

	ccOptions := tkgctl.ListTKGClustersOptions{
		ClusterName: "",
		Namespace:   lc.namespace,
		IncludeMC:   lc.includeMC,
	}

	clusters, err := tkgctlClient.GetClusters(ccOptions)
	if err != nil {
		return err
	}

	var t component.OutputWriter
	if lc.outputFormat == string(component.JSONOutputType) || lc.outputFormat == string(component.YAMLOutputType) {
		t = component.NewObjectWriter(cmd.OutOrStdout(), lc.outputFormat, clusters)
	} else {
		t = component.NewOutputWriter(cmd.OutOrStdout(), lc.outputFormat, "NAME", "NAMESPACE", "STATUS", "CONTROLPLANE", "WORKERS", "KUBERNETES", "ROLES", "PLAN")
		for _, cl := range clusters {
			clusterRoles := "<none>"
			if len(cl.Roles) != 0 {
				clusterRoles = strings.Join(cl.Roles, ",")
			}
			t.AddRow(cl.Name, cl.Namespace, cl.Status, cl.ControlPlaneCount, cl.WorkerCount, cl.K8sVersion, clusterRoles, cl.Plan)
		}
	}
	t.Render()

	return nil
}
