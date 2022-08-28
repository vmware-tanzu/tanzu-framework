// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/component"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"
	tkrconstants "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/types"
	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
)

var OsCmd = &cobra.Command{
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
	getOSCmd.Flags().StringVarP(&outputFormat, "output", "o", "", "Output format (yaml|json|table)")
	OsCmd.AddCommand(getOSCmd)
}

//nolint:gocyclo,funlen
func getOS(cmd *cobra.Command, args []string) error {
	server, err := config.GetCurrentServer()
	if err != nil {
		return err
	}

	if server.IsGlobal() {
		return errors.New("getting TanzuKubernetesRelease with a global server is not implemented yet")
	}

	clusterClientOptions := clusterclient.Options{GetClientInterval: 2 * time.Second, GetClientTimeout: 5 * time.Second}
	clusterClient, err := clusterclient.NewClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context, clusterClientOptions)
	if err != nil {
		return err
	}

	bomConfigMap, err := clusterClient.GetBomConfigMap(args[0])
	if err != nil {
		return err
	}

	bomByte, ok := bomConfigMap.BinaryData[tkrconstants.BomConfigMapContentKey]
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
	case constants.InfrastructureRefVSphere:
		ovas, err := bom.GetOVAInfo()
		if err != nil {
			return errors.Wrap(err, "failed to get vSphere OVA info from the BOM file")
		}
		for _, ova := range ovas {
			osMap[ova.OSInfo.String()] = ova.OSInfo
		}

	case constants.InfrastructureRefAWS:
		amiMap, err := bom.GetAMIInfo()
		if err != nil {
			return errors.Wrap(err, "failed to get AWS AMI info from the BOM file")
		}

		if goo.region == "" {
			return errors.New("you are currently on an AWS management cluster. Please specify a region")
		}

		amis, ok := amiMap[goo.region]
		if !ok {
			return errors.Errorf("failed to find os info in region %s", goo.region)
		}

		for _, ami := range amis {
			osMap[ami.OSInfo.String()] = ami.OSInfo
		}

	case constants.InfrastructureRefAzure:
		azureImages, err := bom.GetAzureInfo()
		if err != nil {
			return errors.Wrap(err, "failed to get Azure image info from the BOM file")
		}
		for i := range azureImages {
			osMap[azureImages[i].OSInfo.String()] = azureImages[i].OSInfo
		}
	}

	t := component.NewOutputWriter(cmd.OutOrStdout(), outputFormat, "NAME", "VERSION", "ARCH")
	for _, os := range osMap {
		t.AddRow(os.Name, os.Version, os.Arch)
	}
	t.Render()

	return nil
}
