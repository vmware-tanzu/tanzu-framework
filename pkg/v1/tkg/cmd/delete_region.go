// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"
)

type deleteRegionOptions struct {
	force              bool
	useExistingCluster bool
	unattended         bool
	timeout            time.Duration
}

var dr = &deleteRegionOptions{}

var deleteRegionCmd = &cobra.Command{
	Use:   "management-cluster CLUSTER_NAME",
	Short: "Delete a management cluster",
	Long:  `Delete a management cluster and tears down the underlying infrastructure`,
	Example: Examples(`
		# tkg delete management-cluster my-mc-cluster
	`),
	Aliases: []string{"mc"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		log.UnsetStdoutStderr()
		displayLogFileLocation()
		err := runDeleteRegion(args[0])
		verifyCommandError(err)
	},
}

func init() {
	deleteRegionCmd.Flags().BoolVarP(&dr.force, "force", "f", false, "Force deletion of the management cluster even if it is managing active Tanzu Kubernetes clusters")
	deleteRegionCmd.Flags().BoolVarP(&dr.useExistingCluster, "use-existing-cleanup-cluster", "e", false, "Use an existing cleanup cluster to delete the management cluster")
	deleteRegionCmd.Flags().BoolVarP(&dr.unattended, "yes", "y", false, "Delete management cluster without asking for confirmation")
	deleteRegionCmd.Flags().DurationVarP(&dr.timeout, "timeout", "t", constants.DefaultLongRunningOperationTimeout, "Time duration to wait for an operation before timeout. Timeout duration in hours(h)/minutes(m)/seconds(s) units or as some combination of them (e.g. 2h, 30m, 2h30m10s)")
	deleteCmd.AddCommand(deleteRegionCmd)
}

func runDeleteRegion(clusterName string) error {
	tkgClient, err := newTKGCtlClient()
	if err != nil {
		return err
	}

	options := tkgctl.DeleteRegionOptions{
		ClusterName:        clusterName,
		Force:              dr.force,
		UseExistingCluster: dr.useExistingCluster,
		SkipPrompt:         dr.unattended || skipPrompt,
		Timeout:            dr.timeout,
	}
	return tkgClient.DeleteRegion(options)
}
