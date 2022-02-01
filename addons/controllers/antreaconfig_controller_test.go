// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	"github.com/vmware-tanzu/tanzu-framework/addons/testutil"
	cniv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cni/v1alpha1"
)

var _ = Describe("AntreaConfig Reconciler", func() {
	var (
		configCRName            string
		clusterResourceFilePath string
	)

	JustBeforeEach(func() {
		// create cluster resources
		By("Creating a cluster and a AntreaConfig")
		f, err := os.Open(clusterResourceFilePath)
		Expect(err).ToNot(HaveOccurred())
		defer f.Close()
		err = testutil.CreateResources(f, cfg, dynamicClient)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		By("Deleting cluster and AntreaConfig")
		f, err := os.Open(clusterResourceFilePath)
		Expect(err).ToNot(HaveOccurred())
		defer f.Close()
		testutil.DeleteResources(f, cfg, dynamicClient, true)
	})

	Context("Reconcile AntreaConfig for management cluster", func() {

		BeforeEach(func() {
			configCRName = "test-cluster-1"
			clusterResourceFilePath = "testdata/antrea-test-1.yaml"
		})

		It("Should reconcile AntreaConfig and create data value secret on management cluster", func() {

			key := client.ObjectKey{
				Namespace: "default",
				Name:      "test-cluster-1",
			}

			cluster := &clusterapiv1beta1.Cluster{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, key, cluster)
				if err != nil {
					return false
				}
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			config := &cniv1alpha1.AntreaConfig{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, key, config)
				if err != nil {
					return false
				}

				// Check owner reference
				if len(config.OwnerReferences) == 0 {
					return false
				}
				Expect(len(config.Finalizers)).Should(Equal(1))
				Expect(len(config.OwnerReferences)).Should(Equal(1))
				Expect(config.OwnerReferences[0].Name).Should(Equal("test-cluster-1"))

				// TODO: Possible to add more checks here
				Expect(config.Spec.Antrea.AntConfig.TrafficEncapMode).Should(Equal("encap"))
				Expect(config.Spec.Antrea.AntConfig.FeatureGates.AntreaTraceflow).Should(Equal(false))
				Expect(config.Spec.Antrea.AntConfig.FeatureGates.AntreaPolicy).Should(Equal(true))
				Expect(config.Spec.Antrea.AntConfig.FeatureGates.FlowExporter).Should(Equal(false))

				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			Eventually(func() bool {
				cluster := &clusterapiv1beta1.Cluster{}
				err := k8sClient.Get(ctx, key, cluster)
				if err != nil {
					return false
				}

				serviceCIDR, serviceCIDRv6, err := util.GetServiceCIDRs(cluster)
				if err != nil {
					return false
				}

				infraProvider, err := util.GetInfraProvider(cluster)
				if err != nil {
					return false
				}

				// Check infraProvider values
				Expect(infraProvider).Should(Equal("docker"))

				// Check ServiceCIDR and ServiceCIDRv6 values
				Expect(serviceCIDR).Should(Equal("192.168.0.0/16"))
				Expect(serviceCIDRv6).Should(Equal("fd00:100:96::/48"))
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			Eventually(func() bool {
				secretKey := client.ObjectKey{
					Namespace: "default",
					Name:      util.GenerateDataValueSecretNameFromAddonAndClusterNames(configCRName, constants.AntreaAddonName),
				}
				secret := &v1.Secret{}
				err := k8sClient.Get(ctx, secretKey, secret)
				if err != nil {

					return false
				}
				Expect(secret.Type).Should(Equal(v1.SecretTypeOpaque))
				secretData := string(secret.Data["values.yaml"])
				Expect(strings.Contains(secretData, "AntreaPolicy: true")).Should(BeTrue())
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			Eventually(func() bool {
				// Check status.secretRef after reconciliation
				config := &cniv1alpha1.AntreaConfig{}
				err := k8sClient.Get(ctx, key, config)
				if err != nil {
					return false
				}
				Expect(config.Status.SecretRef).Should(Equal(util.GenerateDataValueSecretNameFromAddonAndClusterNames(configCRName, constants.AntreaAddonName)))
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

		})

	})

})
