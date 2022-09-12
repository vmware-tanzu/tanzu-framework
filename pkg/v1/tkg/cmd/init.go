// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgctl"
)

type initRegionOptions struct {
	clusterConfigFile           string
	plan                        string
	clusterName                 string
	coreProvider                string
	bootstrapProvider           string
	infrastructureProvider      string
	controlPlaneProvider        string
	targetNamespace             string
	watchingNamespace           string
	size                        string
	controlPlaneSize            string
	workerSize                  string
	ceipOptIn                   string
	cniType                     string
	bind                        string
	browser                     string
	vsphereControlPlaneEndpoint string
	featureFlags                map[string]string
	timeout                     time.Duration
	unattended                  bool
	ui                          bool
	useExistingCluster          bool
	enableTKGSOnVsphere7        bool
	deployTKGonVsphere7         bool
}

// Warningvsphere7WithoutPacific ...
var Warningvsphere7WithoutPacific = `
vSphere 7.0 Environment Detected.

You have connected to a vSphere 7.0 environment which does not have vSphere with Tanzu enabled. vSphere with Tanzu includes
an integrated Tanzu Kubernetes Grid Service which turns a vSphere cluster into a platform for running Kubernetes workloads in dedicated
resource pools. Configuring Tanzu Kubernetes Grid Service is done through vSphere HTML5 client.

Tanzu Kubernetes Grid Service is the preferred way to consume Tanzu Kubernetes Grid in vSphere 7.0 environments. Alternatively you may
deploy a non-integrated Tanzu Kubernetes Grid instance on vSphere 7.0.`

// Warningvsphere7WithPacific ...
var Warningvsphere7WithPacific = `
vSphere 7.0 with Tanzu Detected.

You have connected to a vSphere 7.0 with Tanzu environment that includes an integrated Tanzu Kubernetes Grid Service which
turns a vSphere cluster into a platform for running Kubernetes workloads in dedicated resource pools. Configuring Tanzu
Kubernetes Grid Service is done through the vSphere HTML5 Client.

Tanzu Kubernetes Grid Service is the preferred way to consume Tanzu Kubernetes Grid in vSphere 7.0 environments. Alternatively you may
deploy a non-integrated Tanzu Kubernetes Grid instance on vSphere 7.0.`

var iro = &initRegionOptions{}

// InitCmd defines init command
var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a Tanzu Kubernetes Grid management cluster",
	Long: LongDesc(`
			Create a Tanzu Kubernetes Grid management cluster including initializing it with Cluster API components
			appropriate for the target infrastructure.
		`),

	Example: Examples(`
		# Create a management cluster on AWS infrastructure, initializing it with
		# components required to create workload clusters through it on the same infrastructure
		# by bootstrapping through a self-provisioned bootstrap cluster.
		tkg init --infrastructure=aws

		# Create a management cluster, but on vSphere infrastructure instead,
		# using cluster plan 'prod'.
		tkg init --infrastructure=vsphere --plan prod

		# Launch an interactive UI to configure the settings necessary to create a
		# management cluster on vSphere infrastructure.
		tkg init --infrastructure=vsphere --ui

		# Create a management cluster on AWS infrastructure by using an existing
		# bootstrapper cluster. The current kube context should point to that
		# of the existing bootstrap cluster.
		tkg init --use-existing-bootstrap-cluster --infrastructure=aws

		# Create a management cluster on AWS using a particular plan using 'i3.large' instance size
		# for control plane nodes, and 'm5.large' size for worker nodes.
		tkg init --infrastructure=aws --controlplane-size i3.large --worker-size m5.large -p prod

		Note: The current cluster pointed to by the kubeconfig file will only
		be used to bootstrap the creation of the management cluster if
		--use-existing-bootstrap-cluster argument is supplied

		[+] : instance size options available for vSphere are as follows:
		vSphere: [extra-large,large,medium,small]


		[*] : VMware's Customer Experience Improvement Program ("CEIP") provides VMware with information that enables
		VMware to improve its products and services and fix problems. By choosing to participate in CEIP, you agree that
		VMware may collect technical information about your use of VMware products and services on a regular basis. This
		information does not personally identify you. Set this flag to false if you choose not to participate in the program.
		If this flag is not set, then the value from the cluster-config is used to determine the CEIP participation status.
  		CEIP participation is true by default.`),

	Run: func(cmd *cobra.Command, args []string) {
		log.UnsetStdoutStderr()
		displayLogFileLocation()
		err := runInit()
		verifyCommandError(err)
	},
}

