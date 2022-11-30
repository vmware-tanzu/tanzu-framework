// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"time"

	"github.com/spf13/cobra"

	cliconfig "github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/config"
	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/pluginmanager"
	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgctl"
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
	SilenceUsage: true,
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

func runUpgradeRegion(server *configapi.Server) error {
	forceUpdateTKGCompatibilityImage := false
	tkgClient, err := newTKGCtlClient(forceUpdateTKGCompatibilityImage)
	if err != nil {
		return err
	}

	edition, err := config.GetEdition()
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
		Edition:             edition,
	}

	err = tkgClient.UpgradeRegion(options)
	if err != nil {
		return err
	}

	// Sync plugins if management-cluster upgrade is successful
	if config.IsFeatureActivated(cliconfig.FeatureContextAwareCLIForPlugins) {
		err = pluginmanager.SyncPlugins()
		if err != nil {
			log.Warningf("unable to sync plugins after management cluster upgrade. Please run `tanzu plugin sync` command manually to install/update plugins")
		}
	}

	return nil
}
