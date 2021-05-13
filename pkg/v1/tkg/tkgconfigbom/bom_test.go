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
)

var (
	testingDir string
	_          = Describe("Unit tests for bom client", func() {
		var (
			bomClient         tkgconfigbom.Client
			fakeRegistry      *fakes.Registry
			tkgConfigDir      string
			clusterConfigFile string
		)

		JustBeforeEach(func() {
			tkgConfigReaderWriter, err := tkgconfigreaderwriter.NewReaderWriterFromConfigFile(clusterConfigFile, filepath.Join(tkgConfigDir, "config.yaml"))
			Expect(err).NotTo(HaveOccurred())

			bomClient = tkgconfigbom.New(tkgConfigDir, tkgConfigReaderWriter)
			tkgconfigpaths.TKGDefaultBOMImageTag = "v1.3.1"
		})
		Describe("When upgrading cluster with autoscaler enabled", func() {
			BeforeEach(func() {
				tkgConfigDir = "../fakes/config"
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
