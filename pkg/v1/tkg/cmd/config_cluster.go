// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/utils"
)

var configClusterCmd = &cobra.Command{
	Use:   "cluster CLUSTER_NAME",
	Short: "Generate a cluster plan for creating a workload clusters",
	Long: LongDesc(`
	Generates a cluster plan representing the desired state of a Tanzu Kubernetes cluster that can then be applied to a management cluster.`),

	Example: Examples(`
		# Generates a yaml file for creating a development cluster with TKr v1.17.3---vmware.2-tkr.1
		tkg config cluster my-cluster --plan=dev --tkr=v1.17.3---vmware.2-tkr.1

		# Generates a yaml file for creating a production cluster with custom number of nodes
		tkg config cluster my-cluster --plan=prod --controlplane-machine-count=3 --worker-machine-count=10


		# Generates a yaml file for creating a production cluster on vSphere using 'medium' instance size
		# for control plane nodes, and 'large' size for worker nodes.
		tkg config cluster my-cluster --size medium --worker-size large -p prod

		[+] : instance size options available for vSphere are as follows:
		vSphere: [extra-large,large,medium,small]`),

	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := runGenerateClusterConfig(args[0], configclusteroption)
		verifyCommandError(err)
	},
}

var configclusteroption = &createClusterOptions{}

func init() {
	// --------------------------------------------------------------TODO-----V
	configClusterCmd.Flags().StringVarP(&configclusteroption.plan, "plan", "p", "", "The cluster plan to be used for creating the workload cluster")
	if err := configClusterCmd.MarkFlagRequired("plan"); err != nil {
		log.Fatal(err, "")
	}
	configClusterCmd.Flags().StringVarP(&configclusteroption.clusterConfigFile, "file", "", "", "The cluster configuration file (default \"$HOME/.tkg/cluster-config.yaml\")")

	configClusterCmd.Flags().StringVarP(&configclusteroption.tkrName, "tkr", "", "", "TanzuKubernetesRelease(TKr) to be used for creating the workload cluster")
	configClusterCmd.Flags().IntVarP(&configclusteroption.controlPlaneMachineCount, "controlplane-machine-count", "c", 0, "The number of control plane machines to be added to the workload cluster (default 1 or 3 depending on dev or prod plan)")
	configClusterCmd.Flags().IntVarP(&configclusteroption.workerMachineCount, "worker-machine-count", "w", 0, "The number of worker machines to be added to the workload cluster (default 1 or 3 depending on dev or prod plan)")
	configClusterCmd.Flags().StringVarP(&configclusteroption.namespace, "namespace", "n", "", "The namespace where the cluster should be deployed. Assumes 'default' if not specified")
	configClusterCmd.Flags().StringVarP(&configclusteroption.size, "size", "", "", "Specify size for all nodes including control plane and worker nodes. It can be overridden by --controlplane-size and --worker-size options. (See [+])")
	configClusterCmd.Flags().StringVarP(&configclusteroption.controlPlaneSize, "controlplane-size", "", "", "Specify size for the control plane node. (See [+])")
	configClusterCmd.Flags().StringVarP(&configclusteroption.workerSize, "worker-size", "", "", "Specify size of the worker node. (See [+])")
	configClusterCmd.Flags().StringVarP(&configclusteroption.cniType, "cni", "", "", "Specify the CNI provider the cluster should use ['antrea' (default), 'calico', 'none'].")
	configClusterCmd.Flags().StringVarP(&configclusteroption.enableClusterOptions, "enable-cluster-options", "", "", "List of comma separated cluster options to be enabled")
	configClusterCmd.Flags().StringVarP(&configclusteroption.vsphereControlPlaneEndpoint, "vsphere-controlplane-endpoint", "", "", "Virtual IP address or FQDN for the cluster's control plane nodes")
	// Usually not needed as they are implied from configuration of the management cluster.
	configClusterCmd.Flags().StringVarP(&configclusteroption.infrastructureProvider, "infrastructure", "i", "", "The target infrastructure on which to deploy the workload cluster.")
	configClusterCmd.Flags().MarkHidden("infrastructure") //nolint

	configClusterCmd.Flags().SetNormalizeFunc(aliasNormalizeFunc)
	configCmd.AddCommand(configClusterCmd)
}

func runGenerateClusterConfig(name string, options *createClusterOptions) error {
	tkgClient, err := newTKGCtlClient()
	if err != nil {
		return err
	}

	tkrVersion := ""
	if options.tkrName != "" {
		tkrVersion = utils.GetTKRVersionFromTKRName(options.tkrName)
	}

	return tkgClient.ConfigCluster(buildCreateClusterOption(name, tkrVersion, options))
}

func buildCreateClusterOption(name, tkrVersion string, options *createClusterOptions) tkgctl.CreateClusterOptions {
	edition, _ := config.GetEdition()
	return tkgctl.CreateClusterOptions{
		ClusterConfigFile:           options.clusterConfigFile,
		ClusterName:                 name,
		Namespace:                   options.namespace,
		Plan:                        options.plan,
		InfrastructureProvider:      options.infrastructureProvider,
		TkrVersion:                  tkrVersion,
		ControlPlaneMachineCount:    options.controlPlaneMachineCount,
		WorkerMachineCount:          options.workerMachineCount,
		GenerateOnly:                options.generateOnly,
		Size:                        options.size,
		ControlPlaneSize:            options.controlPlaneSize,
		WorkerSize:                  options.workerSize,
		CniType:                     options.cniType,
		EnableClusterOptions:        options.enableClusterOptions,
		VsphereControlPlaneEndpoint: options.vsphereControlPlaneEndpoint,
		SkipPrompt:                  options.unattended || skipPrompt,
		Timeout:                     options.timeout,
		Edition:                     edition,
	}
}
