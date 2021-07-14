// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"encoding/base64"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	kerrors "k8s.io/apimachinery/pkg/util/errors"

	"github.com/go-openapi/swag"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	clusterctl "sigs.k8s.io/cluster-api/cmd/clusterctl/client"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/yamlprocessor"
)

// UpgradeAddonOptions upgrade addon options
type UpgradeAddonOptions struct {
	AddonNames        []string
	ClusterName       string
	Namespace         string
	Kubeconfig        string
	IsRegionalCluster bool
}

// UpgradeAddon upgrades addons
func (c *TkgClient) UpgradeAddon(options *UpgradeAddonOptions) error {
	if options == nil {
		return errors.New("invalid upgrade addon options nil")
	}

	currentRegion, err := c.GetCurrentRegionContext()
	if err != nil {
		return errors.Wrap(err, "cannot get current management cluster context")
	}
	options.Kubeconfig = currentRegion.SourceFilePath

	log.V(4).Info("Creating management cluster client...")
	regionalClusterClient, err := clusterclient.NewClient(currentRegion.SourceFilePath, currentRegion.ContextName, clusterclient.Options{OperationTimeout: c.timeout})
	if err != nil {
		return errors.Wrap(err, "unable to get cluster client while upgrading addon")
	}

	isPacific, err := regionalClusterClient.IsPacificRegionalCluster()
	if err != nil {
		return errors.Wrap(err, "error determining 'Tanzu Kubernetes Cluster service for vSphere' management cluster")
	}
	if isPacific {
		return errors.Wrap(err, "addons upgrade for 'Tanzu Kubernetes Cluster service for vSphere' management cluster is not yet supported")
	}

	// get the management cluster name and namespace in case of management cluster upgrade
	if options.IsRegionalCluster {
		clusterName, namespace, err := c.getRegionalClusterNameAndNamespace(regionalClusterClient)
		if err != nil {
			return errors.Wrap(err, "unable to get current management cluster information")
		}
		options.ClusterName = clusterName
		options.Namespace = namespace
	}

	var currentClusterClient clusterclient.Client
	if options.IsRegionalCluster {
		currentClusterClient = regionalClusterClient
	} else {
		log.V(4).Info("Creating workload cluster client...")
		currentClusterClient, err = c.getWorkloadClusterClient(options.ClusterName, options.Namespace)
		if err != nil {
			return errors.Wrap(err, "unable to get workload cluster client")
		}
	}

	return c.DoUpgradeAddon(regionalClusterClient, currentClusterClient, options, c.GetClusterConfiguration)
}

