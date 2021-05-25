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

var _ = Describe("Unit test for describe cluster", func() {
	var (
		ctl       tkgctl
		tkgClient = &fakes.Client{}
		ops       = DescribeTKGClustersOptions{
			ClusterName: "my-cluster",
			Namespace:   "",
		}
		err error
	)

	JustBeforeEach(func() {
		ctl = tkgctl{
			configDir:  testingDir,
			tkgClient:  tkgClient,
			kubeconfig: "./kube",
		}
		_, err = ctl.DescribeCluster(ops)
	})

	Context("when failed to list tkg clusters", func() {
		BeforeEach(func() {
			tkgClient.ListTKGClustersReturns(nil, errors.New("failed to list clusters"))
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when it is a failed management cluster", func() {
		BeforeEach(func() {
			tkgClient.ListTKGClustersReturns([]client.ClusterInfo{{Name: "my-cluster", Roles: []string{"management"}}}, nil)
			tkgClient.IsManagementClusterAKindClusterReturns(true, nil)
			tkgClient.DescribeClusterReturns(nil, nil, nil, nil)
		})
		It("should not return an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("when tkgClient failed to describe the cluster", func() {
		BeforeEach(func() {
			tkgClient.ListTKGClustersReturns([]client.ClusterInfo{{Name: "my-cluster", Roles: []string{"<none>"}}}, nil)
			tkgClient.DescribeClusterReturns(nil, nil, nil, errors.New("failed to describe cluster"))
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})
})
