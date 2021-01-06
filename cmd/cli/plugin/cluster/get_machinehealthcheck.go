package main

import (
	"encoding/json"
	"fmt"
	"github.com/vmware-tanzu-private/core/apis/client/v1alpha1"
	"github.com/vmware-tanzu-private/core/pkg/v1/client"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/vmware-tanzu-private/tkg-cli/pkg/tkgctl"
)

type getMachineHealthCheckOptions struct {
	machinehealthCheckName string
	namespace              string
}

var getMHC = &getMachineHealthCheckOptions{}

var getMachineHealthCheckCmd = &cobra.Command{
	Use:     "get CLUSTER_NAME",
	Short:   "Get MachineHealthCheck object",
	Long:    "Get a MachineHealthCheck object for the given cluster",
	Args:    cobra.ExactArgs(1),
	RunE: getMachineHealthCheck,
}

func init() {
	getMachineHealthCheckCmd.Flags().StringVarP(&getMHC.machinehealthCheckName, "mhc-name", "m", "", "Name of the MachineHealthCheck object")
	getMachineHealthCheckCmd.Flags().StringVarP(&getMHC.namespace, "namespace", "n", "", "The namespace where the MachineHealthCheck object was created.")
	machineHealthCheckCmd.AddCommand(getMachineHealthCheckCmd)
}

func getMachineHealthCheck(cmd *cobra.Command, args []string) error {
	server, err := client.GetCurrentServer()
	if err != nil {
		return err
	}

	if server.IsGlobal() {
		return errors.New("getting machine healthcheck with a global server is not implemented yet")
	}
	return runGetMachineHealthCheck(server, args[0])

}

func runGetMachineHealthCheck(server *v1alpha1.Server, clusterName string) error {
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
	options := tkgctl.GetMachineHealthCheckOptions{
		ClusterName:            clusterName,
		Namespace:              getMHC.namespace,
		MachineHealthCheckName: getMHC.machinehealthCheckName,
	}

	mhcList, err := tkgctlClient.GetMachineHealthCheck(options)
	if err != nil {
		return err
	}

	bytes, err := json.MarshalIndent(mhcList, "", "    ")
	if err != nil {
		return errors.Wrap(err, "error marshaling the list of MachineHealthCheck objects")
	}

	fmt.Println(string(bytes))

	return nil
}
