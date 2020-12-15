package main

import (
	"errors"

	"github.com/spf13/cobra"
	"github.com/vmware-tanzu-private/core/apis/client/v1alpha1"
	"github.com/vmware-tanzu-private/core/pkg/v1/client"
	"github.com/vmware-tanzu-private/tkg-cli/pkg/tkgctl"
)

type scaleClustersOptions struct {
	namespace         string
	workerCount       int32
	controlPlaneCount int32
}

var sc = &scaleClustersOptions{}

var scaleClusterCmd = &cobra.Command{
	Use:   "scale CLUSTER_NAME",
	Short: "Scale a cluster",
	Args:  cobra.ExactArgs(1),
	RunE:  scale,
}

func init() {
	scaleClusterCmd.Flags().Int32VarP(&sc.workerCount, "worker-machine-count", "w", 0, "The number of worker nodes to scale to. Assumes unchanged if not specified")
	scaleClusterCmd.Flags().Int32VarP(&sc.controlPlaneCount, "controlplane-machine-count", "c", 0, "The number of control plane nodes to scale to. Assumes unchanged if not specified")
	scaleClusterCmd.Flags().StringVarP(&sc.namespace, "namespace", "n", "", "The namespace where the workload cluster was created. Assumes 'default' if not specified.")
}

func scale(cmd *cobra.Command, args []string) error {
	server, err := client.GetCurrentServer()
	if err != nil {
		return err
	}

	if server.IsGlobal() {
		return errors.New("scaling cluster with a global server is not implemented yet")
	}
	return scaleCluster(server, args[0])
}

func scaleCluster(server *v1alpha1.Server, clusterName string) error {
	configDir, err := getConfigDir()
	if err != nil {
		return err
	}

	tkgctlClient, err := tkgctl.New(tkgctl.Options{
		ConfigDir:   configDir,
		KubeConfig:  server.ManagementClusterOpts.Path,
		KubeContext: server.ManagementClusterOpts.Context,
	})
	if err != nil {
		return err
	}

	scaleClusterOptions := tkgctl.ScaleClusterOptions{
		ClusterName:       clusterName,
		ControlPlaneCount: sc.controlPlaneCount,
		WorkerCount:       sc.workerCount,
		Namespace:         sc.namespace,
	}

	return tkgctlClient.ScaleCluster(scaleClusterOptions)
}
