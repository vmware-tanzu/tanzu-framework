// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package unit

import (
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/vmware-tanzu/tanzu-framework/test/pkg/matchers"
	"github.com/vmware-tanzu/tanzu-framework/test/pkg/ytt"
)

const (
	packageRoot = "../.."
)

var _ = Describe("Cluster-API Provider vSphere Ytt Templating", func() {
	var paths []string
	var values string

	BeforeEach(func() {
		paths = []string{
			filepath.Join(packageRoot, "bundle", "config"),
		}
	})

	It("enables AntiAffinity feature gate on capv-controller-manager Deployment", func() {
		output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, nil)
		Expect(err).NotTo(HaveOccurred())
		docs, err := FindDocsMatchingYAMLPath(output, map[string]string{
			"$.kind":          "Deployment",
			"$.metadata.name": "capv-controller-manager",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(docs).To(HaveLen(1))
		Expect(docs[0]).To(HaveYAMLPathWithValue(
			"$.spec.template.spec.containers[0].args[2]",
			"--feature-gates=NodeAntiAffinity=${NODE_ANTI_AFFINITY:=true}",
		))
	})

	Context("when the httpProxy data value is provided", func() {
		BeforeEach(func() {
			values = `#@data/values
---
capvControllerManager:
  httpProxy: http://127.0.0.1:3124`
		})
		It("adds HTTP_PROXY environment variable to the deployment container", func() {
			output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
			Expect(err).NotTo(HaveOccurred())
			docs, err := FindDocsMatchingYAMLPath(output, map[string]string{
				"$.kind":          "Deployment",
				"$.metadata.name": "capv-controller-manager",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(docs).To(HaveLen(2))
			for _, doc := range docs {
				Expect(doc).To(HaveYAMLPathWithValue(
					"$.spec.template.spec.containers[0].env[?(@.name=='HTTP_PROXY')].value",
					"http://127.0.0.1:3124",
				))
				Expect(doc).NotTo(HaveYAMLPath("$.spec.template.spec.containers[0].env[?(@.name=='HTTPS_PROXY')]"))
				Expect(doc).NotTo(HaveYAMLPath("$.spec.template.spec.containers[0].env[?(@.name=='NO_PROXY')]"))
			}
		})
	})

	Context("when the httpsProxy data value is provided", func() {
		BeforeEach(func() {
			values = `#@data/values
---
capvControllerManager:
  httpsProxy: https://127.0.0.1:3124`
		})
		It("adds HTTPS_PROXY environment variable to the deployment container", func() {
			output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
			Expect(err).NotTo(HaveOccurred())
			docs, err := FindDocsMatchingYAMLPath(output, map[string]string{
				"$.kind":          "Deployment",
				"$.metadata.name": "capv-controller-manager",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(docs).To(HaveLen(2))
			for _, doc := range docs {
				Expect(doc).NotTo(HaveYAMLPath("$.spec.template.spec.containers[0].env[?(@.name=='HTTP_PROXY')]"))
				Expect(doc).To(HaveYAMLPathWithValue(
					"$.spec.template.spec.containers[0].env[?(@.name=='HTTPS_PROXY')].value",
					"https://127.0.0.1:3124",
				))
				Expect(doc).NotTo(HaveYAMLPath("$.spec.template.spec.containers[0].env[?(@.name=='NO_PROXY')]"))
			}
		})
	})

	Context("when the noProxy data value is provided", func() {
		BeforeEach(func() {
			values = `#@data/values
---
capvControllerManager:
  noProxy: 10.0.0.0/16,127.0.0.1`
		})
		It("adds NO_PROXY environment variable to the deployment container", func() {
			output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, strings.NewReader(values))
			Expect(err).NotTo(HaveOccurred())
			docs, err := FindDocsMatchingYAMLPath(output, map[string]string{
				"$.kind":          "Deployment",
				"$.metadata.name": "capv-controller-manager",
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(docs).To(HaveLen(2))
			for _, doc := range docs {
				Expect(doc).NotTo(HaveYAMLPath("$.spec.template.spec.containers[0].env[?(@.name=='HTTP_PROXY')]"))
				Expect(doc).NotTo(HaveYAMLPath("$.spec.template.spec.containers[0].env[?(@.name=='HTTPS_PROXY')]"))
				Expect(doc).To(HaveYAMLPathWithValue(
					"$.spec.template.spec.containers[0].env[?(@.name=='NO_PROXY')].value",
					"10.0.0.0/16,127.0.0.1",
				))
			}
		})
	})
})
