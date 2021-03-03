// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/component"
	"github.com/vmware-tanzu-private/core/pkg/v1/client"
	"github.com/vmware-tanzu-private/core/pkg/v1/clusterclient"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkr/pkg/constants"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkr/pkg/types"
)

var osCmd = &cobra.Command{
	Use:   "os",
	Short: "Get the OS information for a Tanzu Kubernetes Release",
	Long:  `Get the OS information for a Tanzu Kubernetes Release`,
}

type getOSOptions struct {
	region string
}

var goo = &getOSOptions{}

var getOSCmd = &cobra.Command{
	Use:   "get TKR_NAME",
	Short: "Get the OSes that are available for a Tanzu Kubernetes Release",
	Long:  `Get the OSes that are available for a Tanzu Kubernetes Release`,
	Args:  cobra.ExactArgs(1),
	RunE:  getOS,
}

func init() {
	getOSCmd.Flags().StringVarP(&goo.region, "region", "", "", "The AWS region where AMIs are available")
	osCmd.AddCommand(getOSCmd)
}

//nolint
func getOS(cmd *cobra.Command, args []string) error {

	server, err := client.GetCurrentServer()
	if err != nil {
		return err
	}

	if server.IsGlobal() {
		return errors.New("getting TanzuKubernetesRelease with a global server is not implemented yet")
	}

	clusterClient, err := clusterclient.NewClusterClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context)
	if err != nil {
		return err
	}

	bomConfigMap, err := clusterClient.GetBomConfigMap(args[0])
	if err != nil {
		return err
	}

	bomByte, ok := bomConfigMap.BinaryData[constants.BomConfigMapContentKey]
	if !ok {
		return errors.Wrapf(err, "the ConfigMap %s does not contain BOM content", bomConfigMap.Name)
	}
	bom, err := types.NewBom(bomByte)
	if err != nil {
		return errors.Wrap(err, "failed to parse the BOM content")
	}

	infra, err := clusterClient.GetClusterInfrastructure()
	if err != nil {
		return errors.Wrap(err, "failed to get current management cluster infrastructure")
	}

	osMap := make(map[string]types.OSInfo)

	switch infra {
	case clusterclient.InfrastructureRefVSphere:
		ovas, err := bom.GetOVAInfo()
		if err != nil {
			return errors.Wrap(err, "failed to get vSphere OVA info from the BOM file")
		}
		for _, ova := range ovas {
			osMap[ova.OSInfo.String()] = ova.OSInfo
		}

	case clusterclient.InfrastructureRefAWS:
		amiMap, err := bom.GetAMIInfo()
		if err != nil {
			return errors.Wrap(err, "failed to get AWS AMI info from the BOM file")
		}

		if goo.region == "" {
			return errors.New("You are currently on an AWS management cluster. Please specify a region")
		}

		amis, ok := amiMap[goo.region]
		if !ok {
			return errors.Errorf("failed to find os info in region %s", goo.region)
		}

		for _, ami := range amis {
			osMap[ami.OSInfo.String()] = ami.OSInfo
		}

	case clusterclient.InfrastructureRefAzure:
		azureImages, err := bom.GetAzureInfo()
		if err != nil {
			return errors.Wrap(err, "failed to get Azure image info from the BOM file")
		}
		for _, image := range azureImages {
			osMap[image.OSInfo.String()] = image.OSInfo
		}
	}

	t := component.NewTableWriter("NAME", "VERSION", "ARCH")
	for _, os := range osMap {
		t.Append([]string{os.Name, os.Version, os.Arch})
	}
	t.Render()

	return nil
}
