// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/packageclients/pkg/packageclient"
	"github.com/vmware-tanzu/tanzu-framework/packageclients/pkg/packagedatamodel"
)

var repositoryAddCmd = &cobra.Command{
	Use:   "add REPOSITORY_NAME --url REPOSITORY_URL",
	Short: "Add a package repository",
	Args:  cobra.ExactArgs(1),
	Example: `
    # Add a repository in specified namespace which does not exist 	
    tanzu package repository add repo --url projects-stg.registry.vmware.com/tkg/standard-repo:v1.0.0 --namespace test-ns --create-namespace`,
	RunE:         repositoryAdd,
	SilenceUsage: true,
}

func init() {
	repositoryAddCmd.Flags().StringVarP(&repoOp.RepositoryURL, "url", "", "", "OCI registry url for package repository bundle")
	repositoryAddCmd.Flags().BoolVarP(&repoOp.CreateNamespace, "create-namespace", "", false, "Create namespace if the target namespace does not exist, optional")
	repositoryAddCmd.Flags().BoolVarP(&repoOp.Wait, "wait", "", true, "Wait for the package repository reconciliation to complete, optional. To disable wait, specify --wait=false")
	repositoryAddCmd.Flags().DurationVarP(&repoOp.PollInterval, "poll-interval", "", packagedatamodel.DefaultPollInterval, "Time interval between subsequent polls of package repository reconciliation status, optional")
	repositoryAddCmd.Flags().DurationVarP(&repoOp.PollTimeout, "poll-timeout", "", packagedatamodel.DefaultPollTimeout, "Timeout value for polls of package repository reconciliation status, optional")
	repositoryAddCmd.MarkFlagRequired("url") //nolint
	repositoryCmd.AddCommand(repositoryAddCmd)
}

func repositoryAdd(_ *cobra.Command, args []string) error {
	repoOp.RepositoryName = args[0]

	pkgClient, err := packageclient.NewPackageClient(kubeConfig)
	if err != nil {
		return err
	}

	return pkgClient.AddRepositorySync(repoOp, packagedatamodel.OperationTypeInstall)
}
