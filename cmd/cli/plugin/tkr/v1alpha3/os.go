// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha3

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	runv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/component"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"
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
var awsOsImageRefRegionKey = "region"

func init() {
	getOSCmd.Flags().StringVarP(&goo.region, "region", "", "", "The AWS region where AMIs are available")
	getOSCmd.Flags().StringVarP(&outputFormat, "output", "o", "", "Output format (yaml|json|table)")
	OsCmd.AddCommand(getOSCmd)
}

func getOS(cmd *cobra.Command, args []string) error {
	clusterClient, err := getClusterClient()
	if err != nil {
		return err
	}

	tkrName := args[0]
	osInfoMap, err := osInfoByTKR(clusterClient, tkrName, goo)
	if err != nil {
		return nil
	}
	t := component.NewOutputWriter(cmd.OutOrStdout(), outputFormat, "NAME", "VERSION", "ARCH")
	for _, os := range osInfoMap {
		t.AddRow(os.Name, os.Version, os.Arch)
	}
	t.Render()
	return nil
}

func osInfoByTKR(clusterClient clusterclient.Client, tkrName string, options *getOSOptions) (map[string]runv1alpha3.OSInfo, error) {
	osImages, err := osImagesByTKR(clusterClient, tkrName)
	if err != nil {
		return nil, err
	}
	resultOSImages, err := filterOSImagesWithInfraSpecificOptions(clusterClient, options, osImages)
	if err != nil {
		return nil, err
	}
	return osInfoOfImages(resultOSImages), nil
}

type stringSet map[string]struct{}

func (set stringSet) Add(ss ...string) stringSet {
	for _, s := range ss {
		set[s] = struct{}{}
	}
	return set
}

func (set stringSet) Has(s string) bool {
	_, exists := set[s]
	return exists
}

func osImagesByTKR(clusterClient clusterclient.Client, tkrName string) ([]runv1alpha3.OSImage, error) {
	osImages, err := getOSImages(clusterClient)
	if err != nil {
		return nil, err
	}
	osImageNamesInTKR, err := getOSImageNamesInTKR(clusterClient, tkrName)
	if err != nil {
		return nil, err
	}
	candidates := stringSet{}.Add(osImageNamesInTKR...)
	result := filterOsImages(osImages, func(image *runv1alpha3.OSImage) bool {
		return candidates.Has(image.Name)
	})
	return result, nil
}

func osInfoOfImages(osImages []runv1alpha3.OSImage) map[string]runv1alpha3.OSInfo {
	results := map[string]runv1alpha3.OSInfo{}
	for i := range osImages {
		osInfo := osImages[i].Spec.OS
		results[OsInfoString(osInfo)] = osInfo
	}
	return results
}
func filterOSImagesWithInfraSpecificOptions(clusterClient clusterclient.Client,
	options *getOSOptions, osImages []runv1alpha3.OSImage) ([]runv1alpha3.OSImage, error) {

	infra, err := clusterClient.GetClusterInfrastructure()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get current management cluster infrastructure")
	}
	if infra == constants.InfrastructureRefAWS && options.region != "" {
		osImages = filterOsImages(osImages, func(osImage *runv1alpha3.OSImage) bool {
			if regionName, exists := osImage.Spec.Image.Ref[awsOsImageRefRegionKey]; exists && regionName == options.region {
				return true
			}
			return false
		})
	}
	return osImages, nil
}
func getOSImages(clusterClient clusterclient.Client) ([]runv1alpha3.OSImage, error) {
	var osImageList runv1alpha3.OSImageList
	err := clusterClient.ListResources(&osImageList)
	if err != nil {
		return nil, err
	}
	var osImages []runv1alpha3.OSImage
	for i := range osImageList.Items {
		osImages = append(osImages, osImageList.Items[i])
	}
	return osImages, nil
}

func getOSImageNamesInTKR(clusterClient clusterclient.Client, tkrName string) ([]string, error) {
	var tkr runv1alpha3.TanzuKubernetesRelease
	err := clusterClient.GetResource(&tkr, tkrName, "", nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get TKR ")
	}

	osImageNamesInTKR := make([]string, len(tkr.Spec.OSImages))
	for i := range tkr.Spec.OSImages {
		osImageNamesInTKR[i] = tkr.Spec.OSImages[i].Name
	}
	return osImageNamesInTKR, nil
}

type osImagePredicate func(osImage *runv1alpha3.OSImage) bool

func filterOsImages(osImages []runv1alpha3.OSImage, p osImagePredicate) []runv1alpha3.OSImage {
	result := make([]runv1alpha3.OSImage, 0, len(osImages))
	for i := range osImages {
		if p(&osImages[i]) {
			result = append(result, osImages[i])
		}
	}
	return result
}

func OsInfoString(os runv1alpha3.OSInfo) string {
	return fmt.Sprintf("%s-%s-%s", os.Name, os.Version, os.Arch)
}

func getClusterClient() (clusterclient.Client, error) {
	server, err := config.GetCurrentServer()
	if err != nil {
		return nil, err
	}

	if server.IsGlobal() {
		return nil, errors.New("getting TanzuKubernetesRelease with a global server is not implemented yet")
	}

	clusterClientOptions := clusterclient.Options{GetClientInterval: 2 * time.Second, GetClientTimeout: 5 * time.Second}
	clusterClient, err := clusterclient.NewClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context, clusterClientOptions)
	if err != nil {
		return nil, err
	}
	return clusterClient, nil
}
