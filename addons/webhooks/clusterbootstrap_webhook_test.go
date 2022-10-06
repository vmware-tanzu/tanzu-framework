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
	"sigs.k8s.io/controller-runtime/pkg/client"

	kappctrlv1alph1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	packagev1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	runv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

var (
	tkrName                               = "v1.23.5---vmware.1-tkg.1-zshippable"
	fakeAntreaCarvelPackageRefName        = "antrea-carvel-package"
	fakeCalicoCarvelPackageRefName        = "calico-carvel-package"
	fakeCSICarvelPackageRefName           = "vsphere-pv-csi-carvel-package"
	fakeKappCarvelPackageRefName          = "kapp-controller-carvel-package"
	fakePinnipedCarvelPackageRefName      = "pinniped-carvel-package"
	fakeMetricsServerCarvelPackageRefName = "metrics-server-carvel-package"
	fakeCarvelPackageVersion              = "1.0.0"
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
			Expect(len(clusterBootstrap.Spec.AdditionalPackages)).To(Equal(len(clusterBootstrapTemplate.Spec.AdditionalPackages)))
			Expect(clusterBootstrap.Spec.AdditionalPackages[0].RefName).NotTo(Equal("pinniped*"))
			assertTKRBootstrapPackageNamesContain(tanzuKubernetesRelease, clusterBootstrap.Spec.AdditionalPackages[0].RefName)
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
		Expect(err).NotTo(HaveOccurred())
	}
}
