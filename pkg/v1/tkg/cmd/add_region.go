// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgctl"
)

type addRegionOptions struct {
	overwrite          bool
	useDirectReference bool
}

var (
	ar           = &addRegionOptions{}
	addRegionCmd = &cobra.Command{
		Use:     "management-cluster",
		Short:   "Add an existing management cluster to the config file",
		Long:    "Add an existing management cluster to the config file",
		Aliases: []string{"mc"},
		Run: func(cmd *cobra.Command, args []string) {
			err := runAddRegion()
			verifyCommandError(err)
		},
	}
)

func init() {
	addRegionCmd.Flags().BoolVarP(&ar.overwrite, "overwrite", "", false, "Overwrite management cluster context if already exists")
	addRegionCmd.Flags().BoolVarP(&ar.useDirectReference, "direct-reference", "", false, "reference credentials to the cluster directly via the kubeconfig provided")
	addRegionCmd.Flags().MarkHidden("direct-reference") //nolint
	addCmd.AddCommand(addRegionCmd)
}

func runAddRegion() error {
	tkgClient, err := newTKGCtlClient()
	if err != nil {
		return err
	}

	addRegionOptions := tkgctl.AddRegionOptions{
		Overwrite:          ar.overwrite,
		UseDirectReference: ar.useDirectReference,
	}
	return tkgClient.AddRegion(addRegionOptions)
}
