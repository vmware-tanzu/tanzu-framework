// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgconfigbom_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/fakes"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgconfigbom"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgconfigpaths"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgconfigreaderwriter"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkr/pkg/registry"
)

var (
	testingDir string
	_          = Describe("Unit tests for bom client", func() {
		var (
			bomClient         tkgconfigbom.Client
			fakeRegistry      *fakes.Registry
			tkgConfigDir      string
			clusterConfigFile string
			kubeconfig7Path   = "../fakes/config/config7.yaml"
			fakeDir           = "../fakes/config"
		)

		JustBeforeEach(func() {
			tkgConfigReaderWriter, err := tkgconfigreaderwriter.NewReaderWriterFromConfigFile(clusterConfigFile, filepath.Join(tkgConfigDir, "config.yaml"))
			Expect(err).NotTo(HaveOccurred())

			bomClient = tkgconfigbom.New(tkgConfigDir, tkgConfigReaderWriter)
			tkgconfigpaths.TKGDefaultBOMImageTag = "v1.3.1"
		})
		Describe("When upgrading cluster with autoscaler enabled", func() {
			BeforeEach(func() {
				tkgConfigDir = fakeDir
				clusterConfigFile = "../fakes/config/config.yaml"
			})
			Context("when multiple autoscaler image tags are present for a k8s minor version", func() {
				It("returns an error", func() {
					os.Setenv("DEFAULT_TKG_BOM_FILE_PATH", "../fakes/config/bom/tkg-bom-v1.3.1.yaml")
					_, err := bomClient.GetAutoscalerImageForK8sVersion("v1.18.0+vmware.1")
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("expected one autoscaler image for kubernetes minor version \"v1.18\" but found 2"))
				})
			})

			Context("For k8s version v1.19.0+vmware.1", func() {
				It("returns a valid autoscaler image", func() {
					os.Setenv("DEFAULT_TKG_BOM_FILE_PATH", "../fakes/config/bom/tkg-bom-v1.3.1.yaml")
					image, err := bomClient.GetAutoscalerImageForK8sVersion("v1.19.0+vmware.1")
					Expect(err).NotTo(HaveOccurred())
					Expect(image).To(Equal("projects-stg.registry.vmware.com/tkg/cluster-autoscaler:v1.19.1_vmware.1"))
				})
			})

			Context("For unsupported k8s version", func() {
				It("returns an error", func() {
					os.Setenv("DEFAULT_TKG_BOM_FILE_PATH", "../fakes/config/bom/tkg-bom-v1.3.1.yaml")
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
				f, _ := ioutil.TempFile(testingDir, "config.yaml")
				err := copyFile("../fakes/config/config.yaml", f.Name())
				Expect(err).ToNot(HaveOccurred())
				err = copyFile("../fakes/config/config.yaml", filepath.Join(tkgConfigDir, "config.yaml"))
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
					err := bomClient.DownloadDefaultBOMFilesFromRegistry(fakeRegistry)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("failed to download the BOM file from image name"))
					Expect(err.Error()).To(ContainSubstring("fake GetFile error for TKG BOM file"))
				})
			})

			Context("when downloading the TKG BOM file is success but fails to download TKR BOM file", func() {
				BeforeEach(func() {
					data, err := ioutil.ReadFile("../fakes/config/bom/tkg-bom-v1.3.1.yaml")
					Expect(err).ToNot(HaveOccurred())
					fakeRegistry.GetFileReturnsOnCall(0, data, nil)
					fakeRegistry.GetFileReturnsOnCall(1, nil, errors.New("fake GetFile error for TKR BOM file"))
				})
				It("should return an error", func() {
					err := bomClient.DownloadDefaultBOMFilesFromRegistry(fakeRegistry)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("failed to download the BOM file from image name"))
					Expect(err.Error()).To(ContainSubstring("fake GetFile error for TKR BOM file"))
				})
			})
			Context("when downloading the TKG BOM file and TKR BOM file is success ", func() {
				BeforeEach(func() {
					tkgdata, err := ioutil.ReadFile("../fakes/config/bom/tkg-bom-v1.3.1.yaml")
					Expect(err).ToNot(HaveOccurred())
					tkrdata, err := ioutil.ReadFile("../fakes/config/bom/tkg-bom-v1.3.1.yaml")
					Expect(err).ToNot(HaveOccurred())
					fakeRegistry.GetFileReturnsOnCall(0, tkgdata, nil)
					fakeRegistry.GetFileReturnsOnCall(1, tkrdata, nil)
				})
				It("should return success", func() {
					err := bomClient.DownloadDefaultBOMFilesFromRegistry(fakeRegistry)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})
		Context("When getting BOMConfiguration from TKR version", func() {
			var (
				tkrVersion       string
				bomConfiguration *tkgconfigbom.BOMConfiguration
				err              error
			)
			JustBeforeEach(func() {
				tkgConfigDir = fakeDir
				bomConfiguration, err = bomClient.GetBOMConfigurationFromTkrVersion(tkrVersion)
			})
			Context("When tkr version is missing", func() {
				BeforeEach(func() {
					tkrVersion = ""
				})
				It("Should return the bomConfiguration", func() {
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("No BOM file found with TKR version "))
				})
			})
			Context("When tkr version is found", func() {
				BeforeEach(func() {
					tkrVersion = "v1.18.0+vmware.1-tkg.2"
				})
				It("Should return the BOMConfiguration", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(bomConfiguration).ToNot(BeNil())
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
				BeforeEach(func() {
					tkgConfigDir = fakeDir
				})
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
				BeforeEach(func() {
					tkgConfigDir = fakeDir
				})
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
				BeforeEach(func() {
					tkgConfigDir = fakeDir
				})
				It("Should return the k8s versions", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(actual).To(ContainElements("v1.18.0+vmware.1", "v1.19.3+vmware.1"))
				})
			})
		})
		Context("GetCustomRepositoryCaCertificate", func() {
			var (
				actual []byte
				err    error
			)
			JustBeforeEach(func() {
				actual, err = bomClient.GetCustomRepositoryCaCertificate()
			})
			When("BOM file is present without a Custom Image Repository", func() {
				BeforeEach(func() {
					tkgConfigDir = fakeDir
				})
				It("should return the custom registry", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(actual).To(Equal([]byte{}))
				})
			})
			When("BOM file is present without a Custom Image Repository", func() {
				BeforeEach(func() {
					tkgConfigDir = fakeDir
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
					tkgConfigDir = fakeDir
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
					tkgConfigDir = fakeDir
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
				BeforeEach(func() {
					tkgConfigDir = fakeDir
				})
				It("Should return the default TKR version", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(actual).To(Equal("v1.18.0+vmware.1-tkg.2"))
				})
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
	testingDir, _ = ioutil.TempDir("", "bom_test")
}

func deleteTempDirectory() {
	os.Remove(testingDir)
}

func copyFile(sourceFile, destFile string) error {
	input, err := ioutil.ReadFile(sourceFile)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(destFile, input, constants.ConfigFilePermissions)
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
