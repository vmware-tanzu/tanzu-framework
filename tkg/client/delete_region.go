// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/pkg/errors"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/tkg/region"
	"github.com/vmware-tanzu/tanzu-framework/tkg/utils"
)

// Error message constants
const (
	ErrorMissingRegionalClusterObject = "management cluster object is not present in given management cluster"
	ErrorNoClusterObject              = "no Cluster object present in the given management cluster"
	ErrorGettingClusterObjects        = "unable to get cluster resources, %s. Are you sure the cluster you are using is a management cluster?"
	ErrorDeleteAbort                  = `Deletion is aborted because management cluster is currently managing the following workload clusters:

%s

You need to delete these clusters first before deleting the management cluster.

Alternatively, you can use the --force flag to force the deletion of the management cluster but doing so will orphan the above-mentioned clusters and leave them unmanaged.`
)

// DeleteRegion delete management cluster
func (c *TkgClient) DeleteRegion(options DeleteRegionOptions) error { //nolint:funlen,gocyclo
	var err error
	var isSuccessful = false
	var isStartedRegionalClusterDeletion = false
	var isCleanupClusterCreated = false
	var cleanupClusterName string
	var cleanupClusterKubeconfigPath string

	defer func() {
		// if management cluster deletion is not being started and kind cluster is already created
		if !isSuccessful && isStartedRegionalClusterDeletion {
			c.displayHelpTextOnDeleteRegionFailure(cleanupClusterKubeconfigPath, isCleanupClusterCreated, cleanupClusterName, options.ClusterName)
			return
		}

		if isCleanupClusterCreated {
			if err := c.teardownKindCluster(cleanupClusterName, cleanupClusterKubeconfigPath, options.UseExistingCluster); err != nil {
				log.Warning(err.Error())
			}
		}

		_ = utils.DeleteFile(cleanupClusterKubeconfigPath)
	}()

	if err := c.ValidatePrerequisites(!options.UseExistingCluster, true); err != nil {
		return err
	}

	clusterclientOptions := clusterclient.Options{
		GetClientInterval: 5 * time.Second,
		GetClientTimeout:  10 * time.Second,
		OperationTimeout:  c.timeout,
	}

	contexts, err := c.GetRegionContexts(options.ClusterName)
	if err != nil || len(contexts) == 0 {
		return errors.Errorf("management cluster %s not found", options.ClusterName)
	}
	regionContext := contexts[0]

	regionalClusterClient, err := clusterclient.NewClient(regionContext.SourceFilePath, regionContext.ContextName, clusterclientOptions)
	if err != nil {
		return errors.Wrap(err, "unable to create cluster client for management cluster")
	}

	log.Info("Verifying management cluster...")
	// Verify presence of workload cluster and return management cluster name and namespace
	regionalClusterNamespace, err := c.verifyDeleteRegion(regionalClusterClient, options)
	if err != nil {
		return err
	}

	err = c.RetrieveRegionalClusterConfiguration(regionalClusterClient)
	if err != nil {
		return errors.Wrap(err, "failed to set configurations for deletion")
	}

	isFailure, err := c.IsManagementClusterAKindCluster(options.ClusterName)
	if err != nil {
		return err
	}

	if !isFailure {
		cleanupClusterKubeconfigPath, err = getTKGKubeConfigPath(false)
		if err != nil {
			return errors.Wrap(err, "cannot get cleanup cluster kubeconfig path ")
		}

		// configure variables required to deploy providers
		if err := c.configureVariablesForProvidersInstallation(regionalClusterClient); err != nil {
			return errors.Wrap(err, "unable to configure variables for provider installation")
		}

		// Get the kubeconfig and initOptions for cleanup cluster
		log.V(1).Infof("Using cleanup cluster kubeconfig from path: %v", cleanupClusterKubeconfigPath)
		initOptionsForCleanupCluster, err := c.getCleanupClusterOptions(regionalClusterClient, cleanupClusterKubeconfigPath)
		if err != nil {
			return err
		}

		log.Info("Setting up cleanup cluster...")

		// Create cleanup kind cluster and backup the kubeconfig under ./kube-tkg/tmp/
		if cleanupClusterName, err = c.ensureKindCluster(options.Kubeconfig, options.UseExistingCluster, cleanupClusterKubeconfigPath); err != nil {
			return errors.Wrap(err, "unable to create cleanup cluster")
		}

		isCleanupClusterCreated = true

		cleanupClusterClient, err := clusterclient.NewClient(cleanupClusterKubeconfigPath, "", clusterclientOptions)
		if err != nil {
			return errors.Wrap(err, "cannot create cleanup cluster client")
		}

		// If clusterclass feature flag is enabled then deploy kapp-controller
		if config.IsFeatureActivated(constants.FeatureFlagPackageBasedLCM) {
			if err = c.InstallOrUpgradeKappController(cleanupClusterKubeconfigPath, "", constants.OperationTypeInstall); err != nil {
				return errors.Wrap(err, "unable to install kapp-controller to bootstrap cluster")
			}
		}

		log.Info("Installing providers to cleanup cluster...")
		// Initialize cleanup cluster using same provider name and version from management cluster
		if err = c.InitializeProviders(&initOptionsForCleanupCluster, cleanupClusterClient, cleanupClusterKubeconfigPath); err != nil {
			return errors.Wrap(err, "unable to initialize providers")
		}

		// If clusterclass feature flag is enabled then deploy management components
		if config.IsFeatureActivated(constants.FeatureFlagPackageBasedLCM) {
			// Read config variable from management cluster and set it into cleanup cluster
			configValues, err := c.getUserConfigVariableValueMapFromSecret(regionContext.SourceFilePath, regionContext.ContextName)
			if err != nil {
				return errors.Wrap(err, "unable to get config variables from management cluster")
			}
			for k, v := range configValues {
				// Handle bool, int and string
				c.TKGConfigReaderWriter().Set(k, fmt.Sprint(v))
			}
			if err = c.InstallOrUpgradeManagementComponents(cleanupClusterKubeconfigPath, "", false); err != nil {
				return errors.Wrap(err, "unable to install management components to bootstrap cluster")
			}
		}

		isStartedRegionalClusterDeletion = true

		log.Info("Moving TKR and Cluster API objects from management cluster to cleanup cluster...")
		regionalClusterKubeConfigPath, err := regionalClusterClient.ExportCurrentKubeconfigToFile()
		if err != nil {
			return errors.Wrap(err, "unable to export management cluster's kubeconfig")
		}
		defer func() {
			_ = utils.DeleteFile(regionalClusterKubeConfigPath)
		}()

		// Copy coresponding TKR and OSImages from management cluster to cleanup cluster for all namespaces
		if err = c.CopyNeededTKRAndOSImages(regionalClusterClient, cleanupClusterClient); err != nil {
			return errors.Wrap(err, "unable to copy coresponding TKR and OSImages from management cluster to cleanup cluster")
		}

		// Move all Cluster API objects from management cluster to cleanup cluster for all namespaces
		if err = c.MoveObjects(regionalClusterKubeConfigPath, cleanupClusterKubeconfigPath, regionalClusterNamespace); err != nil {
			return errors.Wrap(err, "unable to move Cluster API objects from management cluster to cleanup cluster")
		}

		log.Info("Waiting for the Cluster API objects to get ready after move...")
		if err := c.WaitForClusterReadyAfterReverseMove(cleanupClusterClient, options.ClusterName, regionalClusterNamespace); err != nil {
			return errors.Wrap(err, "unable to wait for cluster getting ready for move")
		}

		if err = c.cleanUpAVIResourcesInManagementCluster(regionalClusterClient, options.ClusterName, regionalClusterNamespace); err != nil {
			return errors.Wrap(err, "unable to clean up avi resource")
		}
	} else { // remove a management cluster whose deploymentStatus is 'Failed'
		cleanupClusterKubeconfigPath = regionContext.SourceFilePath
		cleanupClusterName = strings.TrimLeft(regionContext.ContextName, "kind-")
		isCleanupClusterCreated = true
		isStartedRegionalClusterDeletion = true
	}

	log.Info("Deleting management cluster...")
	// Delete management cluster, this will start process of deleting cluster and its underlying resources
	if err = c.deleteCluster(cleanupClusterKubeconfigPath, options.ClusterName, regionalClusterNamespace); err != nil {
		return errors.Wrap(err, "unable to delete management cluster")
	}

	// management cluster deletion happens in background and we cannot teardown the cleanup kind cluster until the management cluster is deleted successfully
	if err = c.waitForClusterDeletion(cleanupClusterKubeconfigPath, options.ClusterName, regionalClusterNamespace); err != nil {
		return errors.Wrapf(err, "error waiting for management cluster '%s' to be deleted", options.ClusterName)
	}
	lock, err := utils.GetFileLockWithTimeOut(filepath.Join(c.tkgConfigDir, constants.LocalTanzuFileLock), utils.DefaultLockTimeout)
	if err != nil {
		return errors.Wrap(err, "cannot acquire lock for deleting region context")
	}

	defer func() {
		if err := lock.Unlock(); err != nil {
			log.Warningf("cannot release lock for deleting region context, reason: %v", err)
		}
	}()
	err = c.regionManager.DeleteRegionContext(options.ClusterName)
	if err != nil {
		log.Warningf("Failed to delete management cluster %s context from tkg config file", options.ClusterName)
	}

	log.Infof("Management cluster '%s' deleted.", options.ClusterName)

	// delete management cluster config from default kubeconfig file
	if regionContext.Status != region.Failed {
		userKubeconfigPath := getDefaultKubeConfigFile()
		log.Infof("Deleting the management cluster context from the kubeconfig file '%s'", userKubeconfigPath)
		if err = DeleteContextFromKubeConfig(userKubeconfigPath, regionContext.ContextName); err != nil {
			log.Warningf("Failed to delete management cluster context from the kubeconfig file '%s'", userKubeconfigPath)
		}
	}

	isSuccessful = true

	return nil
}

