// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes"
	fakehelper "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes/helper"
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
		tkgClient, err = CreateTKGClient("../fakes/config/config2.yaml", testingDir, "../fakes/config/bom/tkg-bom-v1.3.1.yaml", 2*time.Millisecond)

		upgradeAddonOptions = &UpgradeAddonOptions{
			ClusterName:       "test-cluster",
			Namespace:         constants.DefaultNamespace,
			Kubeconfig:        "../fakes/config/kubeconfig/config1.yaml",
			IsRegionalCluster: isRegionalCluster,
		}
	})

	Describe("When upgrading addons", func() {
		const (
			clusterName = "regional-cluster-2"
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
			regionalClusterClient.ListResourcesCalls(func(clusterList interface{}, options ...client.ListOption) error {
				if clusterList, ok := clusterList.(*capiv1alpha3.ClusterList); ok {
					clusterList.Items = []capiv1alpha3.Cluster{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name:      clusterName,
								Namespace: constants.DefaultNamespace,
							},
						},
					}
					return nil
				}
				return nil
			})

			regionalClusterClient.GetResourceCalls(func(cluster interface{}, resourceName, namespace string, postVerify clusterclient.PostVerifyrFunc, pollOptions *clusterclient.PollOptions) error {
				if cluster, ok := cluster.(*capiv1alpha3.Cluster); ok && resourceName == clusterName && namespace == constants.DefaultNamespace {
					cluster.Spec = capiv1alpha3.ClusterSpec{
						ClusterNetwork: &capiv1alpha3.ClusterNetwork{
							Services: &capiv1alpha3.NetworkRanges{
								CIDRBlocks: serviceCIDRs,
							},
							Pods: &capiv1alpha3.NetworkRanges{
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

// Test_RetrieveProxySettings tests a previous regression where the proxy settings
// for a workload cluster was being incorrectly inherited from the management cluster.
// Specifically, this test validates retreiving proxy settings for a workload cluster
// with different settings from the management cluster.
func Test_RetrieveProxySettings(t *testing.T) {
	RegisterTestingT(t)
	g := NewWithT(t)

	tkgClient, err := CreateTKGClient("../fakes/config/config2.yaml", fakehelper.CreateTempTestingDirectory(), "../fakes/config/bom/tkg-bom-v1.3.1.yaml", 2*time.Millisecond)
	g.Expect(err).To(BeNil())

	regionalClusterClient := &fakes.ClusterClient{}
	regionalClusterClient.ListResourcesCalls(func(clusterList interface{}, options ...client.ListOption) error {
		if clusterList, ok := clusterList.(*capiv1alpha3.ClusterList); ok {
			clusterList.Items = []capiv1alpha3.Cluster{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "regional-cluster-2",
						Namespace: constants.DefaultNamespace,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "workload-cluster",
						Namespace: constants.DefaultNamespace,
					},
				},
			}
			return nil
		}
		return nil
	})

	regionalClusterClient.GetResourceStub = func(obj interface{}, name string, namespace string, verifyFunc clusterclient.PostVerifyrFunc, pollOptions *clusterclient.PollOptions) error {
		if name == constants.KappControllerConfigMapName {
			cm := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      constants.KappControllerConfigMapName,
					Namespace: constants.KappControllerNamespace,
				},
				Data: map[string]string{
					"httpProxy":  "http://10.0.0.1:8080",
					"httpsProxy": "http://10.0.0.1:8080",
					"noProxy":    "127.0.0.1,foo.com",
				},
			}
			cm.DeepCopyInto(obj.(*corev1.ConfigMap))
		}

		return nil
	}

	workloadClusterClient := &fakes.ClusterClient{}
	workloadClusterClient.GetResourceStub = func(obj interface{}, name string, namespace string, verifyFunc clusterclient.PostVerifyrFunc, pollOptions *clusterclient.PollOptions) error {
		if name == constants.KappControllerConfigMapName {
			cm := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      constants.KappControllerConfigMapName,
					Namespace: constants.KappControllerNamespace,
				},
				Data: map[string]string{
					"httpProxy":  "http://10.0.0.1:8081",
					"httpsProxy": "http://10.0.0.1:8081",
					"noProxy":    "127.0.0.1,bar.com",
				},
			}
			cm.DeepCopyInto(obj.(*corev1.ConfigMap))
		}

		return nil
	}

	// validate retrieving management cluster settings
	err = tkgClient.RetrieveRegionalClusterConfiguration(regionalClusterClient)
	g.Expect(err).To(BeNil())

	httpProxy, err := tkgClient.TKGConfigReaderWriter().Get(constants.TKGHTTPProxy)
	g.Expect(httpProxy).To(Equal("http://10.0.0.1:8080"))
	g.Expect(err).To(BeNil())
	httpsProxy, err := tkgClient.TKGConfigReaderWriter().Get(constants.TKGHTTPSProxy)
	g.Expect(httpsProxy).To(Equal("http://10.0.0.1:8080"))
	g.Expect(err).To(BeNil())
	noProxy, err := tkgClient.TKGConfigReaderWriter().Get(constants.TKGNoProxy)
	g.Expect(noProxy).To(Equal("127.0.0.1,foo.com"))
	g.Expect(err).To(BeNil())

	// validate retrieving workload cluster settings
	err = tkgClient.RetrieveWorkloadClusterConfiguration(regionalClusterClient, workloadClusterClient, "workload-cluster", constants.DefaultNamespace)
	g.Expect(err).To(BeNil())

	httpProxy, err = tkgClient.TKGConfigReaderWriter().Get(constants.TKGHTTPProxy)
	g.Expect(httpProxy).To(Equal("http://10.0.0.1:8081"))
	g.Expect(err).To(BeNil())
	httpsProxy, err = tkgClient.TKGConfigReaderWriter().Get(constants.TKGHTTPSProxy)
	g.Expect(httpsProxy).To(Equal("http://10.0.0.1:8081"))
	g.Expect(err).To(BeNil())
	noProxy, err = tkgClient.TKGConfigReaderWriter().Get(constants.TKGNoProxy)
	g.Expect(noProxy).To(Equal("127.0.0.1,bar.com"))
	g.Expect(err).To(BeNil())
}