// DoUpgradeAddon performs steps for addons upgrade
func (c *TkgClient) DoUpgradeAddon(regionalClusterClient clusterclient.Client, //nolint:funlen,gocyclo
	currentClusterClient clusterclient.Client,
	options *UpgradeAddonOptions,
	clusterConfigurationGetter func(*CreateClusterOptions) ([]byte, error)) error {
	// this is required as clusters created with old version of cli does
	// not have cluster-name label applied to cluster object and this label
	// is required for CRS objects
	err := c.addClusterNameLabel(regionalClusterClient, options)
	if err != nil {
		return err
	}

	k8sVersion, err := c.tkgBomClient.GetDefaultK8sVersion()
	if err != nil {
		return errors.Wrap(err, "unable to get default kubernetes version from BoM files")
	}

	configOptions := ClusterConfigOptions{
		ClusterName:              options.ClusterName,
		Kubeconfig:               clusterctl.Kubeconfig{Path: options.Kubeconfig},
		KubernetesVersion:        k8sVersion,
		ControlPlaneMachineCount: swag.Int64(int64(1)),
		WorkerMachineCount:       swag.Int64(int64(1)),
		TargetNamespace:          options.Namespace,
	}

	// XXX: Using dev plan as for clusters created with old tkg cli does not store plan information anywhere
	// directly for our usage. Also, this information is only used for cluster template generation
	// which is independent as we are just filtering cluster metadata objects at the moment
	// NOTE: Please make sure to verify dev plan is sufficient for the usecase you are trying to add in future
	configOptions.ProviderRepositorySource = &clusterctl.ProviderRepositorySourceOptions{
		InfrastructureProvider: "",
		Flavor:                 constants.PlanDev,
	}

	createClusterOptions := CreateClusterOptions{
		ClusterConfigOptions: configOptions,
	}

	kcp, err := regionalClusterClient.GetKCPObjectForCluster(options.ClusterName, options.Namespace)
	if err != nil {
		return errors.Wrapf(err, "unable to find control plane node object for cluster %s", options.ClusterName)
	}
	if kcp.Spec.InfrastructureTemplate.Kind == constants.VSphereMachineTemplate {
		// As vSphereControlPlaneEndpointIP is required for vSphere cluster for --dry-run
		// while creating workload cluster template, setting this to dummy value currently
		// Note: This value will not be used for addons upgrade
		createClusterOptions.VsphereControlPlaneEndpoint = "unused"
	}

	createClusterOptions.YamlProcessor = yamlprocessor.NewYttProcessorWithConfigDir(c.tkgConfigDir)

	// Skip validation while creating cluster template
	// As we are more interested in generating basic cluster template using default BoM file,
	// we don't want to verify the credentials and other config variables mentioned in
	// user's TKG config file as those are irrelevant for the usecase we are implementing
	createClusterOptions.SkipValidation = true

	if options.IsRegionalCluster {
		createClusterOptions.ClusterType = ManagementCluster
	} else {
		createClusterOptions.ClusterType = WorkloadCluster
	}

	// Only upgrade of cluster metadata, kapp-controller, addons-manager and tkr-controller to the existing cluster
	// during upgrade process and other addons upgrade is not yet supported

	errListAddonUpgrade := []error{}

	for _, addonName := range options.AddonNames {
		crsDisabledAddon := false
		ccOptions := CreateClusterOptions{}
		err := copier.Copy(&ccOptions, &createClusterOptions)
		if err != nil {
			return errors.Wrap(err, "unable to copy createClusterOptions object")
		}

		// apply the addons yaml to management cluster by default
		// override this if operation needs to be performed on
		// current cluster instead of management cluster
		clusterClient := regionalClusterClient

		switch addonName {
		case "metadata/tkg":
			ccOptions.TargetNamespace = constants.TkgPublicNamespace
			crsDisabledAddon = true
			// As tkg metadata yaml is associated per cluster instead
			// of using management cluster client, we need to use
			// current cluster which will point to workload cluster
			// during workload cluster upgrade and management cluster
			// during management cluster upgrade
			clusterClient = currentClusterClient
		case "addons-management/kapp-controller":
			crsDisabledAddon = true
		case "addons-management/tanzu-addons-manager":
			if !options.IsRegionalCluster {
				return errors.Errorf("upgrade of '%s' component is only supported on management cluster", addonName)
			}
			crsDisabledAddon = true
		case "addons-management/core-package-repo":
			if !options.IsRegionalCluster {
				return errors.Errorf("upgrade of '%s' component is only supported on management cluster", addonName)
			}
			crsDisabledAddon = true
		case "tkr/tkr-controller":
			if !options.IsRegionalCluster {
				return errors.Errorf("upgrade of '%s' component is only supported on management cluster", addonName)
			}
			crsDisabledAddon = true
			ccOptions.TargetNamespace = constants.TkrNamespace
		default:
			return errors.Errorf("upgrade of '%s' component is not supported", addonName)
		}

		// This variable configuration is important to filter only specific addons objects
		// with ytt templating. There is specific code written on ytt side to do the filtering
		c.TKGConfigReaderWriter().Set(constants.ConfigVariableFilterByAddonType, addonName)

		if crsDisabledAddon {
			// This variable configuration is used to deactivate CRS for addons and generate the raw yaml
			c.TKGConfigReaderWriter().Set(constants.ConfigVaraibleDisableCRSForAddonType, addonName)
		}

		if err := c.setConfigurationForUpgrade(regionalClusterClient); err != nil {
			return errors.Wrap(err, "unable to set cluster configuration")
		}

		yaml, err := clusterConfigurationGetter(&ccOptions)
		if err != nil {
			return errors.Wrap(err, "unable to get cluster configuration")
		}

		log.Infof("updating additional components: '%s' ...", addonName)

		err = clusterClient.Apply(string(yaml))
		if err != nil {
			errListAddonUpgrade = append(errListAddonUpgrade, errors.Wrapf(err, "error while upgrading additional component '%s'", addonName))
		}
	}

	// Aggregate all errors and return error
	err = kerrors.NewAggregate(errListAddonUpgrade)
	if err != nil {
		return err
	}

	return nil
}

