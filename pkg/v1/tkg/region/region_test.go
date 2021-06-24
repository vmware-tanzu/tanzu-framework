// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package region_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/constants"
	. "github.com/vmware-tanzu-private/core/pkg/v1/tkg/region"
)

func TestClusterClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Region manager Suite")
}

var _ = Describe("Region manager", func() {
	var (
		manager       Manager
		err           error
		tkgConfigPath string
	)
	const (
		fakeConfigYAMLFilePath       = "../fakes/config/config.yaml"
		fakeConfig2YAMLFilePath      = "../fakes/config/config2.yaml"
		RegionalCluster2             = "regional-cluster-2"
		User1RegionalCluster2Context = "user1@regional-cluster-2-context"
	)

	Describe("ListRegions", func() {
		var regions []RegionContext
		JustBeforeEach(func() {
			manager, err = New(tkgConfigPath)
			Expect(err).ToNot(HaveOccurred())
			regions, err = manager.ListRegionContexts()
		})

		Context("When regions node is not present in tkg config file", func() {
			BeforeEach(func() {
				tkgConfigPath = fakeConfigYAMLFilePath
			})

			It("should return no regions", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(len(regions)).To(Equal(0))
			})
		})

		Context("When regions are set in tkg config file", func() {
			BeforeEach(func() {
				tkgConfigPath = fakeConfig2YAMLFilePath
			})

			It("should return egions", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(len(regions)).To(Equal(3))
			})
		})
	})

	Describe("DeleteRegion", func() {
		var (
			clusterName  string
			originalFile []byte
		)
		JustBeforeEach(func() {
			originalFile, err = os.ReadFile(tkgConfigPath)
			Expect(err).ToNot(HaveOccurred())
			manager, err = New(tkgConfigPath)
			Expect(err).ToNot(HaveOccurred())
			err = manager.DeleteRegionContext(clusterName)
		})

		Context("When cluster does not exist in tkg config file", func() {
			BeforeEach(func() {
				tkgConfigPath = fakeConfig2YAMLFilePath
				clusterName = "regional-cluster-3"
			})
			It("should not have any impact on tkg config file", func() {
				Expect(err).ToNot(HaveOccurred())

				regions, err := manager.ListRegionContexts()
				Expect(err).ToNot(HaveOccurred())
				Expect(len(regions)).To(Equal(3))
			})
		})

		Context("When given cluster exists in tkg config file", func() {
			BeforeEach(func() {
				tkgConfigPath = fakeConfig2YAMLFilePath
				clusterName = RegionalCluster2
			})
			It("the cluster info should be deleted from tkg config file ", func() {
				Expect(err).ToNot(HaveOccurred())

				regions, err := manager.ListRegionContexts()
				Expect(err).ToNot(HaveOccurred())
				Expect(len(regions)).To(Equal(1))

				_, err = manager.GetCurrentContext()
				Expect(err).To(HaveOccurred())
			})
		})

		AfterEach(func() {
			err = os.WriteFile(tkgConfigPath, originalFile, constants.ConfigFilePermissions)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("SaveRegion", func() {
		var (
			region       RegionContext
			originalFile []byte
		)
		JustBeforeEach(func() {
			originalFile, err = os.ReadFile(tkgConfigPath)
			Expect(err).ToNot(HaveOccurred())
			manager, err = New(tkgConfigPath)
			Expect(err).ToNot(HaveOccurred())
			err = manager.SaveRegionContext(region)
		})

		Context("when regions node is not present in tkg config file", func() {
			BeforeEach(func() {
				tkgConfigPath = fakeConfigYAMLFilePath
				region = RegionContext{
					ClusterName:    "regional-cluster-3",
					ContextName:    "user2@regional-cluster-3-context",
					SourceFilePath: "path/to/kubeconfig",
				}
			})

			It("should create the regions node and save the cluster info into tkg config", func() {
				Expect(err).ToNot(HaveOccurred())

				regions, err := manager.ListRegionContexts()
				Expect(err).ToNot(HaveOccurred())
				Expect(len(regions)).To(Equal(1))
			})
		})

		Context("when a region with the same cluster name and context name already exists", func() {
			BeforeEach(func() {
				tkgConfigPath = fakeConfig2YAMLFilePath
				region = RegionContext{
					ClusterName:    "regional-cluster-1",
					ContextName:    "user1@regional-cluster-1-context",
					SourceFilePath: "path/to/kubeconfig",
				}
			})

			It("return a duplicate error", func() {
				Expect(err).To(HaveOccurred())

				regions, err := manager.ListRegionContexts()
				Expect(err).ToNot(HaveOccurred())
				Expect(len(regions)).To(Equal(3))
			})
		})

		AfterEach(func() {
			err = os.WriteFile(tkgConfigPath, originalFile, constants.ConfigFilePermissions)
			Expect(err).ToNot(HaveOccurred())
		})
	})
	Describe("UpsertRegion", func() {
		var (
			region       RegionContext
			originalFile []byte
		)
		JustBeforeEach(func() {
			originalFile, err = os.ReadFile(tkgConfigPath)
			Expect(err).ToNot(HaveOccurred())
			manager, err = New(tkgConfigPath)
			Expect(err).ToNot(HaveOccurred())
			err = manager.UpsertRegionContext(region)
		})

		Context("when regions node is not present in tkg config file", func() {
			BeforeEach(func() {
				tkgConfigPath = fakeConfigYAMLFilePath
				region = RegionContext{
					ClusterName:    "regional-cluster-3",
					ContextName:    "user2@regional-cluster-3-context",
					SourceFilePath: "path/to/kubeconfig",
				}
			})

			It("should create the regions node and save the cluster info into tkg config", func() {
				Expect(err).ToNot(HaveOccurred())

				regions, err := manager.ListRegionContexts()
				Expect(err).ToNot(HaveOccurred())
				Expect(len(regions)).To(Equal(1))
			})
		})

		Context("when a region with the same cluster name and context name already exists", func() {
			BeforeEach(func() {
				tkgConfigPath = fakeConfig2YAMLFilePath
				region = RegionContext{
					ClusterName:    "regional-cluster-1",
					ContextName:    "user1@regional-cluster-1-context",
					SourceFilePath: "newpath/to/kubeconfig",
					Status:         Success,
				}
			})

			It("should not return error, but update the existing region context", func() {
				Expect(err).ToNot(HaveOccurred())

				regions, err := manager.ListRegionContexts()
				Expect(err).ToNot(HaveOccurred())
				Expect(len(regions)).To(Equal(3))
				Expect(regions[0]).To(Equal(region))
			})
		})

		AfterEach(func() {
			err = os.WriteFile(tkgConfigPath, originalFile, constants.ConfigFilePermissions)
			Expect(err).ToNot(HaveOccurred())
		})
	})
	Describe("SetCurrentContext", func() {
		var (
			contextName  string
			clusterName  string
			originalFile []byte
		)

		JustBeforeEach(func() {
			originalFile, err = os.ReadFile(tkgConfigPath)
			Expect(err).ToNot(HaveOccurred())
			manager, err = New(tkgConfigPath)
			Expect(err).ToNot(HaveOccurred())
			err = manager.SetCurrentContext(clusterName, contextName)
		})

		Context("When context does not exist in tkg config file", func() {
			BeforeEach(func() {
				tkgConfigPath = fakeConfigYAMLFilePath
				clusterName = RegionalCluster2
				contextName = "user2@regional-cluster-2-context"
			})

			It("it should return an context not found error", func() {
				Expect(err).To(HaveOccurred())
			})
		})

		Context("When multiple region with same cluster name is set", func() {
			BeforeEach(func() {
				tkgConfigPath = fakeConfig2YAMLFilePath
				clusterName = RegionalCluster2
				contextName = ""
			})
			It("should return an error, indicating context name should be specified", func() {
				Expect(err).To(HaveOccurred())
			})
		})

		Context("When current-region-context key has no value", func() {
			BeforeEach(func() {
				tkgConfigPath = "../fakes/config/config3.yaml"
				clusterName = RegionalCluster2
				contextName = User1RegionalCluster2Context
			})

			It("should create value node and set the context as current context", func() {
				Expect(err).ToNot(HaveOccurred())
				context, err := manager.GetCurrentContext()
				Expect(err).ToNot(HaveOccurred())
				Expect(context.ContextName).To(Equal(contextName))
			})
		})

		Context("When current-region-context key does not exist", func() {
			BeforeEach(func() {
				tkgConfigPath = "../fakes/config/config8.yaml"
				clusterName = "regional-cluster-2"
				contextName = "user1@regional-cluster-2-context"
			})

			It("should create node and set the context as current context", func() {
				Expect(err).ToNot(HaveOccurred())
				context, err := manager.GetCurrentContext()
				Expect(err).ToNot(HaveOccurred())
				Expect(context.ContextName).To(Equal(contextName))
			})
		})

		Context("When context is present in tkg config file", func() {
			BeforeEach(func() {
				tkgConfigPath = fakeConfig2YAMLFilePath
				clusterName = RegionalCluster2
				contextName = User1RegionalCluster2Context
			})

			It("should set the context as current context", func() {
				Expect(err).ToNot(HaveOccurred())
				context, err := manager.GetCurrentContext()
				Expect(err).ToNot(HaveOccurred())
				Expect(context.ContextName).To(Equal(contextName))
			})
		})

		AfterEach(func() {
			err = os.WriteFile(tkgConfigPath, originalFile, constants.ConfigFilePermissions)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("GetCurrentContext", func() {
		var regionContext RegionContext
		JustBeforeEach(func() {
			manager, err = New(tkgConfigPath)
			Expect(err).ToNot(HaveOccurred())
			regionContext, err = manager.GetCurrentContext()
		})
		Context("When current context is not set in tkg config file", func() {
			BeforeEach(func() {
				tkgConfigPath = fakeConfigYAMLFilePath
			})

			It("should return current context not set error", func() {
				Expect(err).To(HaveOccurred())
			})
		})

		Context("When current context is set in tkg config file", func() {
			BeforeEach(func() {
				tkgConfigPath = fakeConfig2YAMLFilePath
			})
			It("should return the current context", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(regionContext.ContextName).To(Equal("user2@regional-cluster-2-context"))
			})
		})
	})
})
