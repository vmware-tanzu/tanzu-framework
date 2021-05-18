// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgctl"
)

var updateCredentialsClusterCmd = &cobra.Command{
	Use:   "cluster CLUSTER_NAME",
	Short: "Update workload cluster credentials",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		displayLogFileLocation()
		err := runUpdateCredentialsCluster(args[0])
		verifyCommandError(err)
	},
}

type updateCredentialsClusterOptions struct {
	namespace       string
	vsphereUser     string
	vspherePassword string
	timeout         time.Duration
}

var updateCredentialsClusterOpts = &updateCredentialsClusterOptions{}

func init() {
	updateCredentialsClusterCmd.Flags().StringVarP(&updateCredentialsClusterOpts.namespace, "namespace", "n", "", "The namespace where the workload cluster was created. Assumes 'default' if not specified")
	updateCredentialsClusterCmd.Flags().StringVarP(&updateCredentialsClusterOpts.vsphereUser, "vsphere-user", "", "", "Username for vSphere provider")
	updateCredentialsClusterCmd.Flags().StringVarP(&updateCredentialsClusterOpts.vspherePassword, "vsphere-password", "", "", "Password for vSphere provider")
	updateCredentialsClusterCmd.Flags().DurationVarP(&updateCredentialsClusterOpts.timeout, "timeout", "t", constants.DefaultLongRunningOperationTimeout, "Time duration to wait for an operation before timeout. Timeout duration in hours(h)/minutes(m)/seconds(s) units or as some combination of them (e.g. 2h, 30m, 2h30m10s)")
	updateCredsCommand.AddCommand(updateCredentialsClusterCmd)
}

func runUpdateCredentialsCluster(clusterName string) error {
	tkgClient, err := newTKGCtlClient()
	if err != nil {
		return err
	}

	options := tkgctl.UpdateCredentialsClusterOptions{
		ClusterName:     clusterName,
		Namespace:       updateCredentialsClusterOpts.namespace,
		VSphereUsername: updateCredentialsClusterOpts.vsphereUser,
		VSpherePassword: updateCredentialsClusterOpts.vspherePassword,
		Timeout:         updateCredentialsClusterOpts.timeout,
	}
	return tkgClient.UpdateCredentialsCluster(options)
}
