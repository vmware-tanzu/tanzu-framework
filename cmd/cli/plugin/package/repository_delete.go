// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/log"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackageclient"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackagedatamodel"
)

var repoDeleteOp = tkgpackagedatamodel.NewRepositoryDeleteOptions()

var repositoryDeleteCmd = &cobra.Command{
	Use:   "delete REPOSITORY_NAME",
	Short: "Delete a repository",
	Args:  cobra.ExactArgs(1),
	RunE:  repositoryDelete,
}

func init() {
	repositoryDeleteCmd.Flags().BoolVarP(&repoDeleteOp.IsForce, "force", "f", false, "Force deletion of the repository")
	repositoryDeleteCmd.Flags().StringVarP(&repoDeleteOp.KubeConfig, "kubeconfig", "", "", "The path to the kubeconfig file, optional")
	repositoryCmd.AddCommand(repositoryDeleteCmd)
}

func repositoryDelete(_ *cobra.Command, args []string) error {
	if len(args) == 1 {
		repoDeleteOp.RepositoryName = args[0]
	} else {
		return errors.New("incorrect number of input parameters. Usage: tanzu package repository delete REPO_NAME [FLAGS]")
	}

	pkgClient, err := tkgpackageclient.NewTKGPackageClient(repoDeleteOp.KubeConfig)
	if err != nil {
		return err
	}

	found, err := pkgClient.DeleteRepository(repoDeleteOp)
	if !found {
		log.Warningf("Could not find package repository '%s'\n", repoDeleteOp.RepositoryName)
		return nil
	} else if err != nil {
		return err
	}

	log.Infof("Deleted package repository '%s'\n", repoDeleteOp.RepositoryName)

	return nil
}
