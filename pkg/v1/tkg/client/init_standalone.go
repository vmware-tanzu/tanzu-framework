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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

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

func (c *TkgClient) InitStandaloneRegion(options *InitRegionOptions) error { //nolint:gocyclo
	var err error
	var regionalConfigBytes []byte
	var isSuccessful bool = false
	var isStartedRegionalClusterCreation bool = false
	var isBootstrapClusterCreated bool = false
	var bootstrapClusterName string
	var regionContext region.RegionContext

	bootstrapClusterKubeconfigPath, err := getTKGKubeConfigPath(false)
	if err != nil {
		return err
	}
	if options.TmcRegistrationURL != "" {
		InitRegionSteps = append(InitRegionSteps, StepRegisterWithTMC)
	}
	log.SendProgressUpdate(statusRunning, StepValidateConfiguration, InitRegionSteps)
	log.Info("Validating configuration...")

	// at exit, do these things
	defer func() {
		if regionContext != (region.RegionContext{}) {
			err := c.regionManager.SaveRegionContext(regionContext)
			if err != nil {
				log.Warningf("Unable to persist standalone cluster %s info to tkg config", regionContext.ClusterName)
			}

			err = c.regionManager.SetCurrentContext(regionContext.ClusterName, regionContext.ContextName)
			if err != nil {
				log.Warningf("Unable to use context %s as current tkg context", regionContext.ContextName)
			}
		}

		if isSuccessful {
			log.SendProgressUpdate(statusSuccessful, "", InitRegionSteps)
		} else {
			log.SendProgressUpdate(statusFailed, "", InitRegionSteps)
		}

		// if regional cluster creation failed after bootstrap kind cluster was successfully created
		if !isSuccessful && isStartedRegionalClusterCreation {
			c.displayHelpTextOnFailure(options, isBootstrapClusterCreated, bootstrapClusterKubeconfigPath)
			return
		}

		if isBootstrapClusterCreated {
			if err := c.teardownKindCluster(bootstrapClusterName, bootstrapClusterKubeconfigPath, options.UseExistingCluster); err != nil {
				log.Warning(err.Error())
			}
		}
		_ = utils.DeleteFile(bootstrapClusterKubeconfigPath)

	}()

	if customImageRepo, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableCustomImageRepository); err != nil && customImageRepo != "" && tkgconfighelper.IsCustomRepository(customImageRepo) {
		log.Infof("Using custom image repository: %s", customImageRepo)
	}

	// validate docker only if user is not using an existing cluster
	// Note: Validating in client code as well to cover the usecase where users use client code instead of command line.
	if err := c.ValidatePrerequisites(!options.UseExistingCluster, true); err != nil {
		return err
	}

	log.Infof("Using infrastructure provider %s", options.InfrastructureProvider)
	log.SendProgressUpdate(statusRunning, StepGenerateClusterConfiguration, InitRegionSteps)
	log.Info("Generating cluster configuration...")

	// Obtain regional cluster configuration of a provided flavor
	if regionalConfigBytes, options.ClusterName, err = c.BuildRegionalClusterConfiguration(options); err != nil {
		return errors.Wrap(err, "unable to build standalone cluster configuration")
	}

	log.SendProgressUpdate(statusRunning, StepSetupBootstrapCluster, InitRegionSteps)
	log.Info("Setting up bootstrapper...")
	// Ensure bootstrap cluster and copy boostrap cluster kubeconfig to ~/kube-tkg directory
	if bootstrapClusterName, err = c.ensureKindCluster(options.Kubeconfig, options.UseExistingCluster, bootstrapClusterKubeconfigPath); err != nil {
		return errors.Wrap(err, "unable to create bootstrap cluster")
	}

	isBootstrapClusterCreated = true
	log.Infof("Bootstrapper created. Kubeconfig: %s", bootstrapClusterKubeconfigPath)
	bootStrapClusterClient, err := clusterclient.NewClient(bootstrapClusterKubeconfigPath, "", clusterclient.Options{OperationTimeout: c.timeout})
	if err != nil {
		return errors.Wrap(err, "unable to get bootstrap cluster client")
	}

	log.SendProgressUpdate(statusRunning, StepInstallProvidersOnBootstrapCluster, InitRegionSteps)
	log.Info("Installing providers on bootstrapper...")
	// Initialize bootstrap cluster with providers

	if err = c.InitializeProviders(options, bootStrapClusterClient, bootstrapClusterKubeconfigPath); err != nil {
		return errors.Wrap(err, "unable to initialize providers")
	}
	err = SaveInitOptions(options)
	if err != nil {
		return errors.Wrap(err, "unable to save init options")
	}

	isStartedRegionalClusterCreation = true

	targetClusterNamespace := defaultTkgNamespace
	// if options.Namespace != "" {
	// 	targetClusterNamespace = options.Namespace
	// }

	log.SendProgressUpdate(statusRunning, StepCreateManagementCluster, InitRegionSteps)
	log.Info("Start creating standalone cluster...")
	err = c.DoCreateCluster(bootStrapClusterClient, options.ClusterName, targetClusterNamespace, string(regionalConfigBytes))
	if err != nil {
		return errors.Wrap(err, "unable to create standalone cluster")
	}

	// save this context to tkg config incase the standalone cluster creation fails
	regionContext = region.RegionContext{ClusterName: options.ClusterName, ContextName: "kind-" + bootstrapClusterName, SourceFilePath: bootstrapClusterKubeconfigPath, Status: region.Failed}
	fmt.Printf("regionContext:\n%v", regionContext)

	kubeConfigBytes, err := c.WaitForClusterInitializedAndGetKubeConfig(bootStrapClusterClient, options.ClusterName, targetClusterNamespace)
	if err != nil {
		return errors.Wrap(err, "unable to wait for cluster and get the cluster kubeconfig")
	}

	regionalClusterKubeconfigPath, err := getTKGKubeConfigPath(true)
	if err != nil {
		return err
	}

	mergeFile := getDefaultKubeConfigFile()
	log.Infof("Saving standalone cluster kubeconfig into %s", mergeFile)
	// merge the standalone cluster kubeconfig into user input kubeconfig path/default kubeconfig path
	err = MergeKubeConfigWithoutSwitchContext(kubeConfigBytes, mergeFile)
	if err != nil {
		return errors.Wrap(err, "unable to merge standalone cluster kubeconfig")
	}

	// merge the standalone cluster kubeconfig into tkg managed kubeconfig
	kubeContext, err := MergeKubeConfigAndSwitchContext(kubeConfigBytes, regionalClusterKubeconfigPath)
	if err != nil {
		return errors.Wrap(err, "unable to save standalone cluster kubeconfig to TKG managed kubeconfig")
	}

	log.Info("Waiting for bootstrap cluster to get ready for save ...")
	if err := c.WaitForClusterReadyForMove(bootStrapClusterClient, options.ClusterName, targetClusterNamespace); err != nil {
		return errors.Wrap(err, "unable to wait for cluster getting ready for move")
	}

	regionalClusterClient, err := clusterclient.NewClient(regionalClusterKubeconfigPath, kubeContext, clusterclient.Options{OperationTimeout: c.timeout})
	if err != nil {
		return errors.Wrap(err, "unable to get management cluster client")
	}

	log.Info("Waiting for addons installation...")
	if err := c.WaitForAddons(waitForAddonsOptions{
		regionalClusterClient: bootStrapClusterClient,
		workloadClusterClient: regionalClusterClient,
		clusterName:           options.ClusterName,
		namespace:             options.Namespace,
		waitForCNI:            true,
	}); err != nil {
		return errors.Wrap(err, "error waiting for addons to get installed")
	}

	log.SendProgressUpdate(statusRunning, StepMoveClusterAPIObjects, InitRegionSteps)
	log.Info("Moving all Cluster API objects from bootstrap cluster to management cluster...")
	// Move all Cluster API objects from bootstrap cluster to created to regional cluster for all namespaces
	if err = c.SaveObjects(bootstrapClusterKubeconfigPath, targetClusterNamespace); err != nil {
		return errors.Wrap(err, "unable to move Cluster API objects from bootstrap cluster to management cluster")
	}

	regionContext = region.RegionContext{ClusterName: options.ClusterName, ContextName: kubeContext, SourceFilePath: regionalClusterKubeconfigPath, Status: region.Success}
	log.Infof("Context set for standalone cluster %s as '%s'.", options.ClusterName, kubeContext)
	isSuccessful = true
	return nil
}

// SaveObjects saves all the Cluster API objects from all the namespaces to files
func (c *TkgClient) SaveObjects(fromKubeconfigPath, namespace string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	directoryBin := filepath.Join(homeDir, ".tanzu", "tce", "objects")
	err = os.MkdirAll(directoryBin, 0755)
	if err != nil {
		return err
	}

	saveOptions := clusterctl.SaveOptions{
		FromKubeconfig:    clusterctl.Kubeconfig{Path: fromKubeconfigPath},
		Namespace:         namespace,
		DirectoryLocation: directoryBin,
	}

	return c.clusterctlClient.Save(saveOptions)
}

func SaveInitOptions(options *InitRegionOptions) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	directoryBin := filepath.Join(homeDir, ".tanzu", "tce", "init")
	err = os.MkdirAll(directoryBin, 0755)
	if err != nil {
		return err
	}

	byRaw, err := yaml.Marshal(options)
	if err != nil {
		return err
	}

	initFile := filepath.Join(directoryBin, options.ClusterName)
	err = ioutil.WriteFile(initFile, byRaw, 0644)
	if err != nil {
		return err
	}

	return nil
}
