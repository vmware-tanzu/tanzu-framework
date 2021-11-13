// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/pluginmanager"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"

	"github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
)

// upgradeRegionCmd represents the upgrade command
var upgradeRegionCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrades the management cluster",
	Args:  cobra.MaximumNArgs(1), // TODO: deprecate version of command that takes args
	Example: `
  # Upgrade a management cluster
  tanzu management-cluster upgrade

  # Upgrade a management cluster using specific os name (vsphere)
  tanzu management-cluster upgrade mc-1 --os-name photon

  # Upgrade a management cluster using specific os name and version
  tanzu management-cluster upgrade mc-1 --os-name ubuntu --os-version 20.04

  # Upgrade a management cluster using specific os name, version and arch
  tanzu management-cluster upgrade mc-1 --os-name ubuntu --os-version 20.04 --os-arch amd64

  [+] : Options available for: os-name, os-version, os-arch are as follows:
  vSphere: 
    --os-name ubuntu --os-version 20.04 --os-arch amd64
    --os-name photon --os-version 3 --os-arch amd64	
  aws: 
    --os-name ubuntu --os-version 20.04 --os-arch amd64
    --os-name amazon --os-version 2 --os-arch amd64
  azure:
    --os-name ubuntu --os-version 20.04 --os-arch amd64
    --os-name ubuntu --os-version 18.04 --os-arch amd64
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runForCurrentMC(runUpgradeRegion)
	},
}

type upgradeRegionOptions struct {
	vSphereTemplateName string
	timeout             time.Duration
	unattended          bool
	osName              string
	osVersion           string
	osArch              string
}

var ur = &upgradeRegionOptions{}

func init() {
	upgradeRegionCmd.Flags().StringVarP(&ur.vSphereTemplateName, "vsphere-vm-template-name", "", "", "The vSphere VM template to be used with upgraded kubernetes version. Discovered automatically if not provided")
	upgradeRegionCmd.Flags().MarkHidden("vsphere-vm-template-name") //nolint

	upgradeRegionCmd.Flags().DurationVarP(&ur.timeout, "timeout", "t", constants.DefaultLongRunningOperationTimeout, "Time duration to wait for an operation before timeout. Timeout duration in hours(h)/minutes(m)/seconds(s) units or as some combination of them (e.g. 2h, 30m, 2h30m10s)")
	upgradeRegionCmd.Flags().BoolVarP(&ur.unattended, "yes", "y", false, "Upgrade management cluster without asking for confirmation")
	upgradeRegionCmd.Flags().StringVar(&ur.osName, "os-name", "", "OS name to use during management cluster upgrade. Discovered automatically if not provided (See [+])")
	upgradeRegionCmd.Flags().StringVar(&ur.osVersion, "os-version", "", "OS version to use during management cluster upgrade. Discovered automatically if not provided (See [+])")
	upgradeRegionCmd.Flags().StringVar(&ur.osArch, "os-arch", "", "OS arch to use during management cluster upgrade. Discovered automatically if not provided (See [+])")
}

func runUpgradeRegion(server *v1alpha1.Server) error {
	forceUpdateTKGCompatibilityImage := false
	tkgClient, err := newTKGCtlClient(forceUpdateTKGCompatibilityImage)
	if err != nil {
		return err
	}

	options := tkgctl.UpgradeRegionOptions{
		ClusterName:         server.Name,
		VSphereTemplateName: ur.vSphereTemplateName,
		SkipPrompt:          ur.unattended,
		Timeout:             ur.timeout,
		OSName:              ur.osName,
		OSVersion:           ur.osVersion,
		OSArch:              ur.osArch,
		Edition:             BuildEdition,
	}

	err = tkgClient.UpgradeRegion(options)
	if err != nil {
		return err
	}

	// Sync plugins if management-cluster upgrade is successful
	if config.IsFeatureActivated(config.FeatureContextAwareDiscovery) {
		err = pluginmanager.SyncPlugins(server.Name)
		if err != nil {
			log.Warningf("unable to sync plugins after management cluster upgrade. Please run `tanzu plugin sync` command manually to install/update plugins")
		}
	}

	return nil
}
