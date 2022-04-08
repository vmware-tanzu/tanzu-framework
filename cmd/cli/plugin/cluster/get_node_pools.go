// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	tkgsv1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha2"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/component"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"
)

type listNodePoolOptions struct {
	namespace    string
	outputFormat string
}

var lnp = &listNodePoolOptions{}

var listNodePoolsCmd = &cobra.Command{
	Use:   "list [CLUSTER_NAME]",
	Short: "List node pools",
	Args:  cobra.ExactArgs(1),
	RunE:  listNodePools,
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

	isPacific, err := tkgctlClient.IsPacificRegionalCluster()
	if err != nil {
		return errors.Wrap(err, "error determining Tanzu Kubernetes Cluster service for vSphere management cluster ")
	}
	if isPacific {
		return listPacificNodePools(cmd, tkgctlClient, mdOptions)
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

func listPacificNodePools(cmd *cobra.Command, tkgctlClient tkgctl.TKGClient, mdOptions client.GetMachineDeploymentOptions) error {
	var tkcObj *tkgsv1alpha2.TanzuKubernetesCluster

	tkcObj, err := tkgctlClient.GetPacificClusterObject(mdOptions.ClusterName, mdOptions.Namespace)
	if err != nil {
		return errors.Wrap(err, "unable to get Tanzu Kubernetes Cluster object")
	}
	machineDeployments, err := tkgctlClient.GetPacificMachineDeployments(mdOptions)
	if err != nil {
		return err
	}
	// // Pacific TKC has nodepool names, so update the MD names with nodepool names from TKC object before listing nodepools.
	// // This is required because the user would want to use the nodepool names from the list output for nodepool set/delete operations
	err = updateMDsWithTKCNodepoolNames(tkcObj, machineDeployments)
	if err != nil {
		return err
	}

	var t component.OutputWriter
	if lnp.outputFormat == string(component.JSONOutputType) || lnp.outputFormat == string(component.YAMLOutputType) {
		t = component.NewObjectWriter(cmd.OutOrStdout(), lnp.outputFormat, machineDeployments)
	} else {
		t = component.NewOutputWriter(cmd.OutOrStdout(), lnp.outputFormat, "NAME", "NAMESPACE", "PHASE", "REPLICAS", "READY", "UPDATED", "UNAVAILABLE")
		for idx := range machineDeployments {
			md := machineDeployments[idx]
			t.AddRow(md.Name, md.Namespace, md.Status.Phase, md.Status.Replicas, md.Status.ReadyReplicas, md.Status.UpdatedReplicas, md.Status.UnavailableReplicas)
		}
	}
	t.Render()

	return nil
}

func updateMDsWithTKCNodepoolNames(tkcObj *tkgsv1alpha2.TanzuKubernetesCluster, machineDeployments []capi.MachineDeployment) error {
	nodepools := tkcObj.Spec.Topology.NodePools
	for mdIdx := range machineDeployments {
		nodepoolName := getNodePoolNameFromMDName(tkcObj.Name, machineDeployments[mdIdx].Name)
		for npIdx := range nodepools {
			if nodepoolName == nodepools[npIdx].Name {
				machineDeployments[mdIdx].Name = nodepools[npIdx].Name
				break
			}
		}
	}
	return nil
}

func getNodePoolNameFromMDName(clusterName, mdName string) string {
	// Pacific(TKGS) creates a corresponding MachineDeployment for a nodepool in
	// the format {tkc-clustername}-{nodepool-name}-{randomstring}
	trimmedName := strings.TrimPrefix(mdName, fmt.Sprintf("%s-", clusterName))
	lastHypenIdx := strings.LastIndex(trimmedName, "-")
	if lastHypenIdx == -1 {
		return ""
	}
	nodepoolName := trimmedName[:lastHypenIdx]
	return nodepoolName
}
