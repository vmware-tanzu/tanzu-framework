// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	capvvmwarev1beta1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/vmware/v1beta1"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/test/testutil"
	csiv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/csi/v1alpha1"
)

var _ = Describe("VSphereCSIConfig Reconciler", func() {
	const (
		clusterNamespace = "default"
	)

	var (
		key                       client.ObjectKey
		clusterName               string
		clusterResourceFilePath   string
		enduringResourcesFilePath string
	)

	JustBeforeEach(func() {
		By("Creating cluster and VSphereCSIConfig resources")
		key = client.ObjectKey{
			Namespace: clusterNamespace,
			Name:      clusterName,
		}
		if enduringResourcesFilePath != "" {
			fers, err := os.Open(enduringResourcesFilePath)
			Expect(err).ToNot(HaveOccurred())
			defer fers.Close()
			err = testutil.EnsureResources(fers, cfg, dynamicClient)
			Expect(err).ToNot(HaveOccurred())
		}
		f, err := os.Open(clusterResourceFilePath)
		Expect(err).ToNot(HaveOccurred())
		defer f.Close()
		err = testutil.CreateResources(f, cfg, dynamicClient)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		By("Deleting cluster and VSphereCSIConfig resources")
		for _, filePath := range []string{clusterResourceFilePath} {
			f, err := os.Open(filePath)
			Expect(err).ToNot(HaveOccurred())
			err = testutil.DeleteResources(f, cfg, dynamicClient, true)
			Expect(err).ToNot(HaveOccurred())
			Expect(f.Close()).ToNot(HaveOccurred())
		}
	})

	Context("reconcile VSphereCSIConfig manifests in non-paravirtual mode", func() {

		const (
			testClusterCsiName = "test-cluster-csi"
		)

		Context("using fully populated VSphereCSIConfig", func() {
			BeforeEach(func() {
				clusterName = testClusterCsiName
				clusterResourceFilePath = "testdata/test-vsphere-csi-non-paravirtual.yaml"
				enduringResourcesFilePath = ""
			})

			It("Should reconcile VSphereCSIConfig and create data values secret for VSphereCSIConfig on management cluster", func() {
				cluster := &clusterapiv1beta1.Cluster{}
				Eventually(func() error {
					if err := k8sClient.Get(ctx, key, cluster); err != nil {
						return fmt.Errorf("Failed to get Cluster '%v': '%v'", key, err)
					}
					return nil
				}, waitTimeout, pollingInterval).Should(Succeed())

				// the csi config object should be deployed
				config := &csiv1alpha1.VSphereCSIConfig{}
				Eventually(func() error {
					if err := k8sClient.Get(ctx, key, config); err != nil {
						return fmt.Errorf("Failed to get VSphereCSIConfig '%v': '%v'", key, err)
					}
					Expect(config.Spec.VSphereCSI.Mode).Should(Equal("vsphereCSI"))
					Expect(config.Spec.VSphereCSI.NonParavirtualConfig).NotTo(BeZero())
					Expect(config.Spec.VSphereCSI.NonParavirtualConfig.TLSThumbprint).Should(Equal("yadayada"))
					Expect(config.Spec.VSphereCSI.NonParavirtualConfig.Namespace).Should(Equal("default"))
					Expect(config.Spec.VSphereCSI.NonParavirtualConfig.ClusterName).Should(Equal("test-clustername"))
					Expect(config.Spec.VSphereCSI.NonParavirtualConfig.Server).Should(Equal("svr-0"))
					Expect(config.Spec.VSphereCSI.NonParavirtualConfig.Datacenter).Should(Equal("dc0"))
					Expect(config.Spec.VSphereCSI.NonParavirtualConfig.PublicNetwork).Should(Equal("8.2.0.0/16"))
					Expect(config.Spec.VSphereCSI.NonParavirtualConfig.VSphereCredentialLocalObjRef.Name).Should(Equal("csi-vsphere-credential"))
					Expect(config.Spec.VSphereCSI.NonParavirtualConfig.Region).Should(Equal("test-region"))
					Expect(config.Spec.VSphereCSI.NonParavirtualConfig.Zone).Should(Equal("test-zone"))
					Expect(*config.Spec.VSphereCSI.NonParavirtualConfig.InsecureFlag).Should(Equal(false))
					Expect(config.Spec.VSphereCSI.NonParavirtualConfig.UseTopologyCategories).NotTo(BeZero())
					Expect(*config.Spec.VSphereCSI.NonParavirtualConfig.UseTopologyCategories).Should(Equal(true))
					Expect(config.Spec.VSphereCSI.NonParavirtualConfig.ProvisionTimeout).Should(Equal("33s"))
					Expect(config.Spec.VSphereCSI.NonParavirtualConfig.AttachTimeout).Should(Equal("77s"))
					Expect(config.Spec.VSphereCSI.NonParavirtualConfig.ResizerTimeout).Should(Equal("99s"))
					Expect(config.Spec.VSphereCSI.NonParavirtualConfig.VSphereVersion).Should(Equal("8.1.2-release"))
					Expect(config.Spec.VSphereCSI.NonParavirtualConfig.HTTPProxy).Should(Equal("1.1.1.1"))
					Expect(config.Spec.VSphereCSI.NonParavirtualConfig.HTTPSProxy).Should(Equal("2.2.2.2"))
					Expect(config.Spec.VSphereCSI.NonParavirtualConfig.NoProxy).Should(Equal("3.3.3.3"))
					Expect(config.Spec.VSphereCSI.NonParavirtualConfig.DeploymentReplicas).NotTo(BeZero())
					Expect(*config.Spec.VSphereCSI.NonParavirtualConfig.DeploymentReplicas).Should(Equal(int32(3)))
					Expect(config.Spec.VSphereCSI.NonParavirtualConfig.WindowsSupport).NotTo(BeZero())
					Expect(*config.Spec.VSphereCSI.NonParavirtualConfig.WindowsSupport).Should(Equal(true))

					if len(config.OwnerReferences) == 0 {
						return fmt.Errorf("OwnerReferences not yet set")
					}
					Expect(len(config.OwnerReferences)).Should(Equal(1))
					Expect(config.OwnerReferences[0].Name).Should(Equal(clusterName))

					return nil
				}, waitTimeout, pollingInterval).Should(Succeed())

				// the data values secret should be generated
				secret := &v1.Secret{}
				Eventually(func() error {
					secretKey := client.ObjectKey{
						Namespace: clusterNamespace,
						Name:      fmt.Sprintf("%s-%s-data-values", clusterName, constants.CSIAddonName),
					}
					if err := k8sClient.Get(ctx, secretKey, secret); err != nil {
						return fmt.Errorf("Failed to get Secret '%v': '%v'", secretKey, err)
					}
					secretData := string(secret.Data["values.yaml"])
					Expect(len(secretData)).Should(Not(BeZero()))
					fmt.Println(secretData) // debug dump
					Expect(strings.Contains(secretData, "vsphereCSI:")).Should(BeTrue())
					Expect(strings.Contains(secretData, "tlsThumbprint: yadayada")).Should(BeTrue())
					Expect(strings.Contains(secretData, "namespace: default")).Should(BeTrue())
					Expect(strings.Contains(secretData, "server: svr-0")).Should(BeTrue())
					Expect(strings.Contains(secretData, "datacenter: dc0")).Should(BeTrue())
					Expect(strings.Contains(secretData, "publicNetwork: 8.2.0.0/16")).Should(BeTrue())
					Expect(strings.Contains(secretData, "username: foo")).Should(BeTrue())
					Expect(strings.Contains(secretData, "password: bar")).Should(BeTrue())
					Expect(strings.Contains(secretData, "region: test-region")).Should(BeTrue())
					Expect(strings.Contains(secretData, "zone: test-zone")).Should(BeTrue())
					Expect(strings.Contains(secretData, "insecureFlag: false")).Should(BeTrue())
					Expect(strings.Contains(secretData, "useTopologyCategories: true")).Should(BeTrue())
					Expect(strings.Contains(secretData, "provisionTimeout: 33s")).Should(BeTrue())
					Expect(strings.Contains(secretData, "attachTimeout: 77s")).Should(BeTrue())
					Expect(strings.Contains(secretData, "resizerTimeout: 99s")).Should(BeTrue())
					Expect(strings.Contains(secretData, "vSphereVersion: 8.1.2-release")).Should(BeTrue())
					Expect(strings.Contains(secretData, "http_proxy: 1.1.1.1")).Should(BeTrue())
					Expect(strings.Contains(secretData, "https_proxy: 2.2.2.2")).Should(BeTrue())
					Expect(strings.Contains(secretData, "no_proxy: 3.3.3.3")).Should(BeTrue())
					Expect(strings.Contains(secretData, "deployment_replicas: 3")).Should(BeTrue())
					Expect(strings.Contains(secretData, "windows_support: true")).Should(BeTrue())

					return nil
				}, waitTimeout, pollingInterval).Should(Succeed())

				// eventually the secret ref to the data values should be updated
				Eventually(func() error {
					if err := k8sClient.Get(ctx, key, config); err != nil {
						return fmt.Errorf("Failed to get VSphereCSIConfig '%v': '%v'", key, err)
					}
					Expect(*config.Status.SecretRef).To(Equal(fmt.Sprintf("%s-%s-data-values", clusterName, constants.CSIAddonName)))
					return nil
				}, waitTimeout, pollingInterval).Should(Succeed())
			})

		})

		Context("using minimally populated VSphereCSIConfig", func() {
			BeforeEach(func() {
				clusterName = "test-cluster-csi-minimal"
				clusterResourceFilePath = "testdata/test-vsphere-csi-non-paravirtual-minimal.yaml"
				enduringResourcesFilePath = ""
			})

			It("Should reconcile VSphereCSIConfig and create data values secret for VSphereCSIConfig", func() {
				cluster := &clusterapiv1beta1.Cluster{}
				Eventually(func() error {
					objKey := client.ObjectKey{Namespace: "default", Name: "test-cluster-csi-minimal"}
					if err := k8sClient.Get(ctx, objKey, cluster); err != nil {
						return fmt.Errorf("Failed to get Cluster '%v': '%v'", objKey, err)
					}
					return nil
				}, waitTimeout, pollingInterval).Should(Succeed())

				// the csi config object should be deployed
				config := &csiv1alpha1.VSphereCSIConfig{}
				Eventually(func() error {
					objKey := client.ObjectKey{Namespace: "default", Name: "test-cluster-csi-minimal"}
					if err := k8sClient.Get(ctx, objKey, config); err != nil {
						return fmt.Errorf("Failed to get VSphereCSIConfig '%v': '%v'", objKey, err)
					}
					Expect(config.Spec.VSphereCSI.Mode).Should(Equal("vsphereCSI"))
					Expect(config.Spec.VSphereCSI.NonParavirtualConfig).To(BeNil())

					if len(config.OwnerReferences) == 0 {
						return fmt.Errorf("Owner references not yet set")
					}
					Expect(len(config.OwnerReferences)).Should(Equal(1))
					Expect(config.OwnerReferences[0].Name).Should(Equal(clusterName))

					return nil
				}, waitTimeout, pollingInterval).Should(Succeed())

				// the data values secret should be generated
				secret := &v1.Secret{}
				Eventually(func() error {
					secretKey := client.ObjectKey{
						Namespace: clusterNamespace,
						Name:      fmt.Sprintf("%s-%s-data-values", clusterName, constants.CSIAddonName),
					}
					if err := k8sClient.Get(ctx, secretKey, secret); err != nil {
						return fmt.Errorf("Failed to get secret '%v' : '%v'", secretKey, err)
					}
					secretData := string(secret.Data["values.yaml"])
					Expect(len(secretData)).Should(Not(BeZero()))
					fmt.Println(secretData) // debug dump
					Expect(strings.Contains(secretData, "vsphereCSI:")).Should(BeTrue())
					Expect(strings.Contains(secretData, "tlsThumbprint: thumbprint-yadayada")).Should(BeTrue())
					Expect(strings.Contains(secretData, "namespace: kube-system")).Should(BeTrue())
					Expect(strings.Contains(secretData, "server: vsphere-server.local")).Should(BeTrue())
					Expect(strings.Contains(secretData, "datacenter: dc0")).Should(BeTrue())
					Expect(strings.Contains(secretData, "publicNetwork: 3.7.9.0/16")).Should(BeTrue())
					Expect(strings.Contains(secretData, "username: administrator@vsphere.local")).Should(BeTrue())
					Expect(strings.Contains(secretData, "password: Admin!23")).Should(BeTrue())
					Expect(strings.Contains(secretData, "region: mombasa")).Should(BeTrue())
					Expect(strings.Contains(secretData, "zone: kisauni")).Should(BeTrue())
					Expect(strings.Contains(secretData, "insecureFlag: true")).Should(BeTrue())
					Expect(strings.Contains(secretData, "useTopologyCategories: true")).Should(BeFalse())
					Expect(strings.Contains(secretData, "provisionTimeout: 300s")).Should(BeTrue())
					Expect(strings.Contains(secretData, "attachTimeout: 300s")).Should(BeTrue())
					Expect(strings.Contains(secretData, "resizerTimeout: 300s")).Should(BeTrue())
					Expect(strings.Contains(secretData, "vSphereVersion: 8.3.7-release")).Should(BeTrue())
					Expect(strings.Contains(secretData, "http_proxy: foo.com")).Should(BeTrue())
					Expect(strings.Contains(secretData, "https_proxy: bar.com")).Should(BeTrue())
					Expect(strings.Contains(secretData, "no_proxy: foobar.com")).Should(BeTrue())
					Expect(strings.Contains(secretData, "deployment_replicas: 2")).Should(BeTrue())
					Expect(strings.Contains(secretData, "windows_support: true")).Should(BeTrue())

					return nil
				}, waitTimeout, pollingInterval).Should(Succeed())

				// eventually the secret ref to the data values should be updated
				Eventually(func() error {
					objKey := client.ObjectKey{Namespace: "default", Name: "test-cluster-csi-minimal"}
					if err := k8sClient.Get(ctx, objKey, config); err != nil {
						return fmt.Errorf("Failed to get VSphereCSIConfig '%v' : '%v", objKey, err)
					}
					if config.Status.SecretRef == nil {
						return fmt.Errorf("Secret status of VSphereCSIConfig is not yet updated: '%v'", objKey)
					}
					Expect(*config.Status.SecretRef).To(Equal(fmt.Sprintf("%s-%s-data-values", clusterName, constants.CSIAddonName)))
					return nil
				}, waitTimeout, pollingInterval).Should(Succeed())
			})

		})

	})

	Context("reconcile VSphereCSIConfig manifests in paravirtual mode", func() {
		BeforeEach(func() {
			clusterName = "test-cluster-pv-csi"
			clusterResourceFilePath = "testdata/test-vsphere-csi-paravirtual.yaml"
			enduringResourcesFilePath = "testdata/vmware-csi-system-ns.yaml"
		})
		It("Should reconcile VSphereCSIConfig and create data values secret for VSphereCSIConfig on management cluster", func() {
			// the data values secret should be generated
			secret := &v1.Secret{}
			Eventually(func() error {
				secretKey := client.ObjectKey{
					Namespace: clusterNamespace,
					Name:      fmt.Sprintf("%s-%s-data-values", clusterName, constants.PVCSIAddonName),
				}
				if err := k8sClient.Get(ctx, secretKey, secret); err != nil {
					return fmt.Errorf("Failed to get Secret '%v': '%v'", secretKey, err)
				}
				secretData := string(secret.Data["values.yaml"])
				fmt.Println(secretData) // debug dump
				Expect(len(secretData)).Should(Not(BeZero()))
				Expect(strings.Contains(secretData, "vspherePVCSI:")).Should(BeTrue())
				Expect(strings.Contains(secretData, "cluster_name: test-cluster-pv-csi")).Should(BeTrue())
				match, _ := regexp.MatchString("cluster_uid: [a-z0-9]{8}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{12}", secretData)
				Expect(match).Should(BeTrue())
				Expect(strings.Contains(secretData, "namespace: vmware-system-csi")).Should(BeTrue())
				Expect(strings.Contains(secretData, "supervisor_master_endpoint_hostname: supervisor.default.svc")).Should(BeTrue())
				Expect(strings.Contains(secretData, "supervisor_master_port: 6443")).Should(BeTrue())
				Expect(strings.Contains(secretData, "feature_states:")).Should(BeTrue())
				Expect(strings.Contains(secretData, "state1: value1")).Should(BeTrue())
				Expect(strings.Contains(secretData, "state2: value2")).Should(BeTrue())
				Expect(strings.Contains(secretData, "state3: value3")).Should(BeTrue())

				return nil
			}, waitTimeout, pollingInterval).Should(Succeed())

			// eventually the secret ref to the data values should be updated
			config := &csiv1alpha1.VSphereCSIConfig{}
			Eventually(func() error {
				configKey := client.ObjectKey{
					Namespace: clusterNamespace,
					Name:      clusterName,
				}
				if err := k8sClient.Get(ctx, configKey, config); err != nil {
					return fmt.Errorf("Failed to get vsphereconfig '%v': '%v'", configKey, err)
				}
				if config.Status.SecretRef == nil {
					return fmt.Errorf("VSphereConfig status not yet updated: %v", configKey)
				}
				Expect(*config.Status.SecretRef).To(Equal(fmt.Sprintf("%s-%s-data-values", clusterName, constants.PVCSIAddonName)))
				return nil
			}, waitTimeout, pollingInterval).Should(Succeed())
		})

		It("Should reconcile ProviderServiceAccount", func() {
			vsphereClusterName := "test-cluster-pv-csi-kl5tm"
			serviceAccount := &capvvmwarev1beta1.ProviderServiceAccount{}
			Eventually(func() error {
				serviceAccountKey := client.ObjectKey{
					Namespace: clusterNamespace,
					Name:      fmt.Sprintf("%s-%s", vsphereClusterName, "pvcsi"),
				}
				if err := k8sClient.Get(ctx, serviceAccountKey, serviceAccount); err != nil {
					return fmt.Errorf("Failed to get provider service account '%v' : '%v'", serviceAccountKey, err)
				}
				Expect(serviceAccount.Spec.Ref.Name).To(Equal(vsphereClusterName))
				Expect(serviceAccount.Spec.Ref.Namespace).To(Equal(key.Namespace))
				Expect(serviceAccount.Spec.Rules).To(HaveLen(6))
				Expect(serviceAccount.Spec.TargetNamespace).To(Equal("vmware-system-csi"))
				Expect(serviceAccount.Spec.TargetSecretName).To(Equal("pvcsi-provider-creds"))
				return nil
			}, waitTimeout, pollingInterval).Should(Succeed())
		})

		It("Should reconcile aggregated cluster role", func() {
			clusterRole := &rbacv1.ClusterRole{}
			Eventually(func() error {
				key := client.ObjectKey{
					Name: constants.VsphereCSIProviderServiceAccountAggregatedClusterRole,
				}
				if err := k8sClient.Get(ctx, key, clusterRole); err != nil {
					return err
				}
				Expect(clusterRole.Labels).To(Equal(map[string]string{
					constants.CAPVClusterRoleAggregationRuleLabelSelectorKey: constants.CAPVClusterRoleAggregationRuleLabelSelectorValue,
				}))
				Expect(clusterRole.Rules).To(HaveLen(6))
				return nil
			}, waitTimeout, pollingInterval).Should(Succeed())
		})
	})

	Context("Reconcile VSphereCSIConfig used as template", func() {

		BeforeEach(func() {
			clusterName = "test-cluster-csi-template"
			clusterResourceFilePath = "testdata/test-vsphere-csi-template-config.yaml"
			enduringResourcesFilePath = ""
		})

		It("Should skip the reconciliation", func() {

			key.Namespace = addonNamespace
			config := &csiv1alpha1.VSphereCSIConfig{}
			Expect(k8sClient.Get(ctx, key, config)).To(Succeed())

			By("OwnerReferences is not set")
			Expect(len(config.OwnerReferences)).Should(Equal(0))
		})
	})
})
