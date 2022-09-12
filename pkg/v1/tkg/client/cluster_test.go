// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package client_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/region"
)

var _ = Describe("ValidateManagementClusterVersionWithCLI", func() {
	const (
		clusterName = "test-cluster"
		v140        = "v1.4.0"
		v141        = "v1.4.1"
		v150        = "v1.5.0"
	)
	var (
		regionalClient fakes.ClusterClient
		tkgBomClient   fakes.TKGConfigBomClient
		regionManager  fakes.RegionManager
		c              *TkgClient
		err            error
	)
	JustBeforeEach(func() {
		err = c.ValidateManagementClusterVersionWithCLI(&regionalClient)
	})
	BeforeEach(func() {
		regionManager = fakes.RegionManager{}
		regionManager.GetCurrentContextReturns(region.RegionContext{
			ClusterName: clusterName,
			Status:      region.Success,
		}, nil)

		regionalClient = fakes.ClusterClient{}
		regionalClient.ListResourcesStub = func(i interface{}, lo ...client.ListOption) error {
			list := i.(*v1alpha3.ClusterList)
			*list = v1alpha3.ClusterList{
				Items: []v1alpha3.Cluster{
					{
						ObjectMeta: v1.ObjectMeta{
							Name:      clusterName,
							Namespace: "default",
						},
					},
				},
			}
			return nil
		}

		c, err = New(Options{
			TKGConfigUpdater: &fakes.TKGConfigUpdaterClient{},
			TKGBomClient:     &tkgBomClient,
			RegionManager:    &regionManager,
		})
	})
	Context("v1.4.0 management cluster", func() {
		BeforeEach(func() {
			regionalClient.GetManagementClusterTKGVersionReturns(v140, nil)
		})

		When("management cluster version matches cli version", func() {
			BeforeEach(func() {
				tkgBomClient = fakes.TKGConfigBomClient{}
				tkgBomClient.GetDefaultTKGReleaseVersionReturns(v140, nil)
			})
			It("should validate without error", func() {
				Expect(err).To(BeNil())
			})
		})

		When("cli version is a patch version ahead of management cluster", func() {
			BeforeEach(func() {
				tkgBomClient = fakes.TKGConfigBomClient{}
				tkgBomClient.GetDefaultTKGReleaseVersionReturns(v141, nil)
			})
			It("should validate without error", func() {
				Expect(err).To(BeNil())
			})
		})

		When("cli version is a minor version ahead of management cluster", func() {
			BeforeEach(func() {
				tkgBomClient = fakes.TKGConfigBomClient{}
				tkgBomClient.GetDefaultTKGReleaseVersionReturns(v150, nil)
			})
			It("should return an error", func() {
				Expect(err).Should(MatchError("version mismatch between management cluster and cli version. Please upgrade your management cluster to the latest to continue"))
			})
		})
	})
})
