package main

import (
	"github.com/aunum/log"
	"github.com/pkg/errors"

	"encoding/json"

	"github.com/spf13/cobra"
	"github.com/vmware-tanzu-private/core/apis/client/v1alpha1"
	tkgauth "github.com/vmware-tanzu-private/core/pkg/v1/auth/tkg"
	"github.com/vmware-tanzu-private/core/pkg/v1/client"
	tkgclient "github.com/vmware-tanzu-private/tkg-cli/pkg/client"
	"github.com/vmware-tanzu-private/tkg-cli/pkg/tkgctl"
)

type getClusterKubeconfigOptions struct {
	workloadClusterName string
	namespace           string
	exportFile          string
}

const (
	// TKGSystemNamespace is the TKG system namespace.
	TKGSystemNamespace = "tkg-system"

	// DefaultNamespace is the default namespace.
	DefaultNamespace = "default"
)

var getKCOptions = &getClusterKubeconfigOptions{}

var getClusterKubeconfigCmd = &cobra.Command{
	Use:   "get",
	Short: "Get Kubeconfig of a cluster",
	Long:  `Get Kubeconfig of a cluster and merge the context into the default kubeconfig file`,
	Example: `
	# Get management cluster kubeconfig
	tanzu cluster kubeconfig get
	
	# Get workload cluster kubeconfig
	tanzu cluster kubeconfig get -w cluster1`,
	RunE: getKubeconfig,
}

func init() {
	getClusterKubeconfigCmd.Flags().StringVarP(&getKCOptions.workloadClusterName, "workload-clustername", "w", "", "The name of the workload cluster. Assumes management cluster if not specified.")
	getClusterKubeconfigCmd.Flags().StringVarP(&getKCOptions.namespace, "namespace", "n", "", "The namespace where the workload cluster was created. Assumes 'default' if not specified.")
	getClusterKubeconfigCmd.Flags().StringVarP(&getKCOptions.exportFile, "export-file", "", "", "File path to export a standalone kubeconfig for workload cluster")

	clusterKubeconfigCmd.AddCommand(getClusterKubeconfigCmd)
}

func getKubeconfig(cmd *cobra.Command, args []string) error {
	server, err := client.GetCurrentServer()
	if err != nil {
		return err
	}

	if server.IsGlobal() {
		return errors.New("get cluster kubeconfig with a global server is not implemented yet")
	}
	return getClusterKubeconfig(server)
}

func getClusterKubeconfig(server *v1alpha1.Server) error {
	tkgctlClient, err := createTKGClient(server.ManagementClusterOpts.Path, server.ManagementClusterOpts.Context)
	if err != nil {
		return err
	}

	isManagementCluster := false
	if getKCOptions.workloadClusterName == "" {
		isManagementCluster = true
	}

	if !isManagementCluster && getKCOptions.namespace == "" {
		getKCOptions.namespace = DefaultNamespace
	}

	getClusterPinnipedInfoOptions := tkgctl.GetClusterPinnipedInfoOptions{
		ClusterName:         getKCOptions.workloadClusterName,
		Namespace:           getKCOptions.namespace,
		IsManagementCluster: isManagementCluster,
	}

	clusterPinnipedInfo, err := tkgctlClient.GetClusterPinnipedInfo(getClusterPinnipedInfoOptions)
	if err != nil {
		return err
	}

	// for workload cluster the audience would be set to the clustername and for management cluster the audience would be set to IssuerURL
	audience := clusterPinnipedInfo.ClusterName
	if isManagementCluster {
		audience = clusterPinnipedInfo.PinnipedInfo.Data.Issuer
	}
	kubeconfig, err := tkgauth.GetPinnipedKubeconfig(clusterPinnipedInfo.ClusterInfo, clusterPinnipedInfo.PinnipedInfo,
		clusterPinnipedInfo.ClusterName, audience)
	if err != nil {
		return err
	}

	kubeconfigbytes, err := json.Marshal(kubeconfig)
	if err != nil {
		return errors.Wrap(err, "unable to marshall the kubeconfig")
	}
	err = tkgclient.MergeKubeConfigWithoutSwitchContext(kubeconfigbytes, getKCOptions.exportFile)
	if err != nil {
		return errors.Wrap(err, "unable to merge cluster kubeconfig into the current kubeconfig path")
	}

	if getKCOptions.exportFile != "" {
		log.Infof("You can now access the cluster by running 'kubectl config use-context %s' under path '%s' \n", kubeconfig.CurrentContext, getKCOptions.exportFile)
	} else {
		log.Infof("You can now access the cluster by running 'kubectl config use-context %s'\n", kubeconfig.CurrentContext)
	}

	return nil
}