// IsManagementClusterAKindCluster Determining if the management cluster creation is successful based on the presence of Annotation 'TKGVERSION: <version>'.
//
//	Management clusters are annotated with the TKG version if they are successful.
//	If the annotation is not present, the cluster is determined to be failed management cluster (kind cluster) with some un-deleted infrastructure resources
func (c *TkgClient) IsManagementClusterAKindCluster(clusterName string) (bool, error) {
	clusterclientOptions := clusterclient.Options{
		GetClientInterval: 5 * time.Second,
		GetClientTimeout:  10 * time.Second,
		OperationTimeout:  c.timeout,
	}

	contexts, err := c.GetRegionContexts(clusterName)
	if err != nil || len(contexts) == 0 {
		return false, errors.Errorf("management cluster %s not found", clusterName)
	}
	regionContext := contexts[0]

	regionalClusterClient, err := clusterclient.NewClient(regionContext.SourceFilePath, regionContext.ContextName, clusterclientOptions)
	if err != nil {
		return false, errors.Wrap(err, "unable to create cluster client for management cluster")
	}

	mcObject := &capi.Cluster{}
	err = regionalClusterClient.GetResource(mcObject, clusterName, "tkg-system", nil, nil)
	if err != nil {
		return false, errors.Wrapf(err, "unable to get the cluster object for cluster %q", clusterName)
	}
	_, exists := mcObject.Annotations[clusterclient.TKGVersionKey]
	return !exists, nil
}

