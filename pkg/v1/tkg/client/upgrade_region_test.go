// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
	clusterctl "sigs.k8s.io/cluster-api/cmd/clusterctl/client"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes"
)

var _ = Describe("Unit tests for upgrade management cluster", func() {
	var (
		err                   error
		regionalClusterClient *fakes.ClusterClient
		providerUpgradeClient *fakes.ProvidersUpgradeClient
		tkgClient             *TkgClient
		upgradeClusterOptions UpgradeClusterOptions
		context               string
	)

	BeforeEach(func() {
		regionalClusterClient = &fakes.ClusterClient{}
		providerUpgradeClient = &fakes.ProvidersUpgradeClient{}

		tkgClient, err = CreateTKGClient("../fakes/config/config.yaml", testingDir, "../fakes/config/bom/tkg-bom-v1.3.1.yaml", 2*time.Millisecond)
		Expect(err).NotTo(HaveOccurred())

		context = "fakeContext"
		upgradeClusterOptions = UpgradeClusterOptions{
			ClusterName:       "fake-cluster-name",
			Namespace:         "fake-namespace",
			KubernetesVersion: newK8sVersion,
			IsRegionalCluster: true,
			Kubeconfig:        "../fakes/config/kubeconfig/config1.yaml",
		}
	})

	Describe("When upgrading management cluster", func() {
		BeforeEach(func() {
			newK8sVersion = "v1.18.0+vmware.1"
			currentK8sVersion = "v1.17.3+vmware.2"
			setupBomFile("../fakes/config/bom/tkg-bom-v1.3.1.yaml", testingDir)
			regionalClusterClient.IsPacificRegionalClusterReturns(false, nil)
		})

		Context("When upgrading management cluster providers", func() {
			JustBeforeEach(func() {
				err = tkgClient.DoProvidersUpgrade(regionalClusterClient, context, providerUpgradeClient, &upgradeClusterOptions)
			})
			Context("When reading upgrade information from BOM file fails", func() {
				BeforeEach(func() {
					updateDefaultBoMFileName(testingDir, "tkg-bom-v1.3.1-fake.yaml")
				})
				It("should return an error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("unable to read in configuration from BOM file"))
				})
			})
			Context("When getting upgrade information fails due to failure to get the current providers information", func() {
				BeforeEach(func() {
					regionalClusterClient.ListResourcesReturns(errors.New("fake ListResourceError"))
				})
				It("should return an error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("fake ListResourceError"))
				})
			})

			Context("When current providers versions are up to date", func() {
				BeforeEach(func() {
					regionalClusterClient.ListResourcesCalls(func(providers interface{}, options ...crtclient.ListOption) error {
						installedProviders, _ := providers.(*clusterctlv1.ProviderList)
						installedProviders.Items = []clusterctlv1.Provider{
							{ObjectMeta: metav1.ObjectMeta{Namespace: "capi-system", Name: "cluster-api"}, Type: "CoreProvider", Version: "v0.3.11", ProviderName: "cluster-api"},
							{ObjectMeta: metav1.ObjectMeta{Namespace: "capi-kubeadm-bootstrap-system", Name: "bootstrap-kubeadm"}, Type: "BootstrapProvider", Version: "v0.3.11", ProviderName: "kubeadm"},
							{ObjectMeta: metav1.ObjectMeta{Namespace: "capi-kubeadm-control-plane-system", Name: "control-plane-kubeadm"}, Type: "ControlPlaneProvider", Version: "v0.3.11", ProviderName: "kubeadm"},
							{ObjectMeta: metav1.ObjectMeta{Namespace: "capv-system", Name: "infrastructure-vsphere"}, Type: "InfrastructureProvider", Version: "v0.7.1", ProviderName: "vsphere"},
						}
						return nil
					})
				})
				It("should not apply the providers version upgrade", func() {
					Expect(providerUpgradeClient.ApplyUpgradeCallCount()).Should(Equal(1))
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("When providers version upgrade failed", func() {
				BeforeEach(func() {
					regionalClusterClient.ListResourcesCalls(func(providers interface{}, options ...crtclient.ListOption) error {
						installedProviders, _ := providers.(*clusterctlv1.ProviderList)
						installedProviders.Items = []clusterctlv1.Provider{
							{ObjectMeta: metav1.ObjectMeta{Namespace: "capi-system", Name: "cluster-api"}, Type: "CoreProvider", Version: "v0.3.3", ProviderName: "cluster-api"},
							{ObjectMeta: metav1.ObjectMeta{Namespace: "capi-kubeadm-bootstrap-system", Name: "bootstrap-kubeadm"}, Type: "BootstrapProvider", Version: "v0.3.3", ProviderName: "kubeadm"},
							{ObjectMeta: metav1.ObjectMeta{Namespace: "capi-kubeadm-control-plane-system", Name: "control-plane-kubeadm"}, Type: "ControlPlaneProvider", Version: "v0.3.3", ProviderName: "kubeadm"},
							{ObjectMeta: metav1.ObjectMeta{Namespace: "capv-system", Name: "infrastructure-vsphere"}, Type: "InfrastructureProvider", Version: "v0.6.3", ProviderName: "vsphere"},
						}
						return nil
					})
					providerUpgradeClient.ApplyUpgradeReturns(errors.New("fake-providers upgrade failed"))
				})
				It("should return error", func() {
					Expect(providerUpgradeClient.ApplyUpgradeCallCount()).Should(Equal(1))
					Expect(err.Error()).To(ContainSubstring("fake-providers upgrade failed"))
				})
			})
			Context("When providers current versions and the latest versions are same", func() {
				var applyUpgradeOptionsRecvd clusterctl.ApplyUpgradeOptions
				BeforeEach(func() {
					regionalClusterClient.ListResourcesCalls(func(providers interface{}, options ...crtclient.ListOption) error {
						installedProviders, _ := providers.(*clusterctlv1.ProviderList)
						installedProviders.Items = []clusterctlv1.Provider{
							{ObjectMeta: metav1.ObjectMeta{Namespace: "capi-system", Name: "cluster-api"}, Type: "CoreProvider", Version: "v0.3.3", ProviderName: "cluster-api"},
							{ObjectMeta: metav1.ObjectMeta{Namespace: "capi-kubeadm-bootstrap-system", Name: "bootstrap-kubeadm"}, Type: "BootstrapProvider", Version: "v0.3.3", ProviderName: "kubeadm"},
							{ObjectMeta: metav1.ObjectMeta{Namespace: "capi-kubeadm-control-plane-system", Name: "control-plane-kubeadm"}, Type: "ControlPlaneProvider", Version: "v0.3.3", ProviderName: "kubeadm"},
							{ObjectMeta: metav1.ObjectMeta{Namespace: "capv-system", Name: "infrastructure-vsphere"}, Type: "InfrastructureProvider", Version: "v0.6.3", ProviderName: "vsphere"},
						}
						return nil
					})
					providerUpgradeClient.ApplyUpgradeCalls(func(applyUpgradeOptions *clusterctl.ApplyUpgradeOptions) error {
						applyUpgradeOptionsRecvd = *applyUpgradeOptions
						return nil
					})
				})
				It("should still apply the providers version upgrade to the same versions", func() {
					Expect(applyUpgradeOptionsRecvd.CoreProvider).Should(Equal("capi-system/cluster-api:v0.3.11"))
					Expect(applyUpgradeOptionsRecvd.BootstrapProviders[0]).Should(Equal("capi-kubeadm-bootstrap-system/kubeadm:v0.3.11"))
					Expect(applyUpgradeOptionsRecvd.ControlPlaneProviders[0]).Should(Equal("capi-kubeadm-control-plane-system/kubeadm:v0.3.11"))
					Expect(applyUpgradeOptionsRecvd.InfrastructureProviders[0]).Should(Equal("capv-system/vsphere:v0.7.1"))
					Expect(providerUpgradeClient.ApplyUpgradeCallCount()).Should(Equal(1))
					Expect(err).NotTo(HaveOccurred())
				})
			})
			Context("When some(cluster-api, Infrastructure-vsphere) providers current versions are not up to date", func() {
				var applyUpgradeOptionsRecvd clusterctl.ApplyUpgradeOptions
				BeforeEach(func() {
					regionalClusterClient.ListResourcesCalls(func(providers interface{}, options ...crtclient.ListOption) error {
						installedProviders, _ := providers.(*clusterctlv1.ProviderList)
						installedProviders.Items = []clusterctlv1.Provider{
							{ObjectMeta: metav1.ObjectMeta{Namespace: "capi-system", Name: "cluster-api"}, Type: "CoreProvider", Version: "v0.3.2", ProviderName: "cluster-api"},
							{ObjectMeta: metav1.ObjectMeta{Namespace: "capi-kubeadm-bootstrap-system", Name: "bootstrap-kubeadm"}, Type: "BootstrapProvider", Version: "v0.3.11", ProviderName: "kubeadm"},
							{ObjectMeta: metav1.ObjectMeta{Namespace: "capi-kubeadm-control-plane-system", Name: "control-plane-kubeadm"}, Type: "ControlPlaneProvider", Version: "v0.3.11", ProviderName: "kubeadm"},
							{ObjectMeta: metav1.ObjectMeta{Namespace: "capv-system", Name: "infrastructure-vsphere"}, Type: "InfrastructureProvider", Version: "v0.6.2", ProviderName: "vsphere"},
						}
						return nil
					})
					providerUpgradeClient.ApplyUpgradeCalls(func(applyUpgradeOptions *clusterctl.ApplyUpgradeOptions) error {
						applyUpgradeOptionsRecvd = *applyUpgradeOptions
						return nil
					})
				})
				It("should apply the providers version upgrade only for the outdated providers to the latest versions", func() {
					Expect(applyUpgradeOptionsRecvd.CoreProvider).Should(Equal("capi-system/cluster-api:v0.3.11"))
					Expect(len(applyUpgradeOptionsRecvd.BootstrapProviders)).Should(Equal(0))
					Expect(len(applyUpgradeOptionsRecvd.ControlPlaneProviders)).Should(Equal(0))
					Expect(applyUpgradeOptionsRecvd.InfrastructureProviders[0]).Should(Equal("capv-system/vsphere:v0.7.1"))
					Expect(providerUpgradeClient.ApplyUpgradeCallCount()).Should(Equal(1))
					Expect(err).NotTo(HaveOccurred())
				})
			})
			Context("When providers current versions are out dated and providers upgraded successfully", func() {
				var applyUpgradeOptionsRecvd clusterctl.ApplyUpgradeOptions
				BeforeEach(func() {
					regionalClusterClient.ListResourcesCalls(func(providers interface{}, options ...crtclient.ListOption) error {
						installedProviders, _ := providers.(*clusterctlv1.ProviderList)
						installedProviders.Items = []clusterctlv1.Provider{
							{ObjectMeta: metav1.ObjectMeta{Namespace: "capi-system", Name: "cluster-api"}, Type: "CoreProvider", Version: "v0.3.1", ProviderName: "cluster-api"},
							{ObjectMeta: metav1.ObjectMeta{Namespace: "capi-kubeadm-bootstrap-system", Name: "bootstrap-kubeadm"}, Type: "BootstrapProvider", Version: "v0.3.1", ProviderName: "kubeadm"},
							{ObjectMeta: metav1.ObjectMeta{Namespace: "capi-kubeadm-control-plane-system", Name: "control-plane-kubeadm"}, Type: "ControlPlaneProvider", Version: "v0.3.1", ProviderName: "kubeadm"},
							{ObjectMeta: metav1.ObjectMeta{Namespace: "capv-system", Name: "infrastructure-vsphere"}, Type: "InfrastructureProvider", Version: "v0.6.1", ProviderName: "vsphere"},
						}
						return nil
					})
					providerUpgradeClient.ApplyUpgradeCalls(func(applyUpgradeOptions *clusterctl.ApplyUpgradeOptions) error {
						applyUpgradeOptionsRecvd = *applyUpgradeOptions
						return nil
					})
				})
				It("should upgrade providers to the latest versions successfully", func() {
					Expect(applyUpgradeOptionsRecvd.CoreProvider).Should(Equal("capi-system/cluster-api:v0.3.11"))
					Expect(applyUpgradeOptionsRecvd.BootstrapProviders[0]).Should(Equal("capi-kubeadm-bootstrap-system/kubeadm:v0.3.11"))
					Expect(applyUpgradeOptionsRecvd.ControlPlaneProviders[0]).Should(Equal("capi-kubeadm-control-plane-system/kubeadm:v0.3.11"))
					Expect(applyUpgradeOptionsRecvd.InfrastructureProviders[0]).Should(Equal("capv-system/vsphere:v0.7.1"))
					Expect(providerUpgradeClient.ApplyUpgradeCallCount()).Should(Equal(1))
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})
})
