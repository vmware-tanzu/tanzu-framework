// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"io"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	runv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/component"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/utils"
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
	tkrs, err := clusterClient.GetTanzuKubernetesReleases(tkrName)
	if err != nil {
		return err
	}

	t := component.NewOutputWriter(gtkr.output, outputFormat, "NAME", "VERSION", "COMPATIBLE", "ACTIVE", "UPDATES AVAILABLE")
	for i := range tkrs {
		compatible := ""
		upgradeAvailable := ""
		for _, condition := range tkrs[i].Status.Conditions {
			if condition.Type == runv1alpha1.ConditionCompatible {
				compatible = string(condition.Status)
			}
			// ConditionUpgradeAvailable is deprecated, but check here for backward compatibility
			if condition.Type == runv1alpha1.ConditionUpdatesAvailable || condition.Type == runv1alpha1.ConditionUpgradeAvailable {
				upgradeAvailable = string(condition.Status)
			}
		}
		labels := tkrs[i].Labels
		activeStatus := "True"
		if labels != nil {
			if _, exists := labels[constants.TanzuKubernetesReleaseInactiveLabel]; exists {
				activeStatus = "False"
			}
		}

		if !gtkr.listAll && (!utils.IsTkrCompatible(&tkrs[i]) || !utils.IsTkrActive(&tkrs[i])) {
			continue
		}
		t.AddRow(tkrs[i].Name, tkrs[i].Spec.Version, compatible, activeStatus, upgradeAvailable)
	}
	t.Render()
	return nil
}
