// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	runv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
	tkr "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/controllers/source"
	tkrutils "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/utils"
	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgctl"

	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
)

// Note: We can remove all this additional options at the time when
// we decide we really want to use `--file` option to create cluster
// and we don't have to support all these flags with create cluster command
type createClusterOptions struct {
	plan                        string
	infrastructureProvider      string
	namespace                   string
	size                        string
	controlPlaneSize            string
	workerSize                  string
	cniType                     string
	enableClusterOptions        string
	vsphereControlPlaneEndpoint string
	clusterConfigFile           string
	tkrName                     string
	controlPlaneMachineCount    int
	workerMachineCount          int
	timeout                     time.Duration
	generateOnly                bool
	unattended                  bool
}

var cc = &createClusterOptions{}

var createClusterCmd = &cobra.Command{
	Use:          "create CLUSTER_NAME",
	Short:        "Create a cluster",
	RunE:         create,
	SilenceUsage: true,
}

func init() {
	createClusterCmd.Flags().StringVarP(&cc.clusterConfigFile, "file", "f", "", "Configuration file or Cluster objects from which to create a cluster")
	createClusterCmd.Flags().StringVarP(&cc.tkrName, "tkr", "", "", "TanzuKubernetesRelease(TKr) to be used for creating the workload cluster. If TKr name prefix is provided, the latest compatible TKr matching the TKr name prefix would be used")

	createClusterCmd.Flags().StringVarP(&cc.plan, "plan", "p", "", "The plan to be used for creating the workload cluster")
	createClusterCmd.Flags().IntVarP(&cc.controlPlaneMachineCount, "controlplane-machine-count", "c", 0, "The number of control plane machines to be added to the workload cluster (default 1 or 3 depending on dev or prod plan)")
	createClusterCmd.Flags().IntVarP(&cc.workerMachineCount, "worker-machine-count", "w", 0, "The number of worker machines to be added to the workload cluster (default 1 or 3 depending on dev or prod plan)")
	createClusterCmd.Flags().BoolVarP(&cc.generateOnly, "dry-run", "d", false, "Does not create cluster, but show the deployment YAML instead")
	createClusterCmd.Flags().StringVarP(&cc.namespace, "namespace", "n", "", "The namespace where the cluster should be deployed. Assumes 'default' if not specified")
	createClusterCmd.Flags().StringVarP(&cc.vsphereControlPlaneEndpoint, "vsphere-controlplane-endpoint", "", "", "Virtual IP address or FQDN for the cluster's control plane nodes")
	createClusterCmd.Flags().DurationVarP(&cc.timeout, "timeout", "t", constants.DefaultLongRunningOperationTimeout, "Time duration to wait for an operation before timeout. Timeout duration in hours(h)/minutes(m)/seconds(s) units or as some combination of them (e.g. 2h, 30m, 2h30m10s)")
	createClusterCmd.Flags().StringVarP(&cc.cniType, "cni", "", "", "Specify the CNI provider the workload cluster should use ['antrea' (default), 'calico', 'none'].")
	createClusterCmd.Flags().StringVarP(&cc.size, "size", "", "", "Specify size for all nodes including control plane and worker nodes. It can be overridden by --controlplane-size and --worker-size options. (See [+])")
	createClusterCmd.Flags().StringVarP(&cc.controlPlaneSize, "controlplane-size", "", "", "Specify size for the control plane node. (See [+])")
	createClusterCmd.Flags().StringVarP(&cc.workerSize, "worker-size", "", "", "Specify size of the worker node. (See [+])")
	createClusterCmd.Flags().BoolVarP(&cc.unattended, "yes", "y", false, "Create workload cluster without asking for confirmation")
	createClusterCmd.Flags().StringVarP(&cc.enableClusterOptions, "enable-cluster-options", "", "", "List of comma separated cluster options to be enabled")
	createClusterCmd.Flags().StringVarP(&cc.infrastructureProvider, "infrastructure", "i", "", "The target infrastructure on which to deploy the workload cluster.")

	// Hide some of the variables not relevant to tanzu cli at the moment
	createClusterCmd.Flags().MarkHidden("plan")                          //nolint
	createClusterCmd.Flags().MarkHidden("controlplane-machine-count")    //nolint
	createClusterCmd.Flags().MarkHidden("worker-machine-count")          //nolint
	createClusterCmd.Flags().MarkHidden("namespace")                     //nolint
	createClusterCmd.Flags().MarkHidden("vsphere-controlplane-endpoint") //nolint
	createClusterCmd.Flags().MarkHidden("timeout")                       //nolint
	createClusterCmd.Flags().MarkHidden("cni")                           //nolint
	createClusterCmd.Flags().MarkHidden("size")                          //nolint
	createClusterCmd.Flags().MarkHidden("controlplane-size")             //nolint
	createClusterCmd.Flags().MarkHidden("worker-size")                   //nolint
	createClusterCmd.Flags().MarkHidden("yes")                           //nolint
	createClusterCmd.Flags().MarkHidden("enable-cluster-options")        //nolint
	createClusterCmd.Flags().MarkHidden("infrastructure")                //nolint // Usually not needed as they are implied from configuration of the management cluster.

	createClusterCmd.Flags().SetNormalizeFunc(aliasNormalizeFunc)
}

