// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	appsv1 "k8s.io/api/apps/v1"

	. "github.com/vmware-tanzu/tanzu-framework/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/fakes"
)

var _ = Describe("Unit tests for create cluster", func() {
	var (
		err              error
		clusterClient    *fakes.ClusterClient
		kubeConfigBytes  []byte
		tkgClient        *TkgClient
		name             string
		namespace        string
		manifest         string
		autoscalerDeploy *appsv1.Deployment
	)

	BeforeEach(func() {
		clusterClient = &fakes.ClusterClient{}
		tkgClient, err = CreateTKGClient("../fakes/config/config.yaml", testingDir, defaultTKGBoMFileForTesting, 2*time.Second)
		Expect(err).NotTo(HaveOccurred())
		name = "testClusterName"
		autoscalerDeploy = &appsv1.Deployment{}
		autoscalerDeploy.Name = constants.AutoscalerDeploymentNameSuffix
		manifest = `---
apiVersion: cluster.x-k8s.io/v1alpha3
kind: Cluster
metadata:
  name: tkg-region-aws-11111111111111
  namespace: default
spec:
  clusterNetwork:
    pods:
      cidrBlocks:
      - 192.168.0.0/16
  infrastructureRef:
    apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
    kind: AWSCluster
    name: tkg-region-aws-11111111111111
`
	})

	Describe("When creating cluster", func() {
		JustBeforeEach(func() {
			namespace = "fake-namespace"
			err = tkgClient.DoCreateCluster(clusterClient, name, namespace, manifest)
		})
		Context("When cluster name is invalid", func() {
			BeforeEach(func() {
				name = ""
			})
			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(HavePrefix("invalid cluster name"))
			})
		})
		Context("When manifest is invalid", func() {
			BeforeEach(func() {
				manifest = ""
			})
			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(HavePrefix("invalid cluster manifest"))
			})
		})
		Context("When apply manifest fails", func() {
			BeforeEach(func() {
				clusterClient.ApplyReturns(errors.New("fake-apply-error"))
			})
			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(HavePrefix("unable to apply cluster configuration"))
			})
		})
	})
	Describe("testing WaitingForClusterAndGetKubeConfig method ", func() {
		JustBeforeEach(func() {
			kubeConfigBytes, err = tkgClient.WaitForClusterInitializedAndGetKubeConfig(clusterClient, name, constants.DefaultNamespace)
		})
		Context("When waiting for cluster fails", func() {
			BeforeEach(func() {
				clusterClient.WaitForClusterInitializedReturns(errors.New("fake-wait-error"))
			})
			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(HavePrefix("error waiting for cluster to be provisioned"))
			})
		})
		Context("When get kube config fails", func() {
			BeforeEach(func() {
				clusterClient.GetKubeConfigForClusterReturns(kubeConfigBytes, errors.New("fake-getkubeconfig-error"))
			})
			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(HavePrefix("unable to extract kube config for cluster"))
			})
		})
		Context("With get kube config succeeds", func() {
			BeforeEach(func() {
				clusterClient.GetKubeConfigForClusterReturns([]byte("fake kubeconfig data"), nil)
			})
			It("kube config bytes are returned", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(kubeConfigBytes).To(Equal([]byte("fake kubeconfig data")))
			})
		})
	})

	Describe("WaitForAutoscalerDeployment by legacy mod", func() {
		JustBeforeEach(func() {
			tkgClient.WaitForAutoscalerDeployment(clusterClient, name, constants.DefaultNamespace, false)
		})

		Context("When the value for config variable 'ENABLE_AUTOSCALER' is not set", func() {
			It("should not wait for autoscaler deployment", func() {
				Expect(clusterClient.WaitForAutoscalerDeploymentCallCount()).To(Equal(0))
			})
		})

		Context("When the value for config variable 'ENABLE_AUTOSCALER' is set to 'false'", func() {
			BeforeEach(func() {
				os.Setenv(constants.ConfigVariableEnableAutoscaler, "false")
			})

			It("should not wait for autoscaler deployment", func() {
				Expect(clusterClient.WaitForAutoscalerDeploymentCallCount()).To(Equal(0))
			})
		})

		Context("When the value for config variable 'ENABLE_AUTOSCALER' is set to 'true'", func() {
			BeforeEach(func() {
				os.Setenv(constants.ConfigVariableEnableAutoscaler, "true")
			})

			It("should wait for autoscaler deployment", func() {
				Expect(clusterClient.WaitForAutoscalerDeploymentCallCount()).To(Equal(1))
			})
		})
	})

	Describe("WaitForAutoscalerDeployment by clusterclass mod", func() {
		JustBeforeEach(func() {
			tkgClient.WaitForAutoscalerDeployment(clusterClient, name, constants.DefaultNamespace, true)
		})

		Context("When the deployment of autoscaler not exsits in mgmt cluster", func() {
			BeforeEach(func() {
				clusterClient.GetDeploymentReturns(appsv1.Deployment{}, errors.New("can't get resource of autoscaler deployment"))
			})

			It("should not wait for autoscaler deployment", func() {
				Expect(clusterClient.WaitForAutoscalerDeploymentCallCount()).To(Equal(0))
			})
		})

		Context("When the deployment of autoscaler exsits in mgmt cluster", func() {
			BeforeEach(func() {
				clusterClient.GetDeploymentReturns(*autoscalerDeploy, nil)
			})

			It("should wait for autoscaler deployment", func() {
				Expect(clusterClient.WaitForAutoscalerDeploymentCallCount()).To(Equal(1))
			})
		})
	})
})
