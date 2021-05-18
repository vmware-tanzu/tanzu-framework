// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgctl"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/utils"
)

// upgradeClusterCmd represents the upgrade command
var upgradeClusterCmd = &cobra.Command{
	Use:   "cluster CLUSTER_NAME",
	Short: "Upgrade a Tanzu Kubernetes Grid cluster",
	Args:  cobra.ExactArgs(1),
	Example: Examples(`
	# Upgrade workload cluster
	tkg upgrade cluster wc-1

	# Upgrade workload cluster with specific tkr v1.20.1---vmware.1-tkg.2
	tkg upgrade cluster wc-1 --tkr v1.20.1---vmware.1-tkg.2

	# Upgrade workload cluster using specific os name (vsphere)
	tkg upgrade cluster wc-1 --os-name photon

	# Upgrade workload cluster using specific os name and version
	tkg upgrade cluster wc-1 --os-name ubuntu --os-version 20.04

	# Upgrade workload cluster using specific os name, version and arch
	tkg upgrade cluster wc-1 --os-name ubuntu --os-version 20.04 --os-arch amd64

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
		err := runUpgradeCluster(args[0])
		verifyCommandError(err)
	},
}

type upgradeClusterOptions struct {
	namespace           string
	tkrName             string
	vSphereTemplateName string
	timeout             time.Duration
	unattended          bool
	osName              string
	osVersion           string
	osArch              string
}

var uc = &upgradeClusterOptions{}

func init() {
	upgradeClusterCmd.Flags().StringVarP(&uc.tkrName, "tkr", "", "", "TanzuKubernetesRelease(TKR) to be used for creating the workload cluster")
	upgradeClusterCmd.Flags().StringVarP(&uc.namespace, "namespace", "n", "", "The namespace where the workload cluster was created. Assumes 'default' if not specified")
	upgradeClusterCmd.Flags().StringVarP(&uc.vSphereTemplateName, "vsphere-vm-template-name", "", "", "The vSphere VM template to be used with upgraded kubernetes version. Discovered automatically if not provided")
	upgradeClusterCmd.Flags().DurationVarP(&uc.timeout, "timeout", "t", constants.DefaultLongRunningOperationTimeout, "Time duration to wait for an operation before timeout. Timeout duration in hours(h)/minutes(m)/seconds(s) units or as some combination of them (e.g. 2h, 30m, 2h30m10s)")
	upgradeClusterCmd.Flags().BoolVarP(&uc.unattended, "yes", "y", false, "Upgrade workload cluster without asking for confirmation")

	upgradeClusterCmd.Flags().StringVarP(&uc.osName, "os-name", "", "", "OS name to use during management cluster upgrade. Discovered automatically if not provided (See [+])")
	upgradeClusterCmd.Flags().StringVarP(&uc.osVersion, "os-version", "", "", "OS version to use during management cluster upgrade. Discovered automatically if not provided (See [+])")
	upgradeClusterCmd.Flags().StringVarP(&uc.osArch, "os-arch", "", "", "OS arch to use during management cluster upgrade. Discovered automatically if not provided (See [+])")

	upgradeCmd.AddCommand(upgradeClusterCmd)
}

func runUpgradeCluster(clusterName string) error {
	tkgClient, err := newTKGCtlClient()
	if err != nil {
		return err
	}

	tkrVersion := ""
	if uc.tkrName != "" {
		tkrVersion = utils.GetTKRVersionFromTKRName(uc.tkrName)
	}

	options := tkgctl.UpgradeClusterOptions{
		ClusterName:         clusterName,
		Namespace:           uc.namespace,
		TkrVersion:          tkrVersion,
		VSphereTemplateName: uc.vSphereTemplateName,
		SkipPrompt:          uc.unattended || skipPrompt,
		Timeout:             uc.timeout,
		OSName:              uc.osName,
		OSVersion:           uc.osVersion,
		OSArch:              uc.osArch,
	}
	return tkgClient.UpgradeCluster(options)
}
