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

var _ = Describe("Cluster-API-bootstrap-kubeadm Ytt Templating", func() {
	var paths []string
	BeforeEach(func() {
		paths = []string{
			filepath.Join(packageRoot, "bundle", "config"),
		}
	})

	It("sets the bootstrap token ttl flag to 15m", func() {
		output, err := ytt.RenderYTTTemplate(ytt.CommandOptions{}, paths, nil)
		Expect(err).NotTo(HaveOccurred())
		docs, err := FindDocsMatchingYAMLPath(output, map[string]string{
			"$.kind":          "Deployment",
			"$.metadata.name": "capi-kubeadm-bootstrap-controller-manager",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(docs).To(HaveLen(1))
		Expect(docs[0]).To(HaveYAMLPathWithValue(
			"$.spec.template.spec.containers[?(@.name=='manager')].args[3]",
			"--bootstrap-token-ttl=${CAPBK_BOOTSTRAP_TOKEN_TTL:=15m}",
		))
	})
})
