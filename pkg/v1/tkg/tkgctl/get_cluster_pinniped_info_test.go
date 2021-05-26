// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/fakes"
)

var _ = Describe("Unit test for get Pinniped Info", func() {
	var (
		ctl       tkgctl
		tkgClient = &fakes.Client{}
		ops       = GetClusterPinnipedInfoOptions{
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
		_, err = ctl.GetClusterPinnipedInfo(ops)
	})

	Context("when failed to get pinniped info", func() {
		BeforeEach(func() {
			tkgClient.GetClusterPinnipedInfoReturns(nil, errors.New("failed to get pinniped info"))
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})
	Context("when it is able get the pinniped info ", func() {
		BeforeEach(func() {
			ops.IsManagementCluster = true
			tkgClient.GetClusterPinnipedInfoReturns(nil, nil)
		})
		It("should not return an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
