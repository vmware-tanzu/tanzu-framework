// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/component"
	"github.com/vmware-tanzu/tanzu-framework/packageclients/pkg/packageclient"
	"github.com/vmware-tanzu/tanzu-framework/packageclients/pkg/packagedatamodel"
)

var repositoryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List package repositories",
	Args:  cobra.NoArgs,
	Example: `
    # List repositories across all namespaces 	
    tanzu package repository list -A
	
    # List installed packages from default namespace	
    tanzu package repository list`,
	RunE:         repositoryList,
	SilenceUsage: true,
}

func init() {
	repositoryListCmd.Flags().StringVarP(&outputFormat, "output", "o", "", "Output format (yaml|json|table), optional")
	repositoryListCmd.Flags().BoolVarP(&repoOp.AllNamespaces, "all-namespaces", "A", false, "If present, list the package repositories across all namespaces, optional")
	repositoryCmd.AddCommand(repositoryListCmd)
}

func repositoryList(cmd *cobra.Command, _ []string) error {
	pkgClient, err := packageclient.NewPackageClient(kubeConfig)
	if err != nil {
		return err
	}

	if repoOp.AllNamespaces {
		repoOp.Namespace = ""
	}

	var t component.OutputWriterSpinner

	if repoOp.AllNamespaces {
		t, err = component.NewOutputWriterWithSpinner(cmd.OutOrStdout(), outputFormat,
			"Retrieving repositories...", true, "NAME", "REPOSITORY", "TAG", "STATUS", "DETAILS", "NAMESPACE")
	} else {
		t, err = component.NewOutputWriterWithSpinner(cmd.OutOrStdout(), outputFormat, "Retrieving repositories...", true,
			"NAME", "REPOSITORY", "TAG", "STATUS", "DETAILS")
	}
	if err != nil {
		return err
	}

	packageRepositoryList, err := pkgClient.ListRepositories(repoOp)
	if err != nil {
		t.StopSpinner()
		return err
	}
	for i := range packageRepositoryList.Items {
		packageRepository := packageRepositoryList.Items[i]
		status := packageRepository.Status.FriendlyDescription
		details := packageRepository.Status.UsefulErrorMessage
		imageRepository, tag, _ := packageclient.GetCurrentRepositoryAndTagInUse(&packageRepository)
		if len(status) > packagedatamodel.ShortDescriptionMaxLength {
			status = fmt.Sprintf("%s...", status[:packagedatamodel.ShortDescriptionMaxLength])
		}
		if len(details) > packagedatamodel.ShortDescriptionMaxLength {
			details = fmt.Sprintf("%s...", details[:packagedatamodel.ShortDescriptionMaxLength])
		}

		if repoOp.AllNamespaces {
			t.AddRow(
				packageRepository.Name,
				imageRepository,
				tag,
				status,
				details,
				packageRepository.Namespace)
		} else {
			t.AddRow(
				packageRepository.Name,
				imageRepository,
				tag,
				status,
				details)
		}
	}
	t.RenderWithSpinner()

	return nil
}