func aliasNormalizeFunc(f *pflag.FlagSet, name string) pflag.NormalizedName {
	if name == "vsphere-controlplane-endpoint-ip" {
		name = "vsphere-controlplane-endpoint"
	}
	return pflag.NormalizedName(name)
}

func create(cmd *cobra.Command, args []string) error {
	server, err := config.GetCurrentServer()
	if err != nil {
		// if current server does not exist and user is using generate only
		// option then allow user to proceed by providing dummy management server
		// information.
		// Note: This is only used for testing purpose when management cluster
		// does not exist and we want to test cluster template generation
		if cc.generateOnly {
			server = &configapi.Server{
				Type:                  configapi.ManagementClusterServerType,
				ManagementClusterOpts: &configapi.ManagementClusterServer{},
			}
		} else {
			return err
		}
	}
	clusterName := ""
	if len(args) > 0 {
		clusterName = args[0]
	}

	if server.IsGlobal() {
		return errors.New("creating cluster with a global server is not implemented yet")
	}
	return createCluster(clusterName, server)
}

func createCluster(clusterName string, server *configapi.Server) error {
	tkgctlClient, err := createTKGClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context)
	if err != nil {
		return err
	}

	tkrVersion := ""
	if cc.tkrName != "" {
		clusterClientOptions := clusterclient.Options{GetClientInterval: 2 * time.Second, GetClientTimeout: 5 * time.Second}
		clusterClient, err := clusterclient.NewClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context, clusterClientOptions)
		if err != nil {
			return err
		}

		tkrVersion, err = getTkrVersionForMatchingTkr(clusterClient, cc.tkrName)
		if err != nil {
			return err
		}
	}

	edition, err := config.GetEdition()
	if err != nil {
		return err
	}

	ccOptions := tkgctl.CreateClusterOptions{
		ClusterConfigFile:           cc.clusterConfigFile,
		TkrVersion:                  tkrVersion,
		ClusterName:                 clusterName,
		Namespace:                   cc.namespace,
		Plan:                        cc.plan,
		InfrastructureProvider:      cc.infrastructureProvider,
		ControlPlaneMachineCount:    cc.controlPlaneMachineCount,
		WorkerMachineCount:          cc.workerMachineCount,
		GenerateOnly:                cc.generateOnly,
		Size:                        cc.size,
		ControlPlaneSize:            cc.controlPlaneSize,
		WorkerSize:                  cc.workerSize,
		CniType:                     cc.cniType,
		EnableClusterOptions:        cc.enableClusterOptions,
		VsphereControlPlaneEndpoint: cc.vsphereControlPlaneEndpoint,
		SkipPrompt:                  cc.unattended,
		Timeout:                     cc.timeout,
		Edition:                     edition,
	}

	return tkgctlClient.CreateCluster(ccOptions)
}

