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

var repositoryListOp = tkgpackagedatamodel.NewRepositoryListOptions()

var repositoryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List repositories",
	Args:  cobra.NoArgs,
	RunE:  repositoryList,
}

func init() {
	repositoryListCmd.Flags().StringVarP(&repositoryListOp.KubeConfig, "kubeconfig", "", "", "The path to the kubeconfig file, optional")
	repositoryListCmd.Flags().StringVarP(&repositoryListOp.Namespace, "namespace", "n", "default", "Namespace of repository, optional")
	repositoryListCmd.Flags().StringVarP(&outputFormat, "output", "o", "", "Output format (yaml|json|table)")
	repositoryListCmd.Flags().BoolVarP(&repositoryListOp.AllNamespaces, "all-namespaces", "A", false, "If present, list the repositories across all namespaces.")
	repositoryCmd.AddCommand(repositoryListCmd)
}

func repositoryList(cmd *cobra.Command, _ []string) error {
	pkgClient, err := tkgpackageclient.NewTKGPackageClient(repositoryListOp.KubeConfig)
	if err != nil {
		return err
	}

	if repositoryListOp.AllNamespaces {
		repositoryListOp.Namespace = ""
	}

	packageRepositoryList, err := pkgClient.ListRepositories(repositoryListOp)
	if err != nil {
		return err
	}

	var t component.OutputWriter

	if repositoryListOp.AllNamespaces {
		t = component.NewOutputWriter(cmd.OutOrStdout(), outputFormat, "NAME", "REPOSITORY", "STATUS", "DETAILS", "NAMESPACE")
	} else {
		t = component.NewOutputWriter(cmd.OutOrStdout(), outputFormat, "NAME", "REPOSITORY", "STATUS", "DETAILS")
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
		if repositoryListOp.AllNamespaces {
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
	t.Render()

	return nil
}