func (c *TkgClient) getCleanupClusterOptions(regionalClusterClient clusterclient.Client, cleanupClusterKubeconfigPath string) (InitRegionOptions, error) {
	// Get installed providers list and configuration
	initOptionsForCleanupCluster, err := c.getInitOptionsFromExistingCluster(regionalClusterClient)
	if err != nil {
		return initOptionsForCleanupCluster, errors.Wrap(err, "unable to get management cluster provider information")
	}
	initOptionsForCleanupCluster.Kubeconfig = cleanupClusterKubeconfigPath
	return initOptionsForCleanupCluster, nil
}

func (c *TkgClient) deleteCluster(kubeconfig, clusterName, clusterNamespace string) error {
	var err error
	clusterClient, err := clusterclient.NewClient(kubeconfig, "", clusterclient.Options{})
	if err != nil {
		return err
	}

	clusterObject := &capi.Cluster{}
	clusterObject.Name = clusterName
	clusterObject.Namespace = clusterNamespace

	return clusterClient.DeleteResource(clusterObject)
}

func (c *TkgClient) waitForClusterDeletion(kubeconfig, clusterName, clusterNamespace string) error {
	var err error
	clusterClient, err := clusterclient.NewClient(kubeconfig, "", clusterclient.Options{OperationTimeout: c.timeout})
	if err != nil {
		return errors.Wrap(err, "unable to get cluster client while waiting for deletion of cluster")
	}
	return clusterClient.WaitForClusterDeletion(clusterName, clusterNamespace)
}

