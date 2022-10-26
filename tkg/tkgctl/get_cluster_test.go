// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/tkg/fakes"
)

var _ = Describe("Unit test for get clusters", func() {
	var (
		ctl               tkgctl
		tkgClient         = &fakes.Client{}
		featureGateHelper = &fakes.FakeFeatureGateHelper{}
		ops               = ListTKGClustersOptions{
			ClusterName: "my-cluster",
		}
		err error
	)

	JustBeforeEach(func() {
		ctl = tkgctl{
			configDir:         testingDir,
			tkgClient:         tkgClient,
			kubeconfig:        "./kube",
			featureGateHelper: featureGateHelper,
		}
		_, err = ctl.GetClusters(ops)
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
	Context("when the management cluster is not Pacific(TKGS) supervisor cluster and is able to list clusters with option --AllNamespace false", func() {
		BeforeEach(func() {
			ops.AllNamespaces = false
			ops.Namespace = "default"
			tkgClient.IsPacificManagementClusterReturns(false, nil)
			tkgClient.ListTKGClustersReturns([]client.ClusterInfo{{Name: "my-cluster", Namespace: "default"}}, nil)
		})
		It("should not return an error", func() {
			Expect(err).ToNot(HaveOccurred())
			options := tkgClient.ListTKGClustersArgsForCall(0)
			Expect(options.IsTKGSClusterClassFeatureActivated).To(BeFalse())
		})
	})
	Context("when the management cluster is not Pacific(TKGS) supervisor cluster, but failed to list clusters", func() {
		BeforeEach(func() {
			tkgClient.IsPacificManagementClusterReturns(false, nil)
			tkgClient.ListTKGClustersReturns(nil, errors.New("failed to list clusters"))
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to list clusters"))
			options := tkgClient.ListTKGClustersArgsForCall(1)
			Expect(options.IsTKGSClusterClassFeatureActivated).To(BeFalse())
		})
	})
	Context("when the management cluster is Pacific(TKGS) supervisor cluster but failed to get the cluster class feature activation status", func() {
		BeforeEach(func() {
			tkgClient.IsPacificManagementClusterReturns(true, nil)
			featureGateHelper.FeatureActivatedInNamespaceReturns(false, errors.New("fake-feature-gate-error"))
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("fake-feature-gate-error"))
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
			options := tkgClient.ListTKGClustersArgsForCall(2)
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
	Context("when the management cluster is not Pacific(TKGS) supervisor cluster and the listCluster with --AllNamespace", func() {
		BeforeEach(func() {
			ops.AllNamespaces = true
			tkgClient.IsPacificManagementClusterReturns(false, nil)
			tkgClient.ListTKGClustersReturns([]client.ClusterInfo{{Name: "my-cluster", Namespace: "default"}, {Name: "my-cluster-2", Namespace: "my-system"}}, nil)
		})
		It("should not return an error", func() {
			Expect(err).ToNot(HaveOccurred())
			options := tkgClient.ListTKGClustersArgsForCall(0)
			Expect(options.IsTKGSClusterClassFeatureActivated).To(BeFalse())
		})
	})
})

var _ = Describe("Unit test for IsClusterExists", func() {
	var (
		ctl               tkgctl
		tkgClient         = &fakes.Client{}
		featureGateHelper = &fakes.FakeFeatureGateHelper{}
		clustername       = "my-cluster"
		namespace         = "my-namespace"
		isClusterExists   bool
		err               error
	)

	JustBeforeEach(func() {
		ctl = tkgctl{
			configDir:         testingDir,
			tkgClient:         tkgClient,
			kubeconfig:        "./kube",
			featureGateHelper: featureGateHelper,
		}
		isClusterExists, err = ctl.IsClusterExists(clustername, namespace)
	})

	Context("when list cluster returns error", func() {
		BeforeEach(func() {
			tkgClient.IsPacificManagementClusterReturns(false, nil)
			tkgClient.ListTKGClustersReturns(nil, errors.New("failed to list clusters"))
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to list clusters"))
		})
	})

	Context("when given cluster exists", func() {
		BeforeEach(func() {
			tkgClient.IsPacificManagementClusterReturns(false, nil)
			tkgClient.ListTKGClustersReturns([]client.ClusterInfo{{Name: clustername, Namespace: namespace}, {Name: "my-cluster-2", Namespace: "my-system"}}, nil)
		})
		It("should not return an error and cluster should exists", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(isClusterExists).To(BeTrue())
		})
	})

	Context("when given cluster not exists", func() {
		BeforeEach(func() {
			tkgClient.IsPacificManagementClusterReturns(false, nil)
			tkgClient.ListTKGClustersReturns([]client.ClusterInfo{{Name: "my-cluster-2", Namespace: "my-system"}}, nil)
		})
		It("should not return an error and cluster not exists", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(isClusterExists).To(BeFalse())
		})
	})
})
