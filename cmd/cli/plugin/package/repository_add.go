// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/component"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgpackageclient"
)

var repositoryAddCmd = &cobra.Command{
	Use:   "add REPOSITORY_NAME --url REPOSITORY_URL",
	Short: "Add a package repository",
	Args:  cobra.ExactArgs(1),
	Example: `
    # Add a repository in specified namespace which does not exist 	
    tanzu package repository add repo --url projects-stg.registry.vmware.com/tkg/standard-repo:v1.0.0 --namespace test-ns --create-namespace`,
	RunE: repositoryAdd,
}

func init() {
	repositoryAddCmd.Flags().StringVarP(&repoOp.RepositoryURL, "url", "", "", "OCI registry url for package repository bundle")
	repositoryAddCmd.Flags().BoolVarP(&repoOp.CreateNamespace, "create-namespace", "", false, "Create namespace if the target namespace does not exist, optional")
	repositoryAddCmd.MarkFlagRequired("url") //nolint
	repositoryCmd.AddCommand(repositoryAddCmd)
}

func repositoryAdd(cmd *cobra.Command, args []string) error {
	repoOp.RepositoryName = args[0]

	pkgClient, err := tkgpackageclient.NewTKGPackageClient(repoOp.KubeConfig)
	if err != nil {
		return err
	}

	_, err = component.NewOutputWriterWithSpinner(cmd.OutOrStdout(), outputFormat,
		fmt.Sprintf("Adding package repository '%s'...", repoOp.RepositoryName), true)
	if err != nil {
		return err
	}

	if err := pkgClient.AddRepository(repoOp); err != nil {
		return err
	}

	log.Infof("\n Added package repository '%s'", repoOp.RepositoryName)

	return nil
}
