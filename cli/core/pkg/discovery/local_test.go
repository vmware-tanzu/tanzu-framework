// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	pluginPath = "/tmp/test-discovery-plugin"
)

var _ = Describe("Unit tests for local discovery", func() {
	discovery := &LocalDiscovery{
		path: pluginPath,
		name: "local",
	}

	It("test get local discovery name", func() {
		name := discovery.Name()
		Expect(name).To(Equal(discovery.name))
	})

	It("test get local discovery type", func() {
		localType := discovery.Type()
		Expect(localType).To(Equal("local"))
	})

	It("test manifest deserialization", func() {
		createTestLocalPluginFile()
		expectedPlugin := &Plugin{
			Name:               "test-plugin",
			Description:        "test-plugin",
			RecommendedVersion: "1.0.0",
			Optional:           true,
		}
		plugins, err := discovery.Manifest()
		Expect(err).ToNot(HaveOccurred())

		deleteLocalPluginFile()
		Expect(plugins).ToNot(BeNil())
		Expect(len(plugins)).ToNot(Equal(0))
		Expect(plugins[0].Name).To(Equal(expectedPlugin.Name))
		Expect(plugins[0].Description).To(Equal(expectedPlugin.Description))
		Expect(plugins[0].RecommendedVersion).To(Equal(expectedPlugin.RecommendedVersion))
		Expect(plugins[0].Optional).To(Equal(expectedPlugin.Optional))
	})
})

func createTestLocalPluginFile() {
	contents := `
inline:
metadata:
  name: test-plugin
spec:
  description: test-plugin
  recommendedVersion: 1.0.0
  optional: true`
	_, err := os.Stat(pluginPath)
	if err != nil && os.IsNotExist(err) {
		err := os.Mkdir(pluginPath, 0777)
		Expect(err).ToNot(HaveOccurred())
	}

	err = os.WriteFile(pluginPath+"/test-plugin.yaml", []byte(contents), 0600)
	Expect(err).ToNot(HaveOccurred())
}

func deleteLocalPluginFile() {
	err := os.Remove(pluginPath + "/test-plugin.yaml")
	Expect(err).ToNot(HaveOccurred())
	err = os.Remove(pluginPath)
	Expect(err).ToNot(HaveOccurred())
}
