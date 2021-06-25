// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli/component"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackageclient"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackagedatamodel"
)

var repositoryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List repositories",
	Args:  cobra.NoArgs,
	RunE:  repositoryList,
}

func init() {
	repositoryListCmd.Flags().StringVarP(&outputFormat, "output", "o", "", "Output format (yaml|json|table)")
	repositoryListCmd.Flags().BoolVarP(&repoOp.AllNamespaces, "all-namespaces", "A", false, "If present, list the repositories across all namespaces.")
	repositoryCmd.AddCommand(repositoryListCmd)
}

func repositoryList(cmd *cobra.Command, _ []string) error {
	pkgClient, err := tkgpackageclient.NewTKGPackageClient(repoOp.KubeConfig)
	if err != nil {
		return err
	}

	if repoOp.AllNamespaces {
		repoOp.Namespace = ""
	}

	var t component.OutputWriterSpinner

	if repoOp.AllNamespaces {
		t, err = component.NewOutputWriterWithSpinner(cmd.OutOrStdout(), outputFormat,
			"Retrieving repositories...", true, "NAME", "REPOSITORY", "STATUS", "DETAILS", "NAMESPACE")
	} else {
		t, err = component.NewOutputWriterWithSpinner(cmd.OutOrStdout(), outputFormat, "Retrieving repositories...", true,
			"NAME", "REPOSITORY", "STATUS", "DETAILS")
	}
	if err != nil {
		return err
	}

	packageRepositoryList, err := pkgClient.ListRepositories(repoOp)
	if err != nil {
		t.StopSpinner()
		return err
	}
	for _, packageRepository := range packageRepositoryList.Items { //nolint:gocritic
		status := packageRepository.Status.FriendlyDescription
		details := packageRepository.Status.UsefulErrorMessage
		if len(status) > tkgpackagedatamodel.ShortDescriptionMaxLength {
			status = fmt.Sprintf("%s...", status[:tkgpackagedatamodel.ShortDescriptionMaxLength])
		}
		if len(details) > tkgpackagedatamodel.ShortDescriptionMaxLength {
			details = fmt.Sprintf("%s...", details[:tkgpackagedatamodel.ShortDescriptionMaxLength])
		}
		if repoOp.AllNamespaces {
			t.AddRow(
				packageRepository.Name,
				packageRepository.Spec.Fetch.ImgpkgBundle.Image,
				status,
				details,
				packageRepository.Namespace)
		} else {
			t.AddRow(
				packageRepository.Name,
				packageRepository.Spec.Fetch.ImgpkgBundle.Image,
				status,
				details)
		}
	}
	t.RenderWithSpinner()

	return nil
}
