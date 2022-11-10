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

	cutil "github.com/vmware-tanzu/tanzu-framework/addons/controllers/utils"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	"github.com/vmware-tanzu/tanzu-framework/addons/test/testutil"
	runv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

const (
	testKappController1 = "test-kapp-controller-1"
)

var _ = Describe("KappControllerConfig Reconciler", func() {
	var (
		clusterName             string
		kappConfigCRName        string
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
		By("Deleting cluster and KappControllerConfig")
		f, err := os.Open(clusterResourceFilePath)
		Expect(err).ToNot(HaveOccurred())
		defer f.Close()
		// Best effort resource deletion
		_ = testutil.DeleteResources(f, cfg, dynamicClient, true)
	})

	Context("reconcile KappControllerConfig for management cluster", func() {

		BeforeEach(func() {
			clusterName = testKappController1
			kappConfigCRName = util.GeneratePackageSecretName(clusterName, constants.KappControllerDefaultRefName)
			clusterResourceFilePath = "testdata/test-kapp-controller-1.yaml"
		})

		It("Should reconcile KappControllerConfig and create data value secret on management cluster", func() {

			key := client.ObjectKey{
				Namespace: defaultString,
				Name:      clusterName,
			}
			configKey := client.ObjectKey{
				Namespace: defaultString,
				Name:      kappConfigCRName,
			}

			cluster := &clusterapiv1beta1.Cluster{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, key, cluster)
				return err == nil
			}, waitTimeout, pollingInterval).Should(BeTrue())

			config := &runv1alpha3.KappControllerConfig{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, configKey, config)
				if err != nil {
					return false
				}
				// check owner reference
				return cutil.VerifyOwnerRef(config, clusterName, constants.ClusterKind)

			}, waitTimeout, pollingInterval).Should(BeTrue())

			// check spec values
			Expect(config.Spec.Namespace).Should(Equal("test-ns"))
			Expect(config.Spec.KappController.CreateNamespace).Should(Equal(true))
			Expect(config.Spec.KappController.GlobalNamespace).Should(Equal("tanzu-package-repo-global"))
			Expect(config.Spec.KappController.Deployment.Concurrency).Should(Equal(4))
			Expect(config.Spec.KappController.Deployment.HostNetwork).Should(Equal(true))
			Expect(config.Spec.KappController.Deployment.PriorityClassName).Should(Equal("system-cluster-critical"))
			Expect(config.Spec.KappController.Deployment.APIPort).Should(Equal(10100))
			Expect(config.Spec.KappController.Deployment.MetricsBindAddress).Should(Equal("0"))
			Expect(config.Spec.KappController.Deployment.Tolerations).ShouldNot(BeNil())

			secret := &v1.Secret{}
			Eventually(func() bool {
				secretKey := client.ObjectKey{
					Namespace: defaultString,
					Name:      util.GenerateDataValueSecretName(cluster.Name, constants.KappControllerAddonName),
				}
				err := k8sClient.Get(ctx, secretKey, secret)
				if err != nil {
					return false
				}
				return secret.Type == v1.SecretTypeOpaque
			}, waitTimeout, pollingInterval).Should(BeTrue())

			// check data value secret contents
			secretData := string(secret.Data["values.yaml"])
			Expect(strings.Contains(secretData, "createNamespace: true")).Should(BeTrue())
			Expect(strings.Contains(secretData, "globalNamespace: tanzu-package-repo-global")).Should(BeTrue())
			Expect(strings.Contains(secretData, "concurrency: 4")).Should(BeTrue())
			Expect(strings.Contains(secretData, "hostNetwork: true")).Should(BeTrue())
			Expect(strings.Contains(secretData, "coreDNSIP: 100.64.0.10")).Should(BeTrue())
			Expect(strings.Contains(secretData, "- key: CriticalAddonsOnly")).Should(BeTrue())
			Expect(strings.Contains(secretData, "node-role.kubernetes.io/control-plane: \"\"")).Should(BeTrue())
			Expect(strings.Contains(secretData, "key: node.kubernetes.io/not-ready")).Should(BeTrue())
			Expect(strings.Contains(secretData, "key: node.cloudprovider.kubernetes.io/uninitialized")).Should(BeTrue())
			Expect(strings.Contains(secretData, "apiPort: 10100")).Should(BeTrue())
			Expect(strings.Contains(secretData, "metricsBindAddress: \"0\"")).Should(BeTrue())
			Expect(strings.Contains(secretData, "key: node-role.kubernetes.io/control-plane")).Should(BeTrue())

			Eventually(func() bool {
				if !strings.Contains(secretData, "caCerts: dummyCertificate") ||
					!strings.Contains(secretData, "httpsProxy: bar.com") ||
					!strings.Contains(secretData, "noProxy: foobar.com") ||
					!strings.Contains(secretData, "dangerousSkipTLSVerify: registry1,registry2") {
					return false
				}

				// user input should override cluster-wide config
				if !strings.Contains(secretData, "httpProxy: overwrite.foo.com") {
					return false
				}

				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			Eventually(func() bool {
				// Check status.secretRef after reconciliation
				config := &runv1alpha3.KappControllerConfig{}
				err := k8sClient.Get(ctx, configKey, config)
				if err != nil {
					return false
				}
				return config.Status.SecretRef == util.GenerateDataValueSecretName(cluster.Name, constants.KappControllerAddonName)
			}, waitTimeout, pollingInterval).Should(BeTrue())

		})

	})

	Context("Reconcile KappControllerConfig used as template", func() {

		BeforeEach(func() {
			clusterName = testKappController1
			kappConfigCRName = util.GeneratePackageSecretName(clusterName, constants.KappControllerDefaultRefName)
			clusterResourceFilePath = "testdata/test-kapp-controller-template-config-1.yaml"
		})

		It("Should skip the reconciliation", func() {

			configKey := client.ObjectKey{
				Namespace: addonNamespace,
				Name:      kappConfigCRName,
			}
			config := &runv1alpha3.KappControllerConfig{}
			Expect(k8sClient.Get(ctx, configKey, config)).To(Succeed())

			By("OwnerReferences is not set")
			Expect(len(config.OwnerReferences)).Should(Equal(0))
		})
	})
})
