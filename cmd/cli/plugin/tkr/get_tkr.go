package main

import (
	"context"
	"strings"

	"github.com/pkg/errors"

	"github.com/spf13/cobra"
	"github.com/vmware-tanzu-private/core/apis/client/v1alpha1"
	runv1alpha1 "github.com/vmware-tanzu-private/core/apis/run/v1alpha1"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/component"
	"github.com/vmware-tanzu-private/core/pkg/v1/client"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

var getTanzuKubernetesRleasesCmd = &cobra.Command{
	Use:   "get TKR_NAME",
	Short: "Get available TanzuKubernetesReleases",
	Long:  "Get available TanzuKubernetesReleases",
	RunE:  getKubernetesReleases,
}

func getKubernetesReleases(cmd *cobra.Command, args []string) error {
	server, err := client.GetCurrentServer()
	if err != nil {
		return err
	}

	if server.IsGlobal() {
		return errors.New("getting TanzuKubernetesRelease with a global server is not implemented yet")
	}

	crtClient, err := getManagementClusterClient(server)
	if err != nil {
		return errors.Wrap(err, "failed to create controller")
	}

	tkrList := &runv1alpha1.TanzuKubernetesReleaseList{}
	err = crtClient.List(context.Background(), tkrList)
	if err != nil {
		return errors.Wrap(err, "failed to list current TKRs")
	}

	tkrName := ""
	if len(args) != 0 {
		tkrName = args[0]
	}

	t := component.NewTableWriter("NAME", "VERSION", "COMPATIBLE", "UPGRADEAVAILABLE")
	for _, tkr := range tkrList.Items {
		if tkrName != "" && !strings.HasPrefix(tkr.Name, tkrName) {
			continue
		}

		compatible := ""
		upgradeAvailable := ""

		for _, condition := range tkr.Status.Conditions {
			if condition.Type == runv1alpha1.ConditionCompatible {
				compatible = string(condition.Status)
			}
			if condition.Type == runv1alpha1.ConditionUpgradeAvailable {
				upgradeAvailable = string(condition.Status)
			}
		}

		t.Append([]string{tkr.Name, tkr.Spec.Version, compatible, upgradeAvailable})
	}
	t.Render()

	return nil
}

func getManagementClusterClient(server *v1alpha1.Server) (crtclient.Client, error) {

	var scheme = runtime.NewScheme()

	_ = runv1alpha1.AddToScheme(scheme)

	config, err := clientcmd.LoadFromFile(server.ManagementClusterOpts.Path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load kubeconfig from %s", server.ManagementClusterOpts.Path)
	}
	config.CurrentContext = server.ManagementClusterOpts.Context

	rawConfig, err := clientcmd.Write(*config)
	restConfig, err := clientcmd.RESTConfigFromKubeConfig(rawConfig)
	if err != nil {
		return nil, errors.Errorf("Unable to set up rest config due to : %v", err)
	}

	mapper, err := apiutil.NewDynamicRESTMapper(restConfig, apiutil.WithLazyDiscovery)
	if err != nil {
		return nil, errors.Errorf("Unable to set up rest mapper due to : %v", err)
	}
	return crtclient.New(restConfig, crtclient.Options{Scheme: scheme, Mapper: mapper})

}
