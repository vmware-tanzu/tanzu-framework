// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"sigs.k8s.io/yaml"

	runtanzuv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	. "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes"
)

const bootstrapObject = "../fakes/config/clusterbootstrap.yaml"

var (
	fakeMgtClusterClient *fakes.ClusterClient
	fakeWcClusterClient  *fakes.ClusterClient
	timeout              time.Duration
	clusterBootstrap     *runtanzuv1alpha3.ClusterBootstrap
)

func init() {
	fakeMgtClusterClient = &fakes.ClusterClient{}
	fakeWcClusterClient = &fakes.ClusterClient{}
	timeout = time.Duration(1)

}

var _ = Describe("unit tests for monitor addon's packages installation", func() {
	Context("get bootstrap object for workload cluster", func() {
		When("bootstrap object exists should not return any error", func() {
			BeforeEach(func() {
				fakeMgtClusterClient.GetResourceReturns(nil)
			})
			It("should not return error", func() {
				_, err := GetClusterBootstrap(fakeMgtClusterClient, "cluster1", "namespace1")
				Expect(err).NotTo(HaveOccurred())
			})
		})
		When("bootstrap object not exists should return an error", func() {
			resourceNotExists := "resource not exists"
			BeforeEach(func() {
				fakeMgtClusterClient.GetResourceReturns(errors.New(resourceNotExists))
			})
			It("should return error", func() {
				_, err := GetClusterBootstrap(fakeMgtClusterClient, "cluster1", "namespace1")
				Expect(err).To(HaveOccurred())
			})
		})
	})
	Context("get packages from bootstrap object and monitor packages installation", func() {
		BeforeEach(func() {
			bs, _ := os.ReadFile(bootstrapObject)
			clusterBootstrap = &runtanzuv1alpha3.ClusterBootstrap{}
			Expect(yaml.Unmarshal(bs, clusterBootstrap)).To(Succeed(), "Failed to convert the cluster bootstrap input file to yaml")
		})
		When("package installation successful should not return error", func() {
			BeforeEach(func() {
				fakeMgtClusterClient.WaitForPackageInstallReturns(nil)
				fakeWcClusterClient.WaitForPackageInstallReturns(nil)
			})
			It("should not return error", func() {
				packages := GetCorePackagesFromClusterBootstrap(clusterBootstrap, constants.CorePackagesNamespaceInTKGM)
				err := MonitorAddonsCorePackageInstallation(fakeMgtClusterClient, fakeWcClusterClient, packages, timeout)
				pkg, ns, _ := fakeMgtClusterClient.WaitForPackageInstallArgsForCall(0)
				Expect(pkg).To(Equal(packages[0].ObjectMeta.Name))
				Expect(ns).To(Equal(packages[0].ObjectMeta.Namespace))
				pkg, ns, _ = fakeWcClusterClient.WaitForPackageInstallArgsForCall(0)
				Expect([]string{packages[1].ObjectMeta.Name, packages[2].ObjectMeta.Name, packages[3].ObjectMeta.Name}).Should(ContainElements(pkg))
				Expect(ns).To(Equal(packages[1].ObjectMeta.Namespace))
				Expect(err).NotTo(HaveOccurred())
			})
		})
		When("package installation not successful because of MC packages error, should return error", func() {
			packageNotFound := "package not found"
			BeforeEach(func() {
				fakeMgtClusterClient.WaitForPackageInstallReturns(errors.New(packageNotFound))
			})
			It("should return error", func() {
				packages := GetCorePackagesFromClusterBootstrap(clusterBootstrap, constants.CorePackagesNamespaceInTKGM)
				err := MonitorAddonsCorePackageInstallation(fakeMgtClusterClient, fakeWcClusterClient, packages, timeout)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(packageNotFound))
			})
		})
		When("package installation not successful because of WC packages error, should return error", func() {
			packageNotFound := "package not found"
			BeforeEach(func() {
				fakeWcClusterClient.WaitForPackageInstallReturns(errors.New(packageNotFound))
			})
			It("should return error", func() {
				packages := GetCorePackagesFromClusterBootstrap(clusterBootstrap, constants.CorePackagesNamespaceInTKGM)
				err := MonitorAddonsCorePackageInstallation(fakeMgtClusterClient, fakeWcClusterClient, packages, timeout)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(packageNotFound))
			})
		})
	})

})
