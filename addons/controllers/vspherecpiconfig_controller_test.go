// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	capvv1beta1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/v1beta1"
	capvvmwarev1beta1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/vmware/v1beta1"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	clusterapiutil "sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/controller-runtime/pkg/client"

	controllers "github.com/vmware-tanzu/tanzu-framework/addons/controllers/cpi"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	"github.com/vmware-tanzu/tanzu-framework/addons/test/testutil"
	cpiv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/addonconfigs/cpi/v1alpha1"
)

const (
	defaultString               = "default"
	testSupervisorAPIServerVIP  = "10.0.0.100"
	testSupervisorAPIServerPort = 6883
)

func newTestSupervisorLBService() *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      controllers.SupervisorLoadBalancerSvcName,
			Namespace: controllers.SupervisorLoadBalancerSvcNamespace,
		},
		Spec: v1.ServiceSpec{
			// Note: This will be service with no selectors. The endpoints will be manually created.
			Ports: []v1.ServicePort{
				{
					Name:       controllers.SupervisorLoadBalancerSvcAPIServerPortName,
					Port:       testSupervisorAPIServerPort,
					TargetPort: intstr.FromInt(testSupervisorAPIServerPort),
				},
			},
		},
	}
}

func newTestSupervisorLBServiceStatus() v1.ServiceStatus {
	return v1.ServiceStatus{
		LoadBalancer: v1.LoadBalancerStatus{
			Ingress: []v1.LoadBalancerIngress{
				{
					IP: testSupervisorAPIServerVIP,
					Ports: []v1.PortStatus{
						{
							Port:     int32(testSupervisorAPIServerPort),
							Protocol: v1.ProtocolTCP,
						},
					},
				},
			},
		},
	}
}

func newClusterInfoConfigMap() *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      controllers.ConfigMapClusterInfo,
			Namespace: metav1.NamespacePublic,
		},
		Data: map[string]string{
			"kubeconfig": `apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN5RENDQWJDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKY201bGRHVnpNQjRYRFRJeU1ETXlOVEExTkRNd01Wb1hEVE15TURNeU1qQTFORE13TVZvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYSnVaWFJsY3pDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBT25TClZOTlFzb043bExlekdxcmFpb01EOHJDeU9JSDFVVnJqT0x3YWp5UXZQeHRaYjdRdE9DRTJia282RytJcFJlQzYKZ3hIZi9aQi83VXNvTzVqaXpvcWpJUVpsdjZzSm9zR0I0c1Q4NG9hTGtIT21ISEU5a0w5U09Wa1FuT2J5ZWUzNApFaExnUHZWb1pKc05xMTRVWXJsS1R1dnBxNGhpY2pNU1Vua3ZpQmtiY3BCL0oxaUpBbXpIV2tnSkdlSk5HYTNnCk1lSXUxSzJVZ1I3NWp5bkpJSUsvdHFOMyt3VGl1TEcyNDhzZkVCY29pSWFYNTdoQ0E3d0hKR3RaZVJDQmlhNTAKc2JqbFFGd0hEblJldjBPZlNxeE4wZVE3ZVZwQ1NKWWFIQWtmc0t4ZXBSNXNTMnRrRFBVMlg4cHNmcEVQZGdwaApXUS9MN1JiYWdvbE91VkRzdkdVQ0F3RUFBYU1qTUNFd0RnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFCZmp0dFp6NDNwU3NUelM5ZVFzK0lkL3VjOVcKZkxWL2VCMXhOaUJWbXN6MUNsZldaZ1VIeWFsZ1RtRTlHdGtQVnhja2pSczNXQ0hlRUY4N1psOENobEFFMnFrZwpRcE5BVDMyM3RpOVk1TWk3bWZvOFl2OXdOL0ZPNzRwbnB2OElUVXpoRVlJaGxUZFYrd3RmWHUvTmNzN1Z0akNBCkNWNnExeU5lbG05eU9CVW51RElRNVo0L1AvYXFLRDFzSXNCSFJZZVU3SzVEaklSTGNpN3gvb2dlQ2F0ZVowNFoKeWExd2NXVXB3SnNpaGxIOCtHUkJkZ2h1RzZ4aC9ka1JhNmRLbHpYdVRManpsdVYyRUp6N29hZGthSUUybzM3RQpoazZ2aERBc1JHSCtCdnJtcFZJYjNsYTVGbDZHRkd2VElKMGFNT0pMTTVFa3pHNFlUeWpNRi9qK3Joaz0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    server: https://172.17.0.3:6773
  name: ""
contexts: null
current-context: ""
kind: Config
preferences: {}
users: null`,
		},
	}
}

