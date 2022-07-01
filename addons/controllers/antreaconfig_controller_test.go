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
	"github.com/vmware-tanzu/tanzu-framework/addons/test/testutil"
	cniv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cni/v1alpha1"
)

var _ = Describe("AntreaConfig Reconciler and Webhooks", func() {
	var (
		configCRName            string
		clusterResourceFilePath string
		err                     error
		f                       *os.File
	)

	const (
		antreaManifestsTestFile1               = "testdata/antrea-test-1.yaml"
		antreaTemplateConfigManifestsTestFile1 = "testdata/antrea-test-template-config-1.yaml"
		antreaTestCluster1                     = "test-cluster-4"
	)

	JustBeforeEach(func() {
		// Create the admission webhooks
		f, err = os.Open(cniWebhookManifestFile)
		Expect(err).ToNot(HaveOccurred())
		err = testutil.CreateResources(f, cfg, dynamicClient)
		Expect(err).ToNot(HaveOccurred())
		f.Close()

		// set up the certificates and webhook before creating any objects
		By("Creating and installing new certificates for Antrea Admission Webhooks")
		err = testutil.SetupWebhookCertificates(ctx, k8sClient, k8sConfig, &webhookCertDetails)
		Expect(err).ToNot(HaveOccurred())

		// create cluster resources
		By("Creating a cluster and a AntreaConfig")
		f, err = os.Open(clusterResourceFilePath)
		Expect(err).ToNot(HaveOccurred())
		err = testutil.CreateResources(f, cfg, dynamicClient)
		Expect(err).ToNot(HaveOccurred())
		f.Close()
	})

	AfterEach(func() {
		By("Deleting cluster and AntreaConfig")
		f, err = os.Open(clusterResourceFilePath)
		Expect(err).ToNot(HaveOccurred())
		err = testutil.DeleteResources(f, cfg, dynamicClient, true)
		Expect(err).ToNot(HaveOccurred())
		f.Close()

		By("Deleting the Admission Webhook configuration for Antrea")
		f, err = os.Open(cniWebhookManifestFile)
		Expect(err).ToNot(HaveOccurred())
		err = testutil.DeleteResources(f, cfg, dynamicClient, true)
		Expect(err).ToNot(HaveOccurred())
		f.Close()
	})

	Context("Reconcile AntreaConfig for management cluster", func() {

		BeforeEach(func() {
			configCRName = antreaTestCluster1
			clusterResourceFilePath = antreaManifestsTestFile1
		})

		It("Should reconcile AntreaConfig and create data value secret on management cluster", func() {

			key := client.ObjectKey{
				Namespace: "default",
				Name:      configCRName,
			}

			cluster := &clusterapiv1beta1.Cluster{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, key, cluster)
				return err == nil
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

				Expect(len(config.OwnerReferences)).Should(Equal(1))
				Expect(config.OwnerReferences[0].Name).Should(Equal(configCRName))

				Expect(config.Spec.Antrea.AntreaConfigDataValue.TrafficEncapMode).Should(Equal("encap"))
				Expect(config.Spec.Antrea.AntreaConfigDataValue.FeatureGates.AntreaTraceflow).Should(Equal(false))
				Expect(config.Spec.Antrea.AntreaConfigDataValue.FeatureGates.AntreaPolicy).Should(Equal(true))
				Expect(config.Spec.Antrea.AntreaConfigDataValue.FeatureGates.FlowExporter).Should(Equal(false))
				Expect(config.Spec.Antrea.AntreaConfigDataValue.FeatureGates.AntreaIPAM).Should(Equal(false))
				Expect(config.Spec.Antrea.AntreaConfigDataValue.FeatureGates.ServiceExternalIP).Should(Equal(false))
				Expect(config.Spec.Antrea.AntreaConfigDataValue.FeatureGates.Multicast).Should(Equal(false))

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
					Name:      util.GenerateDataValueSecretName(configCRName, constants.AntreaAddonName),
				}
				secret := &v1.Secret{}
				err := k8sClient.Get(ctx, secretKey, secret)
				if err != nil {
					return false
				}
				Expect(secret.Type).Should(Equal(v1.SecretTypeOpaque))

				// check data value secret contents
				secretData := string(secret.Data["values.yaml"])

				Expect(strings.Contains(secretData, "serviceCIDR: 192.168.0.0/16")).Should(BeTrue())
				Expect(strings.Contains(secretData, "serviceCIDRv6: fd00:100:96::/48")).Should(BeTrue())
				Expect(strings.Contains(secretData, "infraProvider: docker")).Should(BeTrue())

				Expect(strings.Contains(secretData, "trafficEncapMode: encap")).Should(BeTrue())
				Expect(strings.Contains(secretData, "tlsCipherSuites: TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_RSA_WITH_AES_256_GCM_SHA384")).Should(BeTrue())
				Expect(strings.Contains(secretData, "AntreaProxy: true")).Should(BeTrue())
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
				Expect(config.Status.SecretRef).Should(Equal(util.GenerateDataValueSecretName(configCRName, constants.AntreaAddonName)))
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

		})

	})

	Context("Reconcile AntreaConfig used as template", func() {

		BeforeEach(func() {
			configCRName = antreaTestCluster1
			clusterResourceFilePath = antreaTemplateConfigManifestsTestFile1
		})

		It("Should skip the reconciliation", func() {

			key := client.ObjectKey{
				Namespace: addonNamespace,
				Name:      configCRName,
			}
			config := &cniv1alpha1.AntreaConfig{}
			Expect(k8sClient.Get(ctx, key, config)).To(Succeed())

			By("OwnerReferences is not set")
			Expect(len(config.OwnerReferences)).Should(Equal(0))
		})
	})

	Context("Mutating webhooks for AntreaConfig", func() {

		BeforeEach(func() {
			configCRName = antreaTestCluster1
			clusterResourceFilePath = antreaManifestsTestFile1
		})

		It("Should fail mutating webhooks for immutable field for AntreaConfig", func() {

			key := client.ObjectKey{
				Namespace: "default",
				Name:      configCRName,
			}
			config := &cniv1alpha1.AntreaConfig{}
			Expect(k8sClient.Get(ctx, key, config)).To(Succeed())

			By("Trying to update the immutable TrafficEncapMode field in Antrea Spec")
			config.Spec.Antrea.AntreaConfigDataValue.TrafficEncapMode = "noEncap"
			Expect(k8sClient.Update(ctx, config)).ToNot(Succeed())
		})
	})
})
