// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package webhooks_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	apiregistrationv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	kappctrlv1alph1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	packagev1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/webhooks"
	csiv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/addonconfigs/csi/v1alpha1"
	runv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

var (
	tkrName                                      = "v1.23.5---vmware.1-tkg.1-zshippable"
	fakeAntreaCarvelPackageRefName               = "antrea-carvel-package"
	fakeCalicoCarvelPackageRefName               = "calico-carvel-package"
	fakeCSICarvelPackageRefName                  = "vsphere-pv-csi-carvel-package"
	fakeKappCarvelPackageRefName                 = "kapp-controller-carvel-package"
	fakePinnipedCarvelPackageRefName             = "pinniped-carvel-package"
	fakeMetricsServerCarvelPackageRefName        = "metrics-server-carvel-package"
	fakeKubevipcloudproviderCarvelPackageRefName = "kube-vip-cloud-provider"
	fakeCarvelPackageVersion                     = "1.0.0"
)

var _ = Describe("ClusterbootstrapWebhook", func() {

	Context("Verify the logic of clusterbootstrap webhook", func() {
		var (
			clusterBootstrapTemplate  *runv1alpha3.ClusterBootstrapTemplate
			tanzuKubernetesRelease    *runv1alpha3.TanzuKubernetesRelease
			clusterBootstrapName      = "fake-clusterbootstrap"
			clusterBootstrapNamespace = "default"
		)
		AfterEach(func() {
			_ = k8sClient.Delete(ctx, &runv1alpha3.ClusterBootstrap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      clusterBootstrapName,
					Namespace: clusterBootstrapNamespace,
				},
			})
		})

		It("should success to prepare the environment", func() {
			// Prepare the Carvel packages
			createCarvelPackages(ctx, k8sClient)
			// Prepare the ClusterBootstrapTemplate
			clusterBootstrapTemplate = constructClusterBootstrapTemplate()
			err := k8sClient.Create(ctx, clusterBootstrapTemplate)
			Expect(err).NotTo(HaveOccurred())
			// Prepare the TanzuKubernetesRelease
			tanzuKubernetesRelease = constructFakeTanzuKubernetesRelease()
			err = k8sClient.Create(ctx, tanzuKubernetesRelease)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should add defaults to the missing fields when the ClusterBootstrap CR has the predefined annotation", func() {
			// Create a ClusterBootstrap with empty spec
			clusterBootstrap := &runv1alpha3.ClusterBootstrap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      clusterBootstrapName,
					Namespace: clusterBootstrapNamespace,
					Annotations: map[string]string{
						constants.AddCBMissingFieldsAnnotationKey: tkrName,
					},
				},
				Spec: &runv1alpha3.ClusterBootstrapTemplateSpec{},
			}
			err := k8sClient.Create(ctx, clusterBootstrap)
			Expect(err).NotTo(HaveOccurred())

			err = k8sClient.Get(ctx, client.ObjectKey{Namespace: clusterBootstrapNamespace, Name: clusterBootstrapName}, clusterBootstrap)
			Expect(err).NotTo(HaveOccurred())
			Expect(clusterBootstrap.Spec).NotTo(BeNil())
			Expect(clusterBootstrap.Spec.CNI.RefName).To(Equal(clusterBootstrapTemplate.Spec.CNI.RefName))
			Expect(clusterBootstrap.Spec.Kapp.RefName).To(Equal(clusterBootstrapTemplate.Spec.Kapp.RefName))
			Expect(clusterBootstrap.Spec.CSI.RefName).To(Equal(clusterBootstrapTemplate.Spec.CSI.RefName))
			Expect(clusterBootstrap.Spec.AdditionalPackages).NotTo(BeNil())
			Expect(len(clusterBootstrap.Spec.AdditionalPackages)).To(Equal(len(clusterBootstrapTemplate.Spec.AdditionalPackages)))
			Expect(clusterBootstrap.Annotations[constants.AddCBMissingFieldsAnnotationKey]).To(Equal(tkrName))
		})
		It("should add defaults ONLY to the missing fields when the ClusterBootstrap CR has the predefined annotation", func() {
			// Create a ClusterBootstrap with empty spec
			additionalCBPackageRefName := fmt.Sprintf("%s.%s", fakePinnipedCarvelPackageRefName, fakeCarvelPackageVersion)
			clusterBootstrap := &runv1alpha3.ClusterBootstrap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      clusterBootstrapName,
					Namespace: clusterBootstrapNamespace,
					Annotations: map[string]string{
						constants.AddCBMissingFieldsAnnotationKey: tkrName,
					},
				},
				Spec: &runv1alpha3.ClusterBootstrapTemplateSpec{
					CSI: &runv1alpha3.ClusterBootstrapPackage{
						RefName: fmt.Sprintf("%s.%s", fakeCSICarvelPackageRefName, fakeCarvelPackageVersion),
						ValuesFrom: &runv1alpha3.ValuesFrom{
							Inline: map[string]interface{}{
								"foo": "bar",
							},
						},
					},
					AdditionalPackages: []*runv1alpha3.ClusterBootstrapPackage{
						{RefName: additionalCBPackageRefName},
					},
				},
			}
			err := k8sClient.Create(ctx, clusterBootstrap)
			Expect(err).NotTo(HaveOccurred())

			err = k8sClient.Get(ctx, client.ObjectKey{Namespace: clusterBootstrapNamespace, Name: clusterBootstrapName}, clusterBootstrap)
			Expect(err).NotTo(HaveOccurred())
			Expect(clusterBootstrap.Spec).NotTo(BeNil())
			Expect(clusterBootstrap.Spec.CNI.RefName).To(Equal(clusterBootstrapTemplate.Spec.CNI.RefName))
			Expect(clusterBootstrap.Spec.Kapp.RefName).To(Equal(clusterBootstrapTemplate.Spec.Kapp.RefName))
			// CSI should not be touched
			Expect(clusterBootstrap.Spec.CSI.RefName).To(Equal(fmt.Sprintf("%s.%s", fakeCSICarvelPackageRefName, fakeCarvelPackageVersion)))
			Expect(clusterBootstrap.Spec.CSI.ValuesFrom.Inline["foo"]).To(Equal("bar"))
			// Existing additionalPackages should not be touched, the ones in ClusterBootstrapTemplate will be added
			Expect(clusterBootstrap.Spec.AdditionalPackages).NotTo(BeNil())
			// One additional package from original clusterBootstrap, another one will be added from clusterBootstrapTemplate
			Expect(len(clusterBootstrap.Spec.AdditionalPackages)).To(Equal(2))
			for idx := range clusterBootstrap.Spec.AdditionalPackages {
				Expect(clusterBootstrap.Spec.AdditionalPackages[idx].RefName).To(Equal(clusterBootstrapTemplate.Spec.AdditionalPackages[idx].RefName))
				// IMPORTANT: With the new contract, valuesFrom fields will be removed by webhook. ClusterBootstrap Controller
				// will be adding it back after the corresponding ClusterBootstrap Packages are cloned.
				Expect(clusterBootstrap.Spec.AdditionalPackages[idx].ValuesFrom).To(BeNil())
			}
		})
		It("should NOT add defaults to the missing fields when the ClusterBootstrap CR does not have the predefined annotation", func() {
			// Create a ClusterBootstrap with empty spec
			clusterBootstrap := &runv1alpha3.ClusterBootstrap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      clusterBootstrapName,
					Namespace: clusterBootstrapNamespace,
				},
				Spec: &runv1alpha3.ClusterBootstrapTemplateSpec{},
			}
			err := k8sClient.Create(ctx, clusterBootstrap)
			Expect(err).To(HaveOccurred())
			// Validating webhook should reject the request
			Expect(apierrors.IsInvalid(err)).To(BeTrue())
		})
		It("should NOT add defaults to the missing fields when the predefined annotation has invalid value", func() {
			// Create a ClusterBootstrap with empty spec
			clusterBootstrap := &runv1alpha3.ClusterBootstrap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      clusterBootstrapName,
					Namespace: clusterBootstrapNamespace,
					Annotations: map[string]string{
						constants.AddCBMissingFieldsAnnotationKey: "fake-tkr-does-not-exist",
					},
				},
				Spec: &runv1alpha3.ClusterBootstrapTemplateSpec{},
			}
			err := k8sClient.Create(ctx, clusterBootstrap)
			Expect(err).To(HaveOccurred())
			// TKR and CBTemplate not found
			Expect(apierrors.IsNotFound(err)).To(BeTrue())
		})

		It("should complete to the partial filled fields when the ClusterBootstrap CR has the predefined annotation", func() {
			// Create a ClusterBootstrap with empty spec
			clusterBootstrap := &runv1alpha3.ClusterBootstrap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      clusterBootstrapName,
					Namespace: clusterBootstrapNamespace,
					Annotations: map[string]string{
						constants.AddCBMissingFieldsAnnotationKey: tkrName,
					},
				},
				Spec: &runv1alpha3.ClusterBootstrapTemplateSpec{
					CNI: &runv1alpha3.ClusterBootstrapPackage{
						RefName: "antrea*",
					},
					AdditionalPackages: []*runv1alpha3.ClusterBootstrapPackage{
						{RefName: "pinniped*", ValuesFrom: &runv1alpha3.ValuesFrom{Inline: map[string]interface{}{"identity_management_type": "ldap"}}},
						{RefName: "kube-vip-cloud-provider*",
							ValuesFrom: &runv1alpha3.ValuesFrom{
								Inline: map[string]interface{}{
									"foo": "bar",
								},
							},
						},
					},
				},
			}
			err := k8sClient.Create(ctx, clusterBootstrap)
			Expect(err).NotTo(HaveOccurred())

			err = k8sClient.Get(ctx, client.ObjectKey{Namespace: clusterBootstrapNamespace, Name: clusterBootstrapName}, clusterBootstrap)
			Expect(err).NotTo(HaveOccurred())
			Expect(clusterBootstrap.Spec).NotTo(BeNil())
			// clusterBootstrap.Spec.CNI.RefName should be complete by the webhook
			Expect(clusterBootstrap.Spec.CNI.RefName).NotTo(Equal("antrea*"))
			assertTKRBootstrapPackageNamesContain(tanzuKubernetesRelease, clusterBootstrap.Spec.CNI.RefName)
			// clusterBootstrap.Spec.AdditionalPackages[x].RefName should be complete by the webhook
			// extra additionalPackage not in CBT
			Expect(len(clusterBootstrap.Spec.AdditionalPackages)).To(Equal(len(clusterBootstrapTemplate.Spec.AdditionalPackages) + 1))
			Expect(clusterBootstrap.Spec.AdditionalPackages[0].RefName).NotTo(Equal("pinniped*"))
			assertTKRBootstrapPackageNamesContain(tanzuKubernetesRelease, clusterBootstrap.Spec.AdditionalPackages[0].RefName)
			assertTKRBootstrapPackageNamesContain(tanzuKubernetesRelease, clusterBootstrap.Spec.AdditionalPackages[1].RefName)
			assertFindKubeVipInClusterBootstrap(clusterBootstrap)
			// the rest of fields should be added by the webhook
			Expect(clusterBootstrap.Spec.Kapp.RefName).To(Equal(clusterBootstrapTemplate.Spec.Kapp.RefName))
			Expect(clusterBootstrap.Spec.CSI.RefName).To(Equal(clusterBootstrapTemplate.Spec.CSI.RefName))
		})

		It("should complete the partial filled fields when the ClusterBootstrap CR has the predefined annotation", func() {
			// Create a ClusterBootstrap with empty spec
			clusterBootstrap := &runv1alpha3.ClusterBootstrap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      clusterBootstrapName,
					Namespace: clusterBootstrapNamespace,
					Annotations: map[string]string{
						constants.AddCBMissingFieldsAnnotationKey: tkrName,
					},
				},
				Spec: &runv1alpha3.ClusterBootstrapTemplateSpec{
					CNI: &runv1alpha3.ClusterBootstrapPackage{
						RefName: "antrea*",
					},
					CSI: &runv1alpha3.ClusterBootstrapPackage{
						RefName: fmt.Sprintf("%s.%s", fakeCSICarvelPackageRefName, fakeCarvelPackageVersion),
						ValuesFrom: &runv1alpha3.ValuesFrom{
							Inline: map[string]interface{}{"should-not-be-updated": true},
						},
					},
					AdditionalPackages: []*runv1alpha3.ClusterBootstrapPackage{
						{RefName: "pinniped*", ValuesFrom: &runv1alpha3.ValuesFrom{Inline: map[string]interface{}{"identity_management_type": "ldap"}}},
						{RefName: fmt.Sprintf("%s.%s", fakeMetricsServerCarvelPackageRefName, fakeCarvelPackageVersion)},
					},
				},
			}
			err := k8sClient.Create(ctx, clusterBootstrap)
			Expect(err).NotTo(HaveOccurred())

			err = k8sClient.Get(ctx, client.ObjectKey{Namespace: clusterBootstrapNamespace, Name: clusterBootstrapName}, clusterBootstrap)
			Expect(err).NotTo(HaveOccurred())
			Expect(clusterBootstrap.Spec).NotTo(BeNil())
			// clusterBootstrap.Spec.CNI.RefName should be completed by the webhook
			Expect(clusterBootstrap.Spec.CNI.RefName).NotTo(Equal("antrea*"))
			assertTKRBootstrapPackageNamesContain(tanzuKubernetesRelease, clusterBootstrap.Spec.CNI.RefName)
			// clusterBootstrap.Spec.AdditionalPackages[0].RefName should be completed by the webhook
			Expect(len(clusterBootstrap.Spec.AdditionalPackages)).To(Equal(2))
			Expect(clusterBootstrap.Spec.AdditionalPackages[0].RefName).NotTo(Equal("pinniped*"))
			assertTKRBootstrapPackageNamesContain(tanzuKubernetesRelease, clusterBootstrap.Spec.AdditionalPackages[0].RefName)
			// clusterBootstrap.Spec.AdditionalPackages[1].RefName should be untouched
			Expect(clusterBootstrap.Spec.AdditionalPackages[1].RefName).To(Equal(fmt.Sprintf("%s.%s", fakeMetricsServerCarvelPackageRefName, fakeCarvelPackageVersion)))
			// CSI should not be touched
			Expect(clusterBootstrap.Spec.CSI.RefName).To(Equal(fmt.Sprintf("%s.%s", fakeCSICarvelPackageRefName, fakeCarvelPackageVersion)))
			Expect(clusterBootstrap.Spec.CSI.ValuesFrom.Inline[""]).To(BeNil())
			Expect(len(clusterBootstrap.Spec.CSI.ValuesFrom.Inline)).To(Equal(1))
			Expect(clusterBootstrap.Spec.CSI.ValuesFrom.Inline["should-not-be-updated"]).To(BeTrue())
			// Kapp should be added by the webhook
			Expect(clusterBootstrap.Spec.Kapp.RefName).To(Equal(clusterBootstrapTemplate.Spec.Kapp.RefName))

		})
		It("should not fail validation if cluster is marked for deletion", func() {
			timestamp := metav1.Now()
			deletingCluster := &clusterapiv1beta1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:              clusterBootstrapName,
					Namespace:         clusterBootstrapNamespace,
					UID:               uuid.NewUUID(),
					DeletionTimestamp: &timestamp,
				},
			}
			err := clientgoscheme.AddToScheme(scheme)
			Expect(err).ToNot(HaveOccurred())
			err = clusterapiv1beta1.AddToScheme(scheme)
			Expect(err).ToNot(HaveOccurred())
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(deletingCluster).Build()
			cbWebHook := webhooks.ClusterBootstrap{Client: fakeClient}
			err = cbWebHook.SetupWebhookWithManager(context.TODO(), mgr)
			Expect(err).ToNot(HaveOccurred())
			clusterBootstrapOld := &runv1alpha3.ClusterBootstrap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      clusterBootstrapName,
					Namespace: clusterBootstrapNamespace,
				},
				Spec: &runv1alpha3.ClusterBootstrapTemplateSpec{},
			}
			clusterBootstrapNew := &runv1alpha3.ClusterBootstrap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      clusterBootstrapName,
					Namespace: clusterBootstrapNamespace,
				},
				Spec: &runv1alpha3.ClusterBootstrapTemplateSpec{},
			}
			err = cbWebHook.ValidateUpdate(context.TODO(), clusterBootstrapOld, clusterBootstrapNew)
			Expect(err).ToNot(HaveOccurred())

		})
	})
})

