// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/tkg-cli/pkg/constants"
	"github.com/vmware-tanzu-private/tkg-cli/pkg/tkgctl"

	"github.com/vmware-tanzu-private/core/apis/config/v1alpha1"
)

type deleteRegionOptions struct {
	force              bool
	useExistingCluster bool
	unattended         bool
	timeout            time.Duration
}

var dr = &deleteRegionOptions{}

var deleteRegionCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a management cluster",
	Long:  `Delete a management cluster and tears down the underlying infrastructure`,
	Example: `
    # Deletes the management cluster of the current server
    tanzu management-cluster delete`,
	Args: cobra.MaximumNArgs(1), // TODO: deprecate version of command that takes args
	RunE: func(cmd *cobra.Command, args []string) error {
		return runForCurrentMC(runDeleteRegion)
	},
}

func init() {
	deleteRegionCmd.Flags().BoolVar(&dr.force, "force", false, "Force deletion of the management cluster even if it is managing active Tanzu Kubernetes clusters")
	deleteRegionCmd.Flags().BoolVarP(&dr.useExistingCluster, "use-existing-cleanup-cluster", "e", false, "Use an existing cleanup cluster to delete the management cluster")
	deleteRegionCmd.Flags().BoolVarP(&dr.unattended, "yes", "y", false, "Delete management cluster without asking for confirmation")
	deleteRegionCmd.Flags().DurationVarP(&dr.timeout, "timeout", "t", constants.DefaultLongRunningOperationTimeout, "Time duration to wait for an operation before timeout. Timeout duration in hours(h)/minutes(m)/seconds(s) units or as some combination of them (e.g. 2h, 30m, 2h30m10s)")
}

func runDeleteRegion(server *v1alpha1.Server) error {
	tkgClient, err := newTKGCtlClient()
	if err != nil {
		return err
	}

	options := tkgctl.DeleteRegionOptions{
		ClusterName:        server.Name,
		Force:              dr.force,
		UseExistingCluster: dr.useExistingCluster,
		SkipPrompt:         dr.unattended,
		Timeout:            dr.timeout,
	}
	return tkgClient.DeleteRegion(options)
}
