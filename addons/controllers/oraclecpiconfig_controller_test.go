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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/test/testutil"
	cpiv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/addonconfigs/cpi/v1alpha1"
)

var _ = Describe("OracleCPIConfig Reconciler", func() {
	var (
		key                     client.ObjectKey
		clusterName             string
		clusterNamespace        string
		clusterResourceFilePath string
	)

	JustBeforeEach(func() {
		By("Creating cluster and OracleCPIConfig resources")
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
		By("Deleting cluster and OracleCPIConfig resources")
		for _, filePath := range []string{clusterResourceFilePath} {
			f, err := os.Open(filePath)
			Expect(err).ToNot(HaveOccurred())
			if err = testutil.DeleteResources(f, cfg, dynamicClient, true); !apierrors.IsNotFound(err) {
				// namespace has been explicitly deleted using testutil.DeleteNamespace
				// ignore its NotFound error here
				Expect(err).ToNot(HaveOccurred())
			}
			Expect(f.Close()).ToNot(HaveOccurred())
		}
	})

	Context("reconcile OracleCPIConfig manifests", func() {

		capociNamespace := "cluster-api-provider-oci-system"

		BeforeEach(func() {
			clusterName = "test-cluster-cpi"
			clusterNamespace = "default"
			clusterResourceFilePath = "testdata/test-oracle-cpi.yaml"
		})

		JustAfterEach(func() {
			Expect(testutil.DeleteNamespace(ctx, clientSet, capociNamespace)).To(Succeed())
		})

		It("Should reconcile OracleCPIConfig and create data values secret", func() {

			// the cpi config object should be deployed
			config := &cpiv1alpha1.OracleCPIConfig{}
			Eventually(func() error {
				return k8sClient.Get(ctx, key, config)
			}).Should(Succeed())

			// the data values secret should be generated
			secret := &v1.Secret{}
			Eventually(func() bool {
				secretKey := client.ObjectKey{
					Namespace: clusterNamespace,
					Name:      fmt.Sprintf("%s-%s-data-values", clusterName, constants.OracleCPIAddonName),
				}
				if err := k8sClient.Get(ctx, secretKey, secret); err != nil {
					return false
				}
				secretData := string(secret.Data["values.yaml"])
				fmt.Println(secretData) // debug dump

				Expect(len(secretData)).Should(Not(BeZero()))
				Expect(strings.Contains(secretData, "compartment: test-compartment")).Should(BeTrue())

				// expect the authentication credentials to be read
				Expect(strings.Contains(secretData, "region: us-sanjose-1")).Should(BeTrue())
				Expect(strings.Contains(secretData, "tenancy: ocid1.tenancy.oc1..aaaaaaaaxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")).Should(BeTrue())
				Expect(strings.Contains(secretData, "user: ocid1.user.oc1..aaaaaaaaxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")).Should(BeTrue())
				Expect(strings.Contains(secretData, "key: |\n")).Should(BeTrue())
				Expect(strings.Contains(secretData, "-----BEGIN PRIVATE KEY-----")).Should(BeTrue())
				Expect(strings.Contains(secretData, "fingerprint: eb:02")).Should(BeTrue())
				Expect(strings.Contains(secretData, "passphrase:")).Should(BeTrue())

				return true
			}).Should(BeTrue())

			// eventually the secret ref to the data values should be updated
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, key, config); err != nil {
					return false
				}
				Expect(config.Status.SecretRef).To(Equal(fmt.Sprintf("%s-%s-data-values", clusterName, constants.CPIAddonName)))
				return true
			})
		})

	})

})
