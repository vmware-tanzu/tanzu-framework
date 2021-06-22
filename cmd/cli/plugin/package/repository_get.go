// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli/component"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackageclient"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackagedatamodel"
)

var repositoryGetOp = tkgpackagedatamodel.NewRepositoryGetOptions()

var repositoryGetCmd = &cobra.Command{
	Use:   "get REPOSITORY_NAME",
	Short: "Get repository status",
	Args:  cobra.ExactArgs(1),
	RunE:  repositoryGet,
}

func init() {
	repositoryGetCmd.Flags().StringVarP(&repositoryGetOp.Namespace, "namespace", "n", "default", "Target namespace to get the repository, optional")
	repositoryGetCmd.Flags().StringVarP(&repositoryGetOp.KubeConfig, "kubeconfig", "", "", "The path to the kubeconfig file, optional")
	repositoryGetCmd.Flags().StringVarP(&outputFormat, "output", "o", "", "Output format (yaml|json|table)")
	repositoryCmd.AddCommand(repositoryGetCmd)
}

func repositoryGet(cmd *cobra.Command, args []string) error {
	if len(args) == 1 {
		repositoryGetOp.RepositoryName = args[0]
	} else {
		return errors.New("incorrect number of input parameters. Usage: tanzu package repository get REPOSITORY_NAME [FLAGS]")
	}

	pkgClient, err := tkgpackageclient.NewTKGPackageClient(repositoryGetOp.KubeConfig)
	if err != nil {
		return err
	}

	packageRepository, err := pkgClient.GetRepository(repositoryGetOp)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to find package repository '%s'", repositoryGetOp.RepositoryName))
	}
	t := component.NewOutputWriter(cmd.OutOrStdout(), outputFormat)
	t.SetKeys("NAME", "VERSION", "REPOSITORY", "STATUS", "REASON")
	t.AddRow(
		packageRepository.Name,
		packageRepository.ResourceVersion,
		packageRepository.Spec.Fetch.ImgpkgBundle.Image,
		packageRepository.Status.FriendlyDescription,
		packageRepository.Status.UsefulErrorMessage,
	)
	t.Render()
	return nil
}
