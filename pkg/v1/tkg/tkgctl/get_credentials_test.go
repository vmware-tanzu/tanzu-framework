// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/fakes"
)

var _ = Describe("Unit test for get credentials", func() {
	var (
		ctl       tkgctl
		tkgClient = &fakes.Client{}
		ops       = GetWorkloadClusterCredentialsOptions{
			ClusterName: "my-cluster",
		}
		err error
	)

	JustBeforeEach(func() {
		ctl = tkgctl{
			configDir:  testingDir,
			tkgClient:  tkgClient,
			kubeconfig: "./kube",
		}
		err = ctl.GetCredentials(ops)
	})

	Context("when failed to get credentials", func() {
		BeforeEach(func() {
			tkgClient.GetWorkloadClusterCredentialsReturns("", "", errors.New("failed to get credentials"))
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})
	Context("when it is able to get the credentials ", func() {
		BeforeEach(func() {
			tkgClient.GetWorkloadClusterCredentialsReturns("", "", nil)
		})
		It("should not return an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
