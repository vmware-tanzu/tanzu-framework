package main

import (
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
	"github.com/vmware-tanzu-private/core/apis/client/v1alpha1"
	"github.com/vmware-tanzu-private/core/pkg/v1/client"
	"github.com/vmware-tanzu-private/tkg-cli/pkg/tkgctl"
	tkgutils "github.com/vmware-tanzu-private/tkg-cli/pkg/utils"
)

type getClusterCredentialFunc func(tkgctl.GetWorkloadClusterCredentialsOptions) error
type getClusterAdminKubeconfigCmdDeps struct {
	getClusterCredentials func(tkgctl.TKGClient) getClusterCredentialFunc
	getCurrentServer      func() (*v1alpha1.Server, error)
}

func getClusterAdminKubeconfigCmdRealDeps() getClusterAdminKubeconfigCmdDeps {
	return getClusterAdminKubeconfigCmdDeps{
		getClusterCredentials: func(tkgctlClient tkgctl.TKGClient) getClusterCredentialFunc {
			return tkgctlClient.GetCredentials
		},
		getCurrentServer: client.GetCurrentServer,
	}
}

type getClusterAdminKubeconfigOptions struct {
	workloadClusterName string
	namespace           string
	exportFile          string
}

var gAKCOptions = &getClusterAdminKubeconfigOptions{}

func init() {
	clusterAdminKubeconfigCmd.AddCommand(getClusterAdminKubeconfigCmd(getClusterAdminKubeconfigCmdRealDeps()))
}

func getClusterAdminKubeconfigCmd(deps getClusterAdminKubeconfigCmdDeps) *cobra.Command {
	getAKCCmd := &cobra.Command{
		Use:          "get",
		Short:        "Get admin-kubeconfig of a cluster",
		Long:         `Get admin-kubeconfig of a cluster and merge the context into the default kubeconfig file`,
		SilenceUsage: true,
		Example: `
		# Get management cluster admin-kubeconfig
		tanzu cluster admin-kubeconfig get
		
		# Get workload cluster admin-kubeconfig
		tanzu cluster admin-kubeconfig get -w cluster1`,
	}
	getAKCCmd.Flags().StringVarP(&gAKCOptions.workloadClusterName, "workload-clustername", "w", "", "The name of the workload cluster. Assumes management cluster if not specified.")
	getAKCCmd.Flags().StringVarP(&gAKCOptions.namespace, "namespace", "n", "", "The namespace where the workload cluster was created. Assumes 'default' if not specified.")
	getAKCCmd.Flags().StringVarP(&gAKCOptions.exportFile, "export-file", "", "", "File path to export a standalone kubeconfig for workload cluster")

	getAKCCmd.RunE = func(cmd *cobra.Command, args []string) error {
		server, err := deps.getCurrentServer()
		if err != nil {
			return err
		}

		if server.IsGlobal() {
			return errors.New("get cluster admin-kubeconfig with a global server is not implemented yet")
		}
		return getClusterAdminKubeconfig(server, deps)
	}
	return getAKCCmd
}

func getClusterAdminKubeconfig(server *v1alpha1.Server, deps getClusterAdminKubeconfigCmdDeps) error {
	tkgctlClient, err := createTKGClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context)
	if err != nil {
		return err
	}

	isManagementCluster := false
	if gAKCOptions.workloadClusterName == "" {
		isManagementCluster = true
	}

	if isManagementCluster {
		mcClustername, err := tkgutils.GetClusterNameFromKubeconfigAndContext(server.ManagementClusterOpts.Path,
			server.ManagementClusterOpts.Context)
		if err != nil {
			return errors.Wrap(err, "failed to get management cluster name from kubeconfig")
		}
		gAKCOptions.workloadClusterName = mcClustername
		gAKCOptions.namespace = TKGSystemNamespace

	}
	if !isManagementCluster && gAKCOptions.namespace == "" {
		gAKCOptions.namespace = DefaultNamespace
	}

	getClusterCredentialsOptions := tkgctl.GetWorkloadClusterCredentialsOptions{
		ClusterName: gAKCOptions.workloadClusterName,
		Namespace:   gAKCOptions.namespace,
		ExportFile:  gAKCOptions.exportFile,
	}
	getCredential := deps.getClusterCredentials(tkgctlClient)
	err = getCredential(getClusterCredentialsOptions)
	if err != nil {
		return err
	}

	return nil
}
