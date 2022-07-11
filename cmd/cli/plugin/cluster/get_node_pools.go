// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/pkg/errors"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/component"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
)

type listNodePoolOptions struct {
	namespace    string
	outputFormat string
}

var lnp = &listNodePoolOptions{}

var listNodePoolsCmd = &cobra.Command{
	Use:          "list [CLUSTER_NAME]",
	Short:        "List node pools",
	Args:         cobra.ExactArgs(1),
	RunE:         listNodePools,
	SilenceUsage: true,
}

func init() {
	listNodePoolsCmd.Flags().StringVarP(&lnp.namespace, "namespace", "n", "default", "The namespace from which to list workload clusters.")
	listNodePoolsCmd.Flags().StringVarP(&lnp.outputFormat, "output", "o", "", "Output format (yaml|json|table)")
	clusterNodePoolCmd.AddCommand(listNodePoolsCmd)
}

func listNodePools(cmd *cobra.Command, args []string) error {
	server, err := config.GetCurrentServer()
	if err != nil {
		return err
	}

	if server.IsGlobal() {
		return errors.New("listing node pools with a global server is not implemented yet")
	}
	return listNodePoolsInternal(cmd, server, args[0])
}

//nolint:gocritic
func listNodePoolsInternal(cmd *cobra.Command, server *v1alpha1.Server, clusterName string) error {
	tkgctlClient, err := createTKGClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context)
	if err != nil {
		return err
	}

	mdOptions := client.GetMachineDeploymentOptions{
		ClusterName: clusterName,
		Namespace:   lnp.namespace,
	}

	machineDeployments, err := tkgctlClient.GetMachineDeployments(mdOptions)
	if err != nil {
		return err
	}

	var t component.OutputWriter
	if lnp.outputFormat == string(component.JSONOutputType) || lnp.outputFormat == string(component.YAMLOutputType) {
		t = component.NewObjectWriter(cmd.OutOrStdout(), lnp.outputFormat, machineDeployments)
	} else {
		t = component.NewOutputWriter(cmd.OutOrStdout(), lnp.outputFormat, "NAME", "NAMESPACE", "PHASE", "REPLICAS", "READY", "UPDATED", "UNAVAILABLE")
		for _, md := range machineDeployments {
			t.AddRow(md.Name, md.Namespace, md.Status.Phase, md.Status.Replicas, md.Status.ReadyReplicas, md.Status.UpdatedReplicas, md.Status.UnavailableReplicas)
		}
	}
	t.Render()

	return nil
}
