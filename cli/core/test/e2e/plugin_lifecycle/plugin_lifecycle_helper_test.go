// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// plugin provides plugin command specific E2E test cases
package plugin

import (
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/test/e2e/framework"
)

// GenerateAndPublishScriptBasedPluginsToLocalOCIRepo for given plugin metadata, it generates script based plugins and publishes the plugin bundles to local OCI repository, and updates the distribution and discovery urls in plugin metadata.
func GenerateAndPublishScriptBasedPluginsToLocalOCIRepo(tf *framework.Framework, plugins []*framework.PluginMeta) {
	// Initialize config
	err := tf.Config.ConfigInit()
	Expect(err).To(BeNil(), "should initialize config without error")
	// Generate script based plugins
	_, errs := tf.PluginOps.GeneratePluginBinaries(plugins[:])
	for _, err := range errs {
		Expect(err).To(BeNil(), "should not occur any error while generating the plugin binaries")
	}
	// Publish script based plugin binaries to local OCI repository
	_, errs = tf.PluginOps.PublishPluginBinary(plugins[:])
	for _, err := range errs {
		Expect(err).To(BeNil(), "should not occur any error while publishing the plugin binaries")
	}
	// Generate plugin bundles by using the plugin binaries
	paths, errs := tf.PluginOps.GeneratePluginBundle(plugins[:])
	for i := 0; i < len(paths); i++ {
		Expect(errs[i]).To(BeNil(), "should not occur any error while generating the plugin bundle")
		Expect(paths[i]).NotTo(BeNil(), "should return a local files system path for plugin bundle")
	}
	// Publish the generated plugin bundles in previous steps to local oci repository
	discoveryUrls, errs := tf.PluginOps.PublishPluginBundle(plugins[:])
	for i, err := range errs {
		Expect(err).To(BeNil(), "should not occur any error while publishing the plugin bundle in local oci repo")
		Expect(discoveryUrls[i]).NotTo(BeNil(), "should return the discovery url for every plugin published to local oci repository")
	}
}

// ListAndValidatePlugins lists plugins, validate plugin names in discoveryMap for given discovery name, and make sure find all plugin sources in discoveryMap
func ListAndValidatePlugins(tf *framework.Framework, discoveryMap map[string]string) {
	pluginList, err := tf.Plugin.ListPlugins()
	Expect(err).To(BeNil())
	for _, plugin := range pluginList {
		if pluginName, ok := discoveryMap[plugin.Discovery]; ok {
			Expect(pluginName).To(Equal(plugin.Name), "Plugin name should be same as provided in the plugin metadata")
			delete(discoveryMap, plugin.Discovery)
		}
	}
	x := len(discoveryMap)
	Expect(x).To(Equal(0), "should find all plugin sources and plugins as in list plugin output as added in plugin discovery sources")
}

// AddPluginDiscoveryURLToPluginDiscoverySources adds the plugin discovery url to plugin discovery sources
func AddPluginDiscoveryURLToPluginDiscoverySources(tf *framework.Framework, pluginsMeta []*framework.PluginMeta, discoveryMap map[string]string) {
	n := len(pluginsMeta)
	for i := 0; i < n; i++ {
		discoveryName := "local-oci-" + framework.RandomString(3)
		discoveryMap[discoveryName] = pluginsMeta[i].GetName()
		do := &framework.DiscoveryOptions{
			Name:       discoveryName, //discovery
			SourceType: "oci",
			URI:        pluginsMeta[i].GetRegistryDiscoveryURL(),
		}
		err := tf.Plugin.AddPluginDiscoverySource(do)
		Expect(err).To(BeNil(), "should be able to add plugin discovery source without any error")
	}
	Expect(n).To(Equal(len(discoveryMap)), "should have a discovery source for every given pluginMeta")
}
