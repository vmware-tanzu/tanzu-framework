// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package v1alpha3 provides the command definitions for TKR API v1alpha3
package v1alpha3

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"

	runv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
)

var ActivateCmd = &cobra.Command{
	Use:   "activate TKR_NAME",
	Short: "Activate Tanzu Kubernetes Releases",
	Long:  "Activate Tanzu Kubernetes Releases",
	Args:  cobra.ExactArgs(1),
	RunE:  activateKubernetesReleasesCmd,
}

var DeactivateCmd = &cobra.Command{
	Use:   "deactivate TKR_NAME",
	Short: "Deactivate Tanzu Kubernetes Releases",
	Long:  "Deactivate Tanzu Kubernetes Releases",
	Args:  cobra.ExactArgs(1),
	RunE:  deactivateKubernetesReleasesCmd,
}

func activateKubernetesReleasesCmd(cmd *cobra.Command, args []string) error {
	clusterClient, err := getClusterClient()
	if err != nil {
		return err
	}

	err = activateKubernetesReleases(clusterClient, args[0])
	if err != nil {
		return err
	}

	return nil
}

func deactivateKubernetesReleasesCmd(cmd *cobra.Command, args []string) error {
	clusterClient, err := getClusterClient()
	if err != nil {
		return err
	}

	err = deactivateKubernetesReleases(clusterClient, args[0])
	if err != nil {
		return err
	}

	return nil
}

func activateKubernetesReleases(clusterClient clusterclient.Client, tkrName string) error {
	var tkr runv1alpha3.TanzuKubernetesRelease
	patchFormat := `
	{
		"metadata": {
		    "labels": {
			    %q: null
		    }
	    }
	}`
	activateTKRTimeout := 10 * time.Second
	checkResourceInterval := 5 * time.Second
	patchStr := fmt.Sprintf(patchFormat, runv1alpha3.LabelDeactivated)
	pollOptions := &clusterclient.PollOptions{Interval: checkResourceInterval, Timeout: activateTKRTimeout}
	err := clusterClient.PatchResource(&tkr, tkrName, "", patchStr, types.MergePatchType, pollOptions)
	if err != nil {
		return errors.Wrap(err, "unable to patch the TKr object to remove the inactive label")
	}

	return nil
}
func deactivateKubernetesReleases(clusterClient clusterclient.Client, tkrName string) error {
	var tkr runv1alpha3.TanzuKubernetesRelease
	patchFormat := `
	{
		"metadata": {
		    "labels": {
			    %q: ""
		    }
	    }
	}`
	deactivateTKRTimeout := 10 * time.Second
	checkResourceInterval := 5 * time.Second
	patchStr := fmt.Sprintf(patchFormat, runv1alpha3.LabelDeactivated)
	pollOptions := &clusterclient.PollOptions{Interval: checkResourceInterval, Timeout: deactivateTKRTimeout}
	err := clusterClient.PatchResource(&tkr, tkrName, "", patchStr, types.MergePatchType, pollOptions)
	if err != nil {
		return errors.Wrap(err, "unable to patch the TKr object with inactive label")
	}

	return nil
}
