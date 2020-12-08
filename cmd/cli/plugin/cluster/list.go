package main

import (
	"errors"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vmware-tanzu-private/core/apis/client/v1alpha1"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/component"
	"github.com/vmware-tanzu-private/core/pkg/v1/client"
	"github.com/vmware-tanzu-private/tkg-cli/pkg/tkgctl"
	"github.com/vmware-tanzu-private/tkg-cli/pkg/utils"
)

type listClusterOptions struct {
	namespace    string
	includeMC    bool
	outputFormat string
}

var lc = &listClusterOptions{}

var listClustersCmd = &cobra.Command{
	Use:   "list",
	Short: "List clusters",
	RunE:  list,
}

func init() {
	listClustersCmd.Flags().StringVarP(&lc.namespace, "namespace", "n", "", "The namespace from which to list workload clusters. If not provided clusters from all namespaces will be returned")
	listClustersCmd.Flags().BoolVarP(&lc.includeMC, "include-management-cluster", "", false, "Show active management cluster information as well")
	listClustersCmd.Flags().StringVarP(&lc.outputFormat, "output", "o", "", "Output format. Supported formats: json|yaml")
}

func list(cmd *cobra.Command, args []string) error {
	server, err := client.GetCurrentServer()
	if err != nil {
		return err
	}

	if server.IsGlobal() {
		return errors.New("listing cluster with global setting is not implemented yet")
	}
	return listClusters(server)
}

func listClusters(server *v1alpha1.Server) error {
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

	ccOptions := tkgctl.ListTKGClustersOptions{
		ClusterName: "",
		Namespace:   lc.namespace,
		IncludeMC:   lc.includeMC,
	}

	clusters, err := tkgctlClient.GetClusters(ccOptions)
	if err != nil {
		return err
	}

	if lc.outputFormat != "" {
		return utils.RenderOutput(clusters, lc.outputFormat)
	}

	t := component.NewTableWriter("NAME", "NAMESPACE", "STATUS", "CONTROLPLANE", "WORKERS", "KUBERNETES", "ROLES")
	for _, cl := range clusters {
		clusterRoles := "<none>"
		if len(cl.Roles) != 0 {
			clusterRoles = strings.Join(cl.Roles, ",")
		}
		t.Append([]string{cl.Name, cl.Namespace, cl.Status, cl.ControlPlaneCount, cl.WorkerCount, cl.K8sVersion, clusterRoles})
	}
	t.Render()

	return nil
}
