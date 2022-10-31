// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"encoding/base64"
	"fmt"
	"net"
	"strings"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	kerrors "k8s.io/apimachinery/pkg/util/errors"

	"github.com/go-openapi/swag"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	clusterctl "sigs.k8s.io/cluster-api/cmd/clusterctl/client"

	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/tkg/yamlprocessor"
)

// UpgradeAddonOptions upgrade addon options
type UpgradeAddonOptions struct {
	AddonNames        []string
	ClusterName       string
	Namespace         string
	Kubeconfig        string
	IsRegionalCluster bool
	Edition           string
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
		// Build edition is required to perform addons upgrade. The providers/ytt is referencing to build edition for the
		// rendering logic. Setting build edition here is to make sure addons upgrade to work properly.
		Edition: options.Edition,
	}

	kcp, err := regionalClusterClient.GetKCPObjectForCluster(options.ClusterName, options.Namespace)
	if err != nil {
		return errors.Wrapf(err, "unable to find control plane node object for cluster %s", options.ClusterName)
	}
	if kcp.Spec.MachineTemplate.InfrastructureRef.Kind == constants.KindVSphereMachineTemplate {
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
		case "addons-management/standard-package-repo":
			// This ensures that CRS for standard-package-repo gets added
			// for new clusters as well as upgrades. CRS for standard repo
			// package resource will ensure one time create and no upgrades after
			// initial create. The case is empty because we only need to set
			// crsDisabledAddon = false which is the default and ensures that
			// default isn't triggered.
		case "tkr/tkr-controller":
			if !options.IsRegionalCluster {
				return errors.Errorf("upgrade of '%s' component is only supported on management cluster", addonName)
			}
			crsDisabledAddon = true
			ccOptions.TargetNamespace = constants.TkrNamespace
		case "packages/management-package-repo":
			if !options.IsRegionalCluster {
				return errors.Errorf("upgrade of '%s' component is only supported on management cluster", addonName)
			}
			crsDisabledAddon = true
		case "packages/management-package":
			if !options.IsRegionalCluster {
				return errors.Errorf("upgrade of '%s' component is only supported on management cluster", addonName)
			}
			crsDisabledAddon = true
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

		if options.IsRegionalCluster {
			err = c.RetrieveRegionalClusterConfiguration(regionalClusterClient)
		} else {
			err = c.RetrieveWorkloadClusterConfiguration(regionalClusterClient, currentClusterClient, options.ClusterName, options.Namespace)
		}
		if err != nil {
			return errors.Wrap(err, "unable to set cluster configuration")
		}

		yaml, err := clusterConfigurationGetter(&ccOptions)
		if err != nil {
			return errors.Wrap(err, "unable to get cluster configuration")
		}

		// ensure current kapp-controller deployment on the management cluster has last-applied annotation, which is required for future apply operations
		if addonName == "addons-management/kapp-controller" && options.IsRegionalCluster {
			if err := clusterClient.PatchKappControllerLastAppliedAnnotation(options.Namespace); err != nil {
				return errors.Wrap(err, "unable to add last-applied annotation on kapp-controller")
			}
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

// RetrieveRegionalClusterConfiguration gets TKG configurations from regional cluster and updates the in-memory config.
// this is required when we want to mutate the existing regional cluster.
func (c *TkgClient) RetrieveRegionalClusterConfiguration(regionalClusterClient clusterclient.Client) error {
	if err := c.setProxyConfiguration(regionalClusterClient); err != nil {
		return errors.Wrapf(err, "error while getting proxy configuration from cluster and setting it")
	}

	if err := c.setCustomImageRepositoryConfiguration(regionalClusterClient); err != nil {
		return errors.Wrapf(err, "error while getting custom image repository configuration from cluster and setting it")
	}

	clusterName, regionalClusterNamespace, err := c.getRegionalClusterNameAndNamespace(regionalClusterClient)
	if err != nil {
		return errors.Wrap(err, "unable to get name and namespace of current management cluster")
	}

	if err := c.setNetworkingConfiguration(regionalClusterClient, clusterName, regionalClusterNamespace); err != nil {
		return errors.Wrap(err, "error while initializing networking configuration")
	}
	return nil
}

// RetrieveWorkloadClusterConfiguration gets TKG configurations from regional cluster as well as workload cluster
// and updates the in-memory config. This is required when we want to mutate the existing workload cluster.
func (c *TkgClient) RetrieveWorkloadClusterConfiguration(regionalClusterClient, workloadClusterClient clusterclient.Client, clusterName, clusterNamespace string) error {
	if err := c.setProxyConfiguration(workloadClusterClient); err != nil {
		return errors.Wrapf(err, "error while getting proxy configuration from cluster and setting it")
	}

	// Sets custom image repository configuration from Management Cluster even for workload cluster
	// Currently, TKG does not support using different image repositories for management and workload clusters.
	if err := c.setCustomImageRepositoryConfiguration(regionalClusterClient); err != nil {
		return errors.Wrapf(err, "error while getting custom image repository configuration from cluster and setting it")
	}

	if err := c.setNetworkingConfiguration(regionalClusterClient, clusterName, clusterNamespace); err != nil {
		return errors.Wrap(err, "error while initializing networking configuration")
	}

	return nil
}

func (c *TkgClient) setProxyConfiguration(clusterClusterClient clusterclient.Client) (retErr error) {
	// make sure proxy parameters are non-empty
	defer func() {
		if retErr == nil {
			c.SetDefaultProxySettings()
		}
	}()
	configmap := &corev1.ConfigMap{}
	if err := clusterClusterClient.GetResource(configmap, constants.KappControllerConfigMapName, constants.KappControllerNamespace, nil, nil); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return errors.Wrapf(err, "unable to get object '%v' in namespace '%v'", constants.KappControllerConfigMapName, constants.KappControllerNamespace)
	}

	if configmap.Data == nil {
		return nil
	}

	if httpProxy := configmap.Data["httpProxy"]; httpProxy != "" {
		c.TKGConfigReaderWriter().Set(constants.TKGHTTPProxy, httpProxy)
		c.TKGConfigReaderWriter().Set(constants.TKGHTTPProxyEnabled, trueString)
	}

	if httpsProxy := configmap.Data["httpsProxy"]; httpsProxy != "" {
		c.TKGConfigReaderWriter().Set(constants.TKGHTTPSProxy, httpsProxy)
	}

	if noProxy := configmap.Data["noProxy"]; noProxy != "" {
		c.TKGConfigReaderWriter().Set(constants.TKGNoProxy, noProxy)
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

	// Read TKG_CUSTOM_IMAGE_REPOSITORY from configuration first to allow user to provide different image repository during cluster deletion
	if customImageRepository, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableCustomImageRepository); err != nil || customImageRepository == "" {
		if customImageRepository := configmap.Data["imageRepository"]; customImageRepository != "" {
			c.TKGConfigReaderWriter().Set(constants.ConfigVariableCustomImageRepository, customImageRepository)
		}
	}

	// Read TKG_CUSTOM_IMAGE_REPOSITORY_CA_CERTIFICATE from configuration first to allow user to provide different image repository during cluster deletion
	if customImageRepositoryCaCertificate, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableCustomImageRepositoryCaCertificate); err != nil || customImageRepositoryCaCertificate == "" {
		if customImageRepositoryCaCertificate := configmap.Data["caCerts"]; customImageRepositoryCaCertificate != "" {
			customImageRepositoryCaCertificateEncoded := base64.StdEncoding.EncodeToString([]byte(customImageRepositoryCaCertificate))
			c.TKGConfigReaderWriter().Set(constants.ConfigVariableCustomImageRepositoryCaCertificate, customImageRepositoryCaCertificateEncoded)
		}
	}

	return nil
}