func assertTKRBootstrapPackageNamesContain(tkr *runv1alpha3.TanzuKubernetesRelease, name string) {
	var found bool
	for _, pkg := range tkr.Spec.BootstrapPackages {
		if pkg.Name == name {
			found = true
			break
		}
	}
	Expect(found).To(BeTrue())
}

func constructFakeTanzuKubernetesRelease() *runv1alpha3.TanzuKubernetesRelease {
	return &runv1alpha3.TanzuKubernetesRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name: tkrName,
		},
		Spec: runv1alpha3.TanzuKubernetesReleaseSpec{
			BootstrapPackages: []corev1.LocalObjectReference{
				{Name: fmt.Sprintf("%s.%s", fakeAntreaCarvelPackageRefName, fakeCarvelPackageVersion)},
				{Name: fmt.Sprintf("%s.%s", fakeCalicoCarvelPackageRefName, fakeCarvelPackageVersion)},
				{Name: fmt.Sprintf("%s.%s", fakeMetricsServerCarvelPackageRefName, fakeCarvelPackageVersion)},
				{Name: fmt.Sprintf("%s.%s", fakePinnipedCarvelPackageRefName, fakeCarvelPackageVersion)},
				{Name: fmt.Sprintf("%s.%s", fakeKubevipcloudproviderCarvelPackageRefName, fakeCarvelPackageVersion)},
			},
		},
	}
}

