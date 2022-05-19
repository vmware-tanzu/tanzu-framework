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
	tkrName                  = "v1.23.5---vmware.1-tkg.1-zshippable"
	fakeCarvelPackageRefName = "fake-carvel-package"
	fakeCarvelPackageVersion = "1.0.0"
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
						constants.AddCBMissingFieldsAnnotationKey: "v1.23.5---vmware.1-tkg.1-zshippable",
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
		})
		It("should add defaults ONLY to the missing fields when the ClusterBootstrap CR has the predefined annotation", func() {
			// Create a ClusterBootstrap with empty spec
			clusterBootstrap := &runv1alpha3.ClusterBootstrap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      clusterBootstrapName,
					Namespace: clusterBootstrapNamespace,
					Annotations: map[string]string{
						constants.AddCBMissingFieldsAnnotationKey: "v1.23.5---vmware.1-tkg.1-zshippable",
					},
				},
				Spec: &runv1alpha3.ClusterBootstrapTemplateSpec{
					CSI: &runv1alpha3.ClusterBootstrapPackage{
						RefName: fmt.Sprintf("%s.%s", fakeCarvelPackageRefName, fakeCarvelPackageVersion),
						ValuesFrom: &runv1alpha3.ValuesFrom{
							Inline: map[string]interface{}{
								"foo": "bar",
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
			Expect(clusterBootstrap.Spec.CNI.RefName).To(Equal(clusterBootstrapTemplate.Spec.CNI.RefName))
			Expect(clusterBootstrap.Spec.Kapp.RefName).To(Equal(clusterBootstrapTemplate.Spec.Kapp.RefName))
			Expect(clusterBootstrap.Spec.CSI.RefName).To(Equal(clusterBootstrapTemplate.Spec.CSI.RefName))
			Expect(clusterBootstrap.Spec.CSI.ValuesFrom.Inline["foo"]).To(Equal("bar"))
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
			Expect(apierrors.IsNotFound(err)).To(BeTrue())
		})

		// TODO: Add more tests to verify the CompleteCBPackageRefNamesFromTKR() logic.
		// We don't have that test right now is because envtest has some issues to create TanzuKubernetesRelease resource
		// by using client.Create(). The creation succeeds but the tkr.Spec.BootstrapPackages becomes empty after the
		// creation. We to revisit and make this tests comprehensive.
	})
})

func constructFakeTanzuKubernetesRelease() *runv1alpha3.TanzuKubernetesRelease {
	return &runv1alpha3.TanzuKubernetesRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name: tkrName,
		},
		Spec: runv1alpha3.TanzuKubernetesReleaseSpec{
			BootstrapPackages: []corev1.LocalObjectReference{
				{Name: "antrea.tanzu.vmware.com.1.2.3+vmware.4-tkg.2-advanced-zshippable"},
				{Name: "calico.tanzu.vmware.com.3.22.1+vmware.1-tkg.1-zshippable"},
			},
		},
	}
}

func constructClusterBootstrapTemplate() *runv1alpha3.ClusterBootstrapTemplate {
	fakeCarvelPackageName := fmt.Sprintf("%s.%s", fakeCarvelPackageRefName, fakeCarvelPackageVersion)
	return &runv1alpha3.ClusterBootstrapTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tkrName, // CBT and TKR share the same name
			Namespace: SystemNamespace,
		},
		Spec: &runv1alpha3.ClusterBootstrapTemplateSpec{
			CNI:  &runv1alpha3.ClusterBootstrapPackage{RefName: fakeCarvelPackageName},
			Kapp: &runv1alpha3.ClusterBootstrapPackage{RefName: fakeCarvelPackageName},
			CSI: &runv1alpha3.ClusterBootstrapPackage{
				RefName: fakeCarvelPackageName,
				ValuesFrom: &runv1alpha3.ValuesFrom{
					Inline: map[string]interface{}{
						"fake-key": "fak-value",
					},
				}},
		},
	}
}

func createCarvelPackages(ctx context.Context, client client.Client) {
	fakeCarvelPackage := &packagev1alpha1.Package{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s.%s", fakeCarvelPackageRefName, fakeCarvelPackageVersion),
			Namespace: SystemNamespace,
		},
		Spec: packagev1alpha1.PackageSpec{
			RefName: fakeCarvelPackageRefName,
			Version: fakeCarvelPackageVersion,
			Template: packagev1alpha1.AppTemplateSpec{
				Spec: &kappctrlv1alph1.AppSpec{},
			},
		},
	}
	err := client.Create(ctx, fakeCarvelPackage)
	Expect(err).NotTo(HaveOccurred())
}
