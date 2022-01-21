// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"fmt"
	"os"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"

	"sigs.k8s.io/cluster-api/util/secret"

	addonsv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/addons/v1alpha1"

	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/addons/testutil"
)

const (
	waitTimeout     = time.Second * 15
	pollingInterval = time.Second * 2
)

var _ = Describe("SecretGenControllerConfig Reconciler", func() {
	var (
		clusterName             string
		clusterResourceFilePath string
	)

	JustBeforeEach(func() {
		//create cluster resources
		By("Creating a cluster and a SecretGenControllerConfig")
		f, err := os.Open(clusterResourceFilePath)
		Expect(err).ToNot(HaveOccurred())
		defer f.Close()
		err = testutil.CreateResources(f, cfg, dynamicClient)
		Expect(err).ToNot(HaveOccurred())

		By("Creating kubeconfig for cluster")
		Expect(testutil.CreateKubeconfigSecret(cfg, clusterName, "default", k8sClient)).To(Succeed())
	})

	AfterEach(func() {
		By("Deletingcluster and SecretGenControllerConfig")
		f, err := os.Open(clusterResourceFilePath)
		Expect(err).ToNot(HaveOccurred())
		defer f.Close()
		testutil.DeleteResources(f, cfg, dynamicClient, true)

		By("Deleting kubeconfig for cluster")
		key := client.ObjectKey{
			Namespace: "default",
			Name:      secret.Name(clusterName, secret.Kubeconfig),
		}
		s := &v1.Secret{}
		Expect(k8sClient.Get(ctx, key, s)).To(Succeed())
		Expect(k8sClient.Delete(ctx, s)).To(Succeed())
	})

	Context("reconcile SecretGenControllerConfig for management cluster", func() {

		BeforeEach(func() {
			clusterName = "test-cluster-1"
			clusterResourceFilePath = "testdata/test-1.yaml"
		})

		It("Should reconcile SecretGenControllerConfig and create data value secret on management cluster", func() {

			key := client.ObjectKey{
				Namespace: "default",
				Name:      "test-cluster-1",
			}

			Eventually(func() bool {
				cluster := &clusterapiv1beta1.Cluster{}
				err := k8sClient.Get(ctx, key, cluster)
				if err != nil {
					return false
				}
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			Eventually(func() bool {
				config := &addonsv1alpha1.SecretGenControllerConfig{}
				err := k8sClient.Get(ctx, key, config)
				if err != nil {
					return false
				}
				Expect(config.Spec.Namespace).Should(Equal("test-ns"))
				Expect(config.Spec.CreateNamespace).Should(Equal(false))
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			Eventually(func() bool {
				secretKey := client.ObjectKey{
					Namespace: "default",
					Name:      fmt.Sprintf("%s-secretgen-controller-data-values", clusterName),
				}
				secret := &v1.Secret{}
				err := k8sClient.Get(ctx, secretKey, secret)
				if err != nil {
					return false
				}
				Expect(secret.Type).Should(Equal(v1.SecretTypeOpaque))
				secretData := string(secret.Data["values.yaml"])
				Expect(strings.Contains(secretData, "namespace: test-ns")).Should(BeTrue())
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

		})

		It("Should reconcile SecretGenControllerConfig deletion in management cluster", func() {

			key := client.ObjectKey{
				Namespace: "default",
				Name:      "test-cluster-1",
			}

			Eventually(func() bool {
				config := &addonsv1alpha1.SecretGenControllerConfig{}
				err := k8sClient.Get(ctx, key, config)
				if err != nil {
					if errors.IsNotFound(err) {
						return true
					}
					return false
				}
				err = k8sClient.Delete(ctx, config)
				if err != nil {
					return false
				}
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			Eventually(func() bool {
				secretKey := client.ObjectKey{
					Namespace: "default",
					Name:      fmt.Sprintf("%s-secretgen-controller-data-values", clusterName),
				}
				secret := &v1.Secret{}
				err := k8sClient.Get(ctx, secretKey, secret)
				if err != nil && errors.IsNotFound(err) {
					return true
				}
				return false
			}, waitTimeout, pollingInterval).Should(BeTrue())

		})
	})

})
