// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main // nolint:dupl

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"
)

type getMachineHealthCheckNodeOptions struct {
	machinehealthCheckName string
	namespace              string
}

var getMHCNode = &getMachineHealthCheckNodeOptions{}

var getMachineHealthCheckNodeCmd = &cobra.Command{
	Use:          "get CLUSTER_NAME",
	Short:        "Get a MachineHealthCheck object of the nodes of a cluster",
	Long:         "Get a MachineHealthCheck object of the nodes for the given cluster",
	Args:         cobra.ExactArgs(1),
	RunE:         getMachineHealthCheckNode,
	SilenceUsage: true,
}

func init() {
	getMachineHealthCheckNodeCmd.Flags().StringVarP(&getMHCNode.machinehealthCheckName, "mhc-name", "m", "", "Name of the MachineHealthCheck object")
	getMachineHealthCheckNodeCmd.Flags().StringVarP(&getMHCNode.namespace, "namespace", "n", "", "The namespace where the MachineHealthCheck object was created.")
	machineHealthCheckNodeCmd.AddCommand(getMachineHealthCheckNodeCmd)
}

func getMachineHealthCheckNode(cmd *cobra.Command, args []string) error {
	server, err := config.GetCurrentServer()
	if err != nil {
		return err
	}

	if server.IsGlobal() {
		return errors.New("getting machine healthcheck with a global server is not implemented yet")
	}
	return runGetMachineHealthCheckNode(server, args[0])
}

func runGetMachineHealthCheckNode(server *v1alpha1.Server, clusterName string) error {
	tkgctlClient, err := createTKGClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context)
	if err != nil {
		return err
	}

	options := tkgctl.GetMachineHealthCheckOptions{
		ClusterName:            clusterName,
		Namespace:              getMHCNode.namespace,
		MachineHealthCheckName: getMHCNode.machinehealthCheckName,
		MatchLabel:             "",
	}

	mhcList, err := tkgctlClient.GetMachineHealthCheck(options)
	if err != nil {
		return err
	}

	var filtered []client.MachineHealthCheck
	for i := range mhcList {
		if _, ok := mhcList[i].Spec.Selector.MatchLabels[nodePoolLabel]; ok {
			filtered = append(filtered, mhcList[i])
		} else if _, ok := mhcList[i].Spec.Selector.MatchLabels[machineDeploymentLabel]; ok {
			filtered = append(filtered, mhcList[i])
		}
	}

	bytes, err := json.MarshalIndent(filtered, "", "    ")
	if err != nil {
		return errors.Wrap(err, "error marshaling the list of MachineHealthCheck objects")
	}

	fmt.Println(string(bytes))

	return nil
}
