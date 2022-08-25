// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/tkg/fakes"
)

var _ = Describe("Unit test for describe cluster", func() {
	var (
		ctl               tkgctl
		tkgClient         = &fakes.Client{}
		featureGateHelper = &fakes.FakeFeatureGateHelper{}
		ops               = DescribeTKGClustersOptions{
			ClusterName: "my-cluster",
			Namespace:   "",
		}
		err    error
		result DescribeClusterResult
	)

	JustBeforeEach(func() {
		ctl = tkgctl{
			configDir:         testingDir,
			tkgClient:         tkgClient,
			kubeconfig:        "./kube",
			featureGateHelper: featureGateHelper,
		}
		result, err = ctl.DescribeCluster(ops)
	})

	Context("when failed to determine the management cluster is Pacific(TKGS) supervisor cluster ", func() {
		BeforeEach(func() {
			tkgClient.IsPacificManagementClusterReturns(false, errors.New("fake-error"))
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-error"))
		})
	})
	Context("when the management cluster is not Pacific(TKGS) supervisor cluster and failed to list tkg clusters", func() {
		BeforeEach(func() {
			tkgClient.IsPacificManagementClusterReturns(false, nil)
			tkgClient.ListTKGClustersReturns(nil, errors.New("failed to list clusters"))
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when the management cluster is not Pacific(TKGS) supervisor cluster and it is a failed management cluster", func() {
		BeforeEach(func() {
			tkgClient.IsPacificManagementClusterReturns(false, nil)
			tkgClient.ListTKGClustersReturns([]client.ClusterInfo{{Name: "my-cluster", Roles: []string{"management"}}}, nil)
			tkgClient.IsManagementClusterAKindClusterReturns(true, nil)
			tkgClient.DescribeClusterReturns(nil, nil, nil, nil)
		})
		It("should not return an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("when the management cluster is not Pacific(TKGS) supervisor cluster and when tkgClient failed to describe the cluster", func() {
		BeforeEach(func() {
			tkgClient.IsPacificManagementClusterReturns(false, nil)
			tkgClient.ListTKGClustersReturns([]client.ClusterInfo{{Name: "my-cluster", Roles: []string{"<none>"}}}, nil)
			tkgClient.DescribeClusterReturns(nil, nil, nil, errors.New("failed to describe cluster"))
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
		})
	})
	Context("when the management cluster is Pacific(TKGS) supervisor cluster with cluster class feature disabled and is able to list the clusters", func() {
		BeforeEach(func() {
			tkgClient.IsPacificManagementClusterReturns(true, nil)
			featureGateHelper.FeatureActivatedInNamespaceReturns(true, nil)
			tkgClient.ListTKGClustersReturns([]client.ClusterInfo{{Name: "my-cluster", Namespace: "default"}, {Name: "my-cluster-2", Namespace: "my-system"}}, nil)
		})
		It("should not return an error", func() {
			Expect(err).ToNot(HaveOccurred())
			options := tkgClient.ListTKGClustersArgsForCall(3)
			Expect(options.IsTKGSClusterClassFeatureActivated).To(BeTrue())

		})
	})
	Context("when the management cluster is Pacific(TKGS) supervisor cluster with cluster class feature enabled and is able to list the clusters", func() {
		BeforeEach(func() {
			tkgClient.IsPacificManagementClusterReturns(true, nil)
			featureGateHelper.FeatureActivatedInNamespaceReturns(true, nil)
			tkgClient.ListTKGClustersReturns([]client.ClusterInfo{{Name: "my-cluster", Namespace: "default"}, {Name: "my-cluster-2", Namespace: "my-system"}}, nil)
		})
		It("should not return an error", func() {
			Expect(err).ToNot(HaveOccurred())
			options := tkgClient.ListTKGClustersArgsForCall(3)
			Expect(options.IsTKGSClusterClassFeatureActivated).To(BeTrue())
		})
	})
	Context("when the management cluster is Pacific(TKGS) supervisor cluster and when tkgClient failed to describe the cluster", func() {
		BeforeEach(func() {
			tkgClient.IsPacificManagementClusterReturns(true, nil)
			tkgClient.ListTKGClustersReturns([]client.ClusterInfo{{Name: "my-cluster", Roles: []string{"<none>"}}}, nil)
			tkgClient.DescribeClusterReturns(nil, nil, nil, errors.New("failed to describe cluster"))
		})
		It("should not return an error but ObjectTree and cluster objects should be nil", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(result.Objs).To(BeNil())
			Expect(result.Cluster).To(BeNil())
		})
	})
})
