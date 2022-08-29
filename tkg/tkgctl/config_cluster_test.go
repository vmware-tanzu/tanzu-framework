// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes"
	fakeproviders "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes/providers"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigpaths"
	"github.com/vmware-tanzu/tanzu-framework/tkg/utils"
)

const (
	defaultTKGBomFileForTesting = "../fakes/config/bom/tkg-bom-v1.3.1.yaml"
	defaultTKRBomFileForTesting = "../fakes/config/bom/tkr-bom-v1.18.0+vmware.1-tkg.2.yaml"
)

var _ = Describe("Unit tests for config cluster", func() {
	var (
		ctl           *tkgctl
		tkgClient     = &fakes.Client{}
		updaterClient = &fakes.TKGConfigUpdaterClient{}
		bomClient     = &fakes.TKGConfigBomClient{}
		configDir     string
		err           error
		ccOps         CreateClusterOptions
	)

	JustBeforeEach(func() {
		configDir, err = os.MkdirTemp("", "test")
		err = os.MkdirAll(testingDir, 0o700)
		Expect(err).ToNot(HaveOccurred())
		prepareConfiDir(configDir)
		options := Options{
			ConfigDir:      configDir,
			ProviderGetter: fakeproviders.FakeProviderGetter(),
		}
		c, createErr := New(options)
		Expect(createErr).ToNot(HaveOccurred())
		ctl, _ = c.(*tkgctl)
		ctl.tkgClient = tkgClient
		ctl.tkgBomClient = bomClient
		ctl.tkgConfigUpdaterClient = updaterClient

		err = ctl.ConfigCluster(ccOps)
	})
	Context("when cluster name is not provided", func() {
		BeforeEach(func() {
			ccOps = CreateClusterOptions{}
			updaterClient.DecodeCredentialsInViperReturns(nil)
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})
	Context("When plan is not provided", func() {
		BeforeEach(func() {
			ccOps = CreateClusterOptions{
				ClusterName: "my-cluster",
			}
			updaterClient.DecodeCredentialsInViperReturns(nil)
		})

		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})

	Context("When creating a Pacific workload cluster", func() {
		BeforeEach(func() {
			ccOps = CreateClusterOptions{
				ClusterName:            "my-cluster",
				Plan:                   "dev",
				InfrastructureProvider: "tkg-service-vsphere",
				TkrVersion:             "1.19.0+vmware.1.tkg.1",
			}
			updaterClient.DecodeCredentialsInViperReturns(nil)
			tkgClient.IsPacificManagementClusterReturns(true, errors.New("unknown"))
			tkgClient.GetClusterConfigurationReturns(nil, nil)
		})

		It("should not return an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("When creating a tkgm workload cluster", func() {
		BeforeEach(func() {
			ccOps = CreateClusterOptions{
				ClusterName: "my-cluster",
				Plan:        "dev",
				TkrVersion:  "1.19.0+vmware.1.tkg.1",
			}
			updaterClient.DecodeCredentialsInViperReturns(nil)
			tkgClient.IsPacificManagementClusterReturns(false, nil)
			bomClient.GetBOMConfigurationFromTkrVersionReturns(nil, nil)
			bomClient.GetK8sVersionFromTkrVersionReturns("1.19.0", nil)
			tkgClient.GetClusterConfigurationReturns(nil, nil)
		})

		It("should not return an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})
	})

	AfterEach(func() {
		os.Remove(configDir)
	})
})

var testTKGCompatibilityFileFmt = `
version: v1
managementClusterPluginVersions:
- version: %s
  supportedTKGBomVersions:
  - imagePath: tkg-bom
    tag: %s
`

func prepareConfiDir(configDir string) {
	bomDir, err := tkgconfigpaths.New(configDir).GetTKGBoMDirectory()
	Expect(err).ToNot(HaveOccurred())
	if _, err := os.Stat(bomDir); os.IsNotExist(err) {
		err = os.MkdirAll(bomDir, 0o700)
		Expect(err).ToNot(HaveOccurred())
	}

	err = utils.CopyFile(defaultTKGBomFileForTesting, filepath.Join(bomDir, filepath.Base(defaultTKGBomFileForTesting)))
	Expect(err).ToNot(HaveOccurred())

	err = utils.CopyFile(defaultTKRBomFileForTesting, filepath.Join(bomDir, filepath.Base(defaultTKRBomFileForTesting)))
	Expect(err).ToNot(HaveOccurred())

	compatibilityDir, err := tkgconfigpaths.New(configDir).GetTKGCompatibilityDirectory()
	Expect(err).ToNot(HaveOccurred())
	if _, err := os.Stat(compatibilityDir); os.IsNotExist(err) {
		err = os.MkdirAll(compatibilityDir, 0o700)
		Expect(err).ToNot(HaveOccurred())
	}

	defaultBomFileTag := utils.GetTKGBoMTagFromFileName(filepath.Base(defaultTKGBomFileForTesting))
	testTKGCompatabilityFileContent := fmt.Sprintf(testTKGCompatibilityFileFmt, tkgconfigpaths.TKGManagementClusterPluginVersion, defaultBomFileTag)

	compatibilityConfigFile, err := tkgconfigpaths.New(configDir).GetTKGCompatibilityConfigPath()
	Expect(err).ToNot(HaveOccurred())
	err = os.WriteFile(compatibilityConfigFile, []byte(testTKGCompatabilityFileContent), constants.ConfigFilePermissions)
	Expect(err).ToNot(HaveOccurred())
}
