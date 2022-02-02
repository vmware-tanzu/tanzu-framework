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
	"github.com/vmware-tanzu/tanzu-framework/addons/testutil"
	cniv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cni/v1alpha1"
)

const testCluster = "test-cluster-calico"

var _ = Describe("CalicoConfig Reconciler", func() {
	var (
		clusterName             string
		clusterResourceFilePath string
	)

	JustBeforeEach(func() {
		By("Creating cluster and CalicoConfig resources")
		f, err := os.Open(clusterResourceFilePath)
		Expect(err).ToNot(HaveOccurred())
		defer f.Close()
		err = testutil.CreateResources(f, cfg, dynamicClient)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		By("Deleting cluster and CalicoConfig resources")
		f, err := os.Open(clusterResourceFilePath)
		Expect(err).ToNot(HaveOccurred())
		defer f.Close()
		err = testutil.DeleteResources(f, cfg, dynamicClient, true)
		Expect(err).ToNot(HaveOccurred())
	})

	Context("reconcile CalicoConfig for management cluster", func() {
		BeforeEach(func() {
			clusterName = testCluster
			clusterResourceFilePath = "testdata/test-calico.yaml"
		})

		It("Should reconcile CalicoConfig and create data values secret for CalicoConfig on management cluster", func() {
			key := client.ObjectKey{
				Namespace: "default",
				Name:      testCluster,
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

				//check spec values
				Expect(config.Spec.Namespace).Should(Equal("kube-system"))
				Expect(config.Spec.Calico.Config.VethMTU).Should(Equal(int64(0)))

				// check owner reference
				if len(config.OwnerReferences) == 0 {
					return false
				}
				Expect(len(config.OwnerReferences)).Should(Equal(1))
				Expect(config.OwnerReferences[0].Name).Should(Equal(testCluster))

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
				Expect(strings.Contains(secretData, "namespace: kube-system")).Should(BeTrue())
				Expect(strings.Contains(secretData, "infraProvider: vsphere")).Should(BeTrue())
				Expect(strings.Contains(secretData, "ipFamily: ipv4,ipv6")).Should(BeTrue())
				Expect(strings.Contains(secretData, "clusterCIDR: 192.168.0.0/16,fd00:100:96::/48")).Should(BeTrue())

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
})
