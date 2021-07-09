// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	. "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes"
)

var _ = Describe("Unit tests for addons upgrade", func() {
	var (
		err                   error
		regionalClusterClient *fakes.ClusterClient
		currentClusterClient  *fakes.ClusterClient
		tkgClient             *TkgClient
		upgradeAddonOptions   *UpgradeAddonOptions
		addonsToBeUpgraded    []string

		clusterTemplateBytes []byte
		clusterTemplateError error
		clusterConfigGetter  func(*CreateClusterOptions) ([]byte, error)

		isRegionalCluster bool
	)

	BeforeEach(func() {
		regionalClusterClient = &fakes.ClusterClient{}
		currentClusterClient = &fakes.ClusterClient{}
		tkgClient, err = CreateTKGClient("../fakes/config/config.yaml", testingDir, "../fakes/config/bom/tkg-bom-v1.3.1.yaml", 2*time.Millisecond)

		upgradeAddonOptions = &UpgradeAddonOptions{
			ClusterName:       "test-cluster",
			Namespace:         constants.DefaultNamespace,
			Kubeconfig:        "../fakes/config/kubeconfig/config1.yaml",
			IsRegionalCluster: isRegionalCluster,
		}
	})

	Describe("When upgrading addons", func() {
		BeforeEach(func() {
			setupBomFile("../fakes/config/bom/tkr-bom-v1.18.0+vmware.1-tkg.2.yaml", testingDir)
			setupBomFile("../fakes/config/bom/tkg-bom-v1.3.1.yaml", testingDir)
			regionalClusterClient.PatchResourceReturns(nil)
			regionalClusterClient.GetKCPObjectForClusterReturns(getDummyKCP(constants.DockerMachineTemplate), nil)
			regionalClusterClient.GetResourceReturns(nil)
			currentClusterClient.GetKubernetesVersionReturns(currentK8sVersion, nil)
			isRegionalCluster = true

			clusterConfigGetter = func(*CreateClusterOptions) ([]byte, error) {
				return clusterTemplateBytes, clusterTemplateError
			}
		})
		JustBeforeEach(func() {
			upgradeAddonOptions.IsRegionalCluster = isRegionalCluster
			upgradeAddonOptions.AddonNames = addonsToBeUpgraded
			err = tkgClient.DoUpgradeAddon(regionalClusterClient, currentClusterClient, upgradeAddonOptions, clusterConfigGetter)
		})
		Context("When unable to patch cluster object with cluster name label", func() {
			BeforeEach(func() {
				regionalClusterClient.PatchClusterObjectReturns(errors.New("fake-error"))
			})
			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unable to patch the cluster object with cluster name label"))
				Expect(err.Error()).To(ContainSubstring("fake-error"))
			})
		})
		Context("When unable to get KCP object for cluster", func() {
			BeforeEach(func() {
				regionalClusterClient.GetKCPObjectForClusterReturns(nil, errors.New("fake-error"))
			})
			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unable to find control plane node object for cluster"))
				Expect(err.Error()).To(ContainSubstring("fake-error"))
			})
		})

		Context("When invalid addon name is passed", func() {
			BeforeEach(func() {
				addonsToBeUpgraded = []string{"invalid-name"}
			})
			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("upgrade of 'invalid-name' component is not supported"))
			})
		})

		Context("When cluster config getter returns error", func() {
			BeforeEach(func() {
				addonsToBeUpgraded = []string{"metadata/tkg"}
				clusterTemplateError = errors.New("fake-error")
			})
			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unable to get cluster configuration"))
				Expect(err.Error()).To(ContainSubstring("fake-error"))
			})
		})

		Context("When cluster template apply returns error", func() {
			BeforeEach(func() {
				addonsToBeUpgraded = []string{"metadata/tkg"}
				clusterTemplateError = nil
				currentClusterClient.ApplyReturns(errors.New("fake-error"))
			})
			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("error while upgrading additional component"))
				Expect(err.Error()).To(ContainSubstring("fake-error"))
			})
		})

		Context("When cluster template apply does not return error and upgrade is successful", func() {
			BeforeEach(func() {
				addonsToBeUpgraded = []string{"metadata/tkg"}
				clusterTemplateError = nil
				currentClusterClient.ApplyReturns(nil)
			})
			It("should not returns an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("When upgrading all addons", func() {
			BeforeEach(func() {
				addonsToBeUpgraded = []string{"metadata/tkg", "addons-management/kapp-controller", "addons-management/tanzu-addons-manager", "tkr/tkr-controller"}
				clusterTemplateError = nil
				currentClusterClient.ApplyReturns(nil)
				regionalClusterClient.ApplyReturns(nil)
			})
			It("should not returns an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("When trying to upgrade 'tkr/tkr-controller' on workload cluster", func() {
			BeforeEach(func() {
				addonsToBeUpgraded = []string{"tkr/tkr-controller"}
				clusterTemplateError = nil
				isRegionalCluster = false
				currentClusterClient.ApplyReturns(nil)
				regionalClusterClient.ApplyReturns(nil)
			})
			It("should returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("upgrade of 'tkr/tkr-controller' component is only supported on management cluster"))
			})
		})
	})
})
