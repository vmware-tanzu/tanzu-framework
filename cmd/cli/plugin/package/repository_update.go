// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/log"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackageclient"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackagedatamodel"
)

var repoUpdateOp = tkgpackagedatamodel.NewRepositoryOptions()

var repositoryUpdateCmd = &cobra.Command{
	Use:   "update REPOSITORY_NAME REPOSITORY_URL",
	Short: "Update repository",
	Args:  cobra.ExactArgs(2),
	RunE:  repositoryUpdate,
}

func init() {
	repositoryUpdateCmd.Flags().BoolVarP(&repoUpdateOp.CreateRepository, "create", "", false, "Creates the repository if it does not exist")
	repositoryUpdateCmd.Flags().StringVarP(&repoUpdateOp.KubeConfig, "kubeconfig", "", "", "The path to the kubeconfig file, optional")
	repositoryCmd.AddCommand(repositoryUpdateCmd)
}

func repositoryUpdate(_ *cobra.Command, args []string) error {
	repoUpdateOp.RepositoryName = args[0]
	repoUpdateOp.RepositoryURL = args[1]

	pkgClient, err := tkgpackageclient.NewTKGPackageClient(repoUpdateOp.KubeConfig)
	if err != nil {
		return err
	}

	if err := pkgClient.UpdateRepository(repoUpdateOp); err != nil {
		return err
	}

	log.Infof("Updated package repository '%s'\n", repoUpdateOp.RepositoryName)

	return nil
}
