// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package cluster_test provides unit test cases for cluster package
package cluster_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/discovery"

	cluster "github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/cluster"
	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/fakes"
)

var (
	crtClientFake              *fakes.CrtClientFake
	discoveryClientFactoryFake *fakes.DiscoveryClientFactory
	dynamicClientFactoryFake   *fakes.DynamicClientFactory
	options                    cluster.Options
	optionsTest                cluster.Options
	kubeconfiFile              string
	clusterClient              cluster.Client
)

func TestCliCorePkgClusterSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cli/core/pkg/cluster Suite")
}

var _ = Describe("New Cluster Client Tests", func() {
	BeforeEach(func() {
		crtClientFake = &fakes.CrtClientFake{}
		discoveryClientFactoryFake = &fakes.DiscoveryClientFactory{}
		dynamicClientFactoryFake = &fakes.DynamicClientFactory{}
		options = cluster.NewOptions(crtClientFake, discoveryClientFactoryFake, dynamicClientFactoryFake)
		kubeconfiFile = "../fakes/config/kubeconfig1.yaml"
	})
	When("kubeconfig and context is valid ", func() {
		BeforeEach(func() {
			discoveryClientFactoryFake.NewDiscoveryClientForConfigReturns(&discovery.DiscoveryClient{}, nil)
			discoveryClientFactoryFake.ServerVersionReturns(nil, nil)
		})
		It("return cluster client", func() {
			client, err := cluster.NewClient(kubeconfiFile, "federal-context", options)
			Expect(client).NotTo(BeNil())
			Expect(err).To(BeNil())
		})
		It("initialize options with default values", func() {
			cluster.InitializeOptions(&optionsTest)
			Expect(optionsTest.CrtClient).NotTo(BeNil())
			Expect(optionsTest.DiscoveryClientFactory).NotTo(BeNil())
			Expect(optionsTest.DynamicClientFactory).NotTo(BeNil())
		})
		Context("when list plugin's don't return any plugins ", func() {
			BeforeEach(func() {
				discoveryClientFactoryFake.NewDiscoveryClientForConfigReturns(&discovery.DiscoveryClient{}, nil)
				discoveryClientFactoryFake.ServerVersionReturns(nil, nil)
				clusterClient, _ = cluster.NewClient(kubeconfiFile, "federal-context", options)
				crtClientFake.ListObjectsReturns(nil)
			})
			It("return empty plugins and no error", func() {
				plugins, err := clusterClient.ListCLIPluginResources()
				Expect(plugins).To(BeNil())
				Expect(err).To(BeNil())
			})
		})
		Context("when BuildClusterQuery() called", func() {
			BeforeEach(func() {
				discoveryClientFactoryFake.NewDiscoveryClientForConfigReturns(&discovery.DiscoveryClient{}, nil)
				discoveryClientFactoryFake.ServerVersionReturns(nil, nil)
				clusterClient, _ = cluster.NewClient(kubeconfiFile, "federal-context", options)
				crtClientFake.ListObjectsReturns(nil)
			})
			It("return clusterQuery object and no errors", func() {
				isExecuted, err := clusterClient.BuildClusterQuery()
				Expect(isExecuted).NotTo(BeNil())
				Expect(err).To(BeNil())
			})
		})
		Context("when GetCLIPluginImageRepositoryOverride() called", func() {
			BeforeEach(func() {
				discoveryClientFactoryFake.NewDiscoveryClientForConfigReturns(&discovery.DiscoveryClient{}, nil)
				discoveryClientFactoryFake.ServerVersionReturns(nil, nil)
				clusterClient, _ = cluster.NewClient(kubeconfiFile, "federal-context", options)
				crtClientFake.ListObjectsReturns(nil)
			})
			It("return empty map and no error", func() {
				mapObj, err := clusterClient.GetCLIPluginImageRepositoryOverride()
				Expect(len(mapObj)).To(Equal(0))
				Expect(err).To(BeNil())
			})
		})
	})

	When("kubeconfig is not valid ", func() {
		BeforeEach(func() {
			discoveryClientFactoryFake.NewDiscoveryClientForConfigReturns(&discovery.DiscoveryClient{}, nil)
			discoveryClientFactoryFake.ServerVersionReturns(nil, nil)
			kubeconfiFile = "invalidkubeconfigfile.yaml"
		})
		It("should return error for NewClient()", func() {
			client, err := cluster.NewClient(kubeconfiFile, "federal-context", options)
			Expect(client).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("Failed to load Kubeconfig file from \"invalidkubeconfigfile.yaml\""))
		})
	})

	Context("Create ClusterClientFactory", func() {
		It("return object for clusterClientFactory", func() {
			ccf := cluster.NewClusterClientFactory()
			Expect(ccf).NotTo(BeNil())
		})
	})

	Context("Create DiscoveryClientFactory", func() {
		It("return object for discoveryClientFactory", func() {
			dcf := cluster.NewDiscoveryClientFactory()
			Expect(dcf).NotTo(BeNil())
		})
	})

	Context("test ConsolidateImageRepoMaps() helper", func() {
		It("should consolidate imageRepoMaps", func() {
			cmList := &corev1.ConfigMapList{}
			cmList.Items = append(cmList.Items, ConfigMapObject())

			irm, err := cluster.ConsolidateImageRepoMaps(cmList)
			Expect(irm).NotTo(BeNil())
			Expect(len(irm)).To(Equal(2))
			Expect(err).To(BeNil())
		})
	})
})

func ConfigMapObject() corev1.ConfigMap {
	imageRepoMapString := `staging.repo.com: stage.custom.repo.com
prod.repo.com: prod.custom.repo.com`
	configMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "image-repository-override",
			Namespace: "tanzu-cli-system",
			Labels: map[string]string{
				"cli.tanzu.vmware.com/cliplugin-image-repository-override": "",
			},
		},
		Data: map[string]string{
			"imageRepoMap": imageRepoMapString,
		},
	}
	return configMap
}

/*
var _ = Context("New Cluster Client Tests", func() {
		BeforeEach(func() {
			discoveryClientFactoryFake.NewDiscoveryClientForConfigReturns(&discovery.DiscoveryClient{}, nil)
			discoveryClientFactoryFake.ServerVersionReturns(nil, nil)
		})
		It("return cluster client", func() {
			plugins, err := ListCLIPluginResources()
			Expect(plugins).NotTo(BeNil())
			Expect(err).To(BeNil())
		})
})
*/
