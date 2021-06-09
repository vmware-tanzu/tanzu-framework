// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/utils"
)

type createClusterOptions struct {
	clusterConfigFile           string
	plan                        string
	infrastructureProvider      string
	namespace                   string
	tkrName                     string
	cniType                     string
	enableClusterOptions        string
	vsphereControlPlaneEndpoint string
	size                        string
	controlPlaneSize            string
	workerSize                  string
	timeout                     time.Duration
	controlPlaneMachineCount    int
	workerMachineCount          int
	unattended                  bool
	generateOnly                bool
}

var cc = &createClusterOptions{}

var createClusterCmd = &cobra.Command{
	Use:   "cluster CLUSTER_NAME",
	Short: "Create a Tanzu Kubernetes cluster",
	Long: LongDesc(`
		Use the management cluster to create a Tanzu Kubernetes cluster.`),

	Example: Examples(`
		# Create a workload cluster with a particular plan
		tkg create cluster my-cluster --plan=dev

		# Create a workload cluster with a particular plan and TKr
		tkg create cluster my-cluster --plan=dev --tkr=v1.17.3---vmware.2-tkr.1

		# Generate a cluster manifest identical to that used in the previous example
		tkg create cluster -d my-cluster --plan=dev --tkr=v1.17.3---vmware.2-tkr.1 > create-my-cluster.yaml

		# Create a workload cluster using a particular plan with a custom number of control plane and worker nodes
		tkg create cluster my-cluster -p prod -c 3 -w 5
		tkg create cluster my-cluster --plan=prod --controlplane-machine-count=3 --worker-machine-count=5

		# Create a vSphere workload cluster using a particular plan using 'medium' instance size
		# for control plane nodes, and 'large' size for worker nodes.
		tkg create cluster my-cluster --size medium --worker-size large -p prod

		[+] : instance size options available for vSphere are as follows:
		vSphere: [extra-large,large,medium,small]`),

	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		displayLogFileLocation()
		err = runCreateCluster(args[0])
		verifyCommandError(err)
	},
}

func init() {
	createClusterCmd.Flags().StringVarP(&cc.clusterConfigFile, "file", "", "", "The cluster configuration file (default \"$HOME/.tkg/cluster-config.yaml\")")

	createClusterCmd.Flags().StringVarP(&cc.plan, "plan", "p", "", "The plan to be used for creating the workload cluster")
	createClusterCmd.Flags().StringVarP(&cc.tkrName, "tkr", "", "", "TanzuKubernetesRelease(TKr) to be used for creating the workload cluster")
	createClusterCmd.Flags().IntVarP(&cc.controlPlaneMachineCount, "controlplane-machine-count", "c", 0, "The number of control plane machines to be added to the workload cluster (default 1 or 3 depending on dev or prod plan)")
	createClusterCmd.Flags().IntVarP(&cc.workerMachineCount, "worker-machine-count", "w", 0, "The number of worker machines to be added to the workload cluster (default 1 or 3 depending on dev or prod plan)")
	createClusterCmd.Flags().BoolVarP(&cc.generateOnly, "dry-run", "d", false, "Does not create cluster but show the deployment YAML instead")
	createClusterCmd.Flags().StringVarP(&cc.namespace, "namespace", "n", "", "The namespace where the cluster should be deployed. Assumes 'default' if not specified")
	createClusterCmd.Flags().DurationVarP(&cc.timeout, "timeout", "t", constants.DefaultLongRunningOperationTimeout, "Time duration to wait for an operation before timeout. Timeout duration in hours(h)/minutes(m)/seconds(s) units or as some combination of them (e.g. 2h, 30m, 2h30m10s)")
	createClusterCmd.Flags().StringVarP(&cc.cniType, "cni", "", "", "Specify the CNI provider the workload cluster should use ['antrea' (default), 'calico', 'none'].")

	createClusterCmd.Flags().StringVarP(&cc.size, "size", "", "", "Specify size for all nodes including control plane and worker nodes. It can be overridden by --controlplane-size and --worker-size options. (See [+])")
	createClusterCmd.Flags().StringVarP(&cc.controlPlaneSize, "controlplane-size", "", "", "Specify size for the control plane node. (See [+])")
	createClusterCmd.Flags().StringVarP(&cc.workerSize, "worker-size", "", "", "Specify size of the worker node. (See [+])")
	createClusterCmd.Flags().BoolVarP(&cc.unattended, "yes", "y", false, "Create workload cluster without asking for confirmation")
	createClusterCmd.Flags().StringVarP(&cc.enableClusterOptions, "enable-cluster-options", "", "", "List of comma separated cluster options to be enabled")

	createClusterCmd.Flags().StringVarP(&cc.vsphereControlPlaneEndpoint, "vsphere-controlplane-endpoint", "", "", "Virtual IP address or FQDN for the cluster's control plane nodes")
	// Usually not needed as they are implied from configuration of the management cluster.
	createClusterCmd.Flags().StringVarP(&cc.infrastructureProvider, "infrastructure", "i", "", "The target infrastructure on which to deploy the workload cluster.")
	createClusterCmd.Flags().MarkHidden("infrastructure") //nolint

	createClusterCmd.Flags().SetNormalizeFunc(aliasNormalizeFunc)
	createCmd.AddCommand(createClusterCmd)
}

func runCreateCluster(name string) error {
	tkgClient, err := newTKGCtlClient()
	if err != nil {
		return err
	}

	tkrVersion := ""
	if cc.tkrName != "" {
		tkrVersion = utils.GetTKRVersionFromTKRName(cc.tkrName)
	}

	return tkgClient.CreateCluster(buildCreateClusterOption(name, tkrVersion, cc))
}