func (c *TkgClient) verifyDeleteRegion(clusterClient clusterclient.Client, options DeleteRegionOptions) (string, error) {
	var errMsg string
	var regionalClusterNamespace string

	isPacific, err := clusterClient.IsPacificRegionalCluster()
	if err != nil {
		return "", errors.Wrap(err, "error determining 'Tanzu Kubernetes Cluster service for vSphere' management cluster")
	}
	if isPacific {
		return "", errors.New("deleting 'Tanzu Kubernetes Cluster service for vSphere' management cluster is not yet supported")
	}

	// Get the all the cluster objects
	clusters := &capi.ClusterList{}
	err = clusterClient.ListResources(clusters, &crtclient.ListOptions{})
	if err != nil {
		return regionalClusterNamespace, errors.Errorf(ErrorGettingClusterObjects, err.Error())
	}

	if len(clusters.Items) == 0 {
		return regionalClusterNamespace, errors.New(ErrorNoClusterObject)
	}
	regionalClusterObjectPresent := false
	var workloadClusters string
	for i := range clusters.Items {
		if clusters.Items[i].Name == options.ClusterName {
			regionalClusterObjectPresent = true
			regionalClusterNamespace = clusters.Items[i].Namespace
		} else {
			workloadClusters += "- " + clusters.Items[i].Name + "\n"
		}
	}

	if !regionalClusterObjectPresent {
		return regionalClusterNamespace, errors.New(ErrorMissingRegionalClusterObject)
	}

	if len(workloadClusters) > 0 {
		errMsg = fmt.Sprintf(ErrorDeleteAbort, workloadClusters)
	}

	if !options.Force && errMsg != "" {
		return regionalClusterNamespace, errors.New(errMsg)
	}

	return regionalClusterNamespace, nil
}

func (c *TkgClient) getInitOptionsFromExistingCluster(regionalClusterClient clusterclient.Client) (InitRegionOptions, error) {
	initOptions := InitRegionOptions{}

	// Get the all the installed provider info
	installedProviders := &clusterctlv1.ProviderList{}
	err := regionalClusterClient.ListResources(installedProviders, &crtclient.ListOptions{})
	if err != nil {
		return initOptions, errors.Wrap(err, "cannot get installed provider config")
	}

	// installedProviderNamespaces is a map to store installed providers namespaces
	installedProviderNamespaces := make(map[string]struct{})
	installedProviderWatchingNamespaces := make(map[string]struct{})

	for i := range installedProviders.Items {
		installedProviderNamespaces[installedProviders.Items[i].Namespace] = struct{}{}

		if len(installedProviders.Items[i].WatchedNamespace) > 0 {
			installedProviderWatchingNamespaces[installedProviders.Items[i].WatchedNamespace] = struct{}{}
		}

		if clusterctlv1.ProviderType(installedProviders.Items[i].Type) == clusterctlv1.InfrastructureProviderType {
			defaultProviderVersion, err := c.tkgConfigUpdaterClient.GetDefaultInfrastructureVersion(installedProviders.Items[i].ProviderName)
			if err != nil {
				return initOptions, errors.Wrap(err, "cannot get default infrastructure provider version from config file")
			}
			if err = c.verifyProviderConfigVariablesExists(installedProviders.Items[i].ProviderName); err != nil {
				return initOptions, errors.Wrap(err, "error verifying config variables")
			}
			initOptions.InfrastructureProvider = installedProviders.Items[i].ProviderName + ":" + defaultProviderVersion
		}
	}

	initOptions.CoreProvider, initOptions.BootstrapProvider, initOptions.ControlPlaneProvider, err = c.tkgBomClient.GetDefaultClusterAPIProviders()
	if err != nil {
		return initOptions, errors.Wrap(err, "unable to get default cluster api provider info")
	}

	// Check if installed Providers namespace is same or not, if same means cluster was
	// deployment with --target-namespace parameter, else each provider is using
	// it's default namespace in this case leave targetNamespace empty
	if len(installedProviderNamespaces) != 1 {
		initOptions.Namespace = ""
	} else {
		for targetNamespace := range installedProviderNamespaces {
			initOptions.Namespace = targetNamespace
		}
	}

	return initOptions, nil
}

func (c *TkgClient) displayHelpTextOnDeleteRegionFailure(cleanupClusterKubeconfigPath string,
	isCleanupClusterCreated bool, cleanupClusterName string, clusterName string) {

	log.Warningf("\n\nFailure while deleting management cluster. Here are some steps to investigate the cause:\n")
	log.Warningf("\nDebug:")
	log.Warningf("    kubectl get po,deploy,cluster,kubeadmcontrolplane,machine,machinedeployment -A --kubeconfig %s", cleanupClusterKubeconfigPath)
	log.Warningf("    kubectl logs deployment.apps/<deployment-name> -n <deployment-namespace> manager --kubeconfig %s", cleanupClusterKubeconfigPath)

	if isCleanupClusterCreated {
		log.Warningf("\nThe cleanup cluster is still running")
		log.Warningf("\nRun docker ps | grep %s to verify that the cleanup cluster is still running", cleanupClusterName)
		log.Warningf("\nTo delete the cleanup cluster and free up resources locally:")
		log.Warningf("    docker rm -v %s-control-plane -f", cleanupClusterName)
	}

	log.Warningf("\nIf you want to remove the management-cluster entry from the existing server list, run 'tanzu config server delete %s'", clusterName)
}

