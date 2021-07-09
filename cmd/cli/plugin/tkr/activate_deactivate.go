// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"
)

var activeCmd = &cobra.Command{
	Use:   "activate TKR_NAME",
	Short: "Activate Tanzu Kubernetes Releases",
	Long:  "Activate Tanzu Kubernetes Releases",
	Args:  cobra.ExactArgs(1),
	RunE:  activateKubernetesReleasesCmd,
}

var deactiveCmd = &cobra.Command{
	Use:   "deactivate TKR_NAME",
	Short: "Deactivate Tanzu Kubernetes Releases",
	Long:  "Deactivate Tanzu Kubernetes Releases",
	Args:  cobra.ExactArgs(1),
	RunE:  deactivateKubernetesReleasesCmd,
}

func activateKubernetesReleasesCmd(cmd *cobra.Command, args []string) error {
	server, err := config.GetCurrentServer()
	if err != nil {
		return err
	}

	if server.IsGlobal() {
		return errors.New("activating TanzuKubernetesRelease with a global server is not implemented yet")
	}

	tkgctlClient, err := createTKGClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context)
	if err != nil {
		return err
	}

	err = tkgctlClient.ActivateTanzuKubernetesReleases(args[0])
	if err != nil {
		return err
	}

	return nil
}

func deactivateKubernetesReleasesCmd(cmd *cobra.Command, args []string) error {
	server, err := config.GetCurrentServer()
	if err != nil {
		return err
	}

	if server.IsGlobal() {
		return errors.New("deactivating TanzuKubernetesRelease with a global server is not implemented yet")
	}

	tkgctlClient, err := createTKGClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context)
	if err != nil {
		return err
	}

	err = tkgctlClient.DeactivateTanzuKubernetesReleases(args[0])
	if err != nil {
		return err
	}

	return nil
}
