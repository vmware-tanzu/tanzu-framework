// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"os"
	"strings"
	"time"

	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	"github.com/vmware-tanzu/tanzu-framework/addons/testutil"
	runv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

const (
	waitTimeout     = time.Second * 15
	pollingInterval = time.Second * 2
)

var _ = Describe("KappControllerConfig Reconciler", func() {
	var (
		clusterName             string
		clusterResourceFilePath string
	)

	JustBeforeEach(func() {
		By("Creating a cluster and a KappControllerConfig")
		f, err := os.Open(clusterResourceFilePath)
		Expect(err).ToNot(HaveOccurred())
		defer f.Close()
		err = testutil.CreateResources(f, cfg, dynamicClient)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		By("Deletingcluster and KappControllerConfig")
		f, err := os.Open(clusterResourceFilePath)
		Expect(err).ToNot(HaveOccurred())
		defer f.Close()
		testutil.DeleteResources(f, cfg, dynamicClient, true)
	})

	Context("reconcile KappControllerConfig for management cluster", func() {

		BeforeEach(func() {
			clusterName = "test-cluster-1"
			clusterResourceFilePath = "testdata/test-1.yaml"
		})

		It("Should reconcile KappControllerConfig and create data value secret on management cluster", func() {

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

			Eventually(func() bool {
				config := &runv1alpha3.KappControllerConfig{}
				err := k8sClient.Get(ctx, key, config)
				if err != nil {
					return false
				}
				// check owner reference
				if len(config.Finalizers) == 0 || len(config.OwnerReferences) == 0 {
					return false
				}
				Expect(len(config.Finalizers)).Should(Equal(1))
				Expect(len(config.OwnerReferences)).Should(Equal(1))
				Expect(config.OwnerReferences[0].Name).Should(Equal("test-cluster-1"))

				// check spec values
				Expect(config.Spec.Namespace).Should(Equal("test-ns"))
				Expect(config.Spec.KappController.CreateNamespace).Should(Equal(true))
				Expect(config.Spec.KappController.GlobalNamespace).Should(Equal("tanzu-package-repo-global"))
				Expect(config.Spec.KappController.Deployment.Concurrency).Should(Equal(4))
				Expect(config.Spec.KappController.Deployment.HostNetwork).Should(Equal(true))
				Expect(config.Spec.KappController.Deployment.PriorityClassName).Should(Equal("system-cluster-critical"))
				Expect(config.Spec.KappController.Deployment.ApiPort).Should(Equal(10100))
				Expect(config.Spec.KappController.Deployment.MetricsBindAddress).Should(Equal("0"))
				Expect(config.Spec.KappController.Deployment.Tolerations).ShouldNot(BeNil())

				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			Eventually(func() bool {
				secretKey := client.ObjectKey{
					Namespace: "default",
					Name:      util.GenerateDataValueSecretNameFromAddonNames(clusterName, KappControllerAddonName),
				}
				secret := &v1.Secret{}
				err := k8sClient.Get(ctx, secretKey, secret)
				if err != nil {
					return false
				}
				Expect(secret.Type).Should(Equal(v1.SecretTypeOpaque))

				// check data value secret contents
				secretData := string(secret.Data["values.yaml"])
				Expect(strings.Contains(secretData, "createNamespace: true")).Should(BeTrue())
				Expect(strings.Contains(secretData, "globalNamespace: tanzu-package-repo-global")).Should(BeTrue())
				Expect(strings.Contains(secretData, "concurrency: 4")).Should(BeTrue())
				Expect(strings.Contains(secretData, "hostNetwork: true")).Should(BeTrue())
				Expect(strings.Contains(secretData, "coreDNSIP: 100.64.0.10")).Should(BeTrue())
				Expect(strings.Contains(secretData, "- key: CriticalAddonsOnly")).Should(BeTrue())
				Expect(strings.Contains(secretData, "key: node-role.kubernetes.io/master")).Should(BeTrue())
				Expect(strings.Contains(secretData, "key: node.kubernetes.io/not-ready")).Should(BeTrue())
				Expect(strings.Contains(secretData, "key: node.cloudprovider.kubernetes.io/uninitialized")).Should(BeTrue())
				Expect(strings.Contains(secretData, "apiPort: 10100")).Should(BeTrue())
				Expect(strings.Contains(secretData, "metricsBindAddress: \"0\"")).Should(BeTrue())

				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			Eventually(func() bool {
				// Check status.secretRef after reconciliation
				config := &runv1alpha3.KappControllerConfig{}
				err := k8sClient.Get(ctx, key, config)
				if err != nil {
					return false
				}
				Expect(config.Status.SecretRef.Name).Should(Equal(util.GenerateDataValueSecretNameFromAddonNames(clusterName, KappControllerAddonName)))
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

		})

		It("Should reconcile KappControllerConfig deletion in management cluster", func() {

			key := client.ObjectKey{
				Namespace: "default",
				Name:      "test-cluster-1",
			}

			Eventually(func() bool {
				config := &runv1alpha3.KappControllerConfig{}
				err := k8sClient.Get(ctx, key, config)
				if err != nil {
					if errors.IsNotFound(err) {
						return true
					}
					return false
				}

				// Wait until the finalizer is added
				if len(config.Finalizers) == 0 || len(config.OwnerReferences) == 0 {
					return false
				}

				// Delete kappControllerConfig
				err = k8sClient.Delete(ctx, config)
				if err != nil {
					return false
				}
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			Eventually(func() bool {
				secretKey := client.ObjectKey{
					Namespace: "default",
					Name:      util.GenerateDataValueSecretNameFromAddonNames(clusterName, KappControllerAddonName),
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
