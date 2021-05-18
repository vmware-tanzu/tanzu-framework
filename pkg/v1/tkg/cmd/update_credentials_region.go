// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgctl"
)

var updateCredentialsRegionCmd = &cobra.Command{
	Use:     "management-cluster CLUSTER_NAME",
	Short:   "Update management cluster credentials",
	Aliases: []string{"mc"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		displayLogFileLocation()
		err := runUpdateCredentialsRegion(args[0])
		verifyCommandError(err)
	},
}

type updateCredentialsRegionOptions struct {
	vsphereUser     string
	vspherePassword string
	isCascading     bool
	timeout         time.Duration
}

var updateCredentialsRegionOpts = &updateCredentialsRegionOptions{}

func init() {
	updateCredentialsRegionCmd.Flags().StringVarP(&updateCredentialsRegionOpts.vsphereUser, "vsphere-user", "", "", "Username for vSphere provider")
	updateCredentialsRegionCmd.Flags().StringVarP(&updateCredentialsRegionOpts.vspherePassword, "vsphere-password", "", "", "Password for vSphere provider")
	updateCredentialsRegionCmd.Flags().BoolVarP(&updateCredentialsRegionOpts.isCascading, "cascading", "", false, "Complete credential rotation for all workload clusters under the management cluster")
	updateCredentialsRegionCmd.Flags().DurationVarP(&updateCredentialsRegionOpts.timeout, "timeout", "t", constants.DefaultLongRunningOperationTimeout, "Time duration to wait for an operation before timeout. Timeout duration in hours(h)/minutes(m)/seconds(s) units or as some combination of them (e.g. 2h, 30m, 2h30m10s)")
	updateCredsCommand.AddCommand(updateCredentialsRegionCmd)
}

func runUpdateCredentialsRegion(clusterName string) error {
	tkgClient, err := newTKGCtlClient()
	if err != nil {
		return err
	}

	options := tkgctl.UpdateCredentialsRegionOptions{
		ClusterName:     clusterName,
		VSphereUsername: updateCredentialsRegionOpts.vsphereUser,
		VSpherePassword: updateCredentialsRegionOpts.vspherePassword,
		IsCascading:     updateCredentialsRegionOpts.isCascading,
		Timeout:         updateCredentialsRegionOpts.timeout,
	}
	return tkgClient.UpdateCredentialsRegion(options)
}