func constructClusterBootstrapTemplate() *runv1alpha3.ClusterBootstrapTemplate {
	return &runv1alpha3.ClusterBootstrapTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tkrName, // CBT and TKR share the same name
			Namespace: SystemNamespace,
		},
		Spec: &runv1alpha3.ClusterBootstrapTemplateSpec{
			CNI:  &runv1alpha3.ClusterBootstrapPackage{RefName: fmt.Sprintf("%s.%s", fakeAntreaCarvelPackageRefName, fakeCarvelPackageVersion)},
			Kapp: &runv1alpha3.ClusterBootstrapPackage{RefName: fmt.Sprintf("%s.%s", fakeKappCarvelPackageRefName, fakeCarvelPackageVersion)},
			CSI: &runv1alpha3.ClusterBootstrapPackage{
				RefName: fmt.Sprintf("%s.%s", fakeCSICarvelPackageRefName, fakeCarvelPackageVersion),
				ValuesFrom: &runv1alpha3.ValuesFrom{
					Inline: map[string]interface{}{
						"fake-key": "fak-value",
					},
				},
			},
			AdditionalPackages: []*runv1alpha3.ClusterBootstrapPackage{
				{RefName: fmt.Sprintf("%s.%s", fakePinnipedCarvelPackageRefName, fakeCarvelPackageVersion), ValuesFrom: &runv1alpha3.ValuesFrom{Inline: map[string]interface{}{"identity_management_type": "oidc"}}},
				{RefName: fmt.Sprintf("%s.%s", fakeMetricsServerCarvelPackageRefName, fakeCarvelPackageVersion)},
			},
		},
	}
}

