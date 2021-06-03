// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/fakes"
)

var _ = Describe("Unit tests for delete cluster", func() {
	var (
		ctl       *tkgctl
		tkgClient = &fakes.Client{}
		err       error
		configDir string
		dcOps     DeleteClustersOptions
	)

	JustBeforeEach(func() {
		configDir, err = os.MkdirTemp("", "test")
		err = os.MkdirAll(testingDir, 0o700)
		Expect(err).ToNot(HaveOccurred())
		prepareConfiDir(configDir)
		options := Options{
			ConfigDir: configDir,
		}
		c, createErr := New(options)
		Expect(createErr).ToNot(HaveOccurred())
		ctl, _ = c.(*tkgctl)
		ctl.tkgClient = tkgClient
		dcOps = DeleteClustersOptions{
			SkipPrompt:  true,
			ClusterName: "my-cluster",
		}
		err = ctl.DeleteCluster(dcOps)
	})
	Context("When cluster exists and can be deleted successfully", func() {
		BeforeEach(func() {
			tkgClient.DeleteWorkloadClusterReturns(nil)
		})
		It("should start deleting the cluster", func() {
			Expect(err).ToNot(HaveOccurred())
		})
	})
	Context("When cluster does not exist", func() {
		BeforeEach(func() {
			tkgClient.DeleteWorkloadClusterReturns(errors.New("cluster does not exist"))
		})
		It("should start deleting the cluster", func() {
			Expect(err).To(HaveOccurred())
		})
	})

	AfterEach(func() {
		os.Remove(configDir)
	})
})
