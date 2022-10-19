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
	csiv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/addonconfigs/csi/v1alpha1"
)

var _ = Describe("AwsEbsCSIConfig Reconciler", func() {
	const (
		clusterNamespace = "default"
	)

	var (
		key                     client.ObjectKey
		clusterName             string
		clusterResourceFilePath string
	)

	JustBeforeEach(func() {
		By("Creating cluster and AwsEbsCSIConfig resources")
		key = client.ObjectKey{
			Namespace: clusterNamespace,
			Name:      clusterName,
		}
		f, err := os.Open(clusterResourceFilePath)
		Expect(err).ToNot(HaveOccurred())
		defer f.Close()
		err = testutil.CreateResources(f, cfg, dynamicClient)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		By("Deleting cluster and AwsEbsCSIConfig resources")
		for _, filePath := range []string{clusterResourceFilePath} {
			f, err := os.Open(filePath)
			Expect(err).ToNot(HaveOccurred())
			err = testutil.DeleteResources(f, cfg, dynamicClient, true)
			Expect(err).ToNot(HaveOccurred())
			Expect(f.Close()).ToNot(HaveOccurred())
		}
	})

	Context("reconcile AwsEbsCSIConfig for management cluster", func() {
		const (
			testClusterName = "test-cluster-aws-ebs-csi"
		)

		BeforeEach(func() {
			clusterName = testClusterName
			clusterResourceFilePath = "testdata/test-aws-ebs-csi-config.yaml"
		})

		It("Should reconcile AwsEbsCSIConfig and create data values secret for AwsEbsCSIConfig on management cluster", func() {
			cluster := &clusterapiv1beta1.Cluster{}
			Eventually(func() error {
				if err := k8sClient.Get(ctx, key, cluster); err != nil {
					return fmt.Errorf("Failed to get Cluster '%v': '%v'", key, err)
				}
				return nil
			}, waitTimeout, pollingInterval).Should(Succeed())

			config := &csiv1alpha1.AwsEbsCSIConfig{}
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, key, config); err != nil {
					return false
				}

				// check spec values
				Expect(*config.Spec.AwsEbsCSI.DeploymentReplicas).Should(Equal(int32(2)))
				// check owner reference
				if len(config.OwnerReferences) == 0 {
					return false
				}
				Expect(len(config.OwnerReferences)).Should(Equal(1))
				Expect(config.OwnerReferences[0].Name).Should(Equal(testClusterName))
				Expect(*config.Status.SecretRef).Should(Equal(fmt.Sprintf("%s-%s-data-values", testClusterName, constants.AwsEbsCSIAddonName)))
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			Eventually(func() bool {
				secretKey := client.ObjectKey{
					Namespace: "default",
					Name:      fmt.Sprintf("%s-%s-data-values", clusterName, constants.AwsEbsCSIAddonName),
				}
				secret := &v1.Secret{}
				if err := k8sClient.Get(ctx, secretKey, secret); err != nil {
					return false
				}
				// check data values secret contents
				Expect(secret.Type).Should(Equal(v1.SecretTypeOpaque))
				secretData := string(secret.Data["values.yaml"])
				Expect(strings.Contains(secretData, "deployment_replicas: 2")).Should(BeTrue())
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())
		})
	})
})
