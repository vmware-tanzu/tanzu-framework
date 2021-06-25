// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli/component"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/log"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackageclient"
)

var repositoryGetCmd = &cobra.Command{
	Use:   "get REPOSITORY_NAME",
	Short: "Get repository status",
	Args:  cobra.ExactArgs(1),
	RunE:  repositoryGet,
}

func init() {
	repositoryGetCmd.Flags().StringVarP(&outputFormat, "output", "o", "", "Output format (yaml|json|table)")
	repositoryCmd.AddCommand(repositoryGetCmd)
}

func repositoryGet(cmd *cobra.Command, args []string) error {
	if len(args) == 1 {
		repoOp.RepositoryName = args[0]
	} else {
		return errors.New("incorrect number of input parameters. Usage: tanzu package repository get REPOSITORY_NAME [FLAGS]")
	}
	pkgClient, err := tkgpackageclient.NewTKGPackageClient(repoOp.KubeConfig)
	if err != nil {
		return err
	}
	t, err := component.NewOutputWriterWithSpinner(cmd.OutOrStdout(), outputFormat,
		fmt.Sprintf("Retrieving repository %s...", repoOp.RepositoryName), true)
	if err != nil {
		return err
	}
	t.SetKeys("NAME", "VERSION", "REPOSITORY", "STATUS", "REASON")

	packageRepository, err := pkgClient.GetRepository(repoOp)
	if err != nil {
		t.StopSpinner()
		if apierrors.IsNotFound(err) {
			log.Infof("failed to find package repository '%s'", repoOp.RepositoryName)
		} else {
			return err
		}
	}

	t.AddRow(
		packageRepository.Name,
		packageRepository.ResourceVersion,
		packageRepository.Spec.Fetch.ImgpkgBundle.Image,
		packageRepository.Status.FriendlyDescription,
		packageRepository.Status.UsefulErrorMessage,
	)
	t.RenderWithSpinner()
	return nil
}
