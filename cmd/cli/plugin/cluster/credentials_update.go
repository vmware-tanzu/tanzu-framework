// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"errors"
	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"
)

type updateCredentialsOptions struct {
	namespace       string
	vSphereUser     string
	vSpherePassword string
}

var updateCredentialsOpts = updateCredentialsOptions{}

var credentialsUpdateCmd = &cobra.Command{
	Use:          "update CLUSTER_NAME",
	Short:        "Update credentials for a cluster",
	Args:         cobra.ExactArgs(1),
	RunE:         updateCredentials,
	SilenceUsage: true,
}

func init() {
	credentialsUpdateCmd.Flags().StringVarP(&updateCredentialsOpts.namespace, "namespace", "n", "", "The namespace of the cluster to be updated")
	credentialsUpdateCmd.Flags().StringVarP(&updateCredentialsOpts.vSphereUser, "vsphere-user", "", "", "Username for vSphere provider")
	credentialsUpdateCmd.Flags().StringVarP(&updateCredentialsOpts.vSpherePassword, "vsphere-password", "", "", "Password for vSphere provider")

	credentialsCmd.AddCommand(credentialsUpdateCmd)
}

func updateCredentials(cmd *cobra.Command, args []string) error {
	server, err := config.GetCurrentServer()
	if err != nil {
		return err
	}

	if server.IsGlobal() {
		return errors.New("creating cluster with a global server is not implemented yet")
	}
	return updateClusterCredentials(args[0], server)
}

func updateClusterCredentials(clusterName string, server *v1alpha1.Server) error {
	tkgctlClient, err := createTKGClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context)
	if err != nil {
		return err
	}

	if updateCredentialsOpts.namespace == "" {
		prompt := &survey.Input{
			Message: "Enter namespace of the cluster:",
			Default: "default",
		}
		err = survey.AskOne(prompt, &updateCredentialsOpts.namespace, cli.SurveyOptions())
		if err != nil {
			return err
		}
	}

	if updateCredentialsOpts.vSphereUser == "" {
		prompt := &survey.Input{
			Message: "Enter vSphere username:",
		}
		err = survey.AskOne(prompt, &updateCredentialsOpts.vSphereUser, cli.SurveyOptions())
		if err != nil {
			return err
		}
	}

	if updateCredentialsOpts.vSpherePassword == "" {
		prompt := &survey.Password{
			Message: "Enter vSphere password:",
		}
		err = survey.AskOne(prompt, &updateCredentialsOpts.vSpherePassword, cli.SurveyOptions())
		if err != nil {
			return err
		}
	}

	uccOptions := tkgctl.UpdateCredentialsClusterOptions{
		ClusterName:     clusterName,
		Namespace:       updateCredentialsOpts.namespace,
		VSphereUsername: updateCredentialsOpts.vSphereUser,
		VSpherePassword: updateCredentialsOpts.vSpherePassword,
	}

	return tkgctlClient.UpdateCredentialsCluster(uccOptions)
}
