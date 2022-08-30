// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgconfigbom_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/clientconfighelpers"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/fakes"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/registry"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigbom"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigpaths"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigreaderwriter"
	"github.com/vmware-tanzu/tanzu-framework/tkg/utils"
)

var (
	testingDir string
	_          = Describe("Unit tests for BOM client", func() {
		var (
			bomClient                   tkgconfigbom.Client
			fakeRegistry                *fakes.Registry
			tkgConfigDir                string
			clusterConfigFile           string
			kubeconfig7Path             = "../fakes/config/config7.yaml"
			defaultTKGBoMFileForTesting = "../fakes/config/bom/tkg-bom-v1.3.1.yaml"
			tkgConfigReaderWriter       tkgconfigreaderwriter.TKGConfigReaderWriter
		)

		JustBeforeEach(func() {
			var err error
			setupTestingFiles(clusterConfigFile, tkgConfigDir, defaultTKGBoMFileForTesting)
			tkgConfigReaderWriter, err = tkgconfigreaderwriter.NewReaderWriterFromConfigFile(clusterConfigFile, filepath.Join(tkgConfigDir, "config.yaml"))
			Expect(err).NotTo(HaveOccurred())

			bomClient = tkgconfigbom.New(tkgConfigDir, tkgConfigReaderWriter)
		})

		Describe("When upgrading cluster with autoscaler enabled", func() {
			BeforeEach(func() {
				createTempDirectory()
				tkgConfigDir = testingDir
				clusterConfigFile = "../fakes/config/config.yaml"
			})
			AfterEach(func() {
				deleteTempDirectory()
			})
			Context("when multiple autoscaler image tags are present for a k8s minor version", func() {
				It("returns an error", func() {
					_, err := bomClient.GetAutoscalerImageForK8sVersion("v1.18.0+vmware.1")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("expected one autoscaler image for kubernetes minor version \"v1.18\" but found 2"))
				})
			})

			Context("For k8s version v1.19.0+vmware.1", func() {
				It("returns a valid autoscaler image", func() {
					image, err := bomClient.GetAutoscalerImageForK8sVersion("v1.19.0+vmware.1")
					Expect(err).NotTo(HaveOccurred())
					Expect(image).To(Equal("projects-stg.registry.vmware.com/tkg/cluster-autoscaler:v1.19.1_vmware.1"))
				})
			})

			Context("For unsupported k8s version", func() {
				It("returns an error", func() {
					_, err := bomClient.GetAutoscalerImageForK8sVersion("v1.16.0+vmware.1")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("autoscaler image not available for kubernetes minor version"))
				})
			})
		})

		Describe("When downloading the default BOM files from registry", func() {
			BeforeEach(func() {
				createTempDirectory()
				tkgConfigDir = testingDir
				f, _ := os.CreateTemp(testingDir, "config.yaml")
				err := utils.CopyFile("../fakes/config/config.yaml", f.Name())
				Expect(err).ToNot(HaveOccurred())
				err = utils.CopyFile("../fakes/config/config.yaml", filepath.Join(tkgConfigDir, "config.yaml"))
				Expect(err).ToNot(HaveOccurred())

				clusterConfigFile = f.Name()
				fakeRegistry = &fakes.Registry{}
			})
			AfterEach(func() {
				deleteTempDirectory()
			})
			Context("when downloading the TKG BOM file fails", func() {
				BeforeEach(func() {
					fakeRegistry.GetFileReturns(nil, errors.New("fake GetFile error for TKG BOM file"))
				})
				It("returns an error", func() {
					err := bomClient.DownloadDefaultBOMFilesFromRegistry("", fakeRegistry)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("failed to download the BOM file from image name"))
					Expect(err.Error()).To(ContainSubstring("fake GetFile error for TKG BOM file"))
				})
			})

			Context("when downloading the TKG BOM file is success but fails to download TKr BOM file", func() {
				BeforeEach(func() {
					data, err := os.ReadFile("../fakes/config/bom/tkg-bom-v1.3.1.yaml")
					Expect(err).ToNot(HaveOccurred())
					fakeRegistry.GetFileReturnsOnCall(0, data, nil)
					fakeRegistry.GetFileReturnsOnCall(1, nil, errors.New("fake GetFile error for TKr BOM file"))
				})
				It("should return an error", func() {
					err := bomClient.DownloadDefaultBOMFilesFromRegistry("", fakeRegistry)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("failed to download the BOM file from image name"))
					Expect(err.Error()).To(ContainSubstring("fake GetFile error for TKr BOM file"))
				})
			})
			Context("when downloading the TKG BOM file and TKr BOM file is success ", func() {
				BeforeEach(func() {
					tkgdata, err := os.ReadFile("../fakes/config/bom/tkg-bom-v1.3.1.yaml")
					Expect(err).ToNot(HaveOccurred())
					tkrdata, err := os.ReadFile("../fakes/config/bom/tkg-bom-v1.3.1.yaml")
					Expect(err).ToNot(HaveOccurred())
					fakeRegistry.GetFileReturnsOnCall(0, tkgdata, nil)
					fakeRegistry.GetFileReturnsOnCall(1, tkrdata, nil)
				})
				It("should return success", func() {
					err := bomClient.DownloadDefaultBOMFilesFromRegistry("", fakeRegistry)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})
		Describe("Downloading the TKG Compatibility file from registry", func() {
			BeforeEach(func() {
				createTempDirectory()
				tkgConfigDir = testingDir
				f, _ := os.CreateTemp(testingDir, "config.yaml")
				err := utils.CopyFile("../fakes/config/config.yaml", f.Name())
				Expect(err).ToNot(HaveOccurred())
				err = utils.CopyFile("../fakes/config/config.yaml", filepath.Join(tkgConfigDir, "config.yaml"))
				Expect(err).ToNot(HaveOccurred())

				clusterConfigFile = f.Name()
				fakeRegistry = &fakes.Registry{}
			})
			AfterEach(func() {
				deleteTempDirectory()
			})
			Context("when listing Image tags from registry fails", func() {
				BeforeEach(func() {
					fakeRegistry.ListImageTagsReturns(nil, errors.New("fake ListImageTags error for TKG Compatibility Image"))
				})
				It("returns an error", func() {
					err := bomClient.DownloadTKGCompatibilityFileFromRegistry("", "", fakeRegistry)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("failed to list TKG compatibility image tags"))
					Expect(err.Error()).To(ContainSubstring("fake ListImageTags error for TKG Compatibility Image"))
				})
			})

			Context("when all the Image tags returned are in invalid format ", func() {
				BeforeEach(func() {
					fakeRegistry.ListImageTagsReturns([]string{"fake1", "fake2", "fake3"}, nil)
					// fakeRegistry.GetFileCalls(func(ImagePath string, ImageTag string, filename string) ([]byte, error) {
					// 	receivedImageTag = ImageTag
					// 	return nil, errors.New("fake GetFile error for TKG Compatibility file")
					// })
				})
				It("should return an error", func() {
					err := bomClient.DownloadTKGCompatibilityFileFromRegistry("", "", fakeRegistry)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("failed to get valid image tags for TKG compatibility image"))

				})
			})
			Context("when multiple Image tags are returned from registry but fails to download TKG Compatibility file", func() {
				var receivedImageTag string
				BeforeEach(func() {
					fakeRegistry.ListImageTagsReturns([]string{"v3", "v1", "v2"}, nil)
					fakeRegistry.GetFileCalls(func(ImageWithTag string, filename string) ([]byte, error) {
						receivedImageTag = strings.Split(ImageWithTag, ":")[1]
						return nil, errors.New("fake GetFile error for TKG Compatibility file")
					})
				})
				It("should return an error", func() {
					err := bomClient.DownloadTKGCompatibilityFileFromRegistry("", "", fakeRegistry)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("failed to download the TKG Compatibility file from image name"))
					Expect(err.Error()).To(ContainSubstring("fake GetFile error for TKG Compatibility file"))
					Expect(receivedImageTag).To(Equal("v3"))
				})
			})
			Context("when downloading the TKG Compatibility file is success ", func() {
				BeforeEach(func() {
					fakeRegistry.ListImageTagsReturns([]string{"v3", "v1", "v2"}, nil)
					testTKGCompatabilityFileContent := fmt.Sprintf(testTKGCompatibilityFileFmt, tkgconfigpaths.TKGManagementClusterPluginVersion, "v1.0.0-fakeVersion")
					fakeRegistry.GetFileReturns([]byte(testTKGCompatabilityFileContent), nil)
				})
				It("should return success", func() {
					err := bomClient.DownloadTKGCompatibilityFileFromRegistry("", "", fakeRegistry)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})
		Context("When getting BOMConfiguration from TKr version", func() {
			var (
				tkrVersion       string
				bomConfiguration *tkgconfigbom.BOMConfiguration
				err              error
			)
			JustBeforeEach(func() {
				bomConfiguration, err = bomClient.GetBOMConfigurationFromTkrVersion(tkrVersion)
			})
			Context("When tkr version is missing", func() {
				BeforeEach(func() {
					createTempDirectory()
					tkgConfigDir = testingDir
					tkrVersion = ""
				})
				AfterEach(func() {
					deleteTempDirectory()
				})
				It("Should return the bomConfiguration", func() {
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("No BOM file found with TKr version "))
				})
			})
			Context("When tkr version(v1.18.0+vmware.1-tkg.2) is found", func() {
				BeforeEach(func() {
					createTempDirectory()
					tkgConfigDir = testingDir
					setupBomFile("../fakes/config/bom/tkr-bom-v1.18.0+vmware.1-tkg.2.yaml", tkgConfigDir)
					setupBomFile("../fakes/config/bom/tkr-bom-v1.19.3+vmware.1-tkg.1.yaml", tkgConfigDir)
					tkrVersion = "v1.18.0+vmware.1-tkg.2"
				})
				AfterEach(func() {
					deleteTempDirectory()
				})
				It("Should return the BOMConfiguration", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(bomConfiguration).ToNot(BeNil())
				})
				It("Should return the correct etcd configuration", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(bomConfiguration.KubeadmConfigSpec.Etcd.Local.DataDir).To(Equal("/var/lib/etcd"))
					Expect(bomConfiguration.KubeadmConfigSpec.Etcd.Local.ImageRepository).To(Equal("registry.tkg.vmware.run"))
					Expect(bomConfiguration.KubeadmConfigSpec.Etcd.Local.ImageTag).To(Equal("v3.4.13_vmware.4"))
					Expect(len(bomConfiguration.KubeadmConfigSpec.Etcd.Local.ExtraArgs)).To(Equal(1))
					Expect(bomConfiguration.KubeadmConfigSpec.Etcd.Local.ExtraArgs["fake-arg"]).To(Equal("fake-arg-value"))
				})
			})
			Context("When tkr version(v1.19.3+vmware.1-tkg.1) is found", func() {
				BeforeEach(func() {
					createTempDirectory()
					tkgConfigDir = testingDir
					setupBomFile("../fakes/config/bom/tkr-bom-v1.18.0+vmware.1-tkg.2.yaml", tkgConfigDir)
					setupBomFile("../fakes/config/bom/tkr-bom-v1.19.3+vmware.1-tkg.1.yaml", tkgConfigDir)
					tkrVersion = "v1.19.3+vmware.1-tkg.1"
				})
				AfterEach(func() {
					deleteTempDirectory()
				})
				It("Should return the BOMConfiguration with valid etcd data when extraArgs not defined in KubeadmConfigSpec.Etcd.Local", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(bomConfiguration).ToNot(BeNil())
					Expect(bomConfiguration.KubeadmConfigSpec.Etcd.Local.DataDir).To(Equal("/var/lib/etcd"))
					Expect(bomConfiguration.KubeadmConfigSpec.Etcd.Local.ImageRepository).To(Equal("registry.tkg.vmware.run"))
					Expect(bomConfiguration.KubeadmConfigSpec.Etcd.Local.ImageTag).To(Equal("v3.4.13_vmware.4"))
					Expect(len(bomConfiguration.KubeadmConfigSpec.Etcd.Local.ExtraArgs)).To(Equal(0))
				})
			})
		})
		When("getting the default k8s version", func() {
			var (
				actual string
				err    error
			)
			JustBeforeEach(func() {
				actual, err = bomClient.GetDefaultK8sVersion()
			})
			When("the version is found", func() {
				It("should return the version", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(actual).To(Equal("v1.18.0+vmware.1"))
				})
			})
		})
		Context("GetDefaultClusterAPIProviders", func() {
			var (
				coreProvider         string
				bootstrapProvider    string
				controlPlaneProvider string
				err                  error
			)
			JustBeforeEach(func() {
				coreProvider, bootstrapProvider, controlPlaneProvider, err = bomClient.GetDefaultClusterAPIProviders()
			})
			When("BoM file is present", func() {
				It("Should return the provider information", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(coreProvider).To(Equal("cluster-api:v0.3.11"))
					Expect(bootstrapProvider).To(Equal("kubeadm:v0.3.11"))
					Expect(controlPlaneProvider).To(Equal("kubeadm:v0.3.11"))
				})
			})
		})
		Context("GetAvailableK8sVersionsFromBOMFiles", func() {
			var (
				actual []string
				err    error
			)
			JustBeforeEach(func() {
				actual, err = bomClient.GetAvailableK8sVersionsFromBOMFiles()
			})
			When("BOM file is present", func() {
				It("Should return the k8s versions", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(actual).To(ContainElements("v1.18.0+vmware.1", "v1.19.3+vmware.1"))
				})
			})
		})
		Context("GetCustomRepositoryCaCertificateForClient", func() {
			var (
				actual []byte
				err    error
			)
			JustBeforeEach(func() {
				actual, err = clientconfighelpers.GetCustomRepositoryCaCertificateForClient(tkgConfigReaderWriter)
			})
			When("BOM file is present without a Custom Image Repository", func() {
				It("should return the custom registry", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(actual).To(Equal([]byte{}))
				})
			})
			When("BOM file is present without a Custom Image Repository", func() {
				BeforeEach(func() {
					clusterConfigFile = kubeconfig7Path
				})
				It("should return the custom registry", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(actual).To(HaveLen(3626))
				})
			})
		})
		Context("InitBOMRegistry", func() {
			var (
				actual registry.Registry
				err    error
			)
			JustBeforeEach(func() {
				actual, err = bomClient.InitBOMRegistry()
			})
			When("Custom Registry is set", func() {
				BeforeEach(func() {
					clusterConfigFile = kubeconfig7Path
				})
				It("Should return a registry", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(actual).ToNot(BeNil())
				})
			})
		})
		Context("IsCustomRepositorySkipTLSVerify", func() {
			var actual bool
			JustBeforeEach(func() {
				actual = bomClient.IsCustomRepositorySkipTLSVerify()
			})
			When("SkipTLSVerify is in the config file", func() {
				BeforeEach(func() {
					clusterConfigFile = "../fakes/config/config7.yaml"
				})
				It("Should return the value of SkipTLSVerify", func() {
					Expect(actual).To(BeTrue())
				})
			})
		})
		Context("GetDefaultTKRVersion", func() {
			var (
				actual string
				err    error
			)
			JustBeforeEach(func() {
				actual, err = bomClient.GetDefaultTKRVersion()
			})
			When("BOM file is present", func() {
				It("Should return the default TKR version", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(actual).To(Equal("v1.18.0+vmware.1-tkg.2"))
				})
			})
		})
		Context("GetDefaultTkgBOMConfiguration", func() {
			It("Should be populated with CAPI version and the supported provider versions", func() {
				bomConfiguration, err := bomClient.GetDefaultTkgBOMConfiguration()
				Expect(err).ToNot(HaveOccurred())
				Expect(bomConfiguration.ProvidersVersionMap["cluster-api"]).To(Equal("v0.3.11-13-ga74685ee9"))
				Expect(bomConfiguration.ProvidersVersionMap["bootstrap-kubeadm"]).To(Equal("v0.3.11-13-ga74685ee9"))
				Expect(bomConfiguration.ProvidersVersionMap["control-plane-kubeadm"]).To(Equal("v0.3.11-13-ga74685ee9"))
				Expect(bomConfiguration.ProvidersVersionMap["infrastructure-docker"]).To(Equal("v0.3.11-13-ga74685ee9"))
				Expect(bomConfiguration.ProvidersVersionMap["infrastructure-azure"]).To(Equal("v0.4.8-47-gfbb2d55b"))
				Expect(bomConfiguration.ProvidersVersionMap["infrastructure-aws"]).To(Equal("v0.6.3"))
				Expect(bomConfiguration.ProvidersVersionMap["infrastructure-vsphere"]).To(Equal("v0.7.1"))
			})
		})
		Context("GetFullImagePath", func() {
			var (
				image               *tkgconfigbom.ImageInfo
				baseImageRepository string
				actual              string
				vmwareBaseImageRepo = "vmware"
			)
			JustBeforeEach(func() {
				actual = tkgconfigbom.GetFullImagePath(image, baseImageRepository)
			})
			When("ImageRepository is empty", func() {
				BeforeEach(func() {
					image = &tkgconfigbom.ImageInfo{
						ImageRepository: "",
						ImagePath:       "cluster-api",
					}
					baseImageRepository = vmwareBaseImageRepo
				})
				It("Should use the baseImageRepository path", func() {
					Expect(actual).To(Equal("vmware/cluster-api"))
				})
			})
			When("ImageRepository is not empty", func() {
				BeforeEach(func() {
					image = &tkgconfigbom.ImageInfo{
						ImageRepository: "azure",
						ImagePath:       "cluster-api",
					}
					baseImageRepository = vmwareBaseImageRepo
				})
				It("Should use the ImageRepository path", func() {
					Expect(actual).To(Equal("azure/cluster-api"))
				})
			})
		})
	})
)