func tryCreateIfNotExists(ctx context.Context, client client.Client, object client.Object) error {
	err := client.Create(ctx, object)
	if apierrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

func getSecretDataValuesForCluster(clusterName string) (string, error) {
	secret := &v1.Secret{}
	secretKey := client.ObjectKey{
		Namespace: defaultString,
		Name:      fmt.Sprintf("%s-%s-data-values", clusterName, constants.CPIAddonName),
	}
	if err := k8sClient.Get(ctx, secretKey, secret); err != nil {
		return "", err
	}
	secretData := string(secret.Data["values.yaml"])
	return secretData, nil
}

var _ = Describe("VSphereCPIConfig Reconciler", func() {
	var (
		key                     client.ObjectKey
		configKey               client.ObjectKey
		clusterName             string
		clusterResourceFilePath string
		vsphereClusterName      string
	)

	JustBeforeEach(func() {
		By("Creating cluster and VSphereCPIConfig resources")
		key = client.ObjectKey{
			Namespace: defaultString,
			Name:      clusterName,
		}
		configKey = client.ObjectKey{
			Namespace: defaultString,
			Name:      util.GeneratePackageSecretName(clusterName, constants.CPIDefaultRefName),
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
			if err = testutil.DeleteResources(f, cfg, dynamicClient, true); !apierrors.IsNotFound(err) {
				// namespace has been explicitly deleted using testutil.DeleteNamespace
				// ignore its NotFound error here
				Expect(err).ToNot(HaveOccurred())
			}
			Expect(f.Close()).ToNot(HaveOccurred())
		}
	})

	Context("reconcile VSphereCPIConfig manifests in non-paravirtual mode, no multitenancy", func() {
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
					Namespace: defaultString,
					Name:      clusterName + "-control-plane-template",
				}, cpMachineTemplate); err != nil {
					return false
				}
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			// the cpi config object should be deployed
			config := &cpiv1alpha1.VSphereCPIConfig{}
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, configKey, config); err != nil {
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
					Namespace: defaultString,
					Name:      fmt.Sprintf("%s-%s-data-values", clusterName, constants.CPIAddonName),
				}
				if err := k8sClient.Get(ctx, secretKey, secret); err != nil {
					return false
				}
				secretData := string(secret.Data["values.yaml"])
				Expect(len(secretData)).ShouldNot(BeZero())
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

				Expect(strings.Contains(secretData, "http_proxy: foo.com")).Should(BeTrue())
				Expect(strings.Contains(secretData, "https_proxy: bar.com")).Should(BeTrue())
				Expect(strings.Contains(secretData, "no_proxy: foobar.com")).Should(BeTrue())

				//assert that there are no paravirt datavalue keys
				Expect(strings.Contains(secretData, "clusterAPIVersion:")).Should(BeFalse())
				Expect(strings.Contains(secretData, "clusterKind:")).Should(BeFalse())
				Expect(strings.Contains(secretData, "clusterName:")).Should(BeFalse())
				Expect(strings.Contains(secretData, "supervisorMasterEndpointIP:")).Should(BeFalse())
				Expect(strings.Contains(secretData, "supervisorMasterPort:")).Should(BeFalse())

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

	Context("reconcile VSphereCPIConfig manifests in non-paravirtual mode, when clusterbootstrapController doesn't add ownerRef to VSphereCPIConfig", func() {

		BeforeEach(func() {
			clusterName = "test-cluster-cpi-enqueue-cluster-event"
			clusterResourceFilePath = "testdata/test-vsphere-cpi-non-paravirtual-enqueue-cluster-event-cluster-spec.yaml"
		})

		It("should not create data values secret until VSphereCPIConfig has an OwnerRef to correct cluster", func() {
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
					Namespace: defaultString,
					Name:      clusterName + "-control-plane-template",
				}, cpMachineTemplate); err != nil {
					return false
				}
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			By("patching cpi with ownerRef")
			config := &cpiv1alpha1.VSphereCPIConfig{}
			cpiConfigKey := client.ObjectKey{
				Namespace: defaultString,
				Name:      "test-cluster-cpi-enqueue-cluster-event-random",
			}
			Consistently(func() bool {
				if err := k8sClient.Get(ctx, cpiConfigKey, config); err != nil {
					return false
				}
				Expect(*config.Spec.VSphereCPI.Mode).Should(Equal("vsphereCPI"))
				Expect(*config.Spec.VSphereCPI.Region).Should(Equal("test-region"))
				Expect(*config.Spec.VSphereCPI.Zone).Should(Equal("test-zone"))

				if len(config.OwnerReferences) > 0 {
					return false
				}
				Expect(len(config.OwnerReferences)).Should(Equal(0))

				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			By("patching cpi with ownerRef as ClusterBootstrapController would do")
			// patch the VSphereCPIConfig with ownerRef
			patchedVSphereCPIConfig := config.DeepCopy()
			ownerRef := metav1.OwnerReference{
				APIVersion: clusterapiv1beta1.GroupVersion.String(),
				Kind:       cluster.Kind,
				Name:       cluster.Name,
				UID:        cluster.UID,
			}

			ownerRef.Kind = "Cluster"
			patchedVSphereCPIConfig.OwnerReferences = clusterapiutil.EnsureOwnerRef(patchedVSphereCPIConfig.OwnerReferences, ownerRef)
			Expect(k8sClient.Patch(ctx, patchedVSphereCPIConfig, client.MergeFrom(config))).ShouldNot(HaveOccurred())

			// the data values secret should be generated
			secret := &v1.Secret{}
			Eventually(func() bool {
				secretKey := client.ObjectKey{
					Namespace: defaultString,
					Name:      util.GenerateDataValueSecretName(clusterName, constants.CPIDefaultRefName),
				}
				if err := k8sClient.Get(ctx, secretKey, secret); err != nil {
					return false
				}
				secretData := string(secret.Data["values.yaml"])
				Expect(len(secretData)).ShouldNot(BeZero())
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

				Expect(strings.Contains(secretData, "http_proxy: foo.com")).Should(BeTrue())
				Expect(strings.Contains(secretData, "https_proxy: bar.com")).Should(BeTrue())
				Expect(strings.Contains(secretData, "no_proxy: foobar.com")).Should(BeTrue())

				//assert that there are no paravirt datavalue keys
				Expect(strings.Contains(secretData, "clusterAPIVersion:")).Should(BeFalse())
				Expect(strings.Contains(secretData, "clusterKind:")).Should(BeFalse())
				Expect(strings.Contains(secretData, "clusterName:")).Should(BeFalse())
				Expect(strings.Contains(secretData, "supervisorMasterEndpointIP:")).Should(BeFalse())
				Expect(strings.Contains(secretData, "supervisorMasterPort:")).Should(BeFalse())

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

	Context("reconcile VSphereCPIConfig manifests in non-paravirtual mode, with multi-tenancy enabled", func() {

		identity := &capvv1beta1.VSphereClusterIdentity{}
		identitySecret := &v1.Secret{}
		identityNamespace := "capv-system"

		BeforeEach(func() {
			clusterName = "test-cluster-cpi-multi-tenancy"
			clusterResourceFilePath = "testdata/test-vsphere-cpi-non-paravirtual-multi-tenancy.yaml"
		})

		JustAfterEach(func() {
			Expect(testutil.DeleteNamespace(ctx, clientSet, identityNamespace)).To(Succeed())
		})

		It("Should reconcile VSphereCPIConfig and create data values secret for VSphereCPIConfig on management cluster", func() {
			identityName := "multi-tenancy"
			Expect(k8sClient.Get(ctx, client.ObjectKey{Name: identityName}, identity)).To(Succeed())
			Expect(k8sClient.Get(ctx, client.ObjectKey{Name: identityName, Namespace: identityNamespace}, identitySecret)).To(Succeed())

			identity.Status.Ready = true
			Expect(k8sClient.Status().Update(ctx, identity)).To(Succeed())

			secret := &v1.Secret{}
			Eventually(func() bool {
				secretKey := client.ObjectKey{
					Namespace: defaultString,
					Name:      fmt.Sprintf("%s-%s-data-values", clusterName, constants.CPIAddonName),
				}
				if err := k8sClient.Get(ctx, secretKey, secret); err != nil {
					return false
				}
				secretData := string(secret.Data["values.yaml"])
				Expect(len(secretData)).ShouldNot(BeZero())
				Expect(strings.Contains(secretData, "username: foo")).Should(BeTrue())
				Expect(strings.Contains(secretData, "password: bar")).Should(BeTrue())

				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())
		})
	})

	Context("reconcile VSphereCPIConfig manifests in paravirtual mode, with floating IP", func() {
		BeforeEach(func() {
			clusterName = "test-cluster-cpi-paravirtual"
			clusterResourceFilePath = "testdata/test-vsphere-cpi-paravirtual.yaml"

			Expect(tryCreateIfNotExists(ctx, k8sClient, newClusterInfoConfigMap())).Should(Succeed())
		})
		It("Should reconcile VSphereCPIConfig", func() {
			By("create data values secret for VSphereCPIConfig on management cluster")
			// the data values secret should be generated
			Eventually(func() bool {
				secretData, err := getSecretDataValuesForCluster(clusterName)
				// allow some time for data value secret to be generated
				if err != nil {
					return false
				}
				Expect(len(secretData)).ShouldNot(BeZero())
				Expect(strings.Contains(secretData, "vsphereCPI:")).Should(BeTrue())
				Expect(strings.Contains(secretData, "mode: vsphereParavirtualCPI")).Should(BeTrue())
				Expect(strings.Contains(secretData, "clusterAPIVersion: cluster.x-k8s.io/v1beta1")).Should(BeTrue())
				Expect(strings.Contains(secretData, "clusterKind: Cluster")).Should(BeTrue())
				Expect(strings.Contains(secretData, "clusterName: test-cluster-cpi-paravirtual")).Should(BeTrue())

				uuidReg, err := regexp.Compile("clusterUID: [0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}")
				Expect(err).ShouldNot(HaveOccurred())
				Expect(uuidReg.FindString(secretData)).To(Not(BeEmpty()))

				Expect(strings.Contains(secretData, "supervisorMasterEndpointIP: 172.17.0.3")).Should(BeTrue())
				Expect(strings.Contains(secretData, "supervisorMasterPort: \"6773\"")).Should(BeTrue())

				// assert that non paravirt data values don't exist, the keys should not exist
				Expect(strings.Contains(secretData, "datacenter:")).Should(BeFalse())
				Expect(strings.Contains(secretData, "server:")).Should(BeFalse())
				Expect(strings.Contains(secretData, "nsxt:")).Should(BeFalse())
				Expect(strings.Contains(secretData, "antreaNSXPodRoutingEnabled: true")).Should(BeTrue())

				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			// eventually the secret ref to the data values should be updated
			config := &cpiv1alpha1.VSphereCPIConfig{}
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, configKey, config); err != nil {
					return false
				}
				Expect(config.Status.SecretRef).To(Equal(fmt.Sprintf("%s-%s-data-values", clusterName, constants.CPIAddonName)))
				return true
			})

			By("create ProviderServiceAccount")
			serviceAccount := &capvvmwarev1beta1.ProviderServiceAccount{}
			Eventually(func() bool {
				serviceAccountKey := client.ObjectKey{
					Namespace: defaultString,
					Name:      fmt.Sprintf("%s-ccm", vsphereClusterName),
				}
				if err := k8sClient.Get(ctx, serviceAccountKey, serviceAccount); err != nil {
					return false
				}
				Expect(serviceAccount.Spec.Ref.Name).To(Equal(vsphereClusterName))
				Expect(serviceAccount.Spec.Ref.Namespace).To(Equal(key.Namespace))
				Expect(serviceAccount.Spec.Rules).To(HaveLen(4))
				Expect(serviceAccount.Spec.TargetNamespace).To(Equal("vmware-system-cloud-provider"))
				Expect(serviceAccount.Spec.TargetSecretName).To(Equal("cloud-provider-creds"))
				return true
			})

			By("create aggregated cluster role")
			clusterRole := &rbacv1.ClusterRole{}
			Eventually(func() bool {
				key = client.ObjectKey{
					Name: constants.VsphereCPIProviderServiceAccountAggregatedClusterRole,
				}
				if err := k8sClient.Get(ctx, key, clusterRole); err != nil {
					return false
				}
				Expect(clusterRole.Labels).To(Equal(map[string]string{
					constants.CAPVClusterRoleAggregationRuleLabelSelectorKey: constants.CAPVClusterRoleAggregationRuleLabelSelectorValue,
				}))
				Expect(clusterRole.Rules).To(HaveLen(4))
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())
		})
	})

	Context("Reconcile VSphereCPIConfig used as template", func() {

		BeforeEach(func() {
			clusterName = "test-cluster-cpi-template"
			clusterResourceFilePath = "testdata/test-vsphere-cpi-template-config.yaml"
		})

		It("Should skip the reconciliation", func() {

			configKey.Name = "test-cluster-cpi-config-template"
			configKey.Namespace = addonNamespace
			config := &cpiv1alpha1.VSphereCPIConfig{}
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, configKey, config); err != nil {
					return false
				}
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			By("OwnerReferences is not set")
			Expect(len(config.OwnerReferences)).Should(Equal(0))
		})
	})
})

var _ = Describe("VSphereCPIConfig Reconciler with existing endpoint from LB service", func() {
	var (
		clusterName             string
		clusterResourceFilePath string
	)

	JustBeforeEach(func() {
		By("Creating cluster and VSphereCPIConfig resources")
		f, err := os.Open(clusterResourceFilePath)
		Expect(err).ToNot(HaveOccurred())
		defer f.Close()
		err = testutil.CreateResources(f, cfg, dynamicClient)
		Expect(err).ToNot(HaveOccurred())
	})

	Context("reconcile VSphereCPIConfig manifests in paravirtual mode, with existing endpoint", func() {
		var svc *v1.Service

		BeforeEach(func() {
			clusterName = "test-cluster-cpi-paravirtual-with-endpoint"
			clusterResourceFilePath = "testdata/test-vsphere-cpi-paravirtual-with-endpoint.yaml"

			svc = newTestSupervisorLBService()
			Expect(k8sClient.Create(ctx, svc)).To(Succeed())
			svc.Status = newTestSupervisorLBServiceStatus()
			Expect(k8sClient.Status().Update(ctx, svc)).To(Succeed())
		})

		AfterEach(func() {
			Expect(k8sClient.Delete(ctx, svc)).To(Succeed())
		})

		It("uses headless service's endpoints", func() {
			Eventually(func() bool {
				secretData, err := getSecretDataValuesForCluster(clusterName)
				// allow some time for data value secret to be generated
				if err != nil {
					return false
				}
				Expect(strings.Contains(secretData, "supervisorMasterEndpointIP: "+testSupervisorAPIServerVIP)).Should(BeTrue())
				Expect(strings.Contains(secretData, "supervisorMasterPort: \""+strconv.Itoa(testSupervisorAPIServerPort)+"\"")).Should(BeTrue())
				return true
			}).Should(BeTrue())
		})
	})

})

var _ = Describe("VSphereCPIConfig Reconciler multi clusters", func() {
	var (
		clusterName             string
		clusterResourceFilePath string
	)

	JustBeforeEach(func() {
		By("Creating cluster and VSphereCPIConfig resources")
		f, err := os.Open(clusterResourceFilePath)
		Expect(err).ToNot(HaveOccurred())
		defer f.Close()
		err = testutil.CreateResources(f, cfg, dynamicClient)
		Expect(err).ToNot(HaveOccurred())
	})

	Context("in paravirtual mode, but two clusters have same name under different namespace", func() {
		anotherClusterNamespace := "another-ns"

		// covers the issue https://github.com/vmware-tanzu/tanzu-framework/issues/2714
		// it was triggered when two clusters with the same name, but under different namespaces
		BeforeEach(func() {
			clusterName = "test-cluster-cpi-paravirtual-same-name"
			clusterResourceFilePath = "testdata/test-vsphere-cpi-paravirtual-clusters-same-name.yaml"

			Expect(tryCreateIfNotExists(ctx, k8sClient, newClusterInfoConfigMap())).Should(Succeed())
		})

		It("should select vSphereCluster with right namespace", func() {
			Eventually(func() error {
				selectedClusterKey := client.ObjectKey{
					Name:      clusterName,
					Namespace: defaultString,
				}
				notSelectedClusterKey := client.ObjectKey{
					Name:      clusterName,
					Namespace: anotherClusterNamespace,
				}

				selectedCluster := &capvvmwarev1beta1.VSphereCluster{}
				notSelectedCluster := &capvvmwarev1beta1.VSphereCluster{}
				var err error

				if err = k8sClient.Get(ctx, selectedClusterKey, selectedCluster); err != nil {
					return err
				}

				if err = k8sClient.Get(ctx, notSelectedClusterKey, notSelectedCluster); err != nil {
					return err
				}
				serviceAccount := &capvvmwarev1beta1.ProviderServiceAccount{}
				serviceAccountKey := client.ObjectKey{
					Namespace: defaultString,
					Name:      fmt.Sprintf("%s-ccm", clusterName),
				}
				if err := k8sClient.Get(ctx, serviceAccountKey, serviceAccount); err != nil {
					return err
				}
				Expect(serviceAccount.ObjectMeta.OwnerReferences).ToNot(BeEmpty())
				Expect(serviceAccount.ObjectMeta.OwnerReferences[0].UID).To(Equal(selectedCluster.ObjectMeta.UID))
				return nil
			}).Should(Succeed())
		})
	})
})
