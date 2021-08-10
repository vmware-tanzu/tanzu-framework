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

var repositoryUpdateCmd = &cobra.Command{
	Use:   "update REPOSITORY_NAME ",
	Short: "Update a repository",
	Args:  cobra.ExactArgs(1),
	Example: `
    # Update repository in default namespace 	
    tanzu package repository update repo --url projects-stg.registry.vmware.com/tkg/standard-repo:v1.0.1 --namespace test-ns`,
	RunE: repositoryUpdate,
}

func init() {
	repositoryUpdateCmd.Flags().StringVarP(&repoOp.RepositoryURL, "url", "", "", "OCI registry url for package repository bundle")
	repositoryUpdateCmd.Flags().BoolVarP(&repoOp.CreateRepository, "create", "", false, "Creates the repository if it does not exist")
	repositoryUpdateCmd.Flags().BoolVarP(&repoOp.CreateNamespace, "create-namespace", "", false, "Create namespace if the target namespace does not exist, optional")
	repositoryUpdateCmd.MarkFlagRequired("url") //nolint
	repositoryCmd.AddCommand(repositoryUpdateCmd)
}

func repositoryUpdate(cmd *cobra.Command, args []string) error {
	repoOp.RepositoryName = args[0]

	pkgClient, err := tkgpackageclient.NewTKGPackageClient(repoOp.KubeConfig)
	if err != nil {
		return err
	}

	_, err = component.NewOutputWriterWithSpinner(cmd.OutOrStdout(), outputFormat,
		fmt.Sprintf("Updating package repository '%s'...", repoOp.RepositoryName), true)
	if err != nil {
		return err
	}

	if err := pkgClient.UpdateRepository(repoOp); err != nil {
		return err
	}

	log.Infof("\n Updated package repository '%s' in namespace '%s'", repoOp.RepositoryName, repoOp.Namespace)

	return nil
}