func createTempDirectory() {
	testingDir, _ = os.MkdirTemp("", "bom_test")
}

func deleteTempDirectory() {
	os.Remove(testingDir)
}

var _ = Describe("GetK8sVersionFromTkrBoM", func() {
	var (
		bomConfiguration *tkgconfigbom.BOMConfiguration
		err              error
	)
	JustBeforeEach(func() {
		_, err = tkgconfigbom.GetK8sVersionFromTkrBoM(bomConfiguration)
	})
	Context("When BOMConfiguration is nil", func() {
		BeforeEach(func() {
			bomConfiguration = nil
		})
		It("Should return an error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("invalid BoM configuration"))
		})
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

func setupTestingFiles(clusterConfigFile string, configDir string, defaultBomFile string) {
	testClusterConfigFile := filepath.Join(configDir, "config.yaml")
	err := utils.CopyFile(clusterConfigFile, testClusterConfigFile)
	Expect(err).ToNot(HaveOccurred())

	bomDir, err := tkgconfigpaths.New(configDir).GetTKGBoMDirectory()
	Expect(err).ToNot(HaveOccurred())
	if _, err := os.Stat(bomDir); os.IsNotExist(err) {
		err = os.MkdirAll(bomDir, 0o700)
		Expect(err).ToNot(HaveOccurred())
	}
	err = utils.CopyFile(defaultBomFile, filepath.Join(bomDir, filepath.Base(defaultBomFile)))
	Expect(err).ToNot(HaveOccurred())

	compatibilityDir, err := tkgconfigpaths.New(configDir).GetTKGCompatibilityDirectory()
	Expect(err).ToNot(HaveOccurred())
	if _, err := os.Stat(compatibilityDir); os.IsNotExist(err) {
		err = os.MkdirAll(compatibilityDir, 0o700)
		Expect(err).ToNot(HaveOccurred())
	}

	defaultBomFileTag := utils.GetTKGBoMTagFromFileName(filepath.Base(defaultBomFile))
	testTKGCompatabilityFileContent := fmt.Sprintf(testTKGCompatibilityFileFmt, tkgconfigpaths.TKGManagementClusterPluginVersion, defaultBomFileTag)

	compatibilityConfigFile, err := tkgconfigpaths.New(configDir).GetTKGCompatibilityConfigPath()
	Expect(err).ToNot(HaveOccurred())
	err = os.WriteFile(compatibilityConfigFile, []byte(testTKGCompatabilityFileContent), constants.ConfigFilePermissions)
	Expect(err).ToNot(HaveOccurred())
}

func setupBomFile(defaultBomFile string, configDir string) {
	bomDir, err := tkgconfigpaths.New(configDir).GetTKGBoMDirectory()
	Expect(err).ToNot(HaveOccurred())
	if _, err := os.Stat(bomDir); os.IsNotExist(err) {
		err = os.MkdirAll(bomDir, 0o700)
		Expect(err).ToNot(HaveOccurred())
	}
	err = utils.CopyFile(defaultBomFile, filepath.Join(bomDir, filepath.Base(defaultBomFile)))
	Expect(err).ToNot(HaveOccurred())
}