func (c *TkgClient) setNetworkingConfiguration(regionalClusterClient clusterclient.Client, clusterName, clusterNamespace string) error {
	cluster := &capi.Cluster{}
	err := regionalClusterClient.GetResource(cluster, clusterName, clusterNamespace, nil, nil)
	if err != nil {
		return errors.Wrapf(err, "unable to get cluster %q from namespace %q", clusterName, clusterNamespace)
	}

	if cluster.Spec.ClusterNetwork != nil {
		if cluster.Spec.ClusterNetwork.Pods != nil && len(cluster.Spec.ClusterNetwork.Pods.CIDRBlocks) > 0 {
			c.TKGConfigReaderWriter().Set(constants.ConfigVariableClusterCIDR, strings.Join(cluster.Spec.ClusterNetwork.Pods.CIDRBlocks, ","))
		}
		if cluster.Spec.ClusterNetwork.Services != nil && len(cluster.Spec.ClusterNetwork.Services.CIDRBlocks) > 0 {
			c.TKGConfigReaderWriter().Set(constants.ConfigVariableServiceCIDR, strings.Join(cluster.Spec.ClusterNetwork.Services.CIDRBlocks, ","))
		}
		ipFamily, err := GetIPFamily(cluster)
		if err != nil {
			return errors.Wrapf(err, "unable to get IPFamily of %q", clusterName)
		}
		switch ipFamily {
		case IPv4IPFamily:
			c.TKGConfigReaderWriter().Set(constants.ConfigVariableIPFamily, constants.IPv4Family)
		case IPv6IPFamily:
			c.TKGConfigReaderWriter().Set(constants.ConfigVariableIPFamily, constants.IPv6Family)
		case DualStackIPFamily:
			ip, _, err := net.ParseCIDR(cluster.Spec.ClusterNetwork.Services.CIDRBlocks[0])
			if err != nil {
				return fmt.Errorf("unable to detect valid IPFamily, could not parse CIDR: %s", err.Error())
			}
			if ip.To4() == nil {
				c.TKGConfigReaderWriter().Set(constants.ConfigVariableIPFamily, constants.DualStackPrimaryIPv6Family)
			} else {
				c.TKGConfigReaderWriter().Set(constants.ConfigVariableIPFamily, constants.DualStackPrimaryIPv4Family)
			}
		default:
			return fmt.Errorf("unable to detect valid IPFamily, found %s", ipFamily)
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
