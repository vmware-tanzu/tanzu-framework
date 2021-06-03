// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/tools/clientcmd"

	. "github.com/vmware-tanzu-private/core/pkg/v1/tkg/client"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/utils"
)

var _ = Describe("Utils", func() {
	var (
		tempKubeConfigPath string
		err                error
		contextName        string
	)
	Describe("DeleteContextFromKubeConfig tests", func() {
		BeforeEach(func() {
			f, err := os.CreateTemp("", "yaml")
			Expect(err).ToNot(HaveOccurred())
			tempKubeConfigPath = f.Name()
			copyFile("../fakes/config/kubeconfig/config1.yaml", tempKubeConfigPath)
		})
		AfterEach(func() {
			_ = utils.DeleteFile(tempKubeConfigPath)
		})

		JustBeforeEach(func() {
			err = DeleteContextFromKubeConfig(tempKubeConfigPath, contextName)
		})
		Context("When context to be deleted is not present in the kubeconfig file", func() {
			BeforeEach(func() {
				contextName = "fake-nonexisting-context"
			})
			It("should not return error", func() {
				Expect(err).ToNot(HaveOccurred())
			})
		})
		Context("When context to be deleted is present in the kubeconfig file", func() {
			BeforeEach(func() {
				contextName = "queen-anne-context"
			})
			It("should not return error and delete the context and cluster from kubeconfig file", func() {
				Expect(err).ToNot(HaveOccurred())
				config, err1 := clientcmd.LoadFromFile(tempKubeConfigPath)
				Expect(err1).ToNot(HaveOccurred())
				_, ok := config.Contexts[contextName]
				Expect(ok).To(Equal(false))
				_, ok = config.Clusters["pig-cluster"]
				Expect(ok).To(Equal(false))
			})
		})
		Context("When context to be deleted is present in the kubeconfig file and is current context", func() {
			BeforeEach(func() {
				contextName = "federal-context"
			})
			It("should not return error and delete the context,cluster and also set the current-context to empty string", func() {
				Expect(err).ToNot(HaveOccurred())
				config, err1 := clientcmd.LoadFromFile(tempKubeConfigPath)
				Expect(err1).ToNot(HaveOccurred())
				_, ok := config.Contexts[contextName]
				Expect(ok).To(Equal(false))
				_, ok = config.Clusters["horse-cluster"]
				Expect(ok).To(Equal(false))
				Expect(config.CurrentContext).To(Equal(""))
			})
		})
	})
})
