// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha3

import (
	"io"
	"strings"
	"time"

	runv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/component"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"
	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
)

var (
	LogLevel     int32
	LogFile      string
	outputFormat string
)

type getTKROptions struct {
	listAll bool
	output  io.Writer
}

var gtkr = &getTKROptions{}

var GetTanzuKubernetesRleasesCmd = &cobra.Command{
	Use:   "get TKR_NAME",
	Short: "Get available Tanzu Kubernetes releases",
	Long:  "Get available Tanzu Kubernetes releases",
	RunE:  getKubernetesReleases,
}

func init() {
	GetTanzuKubernetesRleasesCmd.Flags().BoolVarP(&gtkr.listAll, "all", "a", false, "List all the available Tanzu Kubernetes releases including Incompatible and deactivated")
	GetTanzuKubernetesRleasesCmd.Flags().StringVarP(&outputFormat, "output", "o", "", "Output format (yaml|json|table)")
}

func getKubernetesReleases(cmd *cobra.Command, args []string) error {
	server, err := config.GetCurrentServer()
	if err != nil {
		return err
	}

	if server.IsGlobal() {
		return errors.New("getting Tanzu Kubernetes release with a global server is not implemented yet")
	}

	clusterClientOptions := clusterclient.Options{GetClientInterval: 2 * time.Second, GetClientTimeout: 5 * time.Second}
	clusterClient, err := clusterclient.NewClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context, clusterClientOptions)
	if err != nil {
		return err
	}
	tkrName := ""
	if len(args) != 0 {
		tkrName = args[0]
	}
	gtkr.output = cmd.OutOrStdout()
	return runGetKubernetesReleases(clusterClient, tkrName)
}

func runGetKubernetesReleases(clusterClient clusterclient.Client, tkrName string) error {
	tkrs, err := getTKRs(clusterClient, tkrName)
	if err != nil {
		return err
	}

	t := component.NewOutputWriter(gtkr.output, outputFormat, "NAME", "VERSION", "COMPATIBLE", "ACTIVE")
	for i := range tkrs {
		compatible := ""
		for _, condition := range tkrs[i].Status.Conditions {
			if condition.Type == runv1alpha3.ConditionCompatible {
				compatible = string(condition.Status)
			}
		}
		labels := tkrs[i].Labels
		activeStatus := "True"
		if labels != nil {
			if _, exists := labels[runv1alpha3.LabelDeactivated]; exists {
				activeStatus = "False"
			}
		}

		if !gtkr.listAll && (!strings.EqualFold(compatible, "true") || !strings.EqualFold(activeStatus, "true")) {
			continue
		}
		t.AddRow(tkrs[i].Name, tkrs[i].Spec.Version, compatible, activeStatus)
	}
	t.Render()
	return nil
}

func getTKRs(clusterClient clusterclient.Client, tkrName string) ([]runv1alpha3.TanzuKubernetesRelease, error) {
	var tkrList runv1alpha3.TanzuKubernetesReleaseList

	err := clusterClient.ListResources(&tkrList)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list TKr's")
	}
	if tkrName == "" {
		return tkrList.Items, nil
	}

	result := []runv1alpha3.TanzuKubernetesRelease{}
	for i := range tkrList.Items {
		if strings.HasPrefix(tkrList.Items[i].Name, tkrName) {
			result = append(result, tkrList.Items[i])
		}
	}
	return result, nil
}
