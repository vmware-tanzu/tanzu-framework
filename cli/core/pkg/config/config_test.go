// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/fakes"
)

var (
	prevValue string
)

const envVar = "test-conf-env"

var _ = Describe("config env variables", func() {
	Context("get env from config", func() {
		BeforeEach(func() {
			cc := &fakes.FakeConfigClientWrapper{}
			configClient = cc
			prevValue = os.Getenv(envVar)
			confEnvMap := map[string]string{envVar: envVar}
			cc.GetEnvConfigurationsReturns(confEnvMap)
		})
		It("env variable should be set with config env", func() {
			ConfigureEnvVariables()
			Expect(os.Getenv(envVar)).To(Equal(envVar))
			os.Setenv(envVar, prevValue)
		})
	})
	Context("config return nil map", func() {
		BeforeEach(func() {
			cc := &fakes.FakeConfigClientWrapper{}
			configClient = cc
			cc.GetEnvConfigurationsReturns(nil)
		})
		It("execute without error", func() {
			ConfigureEnvVariables()
		})
	})
})
