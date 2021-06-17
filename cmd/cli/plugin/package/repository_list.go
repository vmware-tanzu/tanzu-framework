// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
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
	repositoryCmd.AddCommand(repositoryListCmd)
}

func repositoryList(_ *cobra.Command, _ []string) error {
	pkgClient, err := tkgpackageclient.NewTKGPackageClient(repositoryListOp.KubeConfig)
	if err != nil {
		return err
	}

	packageRepositoryList, err := pkgClient.ListRepositories()
	if err != nil {
		return err
	}

	t := component.NewTableWriter("NAME", "REPOSITORY", "STATUS", "DETAILS")
	for _, packageRepository := range packageRepositoryList.Items { //nolint:gocritic
		t.Append([]string{
			packageRepository.Name,
			packageRepository.Spec.Fetch.ImgpkgBundle.Image,
			packageRepository.Status.FriendlyDescription,
			packageRepository.Status.UsefulErrorMessage})
	}
	t.Render()

	return nil
}
