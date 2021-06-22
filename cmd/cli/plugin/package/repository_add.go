// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/log"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackageclient"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackagedatamodel"
)

var repoAddOp = tkgpackagedatamodel.NewRepositoryOptions()

var repositoryAddCmd = &cobra.Command{
	Use:   "add REPOSITORY_NAME REPOSITORY_URL ",
	Short: "Add a repository",
	Args:  cobra.ExactArgs(2),
	RunE:  repositoryAdd,
}

func init() {
	repositoryAddCmd.Flags().StringVarP(&repoAddOp.KubeConfig, "kubeconfig", "", "", "The path to the kubeconfig file, optional")
	repositoryAddCmd.Flags().BoolVarP(&repoAddOp.CreateNamespace, "create-namespace", "", false, "Create namespace if the target namespace does not exist, optional")
	repositoryAddCmd.Flags().StringVarP(&repoAddOp.Namespace, "namespace", "n", "default", "Target namespace to add the repository, optional")

	repositoryCmd.AddCommand(repositoryAddCmd)
}

func repositoryAdd(_ *cobra.Command, args []string) error {
	repoAddOp.RepositoryName = args[0]
	repoAddOp.RepositoryURL = args[1]

	pkgClient, err := tkgpackageclient.NewTKGPackageClient(repoAddOp.KubeConfig)
	if err != nil {
		return err
	}

	if err := pkgClient.AddRepository(repoAddOp); err != nil {
		return err
	}

	log.Infof("Added package repository '%s'", repoAddOp.RepositoryName)

	return nil
}
