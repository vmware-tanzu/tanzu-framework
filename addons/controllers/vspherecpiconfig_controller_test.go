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
	capvv1beta1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/v1beta1"
	capvvmwarev1beta1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/vmware/v1beta1"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/testutil"
	cpiv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cpi/v1alpha1"
)

var _ = Describe("VSphereCPIConfig Reconciler", func() {
	const (
		clusterNamespace = "default"
	)

	var (
		key                     client.ObjectKey
		clusterName             string
		clusterResourceFilePath string
	)

	JustBeforeEach(func() {
		By("Creating cluster and VSphereCPIConfig resources")
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
		By("Deleting cluster and VSphereCPIConfig resources")
		for _, filePath := range []string{clusterResourceFilePath} {
			f, err := os.Open(filePath)
			Expect(err).ToNot(HaveOccurred())
			err = testutil.DeleteResources(f, cfg, dynamicClient, true)
			Expect(err).ToNot(HaveOccurred())
			Expect(f.Close()).ToNot(HaveOccurred())
		}
	})

	Context("reconcile VSphereCPIConfig manifests in non-paravirtual mode", func() {
		BeforeEach(func() {
			clusterName = "test-cluster-cpi"
			clusterResourceFilePath = "testdata/test-vsphere-cpi-non-paravirtual.yaml"
		})

		It("Should reconcile VSphereCPIConfig and create data values secret for VSphereCPIConfig on management cluster", func() {
			cluster := &clusterapiv1beta1.Cluster{}
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, key, cluster); err != nil {
					return false
				}
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			// the vsphere cluster and vsphere machine template should be provided
			vsphereCluster := &capvv1beta1.VSphereCluster{}
			cpMachineTemplate := &capvv1beta1.VSphereMachineTemplate{}
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, key, vsphereCluster); err != nil {
					return false
				}
				if err := k8sClient.Get(ctx, client.ObjectKey{
					Namespace: clusterNamespace,
					Name:      clusterName + "-control-plane",
				}, cpMachineTemplate); err != nil {
					return false
				}
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			// the cpi config object should be deployed
			config := &cpiv1alpha1.VSphereCPIConfig{}
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, key, config); err != nil {
					return false
				}
				Expect(*config.Spec.VSphereCPI.Mode).Should(Equal("vsphereCPI"))
				Expect(*config.Spec.VSphereCPI.Region).Should(Equal("test-region"))
				Expect(*config.Spec.VSphereCPI.Zone).Should(Equal("test-zone"))

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
					Name:      fmt.Sprintf("%s-%s-data-values", clusterName, constants.CPIAddonName),
				}
				if err := k8sClient.Get(ctx, secretKey, secret); err != nil {
					return false
				}
				secretData := string(secret.Data["values.yaml"])
				Expect(len(secretData)).Should(Not(BeZero()))
				Expect(strings.Contains(secretData, "vsphereCPI:")).Should(BeTrue())
				Expect(strings.Contains(secretData, "mode: vsphereCPI")).Should(BeTrue())
				Expect(strings.Contains(secretData, "datacenter: dc0")).Should(BeTrue())
				Expect(strings.Contains(secretData, "region: test-region")).Should(BeTrue())
				Expect(strings.Contains(secretData, "zone: test-zone")).Should(BeTrue())
				Expect(strings.Contains(secretData, "insecureFlag: true")).Should(BeTrue())
				Expect(strings.Contains(secretData, "ipFamily: ipv6")).Should(BeTrue())
				Expect(strings.Contains(secretData, "vmInternalNetwork: internal-net")).Should(BeTrue())
				Expect(strings.Contains(secretData, "vmExternalNetwork: external-net")).Should(BeTrue())
				Expect(strings.Contains(secretData, "vmExcludeInternalNetworkSubnetCidr: 192.168.3.0/24")).Should(BeTrue())
				Expect(strings.Contains(secretData, "vmExcludeExternalNetworkSubnetCidr: 22.22.3.0/24")).Should(BeTrue())
				Expect(strings.Contains(secretData, "tlsThumbprint: test-thumbprint")).Should(BeTrue())
				Expect(strings.Contains(secretData, "server: vsphere-server.local")).Should(BeTrue())
				Expect(strings.Contains(secretData, "username: foo")).Should(BeTrue())
				Expect(strings.Contains(secretData, "password: bar")).Should(BeTrue())

				Expect(strings.Contains(secretData, "nsxt:")).Should(BeTrue())
				Expect(strings.Contains(secretData, "podRoutingEnabled: true")).Should(BeTrue())
				Expect(strings.Contains(secretData, "routes")).Should(BeTrue())
				Expect(strings.Contains(secretData, "routerPath: test-route-path")).Should(BeTrue())
				Expect(strings.Contains(secretData, "clusterCidr: 192.168.0.1/24")).Should(BeTrue())
				Expect(strings.Contains(secretData, "username: test-nsxt-username")).Should(BeTrue())
				Expect(strings.Contains(secretData, "password: test-nsxt-password")).Should(BeTrue())
				Expect(strings.Contains(secretData, "host: test-nsxt-manager-host")).Should(BeTrue())
				Expect(strings.Contains(secretData, "insecureFlag: true")).Should(BeTrue())
				Expect(strings.Contains(secretData, "remoteAuth: true")).Should(BeTrue())
				Expect(strings.Contains(secretData, "vmcAccessToken: test-vmc-access-token")).Should(BeTrue())
				Expect(strings.Contains(secretData, "vmcAuthHost: test-vmc-auth-host")).Should(BeTrue())
				Expect(strings.Contains(secretData, "clientCertKeyData: test-client-cert-key-data")).Should(BeTrue())
				Expect(strings.Contains(secretData, "clientCertData: test-client-cert-data")).Should(BeTrue())
				Expect(strings.Contains(secretData, "rootCAData: test-root-ca-data-b64")).Should(BeTrue())
				Expect(strings.Contains(secretData, "secretName: test-nsxt-secret-name")).Should(BeTrue())
				Expect(strings.Contains(secretData, "secretNamespace: test-nsxt-secret-namespace")).Should(BeTrue())

				Expect(strings.Contains(secretData, "http_proxy: foo.com")).Should(BeTrue())
				Expect(strings.Contains(secretData, "https_proxy: bar.com")).Should(BeTrue())
				Expect(strings.Contains(secretData, "no_proxy: foobar.com")).Should(BeTrue())
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

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

	Context("reconcile VSphereCPIConfig manifests in paravirtual mode, with floating IP", func() {
		BeforeEach(func() {
			clusterName = "test-cluster-cpi-paravirtual"
			clusterResourceFilePath = "testdata/test-vsphere-cpi-paravirtual.yaml"
		})
		It("Should reconcile VSphereCPIConfig and create data values secret for VSphereCPIConfig on management cluster", func() {
			// the data values secret should be generated
			secret := &v1.Secret{}
			Eventually(func() bool {
				secretKey := client.ObjectKey{
					Namespace: clusterNamespace,
					Name:      fmt.Sprintf("%s-%s-data-values", clusterName, constants.CPIAddonName),
				}
				if err := k8sClient.Get(ctx, secretKey, secret); err != nil {
					return false
				}
				secretData := string(secret.Data["values.yaml"])
				Expect(len(secretData)).Should(Not(BeZero()))
				Expect(strings.Contains(secretData, "vsphereCPI:")).Should(BeTrue())
				Expect(strings.Contains(secretData, "mode: vsphereParavirtualCPI")).Should(BeTrue())
				Expect(strings.Contains(secretData, "clusterAPIVersion: cluster.x-k8s.io/v1beta1")).Should(BeTrue())
				Expect(strings.Contains(secretData, "clusterKind: Cluster")).Should(BeTrue())
				Expect(strings.Contains(secretData, "clusterName: test-cluster-cpi-paravirtual")).Should(BeTrue())

				uuidReg, err := regexp.Compile("clusterUID: [0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(uuidReg.FindString(secretData)).To(Not(BeEmpty()))

				Expect(strings.Contains(secretData, "supervisorMasterEndpointIP: supervisor.default.svc")).Should(BeTrue())
				Expect(strings.Contains(secretData, "supervisorMasterPort: \"6443\"")).Should(BeTrue())

				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			// eventually the secret ref to the data values should be updated
			config := &cpiv1alpha1.VSphereCPIConfig{}
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, key, config); err != nil {
					return false
				}
				Expect(config.Status.SecretRef).To(Equal(fmt.Sprintf("%s-%s-data-values", clusterName, constants.CPIAddonName)))
				return true
			})
		})

		It("Should reconcile ProviderServiceAccount", func() {
			serviceAccount := &capvvmwarev1beta1.ProviderServiceAccount{}
			Eventually(func() bool {
				serviceAccountKey := client.ObjectKey{
					Namespace: clusterNamespace,
					Name:      fmt.Sprintf("%s-ccm", clusterName),
				}
				if err := k8sClient.Get(ctx, serviceAccountKey, serviceAccount); err != nil {
					return false
				}
				Expect(serviceAccount.Spec.Ref.Name).To(Equal(key.Name))
				Expect(serviceAccount.Spec.Ref.Namespace).To(Equal(key.Namespace))
				Expect(serviceAccount.Spec.Rules).To(HaveLen(4))
				Expect(serviceAccount.Spec.TargetNamespace).To(Equal("vmware-system-cloud-provider"))
				Expect(serviceAccount.Spec.TargetSecretName).To(Equal("cloud-provider-creds"))
				return true
			})
		})
	})
})
