// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/fakes"
)

var _ = Describe("Unit tests for create cluster", func() {
	var tkgClient *fakes.Client

	Context("Creating clusters for TKGs", func() {
		It("Set CEIP participation for prod environment", func() {
			kubeConfigPath := getConfigFilePath()

			tkgClient = &fakes.Client{}

			tkgctlClient := &tkgctl{
				tkgClient:  tkgClient,
				kubeconfig: kubeConfigPath,
			}

			err := tkgctlClient.SetCeip("true", "true", "")
			Expect(err).NotTo(HaveOccurred())
		})
		Context("Set CEIP to true on staging environment", func() {
			It("Set CEIP to true on staging environment without labels", func() {
				kubeConfigPath := getConfigFilePath()
				tkgClient = &fakes.Client{}

				tkgctlClient := &tkgctl{
					kubeconfig: kubeConfigPath,
					tkgClient:  tkgClient,
				}

				err := tkgctlClient.SetCeip("true", "false", "")
				Expect(err).NotTo(HaveOccurred())
			})
			It("Set CEIP to true on staging environment with valid labels", func() {
				kubeConfigPath := getConfigFilePath()
				tkgClient = &fakes.Client{}

				tkgctlClient := &tkgctl{
					kubeconfig: kubeConfigPath,
					tkgClient:  tkgClient,
				}

				err := tkgctlClient.SetCeip("true", "false", "entitlement-account-number=foo,env-type=production")
				Expect(err).NotTo(HaveOccurred())
			})
			It("Invalid labels should return error", func() {
				kubeConfigPath := getConfigFilePath()
				tkgClient = &fakes.Client{}

				tkgctlClient := &tkgctl{
					kubeconfig: kubeConfigPath,
					tkgClient:  tkgClient,
				}

				err := tkgctlClient.SetCeip("true", "false", "entitlement-account-number=foo,env-type=prod")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("Invalid error type prod, environment type can be production, development, or test"))
			})
			It("Incorrect number of labels should return error", func() {
				kubeConfigPath := getConfigFilePath()
				tkgClient = &fakes.Client{}

				tkgctlClient := &tkgctl{
					kubeconfig: kubeConfigPath,
					tkgClient:  tkgClient,
				}

				err := tkgctlClient.SetCeip("true", "false", "entitlement-account-number=foo,env-type=prod,extra-label=bar")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("There are more labels provided than are currently supported. The supported labels are entitlement-account-number,and env-type"))
			})
			It("Incorrect entitlement-account-number should return error", func() {
				kubeConfigPath := getConfigFilePath()
				tkgClient = &fakes.Client{}

				tkgctlClient := &tkgctl{
					kubeconfig: kubeConfigPath,
					tkgClient:  tkgClient,
				}

				err := tkgctlClient.SetCeip("true", "false", "entitlement-account-number=!foo-bar_Baz,env-type=production")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("entitlement-account-number: !foo-bar_Baz cannot contain special characters"))
			})
			It("entitlement-account-number should only contain alphanumeric characters", func() {
				kubeConfigPath := getConfigFilePath()
				tkgClient = &fakes.Client{}

				tkgctlClient := &tkgctl{
					kubeconfig: kubeConfigPath,
					tkgClient:  tkgClient,
				}

				err := tkgctlClient.SetCeip("true", "false", "entitlement-account-number=Foo123baR,env-type=production")
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
