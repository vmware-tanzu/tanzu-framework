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
	csiv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/csi/v1alpha1"
)

var _ = Describe("CSIConfig Reconciler", func() {
	const (
		clusterNamespace = "default"
	)

	var (
		key                     client.ObjectKey
		clusterName             string
		clusterResourceFilePath string
	)

	JustBeforeEach(func() {
		By("Creating cluster and CSIConfig resources")
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
		By("Deleting cluster and CSIConfig resources")
		for _, filePath := range []string{clusterResourceFilePath} {
			f, err := os.Open(filePath)
			Expect(err).ToNot(HaveOccurred())
			err = testutil.DeleteResources(f, cfg, dynamicClient, true)
			Expect(err).ToNot(HaveOccurred())
			Expect(f.Close()).ToNot(HaveOccurred())
		}
	})

	Context("reconcile CSIConfig manifests in non-paravirtual mode", func() {
		BeforeEach(func() {
			clusterName = "test-cluster-csi"
			clusterResourceFilePath = "testdata/test-csi-non-paravirtual.yaml"
		})

		It("Should reconcile CSIConfig and create data values secret for CSIConfig on management cluster", func() {
			cluster := &clusterapiv1beta1.Cluster{}
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, key, cluster); err != nil {
					return false
				}
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			// the csi config object should be deployed
			config := &csiv1alpha1.CSIConfig{}
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, key, config); err != nil {
					return false
				}
				Expect(config.Spec.VSphereCSI.Mode).Should(Equal("vsphereCSI"))
				Expect(config.Spec.VSphereCSI.NonParavirtualConfig).NotTo(BeZero())
				Expect(config.Spec.VSphereCSI.NonParavirtualConfig.Namespace).Should(Equal("default"))
				Expect(config.Spec.VSphereCSI.NonParavirtualConfig.ClusterName).Should(Equal("test-clustername"))
				Expect(config.Spec.VSphereCSI.NonParavirtualConfig.Server).Should(Equal("svr-0"))
				Expect(config.Spec.VSphereCSI.NonParavirtualConfig.Datacenter).Should(Equal("dc0"))
				Expect(config.Spec.VSphereCSI.NonParavirtualConfig.PublicNetwork).Should(Equal("8.2.0.0/16"))
				Expect(config.Spec.VSphereCSI.NonParavirtualConfig.Username).Should(Equal("administrator@vsphere.local"))
				Expect(config.Spec.VSphereCSI.NonParavirtualConfig.Password).Should(Equal("test-passwd"))
				Expect(config.Spec.VSphereCSI.NonParavirtualConfig.Region).NotTo(BeZero())
				Expect(*config.Spec.VSphereCSI.NonParavirtualConfig.Region).Should(Equal("test-region"))
				Expect(config.Spec.VSphereCSI.NonParavirtualConfig.Zone).NotTo(BeZero())
				Expect(*config.Spec.VSphereCSI.NonParavirtualConfig.Zone).Should(Equal("test-zone"))
				Expect(config.Spec.VSphereCSI.NonParavirtualConfig.UseTopologyCategories).NotTo(BeZero())
				Expect(*config.Spec.VSphereCSI.NonParavirtualConfig.UseTopologyCategories).Should(Equal(true))
				Expect(config.Spec.VSphereCSI.NonParavirtualConfig.ProvisionTimeout).NotTo(BeZero())
				Expect(*config.Spec.VSphereCSI.NonParavirtualConfig.ProvisionTimeout).Should(Equal("33s"))
				Expect(config.Spec.VSphereCSI.NonParavirtualConfig.AttachTimeout).NotTo(BeZero())
				Expect(*config.Spec.VSphereCSI.NonParavirtualConfig.AttachTimeout).Should(Equal("77s"))
				Expect(config.Spec.VSphereCSI.NonParavirtualConfig.ResizerTimeout).NotTo(BeZero())
				Expect(*config.Spec.VSphereCSI.NonParavirtualConfig.ResizerTimeout).Should(Equal("99s"))
				Expect(config.Spec.VSphereCSI.NonParavirtualConfig.VSphereVersion).NotTo(BeZero())
				Expect(*config.Spec.VSphereCSI.NonParavirtualConfig.VSphereVersion).Should(Equal("8.1.2-release"))
				Expect(config.Spec.VSphereCSI.NonParavirtualConfig.HttpProxy).NotTo(BeZero())
				Expect(*config.Spec.VSphereCSI.NonParavirtualConfig.HttpProxy).Should(Equal("1.1.1.1"))
				Expect(config.Spec.VSphereCSI.NonParavirtualConfig.HttpsProxy).NotTo(BeZero())
				Expect(*config.Spec.VSphereCSI.NonParavirtualConfig.HttpsProxy).Should(Equal("2.2.2.2"))
				Expect(config.Spec.VSphereCSI.NonParavirtualConfig.NoProxy).NotTo(BeZero())
				Expect(*config.Spec.VSphereCSI.NonParavirtualConfig.NoProxy).Should(Equal("3.3.3.3"))
				Expect(config.Spec.VSphereCSI.NonParavirtualConfig.DeploymentReplicas).NotTo(BeZero())
				Expect(*config.Spec.VSphereCSI.NonParavirtualConfig.DeploymentReplicas).Should(Equal(int32(3)))
				Expect(config.Spec.VSphereCSI.NonParavirtualConfig.WindowsSupport).NotTo(BeZero())
				Expect(*config.Spec.VSphereCSI.NonParavirtualConfig.WindowsSupport).Should(Equal(true))
				Expect(config.Spec.VSphereCSI.ParavirtualConfig).To(BeZero())

				if len(config.OwnerReferences) == 0 {
					return false
				}
				Expect(len(config.OwnerReferences)).Should(Equal(1))
				Expect(config.OwnerReferences[0].Name).Should(Equal(clusterName))

				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			// the data values secret should be generated
			secret := &v1.Secret{}
			Eventually(func() bool {
				secretKey := client.ObjectKey{
					Namespace: clusterNamespace,
					Name:      fmt.Sprintf("%s-%s-data-values", clusterName, constants.CSIAddonName),
				}
				if err := k8sClient.Get(ctx, secretKey, secret); err != nil {
					return false
				}
				secretData := string(secret.Data["values.yaml"])
				Expect(len(secretData)).Should(Not(BeZero()))
				Expect(strings.Contains(secretData, "vsphereCSI:")).Should(BeTrue())
				Expect(strings.Contains(secretData, "namespace: default")).Should(BeTrue())
				Expect(strings.Contains(secretData, "server: svr-0")).Should(BeTrue())
				Expect(strings.Contains(secretData, "datacenter: dc0")).Should(BeTrue())
				Expect(strings.Contains(secretData, "publicNetwork: 8.2.0.0/16")).Should(BeTrue())
				Expect(strings.Contains(secretData, "username: administrator@vsphere.local")).Should(BeTrue())
				Expect(strings.Contains(secretData, "password: test-passwd")).Should(BeTrue())
				Expect(strings.Contains(secretData, "region: test-region")).Should(BeTrue())
				Expect(strings.Contains(secretData, "zone: test-zone")).Should(BeTrue())
				Expect(strings.Contains(secretData, "useTopologyCategories: true")).Should(BeTrue())
				Expect(strings.Contains(secretData, "provisionTimeout: 33s")).Should(BeTrue())
				Expect(strings.Contains(secretData, "attachTimeout: 77s")).Should(BeTrue())
				Expect(strings.Contains(secretData, "resizerTimeout: 99s")).Should(BeTrue())
				Expect(strings.Contains(secretData, "vSphereVersion: 8.1.2-release")).Should(BeTrue())
				Expect(strings.Contains(secretData, "httpProxy: 1.1.1.1")).Should(BeTrue())
				Expect(strings.Contains(secretData, "httpsProxy: 2.2.2.2")).Should(BeTrue())
				Expect(strings.Contains(secretData, "noProxy: 3.3.3.3")).Should(BeTrue())
				Expect(strings.Contains(secretData, "deploymentReplicas: 3")).Should(BeTrue())
				Expect(strings.Contains(secretData, "windowsSupport: true")).Should(BeTrue())

				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			// eventually the secret ref to the data values should be updated
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, key, config); err != nil {
					return false
				}
				Expect(config.Status.SecretRef).To(Equal(fmt.Sprintf("%s-%s-data-values", clusterName, constants.CSIAddonName)))
				return true
			})
		})
	})

	Context("reconcile CSIConfig manifests in paravirtual mode", func() {
		BeforeEach(func() {
			clusterName = "test-cluster-pv-csi"
			clusterResourceFilePath = "testdata/test-csi-paravirtual.yaml"
		})
		It("Should reconcile CSIConfig and create data values secret for CSIConfig on management cluster", func() {
			// the data values secret should be generated
			secret := &v1.Secret{}
			Eventually(func() bool {
				secretKey := client.ObjectKey{
					Namespace: clusterNamespace,
					Name:      fmt.Sprintf("%s-%s-data-values", clusterName, constants.PVCSIAddonName),
				}
				if err := k8sClient.Get(ctx, secretKey, secret); err != nil {
					return false
				}
				secretData := string(secret.Data["values.yaml"])
				Expect(len(secretData)).Should(Not(BeZero()))
				Expect(strings.Contains(secretData, "vspherePVCSI:")).Should(BeTrue())
				Expect(strings.Contains(secretData, "clusterName: test-cluster")).Should(BeTrue())
				Expect(strings.Contains(secretData, "clusterUID: test-uid")).Should(BeTrue())
				Expect(strings.Contains(secretData, "namespace: default")).Should(BeTrue())
				Expect(strings.Contains(secretData, "supervisorMasterEndpointHostname: sababu.svc.local")).Should(BeTrue())
				Expect(strings.Contains(secretData, "supervisorMasterPort: 6443")).Should(BeTrue())

				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			// eventually the secret ref to the data values should be updated
			config := &csiv1alpha1.CSIConfig{}
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, key, config); err != nil {
					return false
				}
				Expect(config.Status.SecretRef).To(Equal(fmt.Sprintf("%s-%s-data-values", clusterName, constants.PVCSIAddonName)))
				return true
			})
		})
	})
})
