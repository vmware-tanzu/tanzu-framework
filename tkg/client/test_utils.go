// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/onsi/gomega"

	"github.com/vmware-tanzu/tanzu-framework/tkg/clientcreator"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/region"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigpaths"
	"github.com/vmware-tanzu/tanzu-framework/tkg/types"
	"github.com/vmware-tanzu/tanzu-framework/tkg/utils"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

var testTKGCompatibilityFileFmt = `
version: v1
managementClusterPluginVersions:
- version: %s
  supportedTKGBomVersions:
  - imagePath: tkg-bom
    tag: %s
`

func copyFile(sourceFile, destFile string) {
	input, _ := os.ReadFile(sourceFile)
	_ = os.WriteFile(destFile, input, constants.ConfigFilePermissions)
}

func createTKGClient(clusterConfigFile, configDir, defaultBomFile string, timeout time.Duration) (*TkgClient, error) {
	return createTKGClientOpts(clusterConfigFile, configDir, defaultBomFile, timeout, func(options Options) Options { return options })
}

func CreateTKGClientOptsMutator(clusterConfigFile, configDir, defaultBomFile string, timeout time.Duration, optMutator func(options Options) Options) (*TkgClient, error) {
	return createTKGClientOpts(clusterConfigFile, configDir, defaultBomFile, timeout, optMutator)
}

func createTKGClientOpts(clusterConfigFile, configDir, defaultBomFile string, timeout time.Duration, optMutator func(options Options) Options) (*TkgClient, error) {
	setupTestingFiles(clusterConfigFile, configDir, defaultBomFile)
	appConfig := types.AppConfig{
		TKGConfigDir: configDir,
		CustomizerOptions: types.CustomizerOptions{
			RegionManagerFactory: region.NewFactory(),
		},
	}
	allClients, err := clientcreator.CreateAllClients(appConfig, nil)
	if err != nil {
		return nil, err
	}

	return New(optMutator(Options{
		ClusterCtlClient:         allClients.ClusterCtlClient,
		ReaderWriterConfigClient: allClients.ConfigClient,
		RegionManager:            allClients.RegionManager,
		TKGConfigDir:             configDir,
		Timeout:                  timeout,
		FeaturesClient:           allClients.FeaturesClient,
		TKGConfigProvidersClient: allClients.TKGConfigProvidersClient,
		TKGBomClient:             allClients.TKGBomClient,
		TKGConfigUpdater:         allClients.TKGConfigUpdaterClient,
		TKGPathsClient:           allClients.TKGConfigPathsClient,
		FeatureFlagClient:        &configapi.ClientConfig{},
	}))
}

func setupTestingFiles(clusterConfigFile, configDir, defaultBomFile string) {
	testClusterConfigFile := filepath.Join(configDir, "config.yaml")
	err := utils.CopyFile(clusterConfigFile, testClusterConfigFile)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	bomDir, err := tkgconfigpaths.New(configDir).GetTKGBoMDirectory()
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	if _, err := os.Stat(bomDir); os.IsNotExist(err) {
		err = os.MkdirAll(bomDir, 0o700)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	}
	err = utils.CopyFile(defaultBomFile, filepath.Join(bomDir, filepath.Base(defaultBomFile)))
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	compatibilityDir, err := tkgconfigpaths.New(configDir).GetTKGCompatibilityDirectory()
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	if _, err := os.Stat(compatibilityDir); os.IsNotExist(err) {
		err = os.MkdirAll(compatibilityDir, 0o700)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	}

	defaultBomFileTag := utils.GetTKGBoMTagFromFileName(filepath.Base(defaultBomFile))
	testTKGCompatabilityFileContent := fmt.Sprintf(testTKGCompatibilityFileFmt, tkgconfigpaths.TKGManagementClusterPluginVersion, defaultBomFileTag)

	compatibilityConfigFile, err := tkgconfigpaths.New(configDir).GetTKGCompatibilityConfigPath()
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	err = os.WriteFile(compatibilityConfigFile, []byte(testTKGCompatabilityFileContent), constants.ConfigFilePermissions)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
}
