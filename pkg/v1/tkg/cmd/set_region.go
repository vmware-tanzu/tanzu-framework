// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgctl"
)

type setRegionOptions struct {
	contextName string
}

var sr = &setRegionOptions{}

var setRegionCmd = &cobra.Command{
	Use:     "management-cluster CLUSTER_NAME",
	Short:   "Set the current management cluster context to use",
	Long:    "Set the current management cluster targeted by most `tkg` commands",
	Aliases: []string{"mc"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := runSetRegion(args[0])
		verifyCommandError(err)
	},
}

func init() {
	setRegionCmd.Flags().StringVarP(&sr.contextName, "context", "c", "", "Optional, the context name of the management cluster")

	setCmd.AddCommand(setRegionCmd)
}

func runSetRegion(clusterName string) error {
	tkgClient, err := newTKGCtlClient()
	if err != nil {
		return err
	}

	options := tkgctl.SetRegionOptions{
		ClusterName: clusterName,
		ContextName: sr.contextName,
	}
	return tkgClient.SetRegion(options)
}
