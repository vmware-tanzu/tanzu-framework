// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	. "github.com/vmware-tanzu-private/core/pkg/v1/tkg/client"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/fakes"
	fakehelper "github.com/vmware-tanzu-private/core/pkg/v1/tkg/fakes/helper"
)

var _ = Describe("Unit tests for scalePacificCluster", func() {
	var (
		err                   error
		regionalClusterClient *fakes.ClusterClient
		tkgClient             *TkgClient
		scaleClusterOptions   ScaleClusterOptions
		kubeconfig            string
	)

	BeforeEach(func() {
		regionalClusterClient = &fakes.ClusterClient{}
		tkgClient, err = CreateTKGClient("../fakes/config/config.yaml", testingDir, defaultTKGBoMFileForTesting, 2*time.Second)
		Expect(err).NotTo(HaveOccurred())
		kubeconfig = fakehelper.GetFakeKubeConfigFilePath(testingDir, "../fakes/config/kubeconfig/config1.yaml")
	})

	JustBeforeEach(func() {
		err = tkgClient.ScalePacificCluster(scaleClusterOptions, regionalClusterClient)
	})

	Context("When scaleClusterOptions is all set", func() {
		BeforeEach(func() {
			scaleClusterOptions = ScaleClusterOptions{
				ClusterName:       "my-cluster",
				Namespace:         "namespace-1",
				WorkerCount:       5,
				ControlPlaneCount: 10,
				Kubeconfig:        kubeconfig,
			}
			regionalClusterClient.ScalePacificClusterControlPlaneReturns(nil)
			regionalClusterClient.ScalePacificClusterWorkerNodesReturns(nil)
		})
		It("should not return error", func() {
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("When namespace is empty", func() {
		BeforeEach(func() {
			regionalClusterClient.GetCurrentNamespaceReturns("", errors.New("fake-error"))
			scaleClusterOptions = ScaleClusterOptions{
				ClusterName:       "my-cluster",
				Namespace:         "",
				WorkerCount:       5,
				ControlPlaneCount: 10,
				Kubeconfig:        kubeconfig,
			}
		})
		It("returns an error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to get current namespace"))
		})
	})

	Context("When control plane for workload cluster cannot be scaled", func() {
		BeforeEach(func() {
			scaleClusterOptions = ScaleClusterOptions{
				ClusterName:       "my-cluster",
				Namespace:         "namespace-1",
				WorkerCount:       5,
				ControlPlaneCount: 10,
				Kubeconfig:        kubeconfig,
			}
			regionalClusterClient.ScalePacificClusterControlPlaneReturns(errors.New("fake-error"))
		})
		It("returns an error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unable to scale control plane for workload cluster"))
		})
	})

	Context("When workers nodes for workload cluster cannot be scaled", func() {
		BeforeEach(func() {
			scaleClusterOptions = ScaleClusterOptions{
				ClusterName:       "my-cluster",
				Namespace:         "namespace-1",
				WorkerCount:       5,
				ControlPlaneCount: 10,
				Kubeconfig:        kubeconfig,
			}
			regionalClusterClient.ScalePacificClusterWorkerNodesReturns(errors.New("fake-error"))
		})
		It("returns an error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unable to scale workers nodes for workload cluster"))
		})
	})
})
