// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package unit

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/vmware-tanzu/tanzu-framework/test/pkg/matchers"
	"github.com/vmware-tanzu/tanzu-framework/test/pkg/ytt"
)

const (
	packageRoot = "../.."
)

var _ = Describe("Cluster-API Ytt Templating", func() {
	var paths []string
	BeforeEach(func() {
		paths = []string{
			filepath.Join(packageRoot, "bundle", "config"),
		}
	})

	It("enables ClusterResourceSet and ClusterTopology feature gates on capi-controller-manager Deployment", func() {
		output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, nil)
		Expect(err).NotTo(HaveOccurred())
		docs, err := FindDocsMatchingYAMLPath(output, map[string]string{
			"$.kind":          "Deployment",
			"$.metadata.name": "capi-controller-manager",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(docs).To(HaveLen(1))
		Expect(docs[0]).To(HaveYAMLPathWithValue(
			"$.spec.template.spec.containers[0].args[2]",
			"--feature-gates=MachinePool=${EXP_MACHINE_POOL:=false},ClusterResourceSet=${EXP_CLUSTER_RESOURCE_SET:=true},ClusterTopology=${CLUSTER_TOPOLOGY:=true},RuntimeSDK=${EXP_RUNTIME_SDK:=false}",
		))
	})
})
