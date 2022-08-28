// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgpackageclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgpackagedatamodel"
)

var repositoryDeleteCmd = &cobra.Command{
	Use:   "delete REPOSITORY_NAME",
	Short: "Delete a package repository",
	Args:  cobra.ExactArgs(1),
	Example: `
    # Delete a repository in specified namespace 	
    tanzu package repository delete repo --namespace test-ns`,
	RunE:         repositoryDelete,
	SilenceUsage: true,
}

func init() {
	repositoryDeleteCmd.Flags().BoolVarP(&repoOp.IsForceDelete, "force", "f", false, "Force deletion of the package repository, optional")
	repositoryDeleteCmd.Flags().BoolVarP(&repoOp.Wait, "wait", "", true, "Wait for the package repository reconciliation to complete, optional. To disable wait, specify --wait=false")
	repositoryDeleteCmd.Flags().DurationVarP(&repoOp.PollInterval, "poll-interval", "", tkgpackagedatamodel.DefaultPollInterval, "Time interval between subsequent polls of package repository reconciliation status, optional")
	repositoryDeleteCmd.Flags().DurationVarP(&repoOp.PollTimeout, "poll-timeout", "", tkgpackagedatamodel.DefaultPollTimeout, "Timeout value for polls of package repository reconciliation status, optional")
	repositoryDeleteCmd.Flags().BoolVarP(&repoOp.SkipPrompt, "yes", "y", false, "Delete package repository without asking for confirmation, optional")
	repositoryCmd.AddCommand(repositoryDeleteCmd)
}

func repositoryDelete(cmd *cobra.Command, args []string) error {
	if len(args) == 1 {
		repoOp.RepositoryName = args[0]
	} else {
		return errors.New("incorrect number of input parameters. Usage: tanzu package repository delete REPO_NAME [FLAGS]")
	}

	if !repoOp.SkipPrompt {
		if err := cli.AskForConfirmation(fmt.Sprintf("Deleting package repository '%s' in namespace '%s'. Are you sure?",
			repoOp.RepositoryName, repoOp.Namespace)); err != nil {
			return err
		}
	}

	pkgClient, err := tkgpackageclient.NewTKGPackageClient(kubeConfig)
	if err != nil {
		return err
	}

	return pkgClient.DeleteRepositorySync(repoOp)
}
