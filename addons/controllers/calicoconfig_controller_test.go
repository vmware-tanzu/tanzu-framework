// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"fmt"
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/test/testutil"
	cniv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cni/v1alpha1"
)

const (
	testClusterCalico1            = "test-cluster-calico-1"
	testClusterCalico2            = "test-cluster-calico-2"
	testDataCalico1               = "testdata/test-calico-1.yaml"
	testDataCalico2               = "testdata/test-calico-2.yaml"
	testDataCalicoTemplateConfig1 = "testdata/test-calico-template-config-1.yaml"
)

var _ = Describe("CalicoConfig Reconciler and Webhooks", func() {
	var (
		clusterName             string
		clusterResourceFilePath string
	)

	JustBeforeEach(func() {
		// Create the admission webhooks
		f, err := os.Open(cniWebhookManifestFile)
		Expect(err).ToNot(HaveOccurred())
		err = testutil.CreateResources(f, cfg, dynamicClient)
		Expect(err).ToNot(HaveOccurred())
		f.Close()

		// set up the certificates and webhook before creating any objects
		By("Creating and installing new certificates for Calico Admission Webhooks")
		err = testutil.SetupWebhookCertificates(ctx, k8sClient, k8sConfig, &webhookCertDetails)
		Expect(err).ToNot(HaveOccurred())

		By("Creating cluster and CalicoConfig resources")
		f, err = os.Open(clusterResourceFilePath)
		Expect(err).ToNot(HaveOccurred())
		err = testutil.CreateResources(f, cfg, dynamicClient)
		Expect(err).ToNot(HaveOccurred())
		f.Close()
	})

	AfterEach(func() {
		By("Deleting cluster and CalicoConfig resources")
		f, err := os.Open(clusterResourceFilePath)
		Expect(err).ToNot(HaveOccurred())
		err = testutil.DeleteResources(f, cfg, dynamicClient, true)
		Expect(err).ToNot(HaveOccurred())
		f.Close()

		By("Deleting the Admission Webhook configuration for Calico")
		f, err = os.Open(cniWebhookManifestFile)
		Expect(err).ToNot(HaveOccurred())
		err = testutil.DeleteResources(f, cfg, dynamicClient, true)
		Expect(err).ToNot(HaveOccurred())
		f.Close()
	})

	Context("reconcile default CalicoConfig for management cluster on dual-stack CIDR", func() {
		BeforeEach(func() {
			clusterName = testClusterCalico1
			clusterResourceFilePath = testDataCalico1
		})

		It("Should reconcile CalicoConfig and create data values secret for CalicoConfig on management cluster", func() {
			key := client.ObjectKey{
				Namespace: "default",
				Name:      testClusterCalico1,
			}

			cluster := &clusterapiv1beta1.Cluster{}
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, key, cluster); err != nil {
					return false
				}
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			config := &cniv1alpha1.CalicoConfig{}
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, key, config); err != nil {
					return false
				}

				// check spec values
				Expect(config.Spec.Calico.Config.VethMTU).Should(Equal(int64(0)))
				Expect(config.Spec.Calico.Config.SkipCNIBinaries).Should(BeTrue())

				// check owner reference
				if len(config.OwnerReferences) == 0 {
					return false
				}
				Expect(len(config.OwnerReferences)).Should(Equal(1))
				Expect(config.OwnerReferences[0].Name).Should(Equal(testClusterCalico1))

				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			Eventually(func() bool {
				secretKey := client.ObjectKey{
					Namespace: "default",
					Name:      fmt.Sprintf("%s-%s-data-values", clusterName, constants.CalicoAddonName),
				}
				secret := &v1.Secret{}
				if err := k8sClient.Get(ctx, secretKey, secret); err != nil {
					return false
				}

				// check data values secret contents
				Expect(secret.Type).Should(Equal(v1.SecretTypeOpaque))
				secretData := string(secret.Data["values.yaml"])
				Expect(strings.Contains(secretData, "infraProvider: vsphere")).Should(BeTrue())
				Expect(strings.Contains(secretData, "ipFamily: ipv4,ipv6")).Should(BeTrue())
				Expect(strings.Contains(secretData, "clusterCIDR: 192.168.0.0/16,fd00:100:96::/48")).Should(BeTrue())
				Expect(strings.Contains(secretData, "vethMTU: \"0\"")).Should(BeTrue())
				Expect(strings.Contains(secretData, "skipCNIBinaries: true")).Should(BeTrue())

				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			Eventually(func() bool {
				config := &cniv1alpha1.CalicoConfig{}
				err := k8sClient.Get(ctx, key, config)
				if err != nil {
					return false
				}
				// Check status.secretName after reconciliation
				Expect(config.Status.SecretRef).Should(Equal(fmt.Sprintf("%s-%s-data-values", clusterName, constants.CalicoAddonName)))

				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())
		})
	})

	Context("reconcile mtu customized and cni binaries installation skipped CalicoConfig for management cluster on ipv4 CIDR", func() {
		BeforeEach(func() {
			clusterName = testClusterCalico2
			clusterResourceFilePath = testDataCalico2
		})

		It("Should reconcile CalicoConfig and create data values secret for CalicoConfig on management cluster", func() {
			key := client.ObjectKey{
				Namespace: "default",
				Name:      testClusterCalico2,
			}

			cluster := &clusterapiv1beta1.Cluster{}
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, key, cluster); err != nil {
					return false
				}
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			config := &cniv1alpha1.CalicoConfig{}
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, key, config); err != nil {
					return false
				}

				// check spec values
				Expect(config.Spec.Calico.Config.VethMTU).Should(Equal(int64(1420)))
				Expect(config.Spec.Calico.Config.SkipCNIBinaries).Should(BeFalse())

				// check owner reference
				if len(config.OwnerReferences) == 0 {
					return false
				}
				Expect(len(config.OwnerReferences)).Should(Equal(1))
				Expect(config.OwnerReferences[0].Name).Should(Equal(testClusterCalico2))

				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			Eventually(func() bool {
				secretKey := client.ObjectKey{
					Namespace: "default",
					Name:      fmt.Sprintf("%s-%s-data-values", clusterName, constants.CalicoAddonName),
				}
				secret := &v1.Secret{}
				if err := k8sClient.Get(ctx, secretKey, secret); err != nil {
					return false
				}

				// check data values secret contents
				Expect(secret.Type).Should(Equal(v1.SecretTypeOpaque))
				secretData := string(secret.Data["values.yaml"])
				Expect(strings.Contains(secretData, "infraProvider: docker")).Should(BeTrue())
				Expect(strings.Contains(secretData, "ipFamily: ipv4")).Should(BeTrue())
				Expect(strings.Contains(secretData, "clusterCIDR: 192.168.0.0/16")).Should(BeTrue())
				Expect(strings.Contains(secretData, "vethMTU: \"1420\"")).Should(BeTrue())
				Expect(strings.Contains(secretData, "skipCNIBinaries: false")).Should(BeTrue())

				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			Eventually(func() bool {
				config := &cniv1alpha1.CalicoConfig{}
				err := k8sClient.Get(ctx, key, config)
				if err != nil {
					return false
				}
				// Check status.secretName after reconciliation
				Expect(config.Status.SecretRef).Should(Equal(fmt.Sprintf("%s-%s-data-values", clusterName, constants.CalicoAddonName)))

				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())
		})
	})

	Context("Reconcile CalicoConfig used as template", func() {

		BeforeEach(func() {
			clusterName = testClusterCalico1
			clusterResourceFilePath = testDataCalicoTemplateConfig1
		})

		It("Should skip the reconciliation", func() {

			key := client.ObjectKey{
				Namespace: addonNamespace,
				Name:      clusterName,
			}
			config := &cniv1alpha1.CalicoConfig{}
			Expect(k8sClient.Get(ctx, key, config)).To(Succeed())

			By("OwnerReferences is not set")
			Expect(len(config.OwnerReferences)).Should(Equal(0))
		})
	})
})
