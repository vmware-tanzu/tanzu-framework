// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/log"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackageclient"
)

var repositoryAddCmd = &cobra.Command{
	Use:   "add REPOSITORY_NAME REPOSITORY_URL ",
	Short: "Add a repository",
	Args:  cobra.ExactArgs(2),
	RunE:  repositoryAdd,
}

func init() {
	repositoryAddCmd.Flags().BoolVarP(&repoOp.CreateNamespace, "create-namespace", "", false, "Create namespace if the target namespace does not exist, optional")

	repositoryCmd.AddCommand(repositoryAddCmd)
}

func repositoryAdd(_ *cobra.Command, args []string) error {
	repoOp.RepositoryName = args[0]
	repoOp.RepositoryURL = args[1]

	pkgClient, err := tkgpackageclient.NewTKGPackageClient(repoOp.KubeConfig)
	if err != nil {
		return err
	}

	if err := pkgClient.AddRepository(repoOp); err != nil {
		return err
	}

	log.Infof("Added package repository '%s'", repoOp.RepositoryName)

	return nil
}
