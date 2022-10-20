// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgconfighelper_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/fakes"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigbom"
	. "github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfighelper"
	"github.com/vmware-tanzu/tanzu-framework/tkg/types"
)

const (
	k8sVersion1dot18dot16vmware4 = "v1.18.16+vmware.4"
	k8sVersion1dot18dot1vmware1  = "v1.18.1+vmware.1"
	k8sVersion1dot19dot1vmware1  = "v1.19.1+vmware.1"
	k8sVersion2dot16dot1vmware1  = "v2.16.1+vmware.1"
	tkgVersion1dot0dot0          = "v1.0.0"
	tkgVersion1dot1dot0          = "v1.1.0"
	tkgVersion1dot1dot0rc1       = "v1.1.0-rc.1"
)

func TestTKGConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tkg config helper Suite")
}

var _ = Describe("ValidateK8sVersionSupport", func() {
	var (
		err                error
		mgmtClusterVersion string
		kubernetesVersion  string
	)

	JustBeforeEach(func() {
		err = ValidateK8sVersionSupport(mgmtClusterVersion, kubernetesVersion)
	})

	Context("when k8s version is not supported by management cluster", func() {
		Context("mgmtClusterVersion= v1.0.0, kubernetesVersion=v1.18.1+vmware.1", func() {
			BeforeEach(func() {
				mgmtClusterVersion = tkgVersion1dot0dot0
				kubernetesVersion = k8sVersion1dot18dot1vmware1
			})
			It("should return error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("kubernetes version v1.18.1+vmware.1 is not supported on current v1.0.0 management cluster. Please upgrade management cluster if you are trying to deploy latest version of kubernetes"))
			})
		})
		Context("mgmtClusterVersion= v1.0.0, kubernetesVersion=v1.18.2", func() {
			BeforeEach(func() {
				mgmtClusterVersion = tkgVersion1dot0dot0
				kubernetesVersion = "v1.18.2"
			})
			It("should return error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("kubernetes version v1.18.2 is not supported on current v1.0.0 management cluster. Please upgrade management cluster if you are trying to deploy latest version of kubernetes"))
			})
		})
		Context("mgmtClusterVersion= v1.0.0, kubernetesVersion=v1.16.1+vmware.1", func() {
			BeforeEach(func() {
				mgmtClusterVersion = tkgVersion1dot0dot0
				kubernetesVersion = "v1.16.1+vmware.1"
			})
			It("should return error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("kubernetes version v1.16.1+vmware.1 is not supported on current v1.0.0 management cluster."))
			})
		})
		Context("mgmtClusterVersion= v1.0.0, kubernetesVersion=v2.18.1+vmware.1", func() {
			BeforeEach(func() {
				mgmtClusterVersion = tkgVersion1dot0dot0
				kubernetesVersion = k8sVersion2dot16dot1vmware1
			})
			It("should return error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("kubernetes version v2.16.1+vmware.1 is not supported on current v1.0.0 management cluster."))
			})
		})

		Context("mgmtClusterVersion= v1.1.0-rc.1, kubernetesVersion=v2.18.1+vmware.1", func() {
			BeforeEach(func() {
				mgmtClusterVersion = "v1.1.0-rc.1"
				kubernetesVersion = k8sVersion2dot16dot1vmware1
			})
			It("should return error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("kubernetes version v2.16.1+vmware.1 is not supported on current v1.1.0-rc.1 management cluster."))
			})
		})
		Context("mgmtClusterVersion= v1.1.0, kubernetesVersion=v2.18.1+vmware.1", func() {
			BeforeEach(func() {
				mgmtClusterVersion = tkgVersion1dot1dot0
				kubernetesVersion = k8sVersion2dot16dot1vmware1
			})
			It("should return error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("kubernetes version v2.16.1+vmware.1 is not supported on current v1.1.0 management cluster."))
			})
		})
		Context("mgmtClusterVersion= v1.1.1, kubernetesVersion=v1.19.1+vmware.1", func() {
			BeforeEach(func() {
				mgmtClusterVersion = "v1.1.1"
				kubernetesVersion = k8sVersion1dot19dot1vmware1
			})
			It("should return error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("kubernetes version v1.19.1+vmware.1 is not supported on current v1.1.1 management cluster."))
			})
		})
		Context("mgmtClusterVersion= v1.2.1, kubernetesVersion=v1.20.1+vmware.1", func() {
			BeforeEach(func() {
				mgmtClusterVersion = "v1.2.1"
				kubernetesVersion = "v1.20.1+vmware.1"
			})
			It("should return error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("kubernetes version v1.20.1+vmware.1 is not supported on current v1.2.1 management cluster."))
			})
		})
		Context("mgmtClusterVersion= v3.6.0, kubernetesVersion=v1.21.1+vmware.1", func() {
			BeforeEach(func() {
				mgmtClusterVersion = "v3.6.0"
				kubernetesVersion = "v1.21.1+vmware.1"
			})
			It("should return error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("only [v1.0 v1.1 v1.2 v1.3 v1.4 v1.5 v1.6 v1.7 v2.1] management cluster versions are supported with current version of TKG CLI. Please upgrade TKG CLI to latest version if you are using it on latest version of management cluster."))
			})
		})
	})

	Context("when k8s version is supported by management cluster", func() {
		Context("mgmtClusterVersion= v1.0.0, kubernetesVersion=v1.17.3+vmware.2", func() {
			BeforeEach(func() {
				mgmtClusterVersion = tkgVersion1dot0dot0
				kubernetesVersion = "v1.17.3+vmware.2"
			})
			It("should not return error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
		Context("mgmtClusterVersion= v1.0.1, kubernetesVersion=v1.17.3+vmware.5", func() {
			BeforeEach(func() {
				mgmtClusterVersion = "v1.0.1"
				kubernetesVersion = "v1.17.3+vmware.5"
			})
			It("should not return error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
		Context("mgmtClusterVersion= v1.0.0, kubernetesVersion=v1.17.5+vmware.2", func() {
			BeforeEach(func() {
				mgmtClusterVersion = tkgVersion1dot0dot0
				kubernetesVersion = "v1.17.5+vmware.2"
			})
			It("should not return error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
		Context("mgmtClusterVersion= v1.0.0, kubernetesVersion=v1.17.12+vmware.1", func() {
			BeforeEach(func() {
				mgmtClusterVersion = tkgVersion1dot0dot0
				kubernetesVersion = "v1.17.12+vmware.1"
			})
			It("should not return error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("mgmtClusterVersion= v1.1.0-rc.1, kubernetesVersion=v1.18.1+vmware.1", func() {
			BeforeEach(func() {
				mgmtClusterVersion = tkgVersion1dot1dot0rc1
				kubernetesVersion = k8sVersion1dot18dot1vmware1
			})
			It("should not return error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
		Context("mgmtClusterVersion= v1.1.0, kubernetesVersion=v1.17.9+vmware.1", func() {
			BeforeEach(func() {
				mgmtClusterVersion = tkgVersion1dot1dot0
				kubernetesVersion = "v1.17.9+vmware.1"
			})
			It("should not return error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
		Context("mgmtClusterVersion= v1.1.4, kubernetesVersion=v1.17.19+vmware.1", func() {
			BeforeEach(func() {
				mgmtClusterVersion = "v1.1.4"
				kubernetesVersion = "v1.17.19+vmware.1"
			})
			It("should not return error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
		Context("mgmtClusterVersion= v1.1.0, kubernetesVersion=v1.18.2+vmware.1", func() {
			BeforeEach(func() {
				mgmtClusterVersion = tkgVersion1dot1dot0
				kubernetesVersion = "v1.18.2+vmware.1"
			})
			It("should not return error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
		Context("mgmtClusterVersion= v1.1.10, kubernetesVersion=v1.18.16+vmware.4", func() {
			BeforeEach(func() {
				mgmtClusterVersion = "v1.1.10"
				kubernetesVersion = k8sVersion1dot18dot16vmware4
			})
			It("should not return error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
		Context("mgmtClusterVersion= v1.2.0, kubernetesVersion=v1.17.19+vmware.4", func() {
			BeforeEach(func() {
				mgmtClusterVersion = "v1.2.0"
				kubernetesVersion = "v1.17.19+vmware.4"
			})
			It("should not return error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
		Context("mgmtClusterVersion= v1.2.4, kubernetesVersion=v1.18.16+vmware.4", func() {
			BeforeEach(func() {
				mgmtClusterVersion = "v1.2.4"
				kubernetesVersion = k8sVersion1dot18dot16vmware4
			})
			It("should not return error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
		Context("mgmtClusterVersion= v1.2.10, kubernetesVersion=v1.19.1+vmware.1", func() {
			BeforeEach(func() {
				mgmtClusterVersion = "v1.2.10"
				kubernetesVersion = k8sVersion1dot19dot1vmware1
			})
			It("should not return error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})

var _ = Describe("GetDefaultOsOptions", func() {
	var providerType string
	var actual tkgconfigbom.OSInfo

	JustBeforeEach(func() {
		actual = GetDefaultOsOptions(providerType)
	})
	Context("When provider type is vsphere", func() {
		BeforeEach(func() {
			providerType = constants.InfrastructureProviderVSphere
		})
		It("should return the correct OSInfo", func() {
			Expect(actual).To(Equal(tkgconfigbom.OSInfo{
				Name:    "ubuntu",
				Version: "20.04",
				Arch:    "amd64",
			}))
		})
	})
	Context("When provider type is aws", func() {
		BeforeEach(func() {
			providerType = constants.InfrastructureProviderAWS
		})
		It("should return the correct OSInfo", func() {
			Expect(actual).To(Equal(tkgconfigbom.OSInfo{
				Name:    "ubuntu",
				Version: "20.04",
				Arch:    "amd64",
			}))
		})
	})
	Context("When provider type is azure", func() {
		BeforeEach(func() {
			providerType = constants.InfrastructureProviderAzure
		})
		It("should return the correct OSInfo", func() {
			Expect(actual).To(Equal(tkgconfigbom.OSInfo{
				Name:    "ubuntu",
				Version: "20.04",
				Arch:    "amd64",
			}))
		})
	})
	Context("When provider type is empty", func() {
		BeforeEach(func() {
			providerType = ""
		})
		It("should return empty OSInfo", func() {
			Expect(actual).To(Equal(tkgconfigbom.OSInfo{
				Name:    "",
				Version: "",
				Arch:    "",
			}))
		})
	})
})

var _ = Describe("GetOSOptionsForProviders", func() {
	var (
		providerType          string
		tkgConfigReaderWriter *fakes.TKGConfigReaderWriter
		actual                tkgconfigbom.OSInfo
	)
	JustBeforeEach(func() {
		actual = GetOSOptionsForProviders(providerType, tkgConfigReaderWriter)
	})
	Context("When providerType is empty", func() {
		BeforeEach(func() {
			providerType = ""
			tkgConfigReaderWriter = &fakes.TKGConfigReaderWriter{}
			tkgConfigReaderWriter.GetReturnsOnCall(0, "ubuntu", nil)
			tkgConfigReaderWriter.GetReturnsOnCall(1, "20.04", nil)
			tkgConfigReaderWriter.GetReturnsOnCall(2, "amd64", nil)
		})
		It("Should populate OSInfo from the TKG Config", func() {
			Expect(actual).ToNot(BeNil())
			Expect(actual.Arch).To(Equal("amd64"))
			Expect(actual.Name).To(Equal("ubuntu"))
			Expect(actual.Version).To(Equal("20.04"))
		})
	})
})

var _ = Describe("GetUserProvidedOsOptions", func() {
	var (
		tkgConfigReaderWriter *fakes.TKGConfigReaderWriter
		actual                tkgconfigbom.OSInfo
	)
	JustBeforeEach(func() {
		actual = GetUserProvidedOsOptions(tkgConfigReaderWriter)
	})
	Context("When tkgConfigReaderWriter is present", func() {
		BeforeEach(func() {
			tkgConfigReaderWriter = &fakes.TKGConfigReaderWriter{}
			tkgConfigReaderWriter.GetReturnsOnCall(0, "ubuntu", nil)
			tkgConfigReaderWriter.GetReturnsOnCall(1, "20.04", nil)
			tkgConfigReaderWriter.GetReturnsOnCall(2, "amd64", nil)
		})
		It("Should populate the OSInfo from the config", func() {
			Expect(actual).ToNot(BeNil())
			Expect(actual.Arch).To(Equal("amd64"))
			Expect(actual.Name).To(Equal("ubuntu"))
			Expect(actual.Version).To(Equal("20.04"))
		})
	})
})

var _ = Describe("SelectTemplateForVsphereProviderBasedonOSOptions", func() {
	var (
		vms                   []*types.VSphereVirtualMachine
		tkgConfigReaderWriter *fakes.TKGConfigReaderWriter
		actual                *types.VSphereVirtualMachine
	)
	JustBeforeEach(func() {
		actual = SelectTemplateForVsphereProviderBasedonOSOptions(vms, tkgConfigReaderWriter)
	})
	Context("When vms is empty", func() {
		BeforeEach(func() {
			vms = []*types.VSphereVirtualMachine{}
		})
		It("should return nil", func() {
			Expect(actual).To(BeNil())
		})
	})
	Context("When no user OS option is present and there is only one vm", func() {
		BeforeEach(func() {
			vms = []*types.VSphereVirtualMachine{
				{
					Name: "photon-3-kube-v1.20.5+vmware.1",
				},
			}
			tkgConfigReaderWriter = &fakes.TKGConfigReaderWriter{}
			tkgConfigReaderWriter.GetReturnsOnCall(0, "", nil)
			tkgConfigReaderWriter.GetReturnsOnCall(1, "", nil)
			tkgConfigReaderWriter.GetReturnsOnCall(2, "", nil)
		})
		It("should return just that VM", func() {
			Expect(actual).ToNot(BeNil())
			Expect(actual.Name).To(Equal("photon-3-kube-v1.20.5+vmware.1"))
		})
	})
	Context("When user options present and there are many vms", func() {
		BeforeEach(func() {
			vms = []*types.VSphereVirtualMachine{
				{
					DistroName:    "ubuntu",
					DistroVersion: "20.04",
					DistroArch:    "amd64",
				},
				{
					DistroName:    "photon",
					DistroVersion: "3.0",
					DistroArch:    "amd64",
				},
			}
			tkgConfigReaderWriter = &fakes.TKGConfigReaderWriter{}
			tkgConfigReaderWriter.GetReturnsOnCall(0, "ubuntu", nil)
			tkgConfigReaderWriter.GetReturnsOnCall(1, "20.04", nil)
			tkgConfigReaderWriter.GetReturnsOnCall(2, "amd64", nil)
		})
		It("Should return a single vm", func() {
			Expect(actual).ToNot(BeNil())
			Expect(actual.DistroArch).To(Equal("amd64"))
			Expect(actual.DistroName).To(Equal("ubuntu"))
			Expect(actual.DistroVersion).To(Equal("20.04"))
		})
	})
	Context("When user options present and there are many vms with the same distro", func() {
		BeforeEach(func() {
			vms = []*types.VSphereVirtualMachine{
				{
					DistroName:    "ubuntu",
					DistroVersion: "20.04",
					DistroArch:    "amd64",
				},
				{
					DistroName:    "ubuntu",
					DistroVersion: "21.01",
					DistroArch:    "arm7",
				},
			}
			tkgConfigReaderWriter = &fakes.TKGConfigReaderWriter{}
			tkgConfigReaderWriter.GetReturnsOnCall(0, "ubuntu", nil)
			tkgConfigReaderWriter.GetReturnsOnCall(1, "", nil)
			tkgConfigReaderWriter.GetReturnsOnCall(2, "", nil)
		})
		It("Should return the first one vm", func() {
			Expect(actual).ToNot(BeNil())
			Expect(actual.DistroArch).To(Equal("amd64"))
			Expect(actual.DistroName).To(Equal("ubuntu"))
			Expect(actual.DistroVersion).To(Equal("20.04"))
		})
	})
})

var _ = Describe("SelectAzureImageBasedonOSOptions", func() {
	var (
		azureImages           []tkgconfigbom.AzureInfo
		tkgConfigReaderWriter *fakes.TKGConfigReaderWriter
		actual                *tkgconfigbom.AzureInfo
	)
	JustBeforeEach(func() {
		actual = SelectAzureImageBasedonOSOptions(azureImages, tkgConfigReaderWriter)
	})
	Context("When azureImages is empty", func() {
		BeforeEach(func() {
			azureImages = []tkgconfigbom.AzureInfo{}
		})
		It("should return nil", func() {
			Expect(actual).To(BeNil())
		})
	})
	Context("When no user OS option is present and there is only one vm", func() {
		BeforeEach(func() {
			azureImages = []tkgconfigbom.AzureInfo{
				{
					Name: "photon-3-kube-v1.20.5+vmware.1",
				},
			}
			tkgConfigReaderWriter = &fakes.TKGConfigReaderWriter{}
			tkgConfigReaderWriter.GetReturnsOnCall(0, "", nil)
			tkgConfigReaderWriter.GetReturnsOnCall(1, "", nil)
			tkgConfigReaderWriter.GetReturnsOnCall(2, "", nil)
		})
		It("should return just that VM", func() {
			Expect(actual).ToNot(BeNil())
			Expect(actual.Name).To(Equal("photon-3-kube-v1.20.5+vmware.1"))
		})
	})
	Context("When user options present and there are many vms", func() {
		BeforeEach(func() {
			azureImages = []tkgconfigbom.AzureInfo{
				{
					OSInfo: tkgconfigbom.OSInfo{
						Name:    "ubuntu",
						Version: "20.04",
						Arch:    "amd64",
					},
				},
				{
					OSInfo: tkgconfigbom.OSInfo{
						Name:    "photon",
						Version: "3.0",
						Arch:    "amd64",
					},
				},
			}
			tkgConfigReaderWriter = &fakes.TKGConfigReaderWriter{}
			tkgConfigReaderWriter.GetReturnsOnCall(0, "ubuntu", nil)
			tkgConfigReaderWriter.GetReturnsOnCall(1, "20.04", nil)
			tkgConfigReaderWriter.GetReturnsOnCall(2, "amd64", nil)
		})
		It("Should return a single vm", func() {
			Expect(actual).ToNot(BeNil())
			Expect(actual.OSInfo.Arch).To(Equal("amd64"))
			Expect(actual.OSInfo.Name).To(Equal("ubuntu"))
			Expect(actual.OSInfo.Version).To(Equal("20.04"))
		})
	})
	Context("When user options present and there are many vms with the same distro", func() {
		BeforeEach(func() {
			azureImages = []tkgconfigbom.AzureInfo{
				{
					OSInfo: tkgconfigbom.OSInfo{
						Name:    "ubuntu",
						Version: "20.04",
						Arch:    "amd64",
					},
				},
				{
					OSInfo: tkgconfigbom.OSInfo{
						Name:    "ubuntu",
						Version: "21.01",
						Arch:    "arm7",
					},
				},
			}
			tkgConfigReaderWriter = &fakes.TKGConfigReaderWriter{}
			tkgConfigReaderWriter.GetReturnsOnCall(0, "ubuntu", nil)
			tkgConfigReaderWriter.GetReturnsOnCall(1, "", nil)
			tkgConfigReaderWriter.GetReturnsOnCall(2, "", nil)
		})
		It("Should return the first one vm", func() {
			Expect(actual).ToNot(BeNil())
			Expect(actual.OSInfo.Arch).To(Equal("amd64"))
			Expect(actual.OSInfo.Name).To(Equal("ubuntu"))
			Expect(actual.OSInfo.Version).To(Equal("20.04"))
		})
	})
})

var _ = Describe("SelectAWSImageBasedonOSOptions", func() {
	var (
		amis                  []tkgconfigbom.AMIInfo
		tkgConfigReaderWriter *fakes.TKGConfigReaderWriter
		actual                *tkgconfigbom.AMIInfo
	)
	JustBeforeEach(func() {
		actual = SelectAWSImageBasedonOSOptions(amis, tkgConfigReaderWriter)
	})
	Context("When amis is empty", func() {
		BeforeEach(func() {
			amis = []tkgconfigbom.AMIInfo{}
		})
		It("should return nil", func() {
			Expect(actual).To(BeNil())
		})
	})
	Context("When no user OS option is present and there is only one vm", func() {
		BeforeEach(func() {
			amis = []tkgconfigbom.AMIInfo{
				{
					OSInfo: tkgconfigbom.OSInfo{
						Name: "photon-3-kube-v1.20.5+vmware.1",
					},
				},
			}
			tkgConfigReaderWriter = &fakes.TKGConfigReaderWriter{}
			tkgConfigReaderWriter.GetReturnsOnCall(0, "", nil)
			tkgConfigReaderWriter.GetReturnsOnCall(1, "", nil)
			tkgConfigReaderWriter.GetReturnsOnCall(2, "", nil)
		})
		It("should return just that VM", func() {
			Expect(actual).ToNot(BeNil())
			Expect(actual.OSInfo.Name).To(Equal("photon-3-kube-v1.20.5+vmware.1"))
		})
	})
	Context("When user options present and there are many vms", func() {
		BeforeEach(func() {
			amis = []tkgconfigbom.AMIInfo{
				{
					OSInfo: tkgconfigbom.OSInfo{
						Name:    "ubuntu",
						Version: "20.04",
						Arch:    "amd64",
					},
				},
				{
					OSInfo: tkgconfigbom.OSInfo{
						Name:    "photon",
						Version: "3.0",
						Arch:    "amd64",
					},
				},
			}
			tkgConfigReaderWriter = &fakes.TKGConfigReaderWriter{}
			tkgConfigReaderWriter.GetReturnsOnCall(0, "ubuntu", nil)
			tkgConfigReaderWriter.GetReturnsOnCall(1, "20.04", nil)
			tkgConfigReaderWriter.GetReturnsOnCall(2, "amd64", nil)
		})
		It("Should return a single vm", func() {
			Expect(actual).ToNot(BeNil())
			Expect(actual.OSInfo.Arch).To(Equal("amd64"))
			Expect(actual.OSInfo.Name).To(Equal("ubuntu"))
			Expect(actual.OSInfo.Version).To(Equal("20.04"))
		})
	})
	Context("When user options present and there are many vms with the same distro", func() {
		BeforeEach(func() {
			amis = []tkgconfigbom.AMIInfo{
				{
					OSInfo: tkgconfigbom.OSInfo{
						Name:    "ubuntu",
						Version: "20.04",
						Arch:    "amd64",
					},
				},
				{
					OSInfo: tkgconfigbom.OSInfo{
						Name:    "ubuntu",
						Version: "21.01",
						Arch:    "arm7",
					},
				},
			}
			tkgConfigReaderWriter = &fakes.TKGConfigReaderWriter{}
			tkgConfigReaderWriter.GetReturnsOnCall(0, "ubuntu", nil)
			tkgConfigReaderWriter.GetReturnsOnCall(1, "", nil)
			tkgConfigReaderWriter.GetReturnsOnCall(2, "", nil)
		})
		It("Should return the first one vm", func() {
			Expect(actual).ToNot(BeNil())
			Expect(actual.OSInfo.Arch).To(Equal("amd64"))
			Expect(actual.OSInfo.Name).To(Equal("ubuntu"))
			Expect(actual.OSInfo.Version).To(Equal("20.04"))
		})
	})
})

var _ = Describe("GetDefaultOsOptionsForTKG12", func() {
	var (
		providerType string
		actual       tkgconfigbom.OSInfo
	)
	JustBeforeEach(func() {
		actual = GetDefaultOsOptionsForTKG12(providerType)
	})
	Context("When providerType is vpshere", func() {
		BeforeEach(func() {
			providerType = constants.InfrastructureProviderVSphere
		})
		It("should return the correct OSInfo", func() {
			Expect(actual).ToNot(BeNil())
			Expect(actual.Arch).To(Equal("amd64"))
			Expect(actual.Name).To(Equal("photon"))
			Expect(actual.Version).To(Equal("3"))
		})
	})
	Context("When providerType is aws", func() {
		BeforeEach(func() {
			providerType = constants.InfrastructureProviderAWS
		})
		It("should return the correct OSInfo", func() {
			Expect(actual).ToNot(BeNil())
			Expect(actual.Arch).To(Equal("amd64"))
			Expect(actual.Name).To(Equal("amazon"))
			Expect(actual.Version).To(Equal("2"))
		})
	})
	Context("When providerType is azure", func() {
		BeforeEach(func() {
			providerType = constants.InfrastructureProviderAzure
		})
		It("should return the correct OSInfo", func() {
			Expect(actual).ToNot(BeNil())
			Expect(actual.Arch).To(Equal("amd64"))
			Expect(actual.Name).To(Equal("ubuntu"))
			Expect(actual.Version).To(Equal("18.04"))
		})
	})
})
