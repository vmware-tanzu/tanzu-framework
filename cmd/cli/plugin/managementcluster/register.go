// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgctl"

	"github.com/vmware-tanzu-private/core/apis/config/v1alpha1"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register the management cluster to Tanzu Mission Control",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runForCurrentMC(registerWithTmc)
	},
}

type registerOptions struct {
	tmcRegistrationURL string
}

var ro = &registerOptions{}

func init() {
	registerCmd.Flags().StringVarP(&ro.tmcRegistrationURL, "tmc-registration-url", "", "", "URL to download the yml which has configuration related to various kubernetes objects to be deployed on the management cluster for it to register to Tanzu Mission Control")
	_ = registerCmd.MarkFlagRequired("tmc-registration-url")
}

func registerWithTmc(server *v1alpha1.Server) error {
	tkgClient, err := newTKGCtlClient()
	if err != nil {
		return err
	}

	options := tkgctl.RegisterOptions{
		ClusterName:        server.Name,
		TMCRegistrationURL: ro.tmcRegistrationURL,
	}
	return tkgClient.RegisterWithTmc(options)
}