func (c *TkgClient) verifyProviderConfigVariablesExists(providerName string) error {
	var missingVariables []string
	if providerName == VSphereProviderName {
		if _, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereUsername); err != nil {
			missingVariables = append(missingVariables, constants.ConfigVariableVsphereUsername)
		}
		if _, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableVspherePassword); err != nil {
			missingVariables = append(missingVariables, constants.ConfigVariableVspherePassword)
		}
	} else if providerName == AWSProviderName {
		if _, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSB64Credentials); err != nil {
			missingVariables = append(missingVariables, constants.ConfigVariableAWSB64Credentials)
			_, errEncode := c.EncodeAWSCredentialsAndGetClient(nil)
			// EncodeAWSCredentialsAndGetClient function may set ConfigVariableAWSB64Credentials in configStore
			// So, verify the existence of ConfigVariableAWSB64Credentials again
			_, errReadVar := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSB64Credentials)
			if errEncode != nil || errReadVar != nil {
				missingVariables = append(missingVariables, constants.ConfigVariableAWSB64Credentials)
			}
		}
	}

	if len(missingVariables) == 0 {
		return nil
	}

	return errors.Errorf("value for variables [%s] is not set. Please set the value using os environment variables or the tkg config file", strings.Join(missingVariables, ","))
}

func (c *TkgClient) cleanUpAVIResourcesInManagementCluster(regionalClusterClient clusterclient.Client, clusterName, clusterNamespace string) error {
	akoAddonSecret := &corev1.Secret{}
	if err := regionalClusterClient.GetResource(akoAddonSecret, constants.AkoAddonName+"-data-values", clusterNamespace, nil, nil); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return errors.Wrapf(err, "unable to get ako add-on secret")
	}
	log.Info("Cleaning up AVI Resources...")
	akoAddonSecretData := akoAddonSecret.Data["values.yaml"]
	var values map[string]interface{}
	err := yaml.Unmarshal(akoAddonSecretData, &values)
	if err != nil {
		return err
	}
	akoInfo, ok := values["loadBalancerAndIngressService"].(map[string]interface{})
	if !ok {
		return errors.Errorf("management cluster %s ako add-on secret yaml data parse error", clusterName)
	}
	akoConfig, ok := akoInfo["config"].(map[string]interface{})
	if !ok {
		return errors.Errorf("management cluster %s ako add-on secret yaml data parse error", clusterName)
	}
	akoSetting, ok := akoConfig["ako_settings"].(map[string]interface{})
	if !ok {
		return errors.Errorf("management cluster %s ako add-on secret yaml data parse error", clusterName)
	}
	akoSetting["delete_config"] = trueStr
	akoAddonSecretData, err = yaml.Marshal(&values)
	if err != nil {
		return err
	}
	akoAddonSecret.Data["values.yaml"] = []byte(constants.TKGDataValueFormatString + string(akoAddonSecretData))
	if err := regionalClusterClient.UpdateResource(akoAddonSecret, constants.AkoAddonName+"-data-values", clusterNamespace); err != nil {
		return errors.Wrapf(err, "unable to update ako add-on secret")
	}
	return c.waitForAVIResourceCleanup(regionalClusterClient)
}

func (c *TkgClient) waitForAVIResourceCleanup(regionalClusterClient clusterclient.Client) error {
	akoStatefulSet := &v1.StatefulSet{}
	if err := regionalClusterClient.GetResource(akoStatefulSet, constants.AkoStatefulSetName, constants.AkoNamespace, nil, nil); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		log.Warning("unable to get ako statefulset")
	}
	if err := regionalClusterClient.WaitForAVIResourceCleanUp(constants.AkoStatefulSetName, constants.AkoNamespace); err != nil {
		if !strings.Contains(err.Error(), "dial tcp") {
			log.Error(err, "clean up AVI resources error")
		}
	}
	return nil
}
