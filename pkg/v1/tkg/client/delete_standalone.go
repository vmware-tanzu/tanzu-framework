/*
Copyright 2020 The TKG Contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package client

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	yaml "github.com/ghodss/yaml"
	"github.com/pkg/errors"

	clusterctl "sigs.k8s.io/cluster-api/cmd/clusterctl/client"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/region"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfighelper"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/utils"
)

// const (
// 	ErrorMissingRegionalClusterObject = "management cluster object is not present in given management cluster"
// 	ErrorNoClusterObject              = "no Cluster object present in the given management cluster"
// 	ErrorGettingClusterObjects        = "unable to get cluster resources, %s. Are you sure the cluster you are using is a management cluster?"
// 	ErrorDeleteAbort                  = `Deletion is aborted because management cluster is currently managing the following workload clusters:

// %s

// You need to delete these clusters first before deleting the management cluster.

// Alternatively, you can use the -f/--force flag to force the deletion of the management cluster but doing so will orphan the above-mentioned clusters and leave them unmanaged.`
// )

func (c *TkgClient) DeleteStandalone(options DeleteRegionOptions) error {
	var err error
	var isSuccessful bool = false
	var isStartedRegionalClusterDeletion bool = false
	var isCleanupClusterCreated bool = false
	var cleanupClusterName string
	var cleanupClusterKubeconfigPath string

	defer func() {
		// if regional cluster deletion is not being started and kind cluster is already created
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

	if customImageRepo, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableCustomImageRepository); err != nil && customImageRepo != "" && tkgconfighelper.IsCustomRepository(customImageRepo) {
		log.Infof("Using custom image repository: %s", customImageRepo)
	}

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
		return errors.Errorf("standalone cluster %s not found", options.ClusterName)
	}
	regionContext := contexts[0]

	regionalClusterNamespace := defaultTkgNamespace
	// if options.Namespace != "" {
	// 	regionalClusterNamespace = options.Namespace
	// }

	cleanupClusterKubeconfigPath, err = getTKGKubeConfigPath(false)
	if err != nil {
		return errors.Wrap(err, "cannot get cleanup cluster kubeconfig path ")
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

	log.Info("Installing providers to cleanup cluster...")
	initOptionsForCleanupCluster, err := RestoreInitOptions(options.ClusterName)
	if err != nil {
		return errors.Wrap(err, "unable to restore init options")
	}
	// set the cluster config for deletion. This enables recieveing credentials for vsphere, AWS,
	// and azure.
	initOptionsForCleanupCluster.ClusterConfigFile = options.ClusterConfig

	// Initialize cleanup cluster using same provider name and version from regional cluster
	if err = c.InitializeProviders(initOptionsForCleanupCluster, cleanupClusterClient, cleanupClusterKubeconfigPath); err != nil {
		return errors.Wrap(err, "unable to initialize providers")
	}

	isStartedRegionalClusterDeletion = true

	// Move all Cluster API objects from files to cleanup cluster for all namespaces
	log.Info("Moving all Cluster API objects from bootstrap cluster to management cluster...")
	if err = c.RestoreObjects(cleanupClusterKubeconfigPath, regionalClusterNamespace, options.ClusterName); err != nil {
		return errors.Wrap(err, "unable to move Cluster API objects from bootstrap cluster to management cluster")
	}

	log.Info("Waiting for the Cluster API objects to be ready after restore ...")
	if err := c.WaitForClusterReadyAfterReverseMove(cleanupClusterClient, options.ClusterName, regionalClusterNamespace); err != nil {
		return errors.Wrap(err, "unable to wait for cluster getting ready for move")
	}

	if err = cleanupClusterClient.DeleteCluster(options.ClusterName, regionalClusterNamespace); err != nil {
		return errors.Wrap(err, "unabe to delete standalone using cleanupClusterClient DeleteCluster")
	}

	// Regional cluster deletion happens in background and we cannot teardown the cleanup kind cluster until the regional cluster is deleted successfully
	log.Info("Deleting standalone cluster...")
	if err = c.waitForClusterDeletion(cleanupClusterKubeconfigPath, options.ClusterName, regionalClusterNamespace); err != nil {
		return errors.Wrapf(err, "error waiting for standalone cluster '%s' to be deleted", options.ClusterName)
	}

	err = c.regionManager.DeleteRegionContext(options.ClusterName)
	if err != nil {
		log.Warningf("Failed to delete standalone cluster %s context from tkg config file", options.ClusterName)
	}

	log.Infof("Standalone cluster '%s' deleted.", options.ClusterName)

	// delete standalone cluster config from default kubeconfig file
	if regionContext.Status != region.Failed {
		userKubeconfigPath := getDefaultKubeConfigFile()
		log.Infof("Deleting the standalone cluster context from the kubeconfig file '%s'", userKubeconfigPath)
		if err = DeleteContextFromKubeConfig(userKubeconfigPath, regionContext.ContextName); err != nil {
			log.Warningf("Failed to delete standalone cluster context from the kubeconfig file '%s'", userKubeconfigPath)
		}
	}

	isSuccessful = true

	return nil
}

// RestoreObjects restores all the Cluster API objects from all the namespaces to files
func (c *TkgClient) RestoreObjects(toKubeconfigPath string, namespace string, standaloneName string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	directoryBin := filepath.Join(homeDir, ".tanzu", "tce", "objects")
	err = os.MkdirAll(directoryBin, 0755)
	if err != nil {
		return err
	}

	restoreOptions := clusterctl.RestoreOptions{
		ToKubeconfig:      clusterctl.Kubeconfig{Path: toKubeconfigPath},
		Namespace:         namespace,
		Glob:              standaloneName,
		DirectoryLocation: directoryBin,
	}

	return c.clusterctlClient.Restore(restoreOptions)
}

func RestoreInitOptions(clusterName string) (*InitRegionOptions, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	initFile := filepath.Join(homeDir, ".tanzu", "tce", "init", clusterName)
	byObj, err := ioutil.ReadFile(initFile)
	if err != nil {
		return nil, err
	}

	options := &InitRegionOptions{}
	err = yaml.Unmarshal(byObj, options)
	if err != nil {
		return nil, err
	}

	return options, nil
}