func getTkrVersionForMatchingTkr(clusterClient clusterclient.Client, tkrName string) (string, error) {
	// get all the TKRs with tkrName prefix matching
	tkrs, err := clusterClient.GetTanzuKubernetesReleases(tkrName)
	if err != nil {
		return "", err
	}

	if len(tkrs) == 0 {
		return "", errors.Errorf("could not find a matching Tanzu Kubernetes release for name %q", tkrName)
	}

	tkrForCreate, err := getMatchingTkrForTkrName(tkrs, tkrName)
	// If the complete TKR name is provided, use it
	if err == nil {
		if !tkrutils.IsTkrActive(&tkrForCreate) {
			return "", errors.Errorf("the Tanzu Kubernetes release %q is deactivated and cannot be used", tkrName)
		}
		if !tkrutils.IsTkrCompatible(&tkrForCreate) {
			fmt.Printf("WARNING: Tanzu Kubernetes release %q is not compatible on the management cluster\n", tkrForCreate.Name)
		}
		if tkrForCreate.Spec.Version == "" {
			return "", errors.Errorf("could not find a matching Tanzu Kubernetes release for name %q", tkrName)
		}
		return tkrForCreate.Spec.Version, nil
	}

	return getLatestTKRVersionMatchingTKRPrefix(tkrName, tkrs)
}

// getLatestTKRVersionMatchingTKRPrefix returns the latest compatible TKR from the prefix name matched TKRs
func getLatestTKRVersionMatchingTKRPrefix(tkrName string, tkrsWithPrefixMatch []runv1alpha1.TanzuKubernetesRelease) (string, error) {
	compatibleTKRs := []runv1alpha1.TanzuKubernetesRelease{}
	for idx := range tkrsWithPrefixMatch {
		if !tkrutils.IsTkrCompatible(&tkrsWithPrefixMatch[idx]) || !tkrutils.IsTkrActive(&tkrsWithPrefixMatch[idx]) {
			continue
		}
		compatibleTKRs = append(compatibleTKRs, tkrsWithPrefixMatch[idx])
	}

	if len(compatibleTKRs) == 0 {
		return "", errors.Errorf("could not find a matching compatible Tanzu Kubernetes release for name %q", tkrName)
	}

	return getLatestTkrVersion(compatibleTKRs)
}

// getLatestTkrVersion returns the latest TKR version from the TKRs list
func getLatestTkrVersion(tkrs []runv1alpha1.TanzuKubernetesRelease) (string, error) {
	sort.SliceStable(tkrs, func(i, j int) bool {
		tkri, err := tkr.NewTKRVersion(tkrs[i].Name)
		if err != nil {
			return true
		}
		tkrj, err := tkr.NewTKRVersion(tkrs[j].Name)
		if err != nil {
			return true
		}
		if tkri.Major != tkrj.Major {
			return tkri.Major < tkrj.Major
		}
		if tkri.Minor != tkrj.Minor {
			return tkri.Minor < tkrj.Minor
		}
		if tkri.Patch != tkrj.Patch {
			return tkri.Patch < tkrj.Patch
		}
		if tkri.VMware != tkrj.VMware {
			return tkri.VMware < tkrj.VMware
		}
		return tkri.TKG < tkrj.TKG
	})

	latestTKRs := []runv1alpha1.TanzuKubernetesRelease{}
	latestTKRsNames := []string{}

	latestTKRs = append(latestTKRs, tkrs[len(tkrs)-1])
	latestTKRVersion, _ := tkr.NewTKRVersion(latestTKRs[0].Name)
	latestTKRsNames = append(latestTKRsNames, latestTKRs[0].Name)

	for i := len(tkrs) - 2; i >= 0; i-- {
		currentTKRVerison, _ := tkr.NewTKRVersion(tkrs[i].Name)
		if reflect.DeepEqual(latestTKRVersion, currentTKRVerison) {
			latestTKRs = append(latestTKRs, tkrs[i])
			latestTKRsNames = append(latestTKRsNames, tkrs[i].Name)
		}
	}

	if len(latestTKRs) > 1 {
		return "", errors.Errorf("found multiple TKrs %v matching the criteria, please specify the TKr name you want to use", latestTKRsNames)
	}

	log.V(4).Infof("Using the TKr version '%s' from TKr name '%s' ", latestTKRs[0].Spec.Version, latestTKRs[0].Name)
	return latestTKRs[0].Spec.Version, nil
}