func init() {
	InitCmd.Flags().StringVarP(&iro.clusterConfigFile, "file", "", "", "The cluster configuration file (default \"$HOME/.tkg/cluster-config.yaml\")")

	InitCmd.Flags().StringVarP(&iro.infrastructureProvider, "infrastructure", "i", "", "Infrastructure to deploy the management cluster on ['aws', 'vsphere', 'azure']")
	InitCmd.Flags().BoolVarP(&iro.ui, "ui", "u", false, "Launch interactive management cluster provisioning UI")
	InitCmd.Flags().StringVarP(&iro.plan, "plan", "p", constants.PlanDev, "Cluster plan to use to deploy the management cluster")

	InitCmd.Flags().BoolVarP(&iro.useExistingCluster, "use-existing-bootstrap-cluster", "e", false, "Use an existing bootstrap cluster to deploy the management cluster")
	InitCmd.Flags().StringVarP(&iro.clusterName, "name", "", "", "Name of the management cluster. One will be generated if not provided")
	InitCmd.Flags().DurationVarP(&iro.timeout, "timeout", "t", constants.DefaultLongRunningOperationTimeout, "Time duration to wait for an operation before timeout. Timeout duration in hours(h)/minutes(m)/seconds(s) units or as some combination of them (e.g. 2h, 30m, 2h30m10s)")

	InitCmd.Flags().StringVarP(&iro.size, "size", "", "", "Specify size for all nodes including control plane and worker nodes. It can be overridden by --controlplane-size and --worker-size options. (See [+])")
	InitCmd.Flags().StringVarP(&iro.controlPlaneSize, "controlplane-size", "", "", "Specify size for the control plane node. (See [+])")
	InitCmd.Flags().StringVarP(&iro.workerSize, "worker-size", "", "", "Specify size of the worker node. (See [+])")
	InitCmd.Flags().StringVarP(&iro.ceipOptIn, "ceip-participation", "", "", "Specify if this management cluster should participate in VMware CEIP. (See [*])")
	InitCmd.Flags().BoolVarP(&iro.deployTKGonVsphere7, "deploy-tkg-on-vSphere7", "", false, "Deploy TKG Management cluster on vSphere 7.0 without prompt")
	InitCmd.Flags().BoolVarP(&iro.enableTKGSOnVsphere7, "enable-tkgs-on-vSphere7", "", false, "Enable TKGS on vSphere 7.0 without prompt")

	InitCmd.Flags().StringVarP(&iro.bind, "bind", "b", "127.0.0.1:8080", "Specify the IP and port to bind the Kickstart UI against (e.g. 127.0.0.1:8080).")
	InitCmd.Flags().StringVarP(&iro.browser, "browser", "", "", "Specify the browser to open the Kickstart UI on. Use 'none' for no browser. Defaults to OS default browser. Supported: ['chrome', 'firefox', 'safari', 'ie', 'edge', 'none']")

	InitCmd.Flags().StringVarP(&iro.vsphereControlPlaneEndpoint, "vsphere-controlplane-endpoint", "", "", "Virtual IP address or FQDN for the cluster's control plane nodes")

	InitCmd.Flags().BoolVarP(&iro.unattended, "yes", "y", false, "Create management cluster without asking for confirmation")

	// Hidden flags, mostly for development and testing

	InitCmd.Flags().StringVarP(&iro.targetNamespace, "target-namespace", "", "", "The target namespace where the providers should be deployed. If not specified, each provider will be installed in a provider's default namespace")
	InitCmd.Flags().MarkHidden("target-namespace") //nolint

	InitCmd.Flags().StringVarP(&iro.cniType, "cni", "", "", "Specify the CNI provider the management cluster should use ['antrea' (default), 'calico', 'none'].")
	InitCmd.Flags().MarkHidden("cni") //nolint

	InitCmd.Flags().StringToStringVarP(&iro.featureFlags, "feature-flags", "", nil, "Activate and deactivate hidden features in the form 'feature1=true,feature2=false'")
	InitCmd.Flags().MarkHidden("feature-flags") //nolint

	InitCmd.Flags().SetNormalizeFunc(aliasNormalizeFunc)
	RootCmd.AddCommand(InitCmd)
}

func aliasNormalizeFunc(f *pflag.FlagSet, name string) pflag.NormalizedName {
	if name == "vsphere-controlplane-endpoint-ip" {
		name = "vsphere-controlplane-endpoint"
	}
	return pflag.NormalizedName(name)
}

func runInit() error {
	tkgClient, err := newTKGCtlClient()
	if err != nil {
		return err
	}

	options := tkgctl.InitRegionOptions{
		ClusterConfigFile:           iro.clusterConfigFile,
		Plan:                        iro.plan,
		UI:                          iro.ui,
		ClusterName:                 iro.clusterName,
		UseExistingCluster:          iro.useExistingCluster,
		CoreProvider:                iro.coreProvider,
		BootstrapProvider:           iro.bootstrapProvider,
		InfrastructureProvider:      iro.infrastructureProvider,
		ControlPlaneProvider:        iro.controlPlaneProvider,
		Namespace:                   iro.targetNamespace,
		WatchingNamespace:           iro.watchingNamespace,
		Size:                        iro.size,
		ControlPlaneSize:            iro.controlPlaneSize,
		WorkerSize:                  iro.workerSize,
		CeipOptIn:                   iro.ceipOptIn,
		CniType:                     iro.cniType,
		FeatureFlags:                iro.featureFlags,
		EnableTKGSOnVsphere7:        iro.enableTKGSOnVsphere7,
		DeployTKGonVsphere7:         iro.deployTKGonVsphere7,
		Bind:                        iro.bind,
		Browser:                     iro.browser,
		VsphereControlPlaneEndpoint: iro.vsphereControlPlaneEndpoint,
		SkipPrompt:                  iro.unattended || skipPrompt,
		Timeout:                     iro.timeout,
	}

	return tkgClient.Init(options)
}
