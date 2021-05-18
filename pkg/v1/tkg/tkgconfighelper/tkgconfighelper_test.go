// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgconfighelper_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgconfighelper"
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
		Context("mgmtClusterVersion= v1.5.0, kubernetesVersion=v1.21.1+vmware.1", func() {
			BeforeEach(func() {
				mgmtClusterVersion = "v1.5.0"
				kubernetesVersion = "v1.21.1+vmware.1"
			})
			It("should return error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("only [v1.0 v1.1 v1.2 v1.3 v1.4] management cluster versions are supported with current version of TKG CLI. Please upgrade TKG CLI to latest version if you are using it on latest version of management cluster."))
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
