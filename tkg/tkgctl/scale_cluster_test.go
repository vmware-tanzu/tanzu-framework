// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes"
)

var _ = Describe("Unit test for scale cluster", func() {
	var (
		ctl       tkgctl
		tkgClient = &fakes.Client{}
		ops       = ScaleClusterOptions{
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
		err = ctl.ScaleCluster(ops)
	})

	Context("when node count is less than zero", func() {
		BeforeEach(func() {
			ops.ControlPlaneCount = -1
			ops.WorkerCount = -1
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when failed to scale the cluster", func() {
		BeforeEach(func() {
			ops.ControlPlaneCount = 1
			ops.WorkerCount = 1
			tkgClient.ScaleClusterReturns(errors.New("region not found"))
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})
	Context("when it is able to scale the cluster", func() {
		BeforeEach(func() {
			ops.ControlPlaneCount = 1
			ops.WorkerCount = 1
			tkgClient.ScaleClusterReturns(nil)
		})
		It("should not return an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
