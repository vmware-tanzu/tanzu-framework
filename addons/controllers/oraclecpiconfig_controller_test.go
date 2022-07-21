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
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/test/testutil"
	cpiv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cpi/v1alpha1"
)

var _ = Describe("OracleCPIConfig Reconciler", func() {
	var (
		key                     client.ObjectKey
		clusterName             string
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
			err = testutil.DeleteResources(f, cfg, dynamicClient, true)
			Expect(err).ToNot(HaveOccurred())
			Expect(f.Close()).ToNot(HaveOccurred())
		}
	})

	Context("reconcile OracleCPIConfig manifests", func() {
		BeforeEach(func() {
			clusterName = "test-cluster-cpi"
			clusterResourceFilePath = "testdata/test-oracle-cpi.yaml"
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
				Expect(len(secretData)).Should(Not(BeZero()))
				Expect(strings.Contains(secretData, "compartment: test-compartment")).Should(BeTrue())
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
