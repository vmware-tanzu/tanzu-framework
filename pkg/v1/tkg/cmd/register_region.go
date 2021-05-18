// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgctl"
)

var registerRegion = &cobra.Command{
	Use:     "management-cluster CLUSTER_NAME",
	Short:   "Register a management cluster to Tanzu Mission Control",
	Aliases: []string{"mc"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := registerWithTmc(args[0])
		verifyCommandError(err)
	},
}

type registerOptions struct {
	tmcRegistrationURL string
	unattended         bool
}

var ro = &registerOptions{}

func init() {
	registerRegion.Flags().StringVarP(&ro.tmcRegistrationURL, "tmc-registration-url", "", "", "URL to download the yml which has configuration related to various kubernetes objects to be deployed on the management cluster for it to register to Tanzu Mission Control")
	registerRegion.Flags().BoolVarP(&ro.unattended, "yes", "y", false, "Register management cluster without asking for confirmation")
	_ = registerRegion.MarkFlagRequired("tmc-registration-url")
	registerCmd.AddCommand(registerRegion)
}

func registerWithTmc(clusterName string) error {
	tkgClient, err := newTKGCtlClient()
	if err != nil {
		return err
	}

	options := tkgctl.RegisterOptions{
		ClusterName:        clusterName,
		TMCRegistrationURL: ro.tmcRegistrationURL,
	}
	return tkgClient.RegisterWithTmc(options)
}
