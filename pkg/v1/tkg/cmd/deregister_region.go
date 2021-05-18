// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgctl"
)

var deRegisterRegion = &cobra.Command{
	Use:     "management-cluster CLUSTER_NAME",
	Short:   "Deregister a management cluster from Tanzu Mission Control",
	Aliases: []string{"mc"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := deregisterFromTmc(args[0])
		verifyCommandError(err)
	},
}

type deRegisterOptions struct {
	unattended bool
}

var uro = &deRegisterOptions{}

func init() {
	deRegisterRegion.Flags().BoolVarP(&uro.unattended, "yes", "y", false, "Deregister management cluster without asking for confirmation")

	// hide the deregister command until TMC supports registering TKG clusters
	deRegisterRegion.Hidden = true
	deRegisterCmd.AddCommand(deRegisterRegion)
}

func deregisterFromTmc(clusterName string) error {
	tkgClient, err := newTKGCtlClient()
	if err != nil {
		return err
	}

	options := tkgctl.DeregisterFromTMCOptions{
		ClusterName: clusterName,
		SkipPrompt:  uro.unattended || skipPrompt,
	}
	return tkgClient.DeregisterFromTmc(options)
}
