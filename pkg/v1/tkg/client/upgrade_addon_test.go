// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	capi "sigs.k8s.io/cluster-api/api/v1alpha3"

	. "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/clusterclient"
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
		const (
			clusterName = "tkg-mgmt"
		)
		var (
			serviceCIDRs []string
			podCIDRs     []string
		)
		BeforeEach(func() {
			serviceCIDRs = []string{"1.2.3.4/16"}
			podCIDRs = []string{"2.3.4.5/16"}

			setupBomFile("../fakes/config/bom/tkr-bom-v1.18.0+vmware.1-tkg.2.yaml", testingDir)
			setupBomFile("../fakes/config/bom/tkg-bom-v1.3.1.yaml", testingDir)
			regionalClusterClient.PatchResourceReturns(nil)
			regionalClusterClient.GetKCPObjectForClusterReturns(getDummyKCP(constants.DockerMachineTemplate), nil)
			currentClusterClient.GetKubernetesVersionReturns(currentK8sVersion, nil)
			regionalClusterClient.GetCurrentKubeContextReturns("context", nil)
			regionalClusterClient.GetCurrentClusterNameReturns(clusterName, nil)
			regionalClusterClient.GetResourceCalls(func(cluster interface{}, resourceName, namespace string, postVerify clusterclient.PostVerifyrFunc, pollOptions *clusterclient.PollOptions) error {
				if cluster, ok := cluster.(*capi.Cluster); ok && resourceName == clusterName && namespace == upgradeAddonOptions.Namespace {
					cluster.Spec = capi.ClusterSpec{
						ClusterNetwork: &capi.ClusterNetwork{
							Services: &capi.NetworkRanges{
								CIDRBlocks: serviceCIDRs,
							},
							Pods: &capi.NetworkRanges{
								CIDRBlocks: podCIDRs,
							},
						},
					}
					return nil
				}
				return nil
			})
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

		Describe("When setting networking configuration", func() {
			It("sets the cluster CIDR in the TKGConfig", func() {
				clusterCIDR, err := tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterCIDR)
				Expect(err).NotTo(HaveOccurred())
				Expect(clusterCIDR).To(Equal("2.3.4.5/16"))
			})
			It("sets the service CIDR in the TKGConfig", func() {
				serviceCIDR, err := tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableServiceCIDR)
				Expect(err).NotTo(HaveOccurred())
				Expect(serviceCIDR).To(Equal("1.2.3.4/16"))
			})
			When("the cluster is ipv4", func() {
				It("sets the IPFamily to ipv4", func() {
					ipFamily, err := tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableIPFamily)
					Expect(err).NotTo(HaveOccurred())
					Expect(ipFamily).To(Equal("ipv4"))
				})
			})
			When("the cluster is ipv6", func() {
				BeforeEach(func() {
					serviceCIDRs = []string{"fd00::/32"}
					podCIDRs = []string{"fd01::/32"}
				})
				It("sets the IPFamily to ipv6", func() {
					ipFamily, err := tkgClient.TKGConfigReaderWriter().Get(constants.ConfigVariableIPFamily)
					Expect(err).NotTo(HaveOccurred())
					Expect(ipFamily).To(Equal("ipv6"))
				})
			})
		})
	})
})