func (c *TkgClient) setConfigurationForUpgrade(regionalClusterClient clusterclient.Client) error {
	if err := c.setProxyConfiguration(regionalClusterClient); err != nil {
		return errors.Wrapf(err, "error while getting proxy configuration from cluster and setting it")
	}

	if err := c.setCustomImageRepositoryConfiguration(regionalClusterClient); err != nil {
		return errors.Wrapf(err, "error while getting custom image repository configuration from cluster and setting it")
	}

	return nil
}

func (c *TkgClient) setProxyConfiguration(regionalClusterClient clusterclient.Client) (retErr error) {
	// make sure proxy parameters are non-empty
	defer func() {
		if retErr == nil {
			c.SetDefaultProxySettings()
		}
	}()
	configmap := &corev1.ConfigMap{}
	if err := regionalClusterClient.GetResource(configmap, constants.KappControllerConfigMapName, constants.KappControllerNamespace, nil, nil); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return errors.Wrapf(err, "unable to get object '%v' in namespace '%v'", constants.KappControllerConfigMapName, constants.KappControllerNamespace)
	}

	if configmap.Data == nil {
		return nil
	}

	if httpProxy, err := c.TKGConfigReaderWriter().Get(constants.HTTPProxy); err != nil || httpProxy == "" {
		if httpProxy := configmap.Data["httpProxy"]; httpProxy != "" {
			c.TKGConfigReaderWriter().Set(constants.TKGHTTPProxy, httpProxy)
		}
	}

	if httpsProxy, err := c.TKGConfigReaderWriter().Get(constants.TKGHTTPSProxy); err != nil || httpsProxy == "" {
		if httpsProxy := configmap.Data["httpsProxy"]; httpsProxy != "" {
			c.TKGConfigReaderWriter().Set(constants.TKGHTTPSProxy, httpsProxy)
		}
	}

	if noProxy, err := c.TKGConfigReaderWriter().Get(constants.TKGNoProxy); err != nil || noProxy == "" {
		if noProxy := configmap.Data["noProxy"]; noProxy != "" {
			c.TKGConfigReaderWriter().Set(constants.TKGNoProxy, noProxy)
		}
	}

	return nil
}

func (c *TkgClient) setCustomImageRepositoryConfiguration(regionalClusterClient clusterclient.Client) error {
	configmap := &corev1.ConfigMap{}
	if err := regionalClusterClient.GetResource(configmap, constants.TkrConfigMapName, constants.TkrNamespace, nil, nil); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return errors.Wrapf(err, "unable to get object '%v' in namespace '%v'", constants.TkrConfigMapName, constants.TkrNamespace)
	}

	if configmap.Data == nil {
		return nil
	}

	if customImageRepository, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableCustomImageRepository); err != nil || customImageRepository == "" {
		if customImageRepository := configmap.Data["imageRepository"]; customImageRepository != "" {
			c.TKGConfigReaderWriter().Set(constants.ConfigVariableCustomImageRepository, customImageRepository)
		}
	}

	if customImageRepositoryCaCertificate, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableCustomImageRepositoryCaCertificate); err != nil || customImageRepositoryCaCertificate == "" {
		if customImageRepositoryCaCertificate := configmap.Data["caCerts"]; customImageRepositoryCaCertificate != "" {
			customImageRepositoryCaCertificateEncoded := base64.StdEncoding.EncodeToString([]byte(customImageRepositoryCaCertificate))
			c.TKGConfigReaderWriter().Set(constants.ConfigVariableCustomImageRepositoryCaCertificate, customImageRepositoryCaCertificateEncoded)
		}
	}

	return nil
}

func (c *TkgClient) addClusterNameLabel(regionalClusterClient clusterclient.Client, options *UpgradeAddonOptions) error {
	patchFormat := `
	{
		"metadata": {
			"labels": {
				"tkg.tanzu.vmware.com/cluster-name": "%s"
			}
		}
	}`
	patchLabel := fmt.Sprintf(patchFormat, options.ClusterName)
	err := regionalClusterClient.PatchClusterObject(options.ClusterName, options.Namespace, patchLabel)
	if err != nil {
		return errors.Wrap(err, "unable to patch the cluster object with cluster name label")
	}
	return nil
}
