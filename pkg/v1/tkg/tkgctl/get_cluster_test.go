// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/client"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/fakes"
)

var _ = Describe("Unit test for get clusters", func() {
	var (
		ctl       tkgctl
		tkgClient = &fakes.Client{}
		ops       = ListTKGClustersOptions{
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
		_, err = ctl.GetClusters(ops)
	})

	Context("when failed to list clusters", func() {
		BeforeEach(func() {
			tkgClient.ListTKGClustersReturns(nil, errors.New("failed to list clusters"))
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})
	Context("when it is able list the clusters ", func() {
		BeforeEach(func() {
			tkgClient.ListTKGClustersReturns([]client.ClusterInfo{{Name: "my-cluster", Namespace: "default"}, {Name: "my-cluster-2", Namespace: "my-system"}}, nil)
		})
		It("should not return an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
