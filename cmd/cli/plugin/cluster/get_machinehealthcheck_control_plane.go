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

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"
)

type getMachineHealthCheckCPOptions struct {
	machinehealthCheckName string
	namespace              string
}

var getMHCCP = &getMachineHealthCheckCPOptions{}

var getMachineHealthCheckCPCmd = &cobra.Command{
	Use:   "get CLUSTER_NAME",
	Short: "Get a MachineHealthCheck object of the control plane of a cluster",
	Long:  "Get a MachineHealthCheck object of the control plane for the given cluster",
	Args:  cobra.ExactArgs(1),
	RunE:  getMachineHealthCheckCP,
}

func init() {
	getMachineHealthCheckCPCmd.Flags().StringVarP(&getMHCCP.machinehealthCheckName, "mhc-name", "m", "", "Name of the MachineHealthCheck object")
	getMachineHealthCheckCPCmd.Flags().StringVarP(&getMHCCP.namespace, "namespace", "n", "", "The namespace where the MachineHealthCheck object was created.")
	machineHealthCheckControlPlaneCmd.AddCommand(getMachineHealthCheckCPCmd)
}

func getMachineHealthCheckCP(cmd *cobra.Command, args []string) error {
	server, err := config.GetCurrentServer()
	if err != nil {
		return err
	}

	if server.IsGlobal() {
		return errors.New("getting machine healthcheck with a global server is not implemented yet")
	}
	return runGetMachineHealthCheckCP(server, args[0])
}

func runGetMachineHealthCheckCP(server *v1alpha1.Server, clusterName string) error {
	tkgctlClient, err := createTKGClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context)
	if err != nil {
		return err
	}

	options := tkgctl.GetMachineHealthCheckOptions{
		ClusterName:            clusterName,
		Namespace:              getMHCCP.namespace,
		MachineHealthCheckName: getMHCCP.machinehealthCheckName,
		MatchLabel:             "cluster.x-k8s.io/control-plane",
	}

	mhcList, err := tkgctlClient.GetMachineHealthCheck(options)
	if err != nil {
		return err
	}

	bytes, err := json.MarshalIndent(mhcList, "", "    ")
	if err != nil {
		return errors.Wrap(err, "error marshaling the list of MachineHealthCheck objects")
	}

	fmt.Println(string(bytes))

	return nil
}
