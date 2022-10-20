// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package managementcomponents_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	packageclientfakes "github.com/vmware-tanzu/tanzu-framework/packageclients/pkg/fakes"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/fakes"
	. "github.com/vmware-tanzu/tanzu-framework/tkg/managementcomponents"
)

const (
	addonsManager   = "addons-manager"
	akoOperatorName = "ako-operator"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ManagementComponent Suite")
}

var _ = Describe("Test InstallManagementPackages", func() {
	var (
		fakePkgClient *packageclientfakes.PackageClient
		mprOptions    ManagementPackageRepositoryOptions
		err           error
	)

	BeforeEach(func() {
		fakePkgClient = &packageclientfakes.PackageClient{}
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
		notFoundError = apierrors.NewNotFound(
			schema.GroupResource{Group: "fakeGroup", Resource: "fakeGroupResource"},
			"fakeGroupResource")
	)

	BeforeEach(func() {
		clusterClient = &fakes.ClusterClient{}
		kcOptions = KappControllerOptions{KappControllerConfigFile: "", KappControllerInstallNamespace: ""}
	})

	Context("when getting kapp-controller deployment throws error other than NotFound err, while upgrading kapp controller", func() {
		JustBeforeEach(func() {
			err = InstallKappController(clusterClient, kcOptions, constants.OperationTypeUpgrade)
		})
		BeforeEach(func() {
			clusterClient.GetResourceReturns(errors.New("fake error getting resource"))
		})
		It("should return error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake error getting resource"))
		})
	})

	Context("when adding last-applied annotation on kapp-controller throws error, while upgrading kapp controller", func() {
		JustBeforeEach(func() {
			err = InstallKappController(clusterClient, kcOptions, constants.OperationTypeUpgrade)
		})
		BeforeEach(func() {
			clusterClient.GetResourceReturns(nil)
			clusterClient.PatchKappControllerLastAppliedAnnotationReturns(errors.New("fake error adding annotation"))
		})
		It("should return error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake error adding annotation"))
			Expect(err.Error()).To(ContainSubstring("error adding last-applied annotation on kapp-controller"))
		})
	})

	Context("when applying kapp-controller config throws error, while upgrading kapp controller", func() {
		JustBeforeEach(func() {
			err = InstallKappController(clusterClient, kcOptions, constants.OperationTypeUpgrade)
		})
		BeforeEach(func() {
			clusterClient.GetResourceReturns(notFoundError)
			clusterClient.ApplyFileReturns(errors.New("fake error applyfile"))
		})
		It("should return error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake error applyfile"))
			Expect(err.Error()).To(ContainSubstring("error installing %s", constants.KappControllerDeploymentName))
		})
	})

	Context("when WaitForDeployment config throws error, while upgrading kapp controller", func() {
		JustBeforeEach(func() {
			err = InstallKappController(clusterClient, kcOptions, constants.OperationTypeUpgrade)
		})
		BeforeEach(func() {
			clusterClient.GetResourceReturns(notFoundError)
			clusterClient.ApplyFileReturns(nil)
			clusterClient.WaitForDeploymentReturns(errors.New("fake error waitfordeployment"))
		})
		It("should return error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake error waitfordeployment"))
			Expect(err.Error()).To(ContainSubstring("error while waiting for deployment %s", constants.KappControllerDeploymentName))
		})
	})

	Context("when kapp-controller is deployed successfully, while upgrading kapp controller", func() {
		JustBeforeEach(func() {
			err = InstallKappController(clusterClient, kcOptions, constants.OperationTypeUpgrade)
		})
		BeforeEach(func() {
			clusterClient.GetResourceReturns(notFoundError)
			clusterClient.ApplyFileReturns(nil)
			clusterClient.WaitForDeploymentReturns(nil)
		})
		It("should return error", func() {
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when applying kapp-controller throws error, while installing kapp controller", func() {
		JustBeforeEach(func() {
			err = InstallKappController(clusterClient, kcOptions, constants.OperationTypeInstall)
		})
		BeforeEach(func() {
			clusterClient.GetResourceReturns(notFoundError)
			clusterClient.ApplyFileReturns(errors.New("fake error applyfile"))
		})
		It("should return error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake error applyfile"))
			Expect(err.Error()).To(ContainSubstring("error installing %s", constants.KappControllerDeploymentName))
		})
	})

	Context("when WaitForDeployment config throws error, while installing kapp controller", func() {
		JustBeforeEach(func() {
			err = InstallKappController(clusterClient, kcOptions, constants.OperationTypeInstall)
		})
		BeforeEach(func() {
			clusterClient.GetResourceReturns(notFoundError)
			clusterClient.ApplyFileReturns(nil)
			clusterClient.WaitForDeploymentReturns(errors.New("fake error waitfordeployment"))
		})
		It("should return error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake error waitfordeployment"))
			Expect(err.Error()).To(ContainSubstring("error while waiting for deployment %s", constants.KappControllerDeploymentName))
		})
	})

	Context("when kapp-controller is deployed successfully, while installing kapp controller", func() {
		JustBeforeEach(func() {
			err = InstallKappController(clusterClient, kcOptions, constants.OperationTypeInstall)
		})
		BeforeEach(func() {
			clusterClient.GetResourceReturns(notFoundError)
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

var _ = Describe("Test PauseAddonLifecycleManagement", func() {
	var (
		clusterClient *fakes.ClusterClient
		err           error
		clusterName   string
		addonName     string
		namespace     string
		notFoundError = apierrors.NewNotFound(
			schema.GroupResource{Group: "fakeGroup", Resource: "fakeGroupResource"},
			"fakeGroupResource")
	)

	BeforeEach(func() {
		clusterClient = &fakes.ClusterClient{}
		clusterName = "mgmtCluster"
		addonName = addonsManager
	})

	JustBeforeEach(func() {
		err = PauseAddonLifecycleManagement(clusterClient, clusterName, addonName, namespace)
	})

	Context("Resource manipulation returns no errors", func() {
		It("should return no error", func() {
			Expect(err).NotTo(HaveOccurred())
		})
	})
	Context("Patching resources returns unknown error", func() {
		BeforeEach(func() {
			clusterClient.PatchResourceReturns(errors.New("Unknown error"))
		})
		It("should return error", func() {
			Expect(err).To(HaveOccurred())
		})
	})
	Context("Patching resources returns not found", func() {
		BeforeEach(func() {
			clusterClient.PatchResourceReturns(notFoundError)
		})
		It("should return no error", func() {
			Expect(err).NotTo(HaveOccurred())
		})
	})

})

var _ = Describe("Test NoopDeletePackageInstall", func() {
	var (
		clusterClient *fakes.ClusterClient
		err           error
		addonName     string
		namespace     string
		notFoundError = apierrors.NewNotFound(
			schema.GroupResource{Group: "fakeGroup", Resource: "fakeGroupResource"},
			"fakeGroupResource")
	)

	BeforeEach(func() {
		clusterClient = &fakes.ClusterClient{}
		addonName = addonsManager
		clusterClient.PatchResourceReturns(nil)
		clusterClient.DeleteResourceReturns(nil)
	})

	JustBeforeEach(func() {
		err = NoopDeletePackageInstall(clusterClient, namespace, fmt.Sprintf("tanzu-%s", addonName))
	})

	Context("Resource manipulation returns no errors", func() {
		It("should return no error", func() {
			Expect(err).NotTo(HaveOccurred())
		})
	})
	Context("Patching resources returns unknown errors", func() {
		BeforeEach(func() {
			clusterClient.PatchResourceReturns(errors.New("Unknown error"))
		})
		It("should return error", func() {
			Expect(err).To(HaveOccurred())
		})
	})
	Context("Patch returns not found", func() {
		BeforeEach(func() {
			clusterClient.PatchResourceReturns(notFoundError)
		})
		It("should return no error", func() {
			Expect(err).NotTo(HaveOccurred())
		})
	})
	Context("Deleting resources returns  unknown errors", func() {
		BeforeEach(func() {
			clusterClient.DeleteResourceReturns(errors.New("Unknown error"))
		})
		It("should return error", func() {
			Expect(err).To(HaveOccurred())
		})
	})
	Context("Deleting resources returns not found", func() {
		BeforeEach(func() {
			clusterClient.DeleteResourceReturns(notFoundError)
		})
		It("should return no error", func() {
			Expect(err).NotTo(HaveOccurred())
		})
	})

})

var _ = Describe("Test DeleteAddonSecret", func() {
	var (
		clusterClient *fakes.ClusterClient
		err           error
		addonName     string
		namespace     string
		notFoundError = apierrors.NewNotFound(
			schema.GroupResource{Group: "fakeGroup", Resource: "fakeGroupResource"},
			"fakeGroupResource")
	)

	BeforeEach(func() {
		clusterClient = &fakes.ClusterClient{}
		addonName = addonsManager
		clusterClient.GetResourceReturns(nil)
		clusterClient.UpdateResourceReturns(nil)
		clusterClient.PatchResourceReturns(nil)
	})

	JustBeforeEach(func() {
		err = DeleteAddonSecret(clusterClient, fmt.Sprintf("fake-cluster-tanzu-%s-addon", addonName), namespace)
	})
	Context("Resource manipulation returns no errors", func() {
		It("should return no error", func() {
			Expect(err).NotTo(HaveOccurred())
		})
	})
	Context("Getting resources returns unknown errors", func() {
		BeforeEach(func() {
			clusterClient.GetResourceReturns(errors.New("Unknown error"))
		})
		It("should return error", func() {
			Expect(err).To(HaveOccurred())
		})
	})
	Context("Getting resources returns not found", func() {
		BeforeEach(func() {
			clusterClient.GetResourceReturns(notFoundError)
		})
		It("should return no error", func() {
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("Updating resources returns unknown errors", func() {
		BeforeEach(func() {
			clusterClient.UpdateResourceReturns(errors.New("Unknown error"))
		})
		It("should return error", func() {
			Expect(err).To(HaveOccurred())
		})
	})
	Context("Updating resources returns not found", func() {
		BeforeEach(func() {
			clusterClient.UpdateResourceReturns(notFoundError)
		})
		It("should return no error", func() {
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("Deleting resources returns unknown errors", func() {
		BeforeEach(func() {
			clusterClient.DeleteResourceReturns(errors.New("Unknown error"))
		})
		It("should return error", func() {
			Expect(err).To(HaveOccurred())
		})
	})
	Context("Deleting resources returns not found", func() {
		BeforeEach(func() {
			clusterClient.DeleteResourceReturns(notFoundError)
		})
		It("should return no error", func() {
			Expect(err).NotTo(HaveOccurred())
		})
	})

})

var _ = Describe("Test AddonSecretExists", func() {

	var (
		clusterClient               *fakes.ClusterClient
		err                         error
		addonName                   string
		pauseAddonsManagerLifecycle bool
		notFoundError               = apierrors.NewNotFound(
			schema.GroupResource{Group: "fakeGroup", Resource: "fakeGroupResource"},
			"fakeGroupResource")
	)

	BeforeEach(func() {
		clusterClient = &fakes.ClusterClient{}
		addonName = addonsManager
		clusterClient.GetResourceReturns(nil)
		clusterClient.UpdateResourceReturns(nil)
		clusterClient.PatchResourceReturns(nil)
	})

	JustBeforeEach(func() {
		pauseAddonsManagerLifecycle, err = AddonSecretExists(clusterClient, fmt.Sprintf("fake-cluster-tanzu-%s-addon", addonName), constants.TkgNamespace)
	})
	Context("Getting resources returns no errors", func() {
		It("should return true, and no error", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(pauseAddonsManagerLifecycle).To(BeTrue())
		})
	})
	Context("Getting resources returns unknown errors", func() {
		BeforeEach(func() {
			clusterClient.GetResourceReturns(errors.New("Unknown error"))
		})
		It("should return false and error", func() {
			Expect(err).To(HaveOccurred())
			Expect(pauseAddonsManagerLifecycle).To(BeFalse())
		})
	})
	Context("Getting resources returns not found", func() {
		BeforeEach(func() {
			clusterClient.GetResourceReturns(notFoundError)
		})
		It("should return false and no error", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(pauseAddonsManagerLifecycle).To(BeFalse())
		})
	})

})

var _ = Describe("Test DeleteLegacyAkoOperatorPackageInstall", func() {
	var (
		clusterClient *fakes.ClusterClient
		err           error
		notFoundError = apierrors.NewNotFound(
			schema.GroupResource{Group: "fakeGroup", Resource: "fakeGroupResource"},
			"fakeGroupResource")
	)

	BeforeEach(func() {
		clusterClient = &fakes.ClusterClient{}
		clusterClient.PatchResourceReturns(nil)
		clusterClient.DeleteResourceReturns(nil)
	})

	JustBeforeEach(func() {
		err = DeleteLegacyAkoOperatorPackageInstall(clusterClient, fmt.Sprintf("fake-cluster-%s-addon", akoOperatorName))
	})

	Context("Resource manipulation returns no errors", func() {
		It("should return no error", func() {
			Expect(err).NotTo(HaveOccurred())
		})
	})
	Context("Patching resources returns unknown errors", func() {
		BeforeEach(func() {
			clusterClient.PatchResourceReturns(errors.New("Unknown error"))
		})
		It("should return error", func() {
			Expect(err).To(HaveOccurred())
		})
	})
	Context("Patching resources returns unknown errors", func() {
		BeforeEach(func() {
			clusterClient.PatchResourceReturnsOnCall(1, errors.New("Unknown error"))
		})
		It("should return error", func() {
			Expect(err).To(HaveOccurred())
		})
	})
	Context("Patch returns not found", func() {
		BeforeEach(func() {
			clusterClient.PatchResourceReturns(notFoundError)
		})
		It("should return no error", func() {
			Expect(err).NotTo(HaveOccurred())
		})
	})
	Context("Deleting resources returns  unknown errors", func() {
		BeforeEach(func() {
			clusterClient.DeleteResourceReturns(errors.New("Unknown error"))
		})
		It("should return error", func() {
			Expect(err).To(HaveOccurred())
		})
	})
	Context("Deleting resources returns not found", func() {
		BeforeEach(func() {
			clusterClient.DeleteResourceReturns(notFoundError)
		})
		It("should return no error", func() {
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
