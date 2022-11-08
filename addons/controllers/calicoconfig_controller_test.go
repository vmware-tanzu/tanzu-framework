// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"fmt"
	"os"
	"strings"

	cutil "github.com/vmware-tanzu/tanzu-framework/addons/controllers/utils"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	runtanzuv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	"github.com/vmware-tanzu/tanzu-framework/addons/test/testutil"
	cniv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/addonconfigs/cni/v1alpha1"
)

const (
	testClusterCalico1            = "test-cluster-calico-1"
	testClusterCalico2            = "test-cluster-calico-2"
	testDataCalico1               = "testdata/test-calico-1.yaml"
	testDataCalico2               = "testdata/test-calico-2.yaml"
	testDataCalicoTemplateConfig1 = "testdata/test-calico-template-config-1.yaml"
	calicoCarvelPkgRefName        = "calico.tanzu.vmware.com"
	calicoCustomTKGSystemFile     = "testdata/calico-custom-tkg-system.yaml"
)

var _ = Describe("CalicoConfig Reconciler and Webhooks", func() {
	var (
		clusterName             string
		clusterNamespace        string
		configName              string
		clusterResourceFilePath string
	)

	JustBeforeEach(func() {
		// Create the admission webhooks
		f, err := os.Open(cniWebhookManifestFile)
		Expect(err).ToNot(HaveOccurred())
		err = testutil.CreateResources(f, cfg, dynamicClient)
		Expect(err).ToNot(HaveOccurred())
		f.Close()

		// set up the certificates and webhook before creating any objects
		By("Creating and installing new certificates for Calico Admission Webhooks")
		err = testutil.SetupWebhookCertificates(ctx, k8sClient, k8sConfig, &webhookCertDetails)
		Expect(err).ToNot(HaveOccurred())

		By("Creating cluster and CalicoConfig resources")
		f, err = os.Open(clusterResourceFilePath)
		Expect(err).ToNot(HaveOccurred())
		err = testutil.CreateResources(f, cfg, dynamicClient)
		Expect(err).ToNot(HaveOccurred())
		f.Close()
	})

	AfterEach(func() {
		By("Deleting cluster and CalicoConfig resources")
		f, err := os.Open(clusterResourceFilePath)
		Expect(err).ToNot(HaveOccurred())
		err = testutil.DeleteResources(f, cfg, dynamicClient, true)
		Expect(err).ToNot(HaveOccurred())
		f.Close()

		By("Deleting the Admission Webhook configuration for Calico")
		f, err = os.Open(cniWebhookManifestFile)
		Expect(err).ToNot(HaveOccurred())
		err = testutil.DeleteResources(f, cfg, dynamicClient, true)
		Expect(err).ToNot(HaveOccurred())
		f.Close()
	})

	Context("reconcile default CalicoConfig for management cluster on dual-stack CIDR", func() {
		BeforeEach(func() {
			clusterNamespace = addonNamespace
			clusterName = testClusterCalico1
			configName = util.GeneratePackageSecretName(clusterName, calicoCarvelPkgRefName)
			clusterResourceFilePath = testDataCalico1
		})

		It("Should reconcile CalicoConfig and create data values secret for CalicoConfig on management cluster", func() {
			key := client.ObjectKey{
				Namespace: clusterNamespace,
				Name:      clusterName,
			}

			configKey := client.ObjectKey{
				Namespace: clusterNamespace,
				Name:      configName,
			}

			cluster := &clusterapiv1beta1.Cluster{}
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, key, cluster); err != nil {
					return false
				}
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			config := &cniv1alpha1.CalicoConfig{}
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, configKey, config); err != nil {
					return false
				}
				// check owner reference
				return cutil.VerifyOwnerRef(config, clusterName, constants.ClusterKind)
			}, waitTimeout, pollingInterval).Should(BeTrue())
			// check spec values
			Expect(config.Spec.Calico.Config.VethMTU).Should(Equal(int64(7)))
			Expect(config.Spec.Calico.Config.SkipCNIBinaries).Should(BeTrue())

			secret := &v1.Secret{}
			Eventually(func() bool {
				secretKey := client.ObjectKey{
					Namespace: clusterNamespace,
					Name:      fmt.Sprintf("%s-%s-data-values", clusterName, constants.CalicoAddonName),
				}
				if err := k8sClient.Get(ctx, secretKey, secret); err != nil {
					return false
				}
				return cutil.VerifyOwnerRef(secret, clusterName, constants.ClusterKind)
			}, waitTimeout, pollingInterval).Should(BeTrue())
			// check data values secret contents
			Expect(secret.Type).Should(Equal(v1.SecretTypeOpaque))
			secretData := string(secret.Data["values.yaml"])
			Expect(strings.Contains(secretData, "infraProvider: vsphere")).Should(BeTrue())
			Expect(strings.Contains(secretData, "ipFamily: ipv4,ipv6")).Should(BeTrue())
			Expect(strings.Contains(secretData, "clusterCIDR: 192.168.0.0/16,fd00:100:96::/48")).Should(BeTrue())
			Expect(strings.Contains(secretData, "vethMTU: \"7\"")).Should(BeTrue())
			Expect(strings.Contains(secretData, "skipCNIBinaries: true")).Should(BeTrue())
			Eventually(func() bool {
				config := &cniv1alpha1.CalicoConfig{}
				err := k8sClient.Get(ctx, configKey, config)
				if err != nil {
					return false
				}
				// Check status.secretName after reconciliation
				return config.Status.SecretRef == util.GenerateDataValueSecretName(clusterName, constants.CalicoAddonName)

			}, waitTimeout, pollingInterval).Should(BeTrue())
		})
	})

	Context("reconcile mtu customized and cni binaries installation skipped CalicoConfig for management cluster on ipv4 CIDR", func() {
		BeforeEach(func() {
			clusterName = testClusterCalico2
			clusterNamespace = addonNamespace
			configName = util.GeneratePackageSecretName(clusterName, constants.CalicoDefaultRefName)
			clusterResourceFilePath = testDataCalico2
		})

		It("Should reconcile CalicoConfig and create data values secret for CalicoConfig on management cluster", func() {
			key := client.ObjectKey{
				Namespace: clusterNamespace,
				Name:      clusterName,
			}

			configKey := client.ObjectKey{
				Namespace: clusterNamespace,
				Name:      configName,
			}

			cluster := &clusterapiv1beta1.Cluster{}
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, key, cluster); err != nil {
					return false
				}
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			config := &cniv1alpha1.CalicoConfig{}
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, configKey, config); err != nil {
					return false
				}
				return cutil.VerifyOwnerRef(config, clusterName, constants.ClusterKind)
			}, waitTimeout, pollingInterval).Should(BeTrue())

			// check spec values
			Expect(config.Spec.Calico.Config.VethMTU).Should(Equal(int64(1420)))
			Expect(config.Spec.Calico.Config.SkipCNIBinaries).Should(BeFalse())

			secret := &v1.Secret{}
			Eventually(func() bool {
				secretKey := client.ObjectKey{
					Namespace: clusterNamespace,
					Name:      util.GenerateDataValueSecretName(clusterName, constants.CalicoAddonName),
				}
				if err := k8sClient.Get(ctx, secretKey, secret); err != nil {
					return false
				}
				return cutil.VerifyOwnerRef(secret, clusterName, constants.ClusterKind)
			}, waitTimeout, pollingInterval).Should(BeTrue())
			// check data values secret contents
			Expect(secret.Type).Should(Equal(v1.SecretTypeOpaque))
			secretData := string(secret.Data["values.yaml"])
			Expect(strings.Contains(secretData, "infraProvider: docker")).Should(BeTrue())
			Expect(strings.Contains(secretData, "ipFamily: ipv4")).Should(BeTrue())
			Expect(strings.Contains(secretData, "clusterCIDR: 192.168.0.0/16")).Should(BeTrue())
			Expect(strings.Contains(secretData, "vethMTU: \"1420\"")).Should(BeTrue())
			Expect(strings.Contains(secretData, "skipCNIBinaries: false")).Should(BeTrue())

			Eventually(func() bool {
				config := &cniv1alpha1.CalicoConfig{}
				err := k8sClient.Get(ctx, configKey, config)
				if err != nil {
					return false
				}
				// Check status.secretName after reconciliation
				expectedName := util.GenerateDataValueSecretName(clusterName, constants.CalicoAddonName)
				return config.Status.SecretRef == expectedName
			}, waitTimeout, pollingInterval).Should(BeTrue())
		})
	})

	Context("Reconcile CalicoConfig used as template", func() {

		BeforeEach(func() {
			clusterName = testClusterCalico1
			clusterResourceFilePath = testDataCalicoTemplateConfig1
			configName = util.GeneratePackageSecretName(clusterName, constants.CalicoDefaultRefName)
		})

		It("Should skip the reconciliation", func() {

			configKey := client.ObjectKey{
				Namespace: addonNamespace,
				Name:      configName,
			}
			config := &cniv1alpha1.CalicoConfig{}
			Expect(k8sClient.Get(ctx, configKey, config)).To(Succeed())

			By("OwnerReferences is not set")
			Expect(len(config.OwnerReferences)).Should(Equal(0))
		})
	})

	When("calico config is for management cluster", func() {

		BeforeEach(func() {
			clusterName = "mgmt-cluster"
			clusterNamespace = "tkg-system"
			configName = util.GeneratePackageSecretName(clusterName, calicoCarvelPkgRefName)
			clusterResourceFilePath = calicoCustomTKGSystemFile

			ns := &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: clusterNamespace,
				},
			}
			err := k8sClient.Create(ctx, ns)
			if err != nil {
				Expect(apierrors.IsAlreadyExists(err)).To(BeTrue())
			}
		})

		It("should add cluster as owner reference to calico config and name config {CLUSTER_NAME}-{short-package-name}-package", func() {
			By("Creating cluster", func() {
				f, err := os.Open("testdata/calico-management-cluster.yaml")
				Expect(err).ToNot(HaveOccurred())
				err = testutil.CreateResources(f, cfg, dynamicClient)
				Expect(err).ToNot(HaveOccurred())
				f.Close()
			})
			clusterKey := client.ObjectKey{
				Namespace: clusterNamespace,
				Name:      clusterName,
			}

			configKey := client.ObjectKey{
				Namespace: clusterNamespace,
				Name:      configName,
			}
			cluster := &clusterapiv1beta1.Cluster{}
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, clusterKey, cluster); err != nil {
					return false
				}
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			By("Verify new config has owner reference", func() {
				config := &cniv1alpha1.CalicoConfig{}
				Eventually(func() bool {
					if err := k8sClient.Get(ctx, configKey, config); err != nil {
						return false
					}
					return cutil.VerifyOwnerRef(config, clusterName, constants.ClusterKind)
				}, waitTimeout, pollingInterval).Should(BeTrue())
			})

			By("Verify new config points to the correct secret ref", func() {
				Eventually(func() bool {
					config := &cniv1alpha1.CalicoConfig{}
					err := k8sClient.Get(ctx, configKey, config)
					if err != nil {
						return false
					}
					// Check status.secretName after reconciliation
					expectedName := util.GenerateDataValueSecretName(clusterName, constants.CalicoAddonName)
					return config.Status.SecretRef == expectedName
				}, waitTimeout, pollingInterval).Should(BeTrue())
			})
		})
	})

	When("calico config is precreated with {CLUSTER_NAME}-{package-short-name}-package in cluster namespace", func() {

		BeforeEach(func() {
			clusterName = "calico-custom-config-cluster"
			clusterNamespace = "calico-custom-config-ns"
			configName = util.GeneratePackageSecretName(clusterName, calicoCarvelPkgRefName)
			clusterResourceFilePath = calicoCustomTKGSystemFile

			ns := &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: clusterNamespace,
				},
			}
			err := k8sClient.Create(ctx, ns)
			if err != nil {
				Expect(apierrors.IsAlreadyExists(err)).To(BeTrue())
			}
		})

		It("should add cluster as owner reference to calico config", func() {
			By("Creating cluster and CalicoConfig resources", func() {
				f, err := os.Open("testdata/calico-custom-config-ns.yaml")
				Expect(err).ToNot(HaveOccurred())
				err = testutil.CreateResources(f, cfg, dynamicClient)
				Expect(err).ToNot(HaveOccurred())
				f.Close()
			})

			By("Verify clusterBootstrap points to pre-created calico config", func() {
				clusterBootstrap := &runtanzuv1alpha3.ClusterBootstrap{}
				Eventually(func() bool {
					err := k8sClient.Get(ctx, client.ObjectKey{Namespace: clusterNamespace, Name: clusterName}, clusterBootstrap)
					if err != nil {
						return false
					}
					if clusterBootstrap.Spec.CNI.ValuesFrom.ProviderRef.Name == configName {
						return true
					}
					return false
				}, waitTimeout, pollingInterval).Should(BeTrue())
			})

			key := client.ObjectKey{
				Namespace: clusterNamespace,
				Name:      clusterName,
			}

			configKey := client.ObjectKey{
				Namespace: clusterNamespace,
				Name:      configName,
			}
			cluster := &clusterapiv1beta1.Cluster{}
			Eventually(func() bool {
				if err := k8sClient.Get(ctx, key, cluster); err != nil {
					return false
				}
				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			key = client.ObjectKey{
				Namespace: clusterNamespace,
				Name:      configName,
			}
			config := &cniv1alpha1.CalicoConfig{}

			Eventually(func() bool {
				if err := k8sClient.Get(ctx, configKey, config); err != nil {
					return false
				}

				// check owner reference
				return cutil.VerifyOwnerRef(config, clusterName, constants.ClusterKind)

			}, waitTimeout, pollingInterval).Should(BeTrue())

			// check spec values
			Expect(config.Spec.Calico.Config.VethMTU).Should(Equal(int64(1420)))
			Expect(config.Spec.Calico.Config.SkipCNIBinaries).Should(BeFalse())
			Eventually(func() bool {
				secretKey := client.ObjectKey{
					Namespace: clusterNamespace,
					Name:      util.GenerateDataValueSecretName(clusterName, constants.CalicoAddonName),
				}
				secret := &v1.Secret{}
				if err := k8sClient.Get(ctx, secretKey, secret); err != nil {
					return false
				}

				// check data values secret contents
				Expect(secret.Type).Should(Equal(v1.SecretTypeOpaque))
				secretData := string(secret.Data["values.yaml"])
				Expect(strings.Contains(secretData, "infraProvider: docker")).Should(BeTrue())
				Expect(strings.Contains(secretData, "ipFamily: ipv4")).Should(BeTrue())
				Expect(strings.Contains(secretData, "clusterCIDR: 192.168.0.0/16")).Should(BeTrue())
				Expect(strings.Contains(secretData, "vethMTU: \"1420\"")).Should(BeTrue())
				Expect(strings.Contains(secretData, "skipCNIBinaries: false")).Should(BeTrue())

				return true
			}, waitTimeout, pollingInterval).Should(BeTrue())

			Eventually(func() bool {
				config := &cniv1alpha1.CalicoConfig{}
				err := k8sClient.Get(ctx, configKey, config)
				if err != nil {
					return false
				}
				// Check status.secretName after reconciliation
				expectedName := util.GenerateDataValueSecretName(clusterName, constants.CalicoAddonName)
				return config.Status.SecretRef == expectedName
			}, waitTimeout, pollingInterval).Should(BeTrue())
		})
	})

	When("custom clusterbootstrap points to precreated calico config in cluster namespace", func() {
		BeforeEach(func() {
			clusterName = "calico-custom-cb-cluster"
			clusterNamespace = "calico-custom-cb-ns"
			configName = "calico-custom-config"
			clusterResourceFilePath = calicoCustomTKGSystemFile

			ns := &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: clusterNamespace,
				},
			}
			err := k8sClient.Create(ctx, ns)
			if err != nil {
				Expect(apierrors.IsAlreadyExists(err)).To(BeTrue())
			}

		})
		It("values from custom config should override those of template", func() {

			// define custom resources in cluster namespace
			f, err := os.Open("testdata/calico-custom-cb-ns.yaml")
			Expect(err).ToNot(HaveOccurred())
			err = testutil.CreateResources(f, cfg, dynamicClient)
			Expect(err).ToNot(HaveOccurred())
			f.Close()

			By("Verify contents of resulting calico config", func() {
				// eventually the secret ref to the data values should be updated
				calicoConfig := &cniv1alpha1.CalicoConfig{}
				configKey := client.ObjectKey{Name: configName, Namespace: clusterNamespace}
				Eventually(func() bool {
					if err := k8sClient.Get(ctx, configKey, calicoConfig); err != nil {
						return false
					}
					return cutil.VerifyOwnerRef(calicoConfig, clusterName, constants.ClusterKind)
				}, waitTimeout, pollingInterval).Should(BeTrue())

				Eventually(func() bool {
					if err := k8sClient.Get(ctx, configKey, calicoConfig); err != nil {
						return false
					}
					expectedSecretRef := util.GenerateDataValueSecretName(clusterName, constants.CalicoDefaultRefName)
					return calicoConfig.Status.SecretRef != "" && calicoConfig.Status.SecretRef == expectedSecretRef
				}, waitTimeout, pollingInterval).Should(BeTrue())

				Expect(k8sClient.Get(ctx, configKey, calicoConfig)).To(Succeed())

				Expect(calicoConfig.Spec.Calico.Config.VethMTU).Should(Equal(int64(1420)))
				//  why  int64??  that would make it architecture specific, as oposed to why not just int?
				// (defined in apis/addonconfigs/cni/v1alpha1/calicoconfig_types.go)
				// seems to also have forced part of this fix: https://github.com/vmware-tanzu/tanzu-framework/pull/2164
			})

		})
	})

	When("custom clusterbootstrap uses wild card to point to precreated calico config", func() {
		BeforeEach(func() {
			clusterName = "calico-wildcard-cb-cluster"
			clusterNamespace = "calico-wildcard-cb-ns"
			configName = "calico-wildcard-config"
			clusterResourceFilePath = calicoCustomTKGSystemFile

			ns := &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: clusterNamespace,
				},
			}
			err := k8sClient.Create(ctx, ns)
			if err != nil {
				Expect(apierrors.IsAlreadyExists(err)).To(BeTrue())
			}

		})
		It("values from custom config should override those of template", func() {

			By("define cluster namespace resources", func() {
				f, err := os.Open("testdata/calico-wildcard-cb-ns.yaml")
				Expect(err).ToNot(HaveOccurred())
				err = testutil.CreateResources(f, cfg, dynamicClient)
				Expect(err).ToNot(HaveOccurred())
				f.Close()
			})

			By("define clusterboostrap with widlcard, config and clsuter", func() {
				// define custom resources in cluster namespace
				f, err := os.Open("testdata/calico-wildcard-user-input.yaml")
				Expect(err).ToNot(HaveOccurred())
				err = testutil.CreateResources(f, cfg, dynamicClient)
				Expect(err).ToNot(HaveOccurred())
				f.Close()
			})
			By("Verify package refName has been updated correctly in the clusterboostrap", func() {
				clusterBootstrap := &runtanzuv1alpha3.ClusterBootstrap{}
				Eventually(func() bool {
					cbKey := client.ObjectKey{Name: clusterName, Namespace: clusterNamespace}
					if err := k8sClient.Get(ctx, cbKey, clusterBootstrap); err != nil {
						return false
					}
					return true
				}, waitTimeout, pollingInterval).Should(BeTrue())
				Expect(clusterBootstrap.Spec.CNI.RefName).Should(Equal("calico.tanzu.vmware.com.1.2.5--vmware.12-tkg.1"))

			})
			By("Verify owner reference of resulting config", func() {
				Eventually(func() bool {
					calicoConfig := &cniv1alpha1.CalicoConfig{}
					configKey := client.ObjectKey{Name: configName, Namespace: clusterNamespace}
					if err := k8sClient.Get(ctx, configKey, calicoConfig); err != nil {
						return false
					}
					// check owner reference
					return cutil.VerifyOwnerRef(calicoConfig, clusterName, constants.ClusterKind)
				}, waitTimeout, pollingInterval).Should(BeTrue())
			})
			By("Verify contents of resulting calico config", func() {
				// eventually the secret ref to the data values should be updated
				calicoConfig := &cniv1alpha1.CalicoConfig{}
				configKey := client.ObjectKey{Name: configName, Namespace: clusterNamespace}
				Eventually(func() bool {
					if err := k8sClient.Get(ctx, configKey, calicoConfig); err != nil {
						return false
					}
					return calicoConfig.Spec.Calico.Config.VethMTU == 1420 && !calicoConfig.Spec.Calico.Config.SkipCNIBinaries
				}, waitTimeout, pollingInterval).Should(BeTrue())

			})
		})
	})
})
