// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgpackageclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgpackagedatamodel"
)

var repositoryUpdateCmd = &cobra.Command{
	Use:   "update REPOSITORY_NAME --url REPOSITORY_URL",
	Short: "Update a package repository",
	Args:  cobra.ExactArgs(1),
	Example: `
    # Update repository in default namespace 	
    tanzu package repository update repo --url projects-stg.registry.vmware.com/tkg/standard-repo:v1.0.1 --namespace test-ns`,
	RunE:         repositoryUpdate,
	SilenceUsage: true,
}

func init() {
	repositoryUpdateCmd.Flags().StringVarP(&repoOp.RepositoryURL, "url", "", "", "OCI registry url for package repository bundle")
	repositoryUpdateCmd.Flags().BoolVarP(&repoOp.CreateRepository, "create", "", false, "Creates the package repository if it does not exist, optional")
	repositoryUpdateCmd.Flags().BoolVarP(&repoOp.CreateNamespace, "create-namespace", "", false, "Create namespace if the target namespace does not exist, optional")
	repositoryUpdateCmd.Flags().BoolVarP(&repoOp.Wait, "wait", "", true, "Wait for the package repository reconciliation to complete, optional. To disable wait, specify --wait=false")
	repositoryUpdateCmd.Flags().DurationVarP(&repoOp.PollInterval, "poll-interval", "", tkgpackagedatamodel.DefaultPollInterval, "Time interval between subsequent polls of package repository reconciliation status, optional")
	repositoryUpdateCmd.Flags().DurationVarP(&repoOp.PollTimeout, "poll-timeout", "", tkgpackagedatamodel.DefaultPollTimeout, "Timeout value for polls of package repository reconciliation status, optional")
	repositoryUpdateCmd.MarkFlagRequired("url") //nolint
	repositoryCmd.AddCommand(repositoryUpdateCmd)
}

func repositoryUpdate(_ *cobra.Command, args []string) error {
	repoOp.RepositoryName = args[0]

	pkgClient, err := tkgpackageclient.NewTKGPackageClient(kubeConfig)
	if err != nil {
		return err
	}

	return pkgClient.UpdateRepositorySync(repoOp, tkgpackagedatamodel.OperationTypeUpdate)
}
