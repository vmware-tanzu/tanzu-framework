// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"
)

// upgradeRegionCmd represents the upgrade command
var upgradeRegionCmd = &cobra.Command{
	Use:     "management-cluster CLUSTER_NAME",
	Short:   "Upgrade a management cluster",
	Aliases: []string{"mc"},
	Args:    cobra.ExactArgs(1),
	Example: Examples(`
	# Upgrade management cluster
	tkg upgrade cluster mc-1

	# Upgrade management cluster using specific os name (vsphere)
	tkg upgrade cluster mc-1 --os-name photon

	# Upgrade management cluster using specific os name and version
	tkg upgrade cluster mc-1 --os-name ubuntu --os-version 20.04

	# Upgrade management cluster using specific os name, version and arch
	tkg upgrade cluster mc-1 --os-name ubuntu --os-version 20.04 --os-arch amd64

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
`),
	Run: func(cmd *cobra.Command, args []string) {
		displayLogFileLocation()
		err := runUpgradeRegion(args[0])
		verifyCommandError(err)
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
	upgradeRegionCmd.Flags().DurationVarP(&ur.timeout, "timeout", "t", constants.DefaultLongRunningOperationTimeout, "Time duration to wait for an operation before timeout. Timeout duration in hours(h)/minutes(m)/seconds(s) units or as some combination of them (e.g. 2h, 30m, 2h30m10s)")
	upgradeRegionCmd.Flags().BoolVarP(&ur.unattended, "yes", "y", false, "Upgrade management cluster without asking for confirmation")

	upgradeRegionCmd.Flags().StringVarP(&ur.osName, "os-name", "", "", "OS name to use during management cluster upgrade. Discovered automatically if not provided (See [+])")
	upgradeRegionCmd.Flags().StringVarP(&ur.osVersion, "os-version", "", "", "OS version to use during management cluster upgrade. Discovered automatically if not provided (See [+])")
	upgradeRegionCmd.Flags().StringVarP(&ur.osArch, "os-arch", "", "", "OS arch to use during management cluster upgrade. Discovered automatically if not provided (See [+])")

	upgradeCmd.AddCommand(upgradeRegionCmd)
}

func runUpgradeRegion(clusterName string) error {
	tkgClient, err := newTKGCtlClient()
	if err != nil {
		return err
	}

	options := tkgctl.UpgradeRegionOptions{
		ClusterName:         clusterName,
		VSphereTemplateName: ur.vSphereTemplateName,
		SkipPrompt:          ur.unattended || skipPrompt,
		Timeout:             ur.timeout,
		OSName:              ur.osName,
		OSVersion:           ur.osVersion,
		OSArch:              ur.osArch,
	}
	return tkgClient.UpgradeRegion(options)
}
