// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package managementcomponents_test

import (
	"errors"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/managementcomponents"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ManagementComponent Suite")
}

var _ = Describe("Test InstallManagementPackages", func() {
	var (
		fakePkgClient *fakes.TKGPackageClient
		mprOptions    ManagementPackageRepositoryOptions
		err           error
	)

	BeforeEach(func() {
		fakePkgClient = &fakes.TKGPackageClient{}
		mprOptions = ManagementPackageRepositoryOptions{ManagementPackageRepoImage: "", TKGPackageValuesFile: ""}
	})

	JustBeforeEach(func() {
		err = InstallManagementPackages(fakePkgClient, mprOptions)
	})

	Context("when update repository throws error", func() {
		BeforeEach(func() {
			fakePkgClient.UpdateRepositorySyncReturns(errors.New("fake error update repository"))
			fakePkgClient.InstallPackageSyncReturns(nil)
		})
		It("should return error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake error update repository"))
		})
	})

	Context("when install package throws error", func() {
		BeforeEach(func() {
			fakePkgClient.UpdateRepositorySyncReturns(nil)
			fakePkgClient.InstallPackageSyncReturns(errors.New("fake error install package"))
		})
		It("should return error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake error install package"))
		})
	})

	Context("when update repository and install package are successful", func() {
		BeforeEach(func() {
			fakePkgClient.UpdateRepositorySyncReturns(nil)
			fakePkgClient.InstallPackageSyncReturns(nil)
		})
		It("should not return error", func() {
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

var _ = Describe("Test InstallKappController", func() {
	var (
		clusterClient *fakes.ClusterClient
		kcOptions     KappControllerOptions
		err           error
	)

	BeforeEach(func() {
		clusterClient = &fakes.ClusterClient{}
		kcOptions = KappControllerOptions{KappControllerConfigFile: "", KappControllerInstallNamespace: ""}
	})

	JustBeforeEach(func() {
		err = InstallKappController(clusterClient, kcOptions)
	})

	Context("when applying kapp-controller config throws error", func() {
		BeforeEach(func() {
			clusterClient.ApplyFileReturns(errors.New("fake error applyfile"))
			clusterClient.WaitForDeploymentReturns(nil)
		})
		It("should return error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake error applyfile"))
		})
	})

	Context("when WaitForDeployment config throws error", func() {
		BeforeEach(func() {
			clusterClient.ApplyFileReturns(nil)
			clusterClient.WaitForDeploymentReturns(errors.New("fake error waitfordeployment"))
		})
		It("should return error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake error waitfordeployment"))
		})
	})

	Context("when kapp-controller is deployed successfully", func() {
		BeforeEach(func() {
			clusterClient.ApplyFileReturns(nil)
			clusterClient.WaitForDeploymentReturns(nil)
		})
		It("should return error", func() {
			Expect(err).NotTo(HaveOccurred())
		})
	})

})
