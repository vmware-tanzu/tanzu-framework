// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// plugin provides plugin command specific E2E test cases
package plugin

import (
	. "github.com/onsi/ginkgo"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/test/e2e/framework"
)

var _ = framework.CLICoreDescribe("[Tests:E2E][Feature:Plugin-lifecycle]", func() {
	var (
		tf           *framework.Framework
		plugins      []*framework.PluginMeta
		discoveryMap map[string]string
	)
	BeforeEach(func() {
		tf = framework.NewFramework()
		pm := framework.NewPluginMeta()
		plugins = make([]*framework.PluginMeta, 1)
		pm.SetName("dummy").SetTarget("k8s").SetVersion("1.0").SetDescription("its dummy plugin").SetSHA("345").SetGroup("admin").SetArch("amd64").SetOS("darwin").SetDiscoveryType("oci").SetHidden(false).SetOptional(false)
		plugins[0] = pm
		GenerateAndPublishScriptBasedPluginsToLocalOCIRepo(tf, plugins[:])
	})
	Context("OCI repository-based plugin lifecycle tests", func() {
		// Test case: add plugin bundle discovery urls as plugin discovery sources
		It("add plugin discovery sources", func() {
			discoveryMap = make(map[string]string)
			AddPluginDiscoveryURLToPluginDiscoverySources(tf, plugins[:], discoveryMap)
		})
		// Test case: List plugins and make sure all plugins added in setup node should be listed
		It("list and validate plugins", func() {
			ListAndValidatePlugins(tf, discoveryMap)
		})
	})
})
