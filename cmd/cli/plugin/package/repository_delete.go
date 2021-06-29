// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/log"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackageclient"
)

var repositoryDeleteCmd = &cobra.Command{
	Use:   "delete REPOSITORY_NAME",
	Short: "Delete a repository",
	Args:  cobra.ExactArgs(1),
	Example: `
    # Delete a repository in specified namespace 	
    tanzu package repository delete repo --namespace test-ns`,
	RunE: repositoryDelete,
}

func init() {
	repositoryDeleteCmd.Flags().BoolVarP(&repoOp.IsForceDelete, "force", "f", false, "Force deletion of the repository")
	repositoryCmd.AddCommand(repositoryDeleteCmd)
}

func repositoryDelete(_ *cobra.Command, args []string) error {
	if len(args) == 1 {
		repoOp.RepositoryName = args[0]
	} else {
		return errors.New("incorrect number of input parameters. Usage: tanzu package repository delete REPO_NAME [FLAGS]")
	}

	pkgClient, err := tkgpackageclient.NewTKGPackageClient(repoOp.KubeConfig)
	if err != nil {
		return err
	}

	found, err := pkgClient.DeleteRepository(repoOp)
	if !found {
		log.Warningf("Could not find package repository '%s' in namespace '%s'\n", repoOp.RepositoryName, repoOp.Namespace)
		return nil
	}
	if err != nil {
		return err
	}

	log.Infof("Deleted package repository '%s' in namespace '%s'\n", repoOp.RepositoryName, repoOp.Namespace)

	return nil
}
