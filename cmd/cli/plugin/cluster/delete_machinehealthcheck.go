package main

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/vmware-tanzu-private/core/apis/client/v1alpha1"
	"github.com/vmware-tanzu-private/core/pkg/v1/client"
	"github.com/vmware-tanzu-private/tkg-cli/pkg/tkgctl"
)

type deleteMachineHealthCheckOptions struct {
	machinehealthCheckName string
	namespace              string
	unattended             bool
}

var deleteMHC = &deleteMachineHealthCheckOptions{}

var deleteMachineHealthCheckCmd = &cobra.Command{
	Use:   "delete CLUSTER_NAME",
	Short: "Delete a MachineHealthCheck object of a cluster",
	Long:  "Delete a MachineHealthCheck object of a cluster",
	Args:    cobra.ExactArgs(1),
	RunE: deleteMachineHealthCheck,
}

func init() {
	deleteMachineHealthCheckCmd.Flags().BoolVarP(&deleteMHC.unattended, "yes", "y", false, "Delete the MachineHealthCheck object without asking for confirmation")
	deleteMachineHealthCheckCmd.Flags().StringVarP(&deleteMHC.machinehealthCheckName, "mhc-name", "m", "", "Name of the MachineHealthCheck object")
	deleteMachineHealthCheckCmd.Flags().StringVarP(&deleteMHC.namespace, "namespace", "n", "", "The namespace where the MachineHealthCheck object was created, default to the cluster's namespace")
	machineHealthCheckCmd.AddCommand(deleteMachineHealthCheckCmd)
}

func deleteMachineHealthCheck(cmd *cobra.Command, args []string) error {
	server, err := client.GetCurrentServer()
	if err != nil {
		return err
	}

	if server.IsGlobal() {
		return errors.New("getting machine healthcheck with a global server is not implemented yet")
	}
	return runDeleteMachineHealthCheck(server, args[0])
}

func runDeleteMachineHealthCheck(server *v1alpha1.Server, clusterName string) error {
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

	options := tkgctl.DeleteMachineHealthCheckOptions{
		ClusterName:            clusterName,
		Namespace:              deleteMHC.namespace,
		MachinehealthCheckName: deleteMHC.machinehealthCheckName,
		SkipPrompt:             deleteMHC.unattended,
	}
	return tkgctlClient.DeleteMachineHealthCheck(options)
}