// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"
)

type getMachineHealthCheckOptions struct {
	machinehealthCheckName string
	namespace              string
}

var getMHC = &getMachineHealthCheckOptions{}

var getMachineHealthCheckCmd = &cobra.Command{
	Use:     "machinehealthcheck CLUSTER_NAME",
	Short:   "Get MachineHealthCheck object",
	Long:    "Get a MachineHealthCheck object for the given cluster",
	Example: Examples("tkg get machinehealthcheck my-cluster"),
	Aliases: []string{"mhc"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		displayLogFileLocation()
		err := runGetMachineHealthCheck(args[0])
		verifyCommandError(err)
	},
}

func init() {
	getMachineHealthCheckCmd.Flags().StringVarP(&getMHC.machinehealthCheckName, "mhc-name", "", "", "Name of the MachineHealthCheck object")
	getMachineHealthCheckCmd.Flags().StringVarP(&getMHC.namespace, "namespace", "n", "", "The namespace where the MachineHealthCheck object was created.")
	getCmd.AddCommand(getMachineHealthCheckCmd)
}

func runGetMachineHealthCheck(clusterName string) error {
	tkgClient, err := newTKGCtlClient()
	if err != nil {
		return err
	}

	options := tkgctl.GetMachineHealthCheckOptions{
		ClusterName:            clusterName,
		Namespace:              getMHC.namespace,
		MachineHealthCheckName: getMHC.machinehealthCheckName,
	}

	mhcList, err := tkgClient.GetMachineHealthCheck(options)
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
