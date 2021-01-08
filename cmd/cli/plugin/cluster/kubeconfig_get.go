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

type getClusterKubeconfigsOptions struct {
	workloadClusterName string
	namespace           string
	exportFile          string
}

var gkc = &getClusterKubeconfigsOptions{}

var getKubeconfigClusterCmd = &cobra.Command{
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
	getKubeconfigClusterCmd.Flags().StringVarP(&gkc.workloadClusterName, "workload-clustername", "w", "", "The name of the workload cluster. Assumes management cluster if not specified.")
	getKubeconfigClusterCmd.Flags().StringVarP(&gkc.namespace, "namespace", "n", "", "The namespace where the workload cluster was created. Assumes 'default' if not specified.")
	getKubeconfigClusterCmd.Flags().StringVarP(&gkc.exportFile, "export-file", "", "", "File path to export a standalone kubeconfig for workload cluster")

	kubeconfigClusterCmd.AddCommand(getKubeconfigClusterCmd)
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
	isManagementCluster := false
	if gkc.workloadClusterName == "" {
		isManagementCluster = true
	}

	if !isManagementCluster && gkc.namespace == "" {
		gkc.namespace = "default"
	}

	getClusterPinnipedInfoOptions := tkgctl.GetClusterPinnipedInfoOptions{
		ClusterName:         gkc.workloadClusterName,
		Namespace:           gkc.namespace,
		IsManagementCluster: isManagementCluster,
	}

	clusterPinnipedInfo, err := tkgctlClient.GetClusterPinnipedInfo(getClusterPinnipedInfoOptions)
	if err != nil {
		return err
	}
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
	err = tkgclient.MergeKubeConfigWithoutSwitchContext(kubeconfigbytes, gkc.exportFile)
	if err != nil {
		return errors.Wrap(err, "unable to merge cluster kubeconfig into the current kubeconfig path")
	}

	if gkc.exportFile != "" {
		log.Infof("You can now access the cluster by running 'kubectl config use-context %s' under path '%s' \n", kubeconfig.CurrentContext, gkc.exportFile)
	} else {
		log.Infof("You can now access the cluster by running 'kubectl config use-context %s'\n", kubeconfig.CurrentContext)
	}

	return nil
}