func createCarvelPackages(ctx context.Context, client client.Client) {
	packageRefNames := []string{
		fakeAntreaCarvelPackageRefName,
		fakeCalicoCarvelPackageRefName,
		fakeCSICarvelPackageRefName,
		fakeKappCarvelPackageRefName,
		fakePinnipedCarvelPackageRefName,
		fakeMetricsServerCarvelPackageRefName,
		fakeKubevipcloudproviderCarvelPackageRefName,
	}

	for _, refName := range packageRefNames {
		err := client.Create(ctx, &packagev1alpha1.Package{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s.%s", refName, fakeCarvelPackageVersion),
				Namespace: SystemNamespace,
			},
			Spec: packagev1alpha1.PackageSpec{
				RefName: refName,
				Version: fakeCarvelPackageVersion,
				Template: packagev1alpha1.AppTemplateSpec{
					Spec: &kappctrlv1alph1.AppSpec{},
				},
			},
		})
		if err != nil && !apierrors.IsAlreadyExists(err) {
			Expect(err).NotTo(HaveOccurred())
		}
	}
}

func deleteCarvelPackages(ctx context.Context, client client.Client) {
	packageRefNames := []string{
		fakeAntreaCarvelPackageRefName,
		fakeCalicoCarvelPackageRefName,
		fakeCSICarvelPackageRefName,
		fakeKappCarvelPackageRefName,
		fakePinnipedCarvelPackageRefName,
		fakeMetricsServerCarvelPackageRefName,
		fakeKubevipcloudproviderCarvelPackageRefName,
	}

	for _, refName := range packageRefNames {
		err := client.Delete(ctx, &packagev1alpha1.Package{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s.%s", refName, fakeCarvelPackageVersion),
				Namespace: SystemNamespace,
			},
			Spec: packagev1alpha1.PackageSpec{
				RefName: refName,
				Version: fakeCarvelPackageVersion,
				Template: packagev1alpha1.AppTemplateSpec{
					Spec: &kappctrlv1alph1.AppSpec{},
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())
	}
}

func assertFindKubeVipInClusterBootstrap(clusterBootstrap *runv1alpha3.ClusterBootstrap) {
	match := false
	var kubevipPackage *runv1alpha3.ClusterBootstrapPackage
	for _, p := range clusterBootstrap.Spec.AdditionalPackages {
		if p.RefName == fmt.Sprintf("%s.%s", fakeKubevipcloudproviderCarvelPackageRefName, fakeCarvelPackageVersion) {
			match = true
			kubevipPackage = p
			break
		}
	}
	Expect(match).To(BeTrue())
	Expect(kubevipPackage.RefName).To(Equal(fmt.Sprintf("%s.%s", fakeKubevipcloudproviderCarvelPackageRefName, fakeCarvelPackageVersion)))
	Expect(kubevipPackage.ValuesFrom.Inline["foo"]).To(Equal("bar"))
}

var _ = Describe("Unmanaged CNI:", func() {
	var (
		clusterBootstrapTemplate  *runv1alpha3.ClusterBootstrapTemplate
		tanzuKubernetesRelease    *runv1alpha3.TanzuKubernetesRelease
		clusterBootstrapName      = "fake-clusterbootstrap"
		clusterBootstrapNamespace = "default"
	)
	BeforeEach(func() {
		// Prepare the Carvel packages
		createCarvelPackages(ctx, k8sClient)
		// Prepare the ClusterBootstrapTemplate
		clusterBootstrapTemplate = constructClusterBootstrapTemplate()
		err := k8sClient.Create(ctx, clusterBootstrapTemplate)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			Expect(err).NotTo(HaveOccurred())
		}
		// Prepare the TanzuKubernetesRelease
		tanzuKubernetesRelease = constructFakeTanzuKubernetesRelease()
		err = k8sClient.Create(ctx, tanzuKubernetesRelease)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			Expect(err).NotTo(HaveOccurred())
		}
	})
	AfterEach(func() {
		// Delete Carvel packages
		deleteCarvelPackages(ctx, k8sClient)

		// Delete the ClusterBootstrapTemplate
		clusterBootstrapTemplate = constructClusterBootstrapTemplate()
		err := k8sClient.Delete(ctx, clusterBootstrapTemplate)
		Expect(err).NotTo(HaveOccurred())

		// Delete the TanzuKubernetesRelease
		tanzuKubernetesRelease = constructFakeTanzuKubernetesRelease()
		err = k8sClient.Delete(ctx, tanzuKubernetesRelease)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("Default webhook:", func() {
		AfterEach(func() {
			_ = k8sClient.Delete(ctx, &runv1alpha3.ClusterBootstrap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      clusterBootstrapName,
					Namespace: clusterBootstrapNamespace,
				},
			})
		})

		When("Clusterbootstrap is NOT annotated and spec is nil", func() {
			It("should return error", func() {
				clusterBootstrap := &runv1alpha3.ClusterBootstrap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      clusterBootstrapName,
						Namespace: clusterBootstrapNamespace,
						Annotations: map[string]string{
							constants.AddCBMissingFieldsAnnotationKey: tkrName,
						},
					},
				}
				err := k8sClient.Create(ctx, clusterBootstrap)
				Expect(err).To(HaveOccurred())
			})
		})
		When("Clusterbootstrap is NOT annotated and CNI is NOT listed", func() {
			It("should copy CNI from template ", func() {
				clusterBootstrap := &runv1alpha3.ClusterBootstrap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      clusterBootstrapName,
						Namespace: clusterBootstrapNamespace,
						Annotations: map[string]string{
							constants.AddCBMissingFieldsAnnotationKey: tkrName,
						},
					},
					Spec: &runv1alpha3.ClusterBootstrapTemplateSpec{
						CNI: nil,
					},
				}
				err := k8sClient.Create(ctx, clusterBootstrap)
				Expect(err).NotTo(HaveOccurred())
				defaultedClusterBootstrap := &runv1alpha3.ClusterBootstrap{}
				err = k8sClient.Get(ctx, client.ObjectKey{Namespace: clusterBootstrapNamespace, Name: clusterBootstrapName}, defaultedClusterBootstrap)
				Expect(err).NotTo(HaveOccurred())
				Expect(defaultedClusterBootstrap.Spec).NotTo(BeNil())
				Expect(defaultedClusterBootstrap.Spec.CNI).NotTo(BeNil())
				Expect(defaultedClusterBootstrap.Spec.CNI.RefName).To(Equal(clusterBootstrapTemplate.Spec.CNI.RefName))
			})
		})
		When("Clusterbootstrap is NOT annotated and CNI is empty", func() {
			It("should use listed CNI", func() {
				clusterBootstrap := &runv1alpha3.ClusterBootstrap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      clusterBootstrapName,
						Namespace: clusterBootstrapNamespace,
						Annotations: map[string]string{
							constants.AddCBMissingFieldsAnnotationKey: tkrName,
						},
					},
					Spec: &runv1alpha3.ClusterBootstrapTemplateSpec{
						CNI: &runv1alpha3.ClusterBootstrapPackage{RefName: fmt.Sprintf("%s.%s", "calico-carvel-package", fakeCarvelPackageVersion)},
					},
				}
				err := k8sClient.Create(ctx, clusterBootstrap)
				Expect(err).NotTo(HaveOccurred())

				defaultedClusterBootstrap := &runv1alpha3.ClusterBootstrap{}
				err = k8sClient.Get(ctx, client.ObjectKey{Namespace: clusterBootstrapNamespace, Name: clusterBootstrapName}, defaultedClusterBootstrap)
				Expect(err).NotTo(HaveOccurred())
				Expect(defaultedClusterBootstrap.Spec).NotTo(BeNil())
				Expect(defaultedClusterBootstrap.Spec.CNI).NotTo(BeNil())
				Expect(defaultedClusterBootstrap.Spec.CNI.RefName).To(Equal(clusterBootstrap.Spec.CNI.RefName))
			})
		})

		When("Clusterbootstrap is annotated and CNI is NOT listed", func() {
			It("should set CNI to empty", func() {
				clusterBootstrap := &runv1alpha3.ClusterBootstrap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      clusterBootstrapName,
						Namespace: clusterBootstrapNamespace,
						Annotations: map[string]string{
							constants.AddCBMissingFieldsAnnotationKey: tkrName,
							constants.UnmanagedCNI:                    "",
						},
					},
					Spec: &runv1alpha3.ClusterBootstrapTemplateSpec{
						CNI: nil,
					},
				}
				err := k8sClient.Create(ctx, clusterBootstrap)
				Expect(err).NotTo(HaveOccurred())

				defaultedClusterBootstrap := &runv1alpha3.ClusterBootstrap{}
				err = k8sClient.Get(ctx, client.ObjectKey{Namespace: clusterBootstrapNamespace, Name: clusterBootstrapName}, defaultedClusterBootstrap)
				Expect(err).NotTo(HaveOccurred())
				Expect(defaultedClusterBootstrap.Spec).NotTo(BeNil())
				Expect(defaultedClusterBootstrap.Spec.CNI).To(BeNil())
			})
		})
		When("Clusterbootstrap is annotated and CNI is empty", func() {
			It("should set CNI to empty", func() {
				clusterBootstrap := &runv1alpha3.ClusterBootstrap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      clusterBootstrapName,
						Namespace: clusterBootstrapNamespace,
						Annotations: map[string]string{
							constants.AddCBMissingFieldsAnnotationKey: tkrName,
							constants.UnmanagedCNI:                    "",
						},
					},
					Spec: &runv1alpha3.ClusterBootstrapTemplateSpec{
						CNI: &runv1alpha3.ClusterBootstrapPackage{RefName: fmt.Sprintf("%s.%s", "calico-carvel-package", fakeCarvelPackageVersion)},
					},
				}
				err := k8sClient.Create(ctx, clusterBootstrap)
				Expect(err).ToNot(HaveOccurred())

				defaultedClusterBootstrap := &runv1alpha3.ClusterBootstrap{}
				err = k8sClient.Get(ctx, client.ObjectKey{Namespace: clusterBootstrapNamespace, Name: clusterBootstrapName}, defaultedClusterBootstrap)
				Expect(err).NotTo(HaveOccurred())
				Expect(defaultedClusterBootstrap.Spec).NotTo(BeNil())
				Expect(defaultedClusterBootstrap.Spec.CNI).To(BeNil())
			})
		})
	})
	Context("ValidationCreate  webhook:", func() {
		AfterEach(func() {
			_ = k8sClient.Delete(ctx, &runv1alpha3.ClusterBootstrap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      clusterBootstrapName,
					Namespace: clusterBootstrapNamespace,
				},
			})
		})
		When("Clusterbootstrap is NOT annotated and CNI is NOT listed", func() {
			It("should reject clusterbootstrap", func() {
				clusterBootstrap := &runv1alpha3.ClusterBootstrap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      clusterBootstrapName,
						Namespace: clusterBootstrapNamespace,
					},
					Spec: &runv1alpha3.ClusterBootstrapTemplateSpec{
						Kapp: &runv1alpha3.ClusterBootstrapPackage{RefName: fmt.Sprintf("%s.%s", fakeKappCarvelPackageRefName, fakeCarvelPackageVersion)},
					},
				}
				err := k8sClient.Create(ctx, clusterBootstrap)
				Expect(err).To(HaveOccurred())
			})
		})
		When("Clusterbootstrap is NOT annotated and CNI is empty", func() {
			It("should accept clusterbootstrap", func() {
				clusterBootstrap := &runv1alpha3.ClusterBootstrap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      clusterBootstrapName,
						Namespace: clusterBootstrapNamespace,
					},
					Spec: &runv1alpha3.ClusterBootstrapTemplateSpec{
						CNI:  &runv1alpha3.ClusterBootstrapPackage{RefName: fmt.Sprintf("%s.%s", "calico-carvel-package", fakeCarvelPackageVersion)},
						Kapp: &runv1alpha3.ClusterBootstrapPackage{RefName: fmt.Sprintf("%s.%s", fakeKappCarvelPackageRefName, fakeCarvelPackageVersion)},
					},
				}
				err := k8sClient.Create(ctx, clusterBootstrap)
				Expect(err).ToNot(HaveOccurred())

				newClusterBootstrap := &runv1alpha3.ClusterBootstrap{}
				Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(clusterBootstrap), newClusterBootstrap)).To(Succeed())
				Expect(newClusterBootstrap.Spec.CNI).To(Equal(clusterBootstrap.Spec.CNI))
			})
		})

		When("Clusterbootstrap is annotated and CNI is NOT listed", func() {
			It("should accept clusterbootstrap", func() {
				clusterBootstrap := &runv1alpha3.ClusterBootstrap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      clusterBootstrapName,
						Namespace: clusterBootstrapNamespace,
						Annotations: map[string]string{
							constants.UnmanagedCNI: "",
						},
					},
					Spec: &runv1alpha3.ClusterBootstrapTemplateSpec{
						Kapp: &runv1alpha3.ClusterBootstrapPackage{RefName: fmt.Sprintf("%s.%s", fakeKappCarvelPackageRefName, fakeCarvelPackageVersion)},
					},
				}
				err := k8sClient.Create(ctx, clusterBootstrap)
				Expect(err).ToNot(HaveOccurred())

				newClusterBootstrap := &runv1alpha3.ClusterBootstrap{}
				Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(clusterBootstrap), newClusterBootstrap)).To(Succeed())
				Expect(newClusterBootstrap.Spec.CNI).To(Equal(clusterBootstrap.Spec.CNI))
			})
		})
		When("Clusterbootstrap is annotated and CNI is empty", func() {
			It("should reject clusterbootstrap", func() {
				clusterBootstrap := &runv1alpha3.ClusterBootstrap{
					ObjectMeta: metav1.ObjectMeta{
						Name:      clusterBootstrapName,
						Namespace: clusterBootstrapNamespace,
						Annotations: map[string]string{
							constants.UnmanagedCNI: "",
						},
					},
					Spec: &runv1alpha3.ClusterBootstrapTemplateSpec{
						CNI:  &runv1alpha3.ClusterBootstrapPackage{RefName: fmt.Sprintf("%s.%s", "calico-carvel-package", fakeCarvelPackageVersion)},
						Kapp: &runv1alpha3.ClusterBootstrapPackage{RefName: fmt.Sprintf("%s.%s", fakeKappCarvelPackageRefName, fakeCarvelPackageVersion)},
					},
				}
				err := k8sClient.Create(ctx, clusterBootstrap)
				Expect(err).To(HaveOccurred())
			})
		})
	})
	Context("ValidationUpdate webhook:", func() {
		BeforeEach(func() {
			clusterBootstrap := &runv1alpha3.ClusterBootstrap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      clusterBootstrapName,
					Namespace: clusterBootstrapNamespace,
				},
				Spec: &runv1alpha3.ClusterBootstrapTemplateSpec{
					CNI:  &runv1alpha3.ClusterBootstrapPackage{RefName: fmt.Sprintf("%s.%s", "calico-carvel-package", fakeCarvelPackageVersion)},
					Kapp: &runv1alpha3.ClusterBootstrapPackage{RefName: fmt.Sprintf("%s.%s", fakeKappCarvelPackageRefName, fakeCarvelPackageVersion)},
				},
			}
			err := k8sClient.Create(ctx, clusterBootstrap)
			Expect(err).ToNot(HaveOccurred())
		})
		AfterEach(func() {
			_ = k8sClient.Delete(ctx, &runv1alpha3.ClusterBootstrap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      clusterBootstrapName,
					Namespace: clusterBootstrapNamespace,
				},
			})
		})
		When("New Clusterbootstrap is NOT annotated and CNI is NOT listed", func() {
			It("should reject clusterbootstrap", func() {
				originalClusterBootstrap := &runv1alpha3.ClusterBootstrap{}
				err := k8sClient.Get(ctx, client.ObjectKey{Namespace: clusterBootstrapNamespace, Name: clusterBootstrapName}, originalClusterBootstrap)
				Expect(err).NotTo(HaveOccurred())
				newClusterBootstrap := originalClusterBootstrap.DeepCopy()
				if _, ok := newClusterBootstrap.Annotations[constants.UnmanagedCNI]; ok {
					delete(newClusterBootstrap.Annotations, constants.UnmanagedCNI)
				}
				newClusterBootstrap.Spec = &runv1alpha3.ClusterBootstrapTemplateSpec{
					Kapp: &runv1alpha3.ClusterBootstrapPackage{RefName: fmt.Sprintf("%s.%s", fakeKappCarvelPackageRefName, fakeCarvelPackageVersion)},
				}

				err = k8sClient.Update(ctx, newClusterBootstrap)
				Expect(err).To(HaveOccurred())
			})
		})
		When("Clusterbootstrap is NOT annotated and CNI is empty", func() {
			It("should accept clusterbootstrap", func() {
				originalClusterBootstrap := &runv1alpha3.ClusterBootstrap{}
				err := k8sClient.Get(ctx, client.ObjectKey{Namespace: clusterBootstrapNamespace, Name: clusterBootstrapName}, originalClusterBootstrap)
				Expect(err).NotTo(HaveOccurred())
				clusterBootstrap := originalClusterBootstrap.DeepCopy()
				if _, ok := clusterBootstrap.Annotations[constants.UnmanagedCNI]; ok {
					delete(clusterBootstrap.Annotations, constants.UnmanagedCNI)
				}
				clusterBootstrap.Spec = &runv1alpha3.ClusterBootstrapTemplateSpec{
					CNI:  &runv1alpha3.ClusterBootstrapPackage{RefName: fmt.Sprintf("%s.%s", "calico-carvel-package", fakeCarvelPackageVersion)},
					Kapp: &runv1alpha3.ClusterBootstrapPackage{RefName: fmt.Sprintf("%s.%s", fakeKappCarvelPackageRefName, fakeCarvelPackageVersion)},
				}

				err = k8sClient.Update(ctx, clusterBootstrap)
				Expect(err).ToNot(HaveOccurred())

				newClusterBootstrap := &runv1alpha3.ClusterBootstrap{}
				Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(clusterBootstrap), newClusterBootstrap)).To(Succeed())
				Expect(newClusterBootstrap.Spec.CNI).To(Equal(clusterBootstrap.Spec.CNI))
			})
		})

		When("Clusterbootstrap is annotated and CNI is NOT listed", func() {
			It("should accept clusterbootstrap", func() {
				originalClusterBootstrap := &runv1alpha3.ClusterBootstrap{}
				err := k8sClient.Get(ctx, client.ObjectKey{Namespace: clusterBootstrapNamespace, Name: clusterBootstrapName}, originalClusterBootstrap)
				Expect(err).NotTo(HaveOccurred())
				clusterBootstrap := originalClusterBootstrap.DeepCopy()
				if _, ok := clusterBootstrap.Annotations[constants.UnmanagedCNI]; !ok {
					clusterBootstrap.Annotations = map[string]string{constants.UnmanagedCNI: ""}
				}
				clusterBootstrap.Spec = &runv1alpha3.ClusterBootstrapTemplateSpec{
					Kapp: &runv1alpha3.ClusterBootstrapPackage{RefName: fmt.Sprintf("%s.%s", fakeKappCarvelPackageRefName, fakeCarvelPackageVersion)},
				}

				err = k8sClient.Update(ctx, clusterBootstrap)
				Expect(err).ToNot(HaveOccurred())

				newClusterBootstrap := &runv1alpha3.ClusterBootstrap{}
				Expect(k8sClient.Get(ctx, client.ObjectKeyFromObject(clusterBootstrap), newClusterBootstrap)).To(Succeed())
				Expect(newClusterBootstrap.Spec.CNI).To(Equal(clusterBootstrap.Spec.CNI))
			})
		})
		When("Clusterbootstrap is annotated and CNI is empty", func() {
			It("reject clusterbootstrap", func() {
				originalClusterBootstrap := &runv1alpha3.ClusterBootstrap{}
				err := k8sClient.Get(ctx, client.ObjectKey{Namespace: clusterBootstrapNamespace, Name: clusterBootstrapName}, originalClusterBootstrap)
				Expect(err).NotTo(HaveOccurred())
				newClusterBootstrap := originalClusterBootstrap.DeepCopy()
				if _, ok := newClusterBootstrap.Annotations[constants.UnmanagedCNI]; !ok {
					newClusterBootstrap.Annotations = map[string]string{constants.UnmanagedCNI: ""}
				}
				newClusterBootstrap.Spec = &runv1alpha3.ClusterBootstrapTemplateSpec{
					CNI:  &runv1alpha3.ClusterBootstrapPackage{RefName: fmt.Sprintf("%s.%s", "calico-carvel-package", fakeCarvelPackageVersion)},
					Kapp: &runv1alpha3.ClusterBootstrapPackage{RefName: fmt.Sprintf("%s.%s", fakeKappCarvelPackageRefName, fakeCarvelPackageVersion)},
				}
				err = k8sClient.Update(ctx, newClusterBootstrap)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})

var _ = Describe("Clusterbootstrap", func() {
	var (
		clusterBootstrapTemplate  *runv1alpha3.ClusterBootstrapTemplate
		tanzuKubernetesRelease    *runv1alpha3.TanzuKubernetesRelease
		clusterBootstrapName      = "fake-clusterbootstrap"
		clusterBootstrapNamespace = "default"
	)
	BeforeEach(func() {
		// Prepare the Carvel packages
		createCarvelPackages(ctx, k8sClient)
		// Prepare the ClusterBootstrapTemplate
		clusterBootstrapTemplate = constructClusterBootstrapTemplate()
		err := k8sClient.Create(ctx, clusterBootstrapTemplate)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			Expect(err).NotTo(HaveOccurred())
		}
		// Prepare the TanzuKubernetesRelease
		tanzuKubernetesRelease = constructFakeTanzuKubernetesRelease()
		err = k8sClient.Create(ctx, tanzuKubernetesRelease)
		if err != nil && !apierrors.IsAlreadyExists(err) {
			Expect(err).NotTo(HaveOccurred())
		}
	})
	AfterEach(func() {
		// Delete Carvel packages
		deleteCarvelPackages(ctx, k8sClient)

		// Delete the ClusterBootstrapTemplate
		clusterBootstrapTemplate = constructClusterBootstrapTemplate()
		err := k8sClient.Delete(ctx, clusterBootstrapTemplate)
		Expect(err).NotTo(HaveOccurred())

		// Delete the TanzuKubernetesRelease
		tanzuKubernetesRelease = constructFakeTanzuKubernetesRelease()
		err = k8sClient.Delete(ctx, tanzuKubernetesRelease)
		Expect(err).NotTo(HaveOccurred())
	})

	Context("Create Validation", func() {
		AfterEach(func() {
			_ = k8sClient.Delete(ctx, &runv1alpha3.ClusterBootstrap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      clusterBootstrapName,
					Namespace: clusterBootstrapNamespace,
				},
			})
		})
		It("should not deny create request when some APIServices have no endpoint running", func() {
			// Create an external APIService which has no endpoints running
			var port int32 = 443
			customAPIService := &apiregistrationv1.APIService{
				ObjectMeta: metav1.ObjectMeta{
					Name: "v1beta1.custom.k8s.io",
				},
				Spec: apiregistrationv1.APIServiceSpec{
					Group:                 "custom.k8s.io",
					GroupPriorityMinimum:  100,
					InsecureSkipTLSVerify: true,
					Service: &apiregistrationv1.ServiceReference{
						Name:      "custom-server",
						Namespace: "kube-system",
						Port:      &port,
					},
					Version:         "v1beta1",
					VersionPriority: 100,
				},
			}
			err := k8sClient.Create(ctx, customAPIService)
			Expect(err).NotTo(HaveOccurred())
			// Create a VSphereCSIConfig
			vSphereCSIConfigName := "fake-csiconfig"
			vSphereCSIConfig := &csiv1alpha1.VSphereCSIConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      vSphereCSIConfigName,
					Namespace: clusterBootstrapNamespace,
				},
				Spec: csiv1alpha1.VSphereCSIConfigSpec{
					VSphereCSI: csiv1alpha1.VSphereCSI{
						Mode: "vsphereParavirtualCSI",
					},
				},
			}
			err = k8sClient.Create(ctx, vSphereCSIConfig)
			Expect(err).NotTo(HaveOccurred())
			// Create a ClusterBootstrap with empty spec
			clusterBootstrap := &runv1alpha3.ClusterBootstrap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      clusterBootstrapName,
					Namespace: clusterBootstrapNamespace,
					Annotations: map[string]string{
						constants.AddCBMissingFieldsAnnotationKey: tkrName,
					},
				},
				Spec: &runv1alpha3.ClusterBootstrapTemplateSpec{
					CSI: &runv1alpha3.ClusterBootstrapPackage{
						RefName: fmt.Sprintf("%s.%s", fakeCSICarvelPackageRefName, fakeCarvelPackageVersion),
						ValuesFrom: &runv1alpha3.ValuesFrom{
							ProviderRef: &corev1.TypedLocalObjectReference{
								Name:     vSphereCSIConfigName,
								APIGroup: &csiv1alpha1.GroupVersion.Group,
								Kind:     "VSphereCSIConfig",
							},
						},
					},
				},
			}
			err = k8sClient.Create(ctx, clusterBootstrap)
			Expect(err).NotTo(HaveOccurred())

			err = k8sClient.Get(ctx, client.ObjectKey{Namespace: clusterBootstrapNamespace, Name: clusterBootstrapName}, clusterBootstrap)
			Expect(err).NotTo(HaveOccurred())
		})
	})

})
