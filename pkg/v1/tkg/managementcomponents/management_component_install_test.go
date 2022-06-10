// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package managementcomponents_test

import (
	"errors"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes"
	. "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/managementcomponents"
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

var _ = Describe("Test WaitForManagementPackages", func() {
	var (
		clusterClient *fakes.ClusterClient
		err           error
		timeout       time.Duration
	)

	BeforeEach(func() {
		clusterClient = &fakes.ClusterClient{}
		timeout = time.Duration(1)
	})

	JustBeforeEach(func() {
		err = WaitForManagementPackages(clusterClient, timeout)
	})

	Context("when listing packageinstall throws error", func() {
		BeforeEach(func() {
			clusterClient.ListResourcesReturns(errors.New("fake error"))
			clusterClient.WaitForPackageInstallReturns(nil)
		})
		It("should return error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unable to list PackageInstalls"))
			Expect(err.Error()).To(ContainSubstring("fake error"))
		})
	})

	Context("when there is an error while waiting for packageinstall to reconcile successfully", func() {
		BeforeEach(func() {
			clusterClient.ListResourcesReturns(nil)
			clusterClient.WaitForPackageInstallReturns(errors.New("fake error"))
		})
		It("should return error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("error while waiting for management packages to be installed"))
			Expect(err.Error()).To(ContainSubstring("fake error"))
		})
	})

	Context("when packages gets reconciled successfully", func() {
		BeforeEach(func() {
			clusterClient.ListResourcesReturns(nil)
			clusterClient.WaitForPackageInstallReturns(nil)
		})
		It("should return error", func() {
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
