// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"

	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgctl"
)

type getMachineHealthCheckCPOptions struct {
	machinehealthCheckName string
	namespace              string
}

var getMHCCP = &getMachineHealthCheckCPOptions{}

var getMachineHealthCheckCPCmd = &cobra.Command{
	Use:          "get CLUSTER_NAME",
	Short:        "Get a MachineHealthCheck object",
	Long:         "Get a MachineHealthCheck object of the control plane for the given cluster",
	Args:         cobra.ExactArgs(1),
	RunE:         getMachineHealthCheckCP,
	SilenceUsage: true,
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

func runGetMachineHealthCheckCP(server *configapi.Server, clusterName string) error {
	tkgctlClient, err := createTKGClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context)
	if err != nil {
		return err
	}

	options := tkgctl.GetMachineHealthCheckOptions{
		ClusterName:            clusterName,
		Namespace:              getMHCCP.namespace,
		MachineHealthCheckName: getMHCCP.machinehealthCheckName,
		MatchLabel:             controlPlaneLabel,
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
