// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery_test

import (
	"errors"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/discovery"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/plugin"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes"
	fakehelper "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes/helper"
)

func TestClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Discovery Suite")
}

var _ = Describe("Unit tests for kubernetes discovery", func() {
	var (
		err                  error
		currentClusterClient *fakes.ClusterClient
		kd                   *discovery.KubernetesDiscovery
		plugins              []plugin.Discovered
		cliplugins           []v1alpha1.CLIPlugin
	)

	Describe("When Getting Discovered plugins from k8s cluster", func() {
		BeforeEach(func() {
			currentClusterClient = &fakes.ClusterClient{}
			kd = &discovery.KubernetesDiscovery{}
		})

		JustBeforeEach(func() {
			plugins, err = kd.GetDiscoveredPlugins(currentClusterClient)
		})

		Context("When CLIPlugin CRD verification throws error", func() {
			BeforeEach(func() {
				currentClusterClient.VerifyCLIPluginCRDReturns(false, errors.New("fake error"))
			})
			It("should return empty plugin list without an error", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(len(plugins)).To(Equal(0))
			})
		})
		Context("When CLIPlugin CRD verification returns false and no error", func() {
			BeforeEach(func() {
				currentClusterClient.VerifyCLIPluginCRDReturns(false, nil)
			})
			It("should return empty plugin list without an error", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(len(plugins)).To(Equal(0))
			})
		})
		Context("When CLIPlugin CRD verification returns true with error", func() {
			BeforeEach(func() {
				currentClusterClient.VerifyCLIPluginCRDReturns(true, errors.New("fake error"))
			})
			It("should return empty plugin list without an error", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(len(plugins)).To(Equal(0))
			})
		})

		Context("When ListCLIPluginResources returns error", func() {
			BeforeEach(func() {
				currentClusterClient.VerifyCLIPluginCRDReturns(true, nil)
				currentClusterClient.ListCLIPluginResourcesReturns(nil, errors.New("fake error"))
			})
			It("should return empty plugin list with an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(len(plugins)).To(Equal(0))
			})
		})
		Context("When ListCLIPluginResources list of CLIPlugin resources", func() {
			BeforeEach(func() {
				cliplugin1 := fakehelper.NewCLIPlugin(fakehelper.TestCLIPluginOption{Name: "plugin1", Description: "plugin1 desc", RecommendedVersion: "v0.0.1"})
				cliplugin2 := fakehelper.NewCLIPlugin(fakehelper.TestCLIPluginOption{Name: "plugin2", Description: "plugin2 desc", RecommendedVersion: "v0.0.2"})
				cliplugins = []v1alpha1.CLIPlugin{}
				cliplugins = append(cliplugins, cliplugin1, cliplugin2)
				currentClusterClient.VerifyCLIPluginCRDReturns(true, nil)
				currentClusterClient.ListCLIPluginResourcesReturns(cliplugins, nil)
				currentClusterClient.GetCLIPluginImageRepositoryOverrideReturns(map[string]string{}, nil)
			})
			It("should return ordered list of plugins without error", func() {
				Expect(len(plugins)).To(Equal(2))
				Expect(plugins[0].Name).To(Equal("plugin1"))
				Expect(plugins[0].RecommendedVersion).To(Equal("v0.0.1"))
				Expect(plugins[1].Name).To(Equal("plugin2"))
				Expect(plugins[1].RecommendedVersion).To(Equal("v0.0.2"))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("When ListCLIPluginResources list of CLIPlugin resources but GetCLIPluginImageRepositoryOverrideReturns return error", func() {
			BeforeEach(func() {
				cliplugin1 := fakehelper.NewCLIPlugin(fakehelper.TestCLIPluginOption{Name: "plugin1", Description: "plugin1 desc", RecommendedVersion: "v0.0.1"})
				cliplugin2 := fakehelper.NewCLIPlugin(fakehelper.TestCLIPluginOption{Name: "plugin2", Description: "plugin2 desc", RecommendedVersion: "v0.0.2"})
				cliplugins = []v1alpha1.CLIPlugin{}
				cliplugins = append(cliplugins, cliplugin1, cliplugin2)
				currentClusterClient.VerifyCLIPluginCRDReturns(true, nil)
				currentClusterClient.ListCLIPluginResourcesReturns(cliplugins, nil)
				currentClusterClient.GetCLIPluginImageRepositoryOverrideReturns(map[string]string{}, errors.New("fake-error-image-repo-override"))
			})
			It("should return ordered list of plugins without error", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(plugins[0].Name).To(Equal("plugin1"))
				Expect(plugins[0].RecommendedVersion).To(Equal("v0.0.1"))
				Expect(plugins[1].Name).To(Equal("plugin2"))
				Expect(plugins[1].RecommendedVersion).To(Equal("v0.0.2"))
				artifact, errArtifact := plugins[0].Distribution.DescribeArtifact("v1.0.0", "darwin", "amd64")
				Expect(errArtifact).NotTo(HaveOccurred())
				Expect(artifact.Image).To(Equal("fake.image.repo.com/tkg/plugin/test-darwin-plugin:v1.4.0"))
			})
		})

		Context("When ListCLIPluginResources list of CLIPlugin resources and GetCLIPluginImageRepositoryOverrideReturns imageRepoOverride map ", func() {
			BeforeEach(func() {
				cliplugin1 := fakehelper.NewCLIPlugin(fakehelper.TestCLIPluginOption{Name: "plugin1", Description: "plugin1 desc", RecommendedVersion: "v0.0.1"})
				cliplugin2 := fakehelper.NewCLIPlugin(fakehelper.TestCLIPluginOption{Name: "plugin2", Description: "plugin2 desc", RecommendedVersion: "v0.0.2"})
				cliplugins = []v1alpha1.CLIPlugin{}
				cliplugins = append(cliplugins, cliplugin1, cliplugin2)
				imageOverrideMap := map[string]string{
					"fake.image.repo.com": "custom.repo.com",
				}
				currentClusterClient.VerifyCLIPluginCRDReturns(true, nil)
				currentClusterClient.ListCLIPluginResourcesReturns(cliplugins, nil)
				currentClusterClient.GetCLIPluginImageRepositoryOverrideReturns(imageOverrideMap, nil)
			})
			It("should return ordered list of plugins and image override should be done in the artifact", func() {
				Expect(len(plugins)).To(Equal(2))
				Expect(plugins[0].Name).To(Equal("plugin1"))
				Expect(plugins[0].RecommendedVersion).To(Equal("v0.0.1"))
				Expect(plugins[1].Name).To(Equal("plugin2"))
				Expect(plugins[1].RecommendedVersion).To(Equal("v0.0.2"))
				Expect(err).NotTo(HaveOccurred())
				artifact, errArtifact := plugins[0].Distribution.DescribeArtifact("v1.0.0", "darwin", "amd64")
				Expect(errArtifact).NotTo(HaveOccurred())
				Expect(artifact.Image).To(Equal("custom.repo.com/tkg/plugin/test-darwin-plugin:v1.4.0"))
				artifact, errArtifact = plugins[0].Distribution.DescribeArtifact("v1.0.0", "windows", "amd64")
				Expect(errArtifact).NotTo(HaveOccurred())
				Expect(artifact.Image).To(Equal("custom.repo.com/tkg/plugin/test-windows-plugin:v1.4.0"))
				artifact, errArtifact = plugins[0].Distribution.DescribeArtifact("v1.0.0", "linux", "amd64")
				Expect(errArtifact).NotTo(HaveOccurred())
				Expect(artifact.Image).To(Equal("custom.repo.com/tkg/plugin/test-linux-plugin:v1.4.0"))
			})
		})
	})
})
