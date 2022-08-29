// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes"
)

var _ = Describe("Unit test for get tkr", func() {
	var (
		ctl       tkgctl
		tkgClient = &fakes.Client{}
		err       error
	)

	JustBeforeEach(func() {
		ctl = tkgctl{
			configDir:  testingDir,
			tkgClient:  tkgClient,
			kubeconfig: "./kube",
		}
		_, err = ctl.GetTanzuKubernetesReleases("")
	})
	Context("when failed to get tkrs", func() {
		BeforeEach(func() {
			tkgClient.GetTanzuKubernetesReleasesReturns(nil, errors.New("region not found"))
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})
	Context("when it is able to get tkrs", func() {
		BeforeEach(func() {
			tkgClient.GetTanzuKubernetesReleasesReturns(nil, nil)
		})
		It("should not return an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
