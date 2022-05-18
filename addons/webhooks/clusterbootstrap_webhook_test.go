// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package webhooks_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kappctrlv1alph1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	packagev1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"

	runv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

var _ = Describe("ClusterbootstrapWebhook", func() {

	Context("Verify the logic of clusterbootstrap webhook", func() {
		var (
			clusterBootstrapTemplate  *runv1alpha3.ClusterBootstrapTemplate
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
		})
		It("should add defaults to the missing fields when the ClusterBootstrap CR has the predefined annotation", func() {
			// Create a ClusterBootstrap with empty spec
			clusterBootstrap := &runv1alpha3.ClusterBootstrap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      clusterBootstrapName,
					Namespace: clusterBootstrapNamespace,
					Annotations: map[string]string{
						constants.AddCBMissingFieldsAnnotationKey: clusterBootstrapTemplate.Namespace + "/" + clusterBootstrapTemplate.Name,
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
						constants.AddCBMissingFieldsAnnotationKey: clusterBootstrapTemplate.Namespace + "/" + clusterBootstrapTemplate.Name,
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
		It("should NOT add defaults to the missing fields when the ClusterBootstrap CR does not have the predefined annotation", func() {
			// Create a ClusterBootstrap with empty spec
			clusterBootstrap := &runv1alpha3.ClusterBootstrap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      clusterBootstrapName,
					Namespace: clusterBootstrapNamespace,
					Annotations: map[string]string{
						constants.AddCBMissingFieldsAnnotationKey: "invalid",
					},
				},
				Spec: &runv1alpha3.ClusterBootstrapTemplateSpec{},
			}
			err := k8sClient.Create(ctx, clusterBootstrap)
			Expect(err).To(HaveOccurred())
			Expect(apierrors.IsForbidden(err)).To(BeTrue())
		})
	})
})

func constructClusterBootstrapTemplate() *runv1alpha3.ClusterBootstrapTemplate {
	fakeCarvelPackageName := fmt.Sprintf("%s.%s", fakeCarvelPackageRefName, fakeCarvelPackageVersion)
	return &runv1alpha3.ClusterBootstrapTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake-clusterbootstrap-template",
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

var fakeCarvelPackageRefName = "fake-carvel-package"
var fakeCarvelPackageVersion = "1.0.0"

func createCarvelPackages(ctx context.Context, client client.Client) {
	fakeCarvelPackage := &packagev1alpha1.Package{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s.%s", fakeCarvelPackageRefName, fakeCarvelPackageVersion),
			Namespace: SystemNamespace,
		},
		Spec: packagev1alpha1.PackageSpec{
			RefName: "fake-carvel-package",
			Version: "1.0.0",
			Template: packagev1alpha1.AppTemplateSpec{
				Spec: &kappctrlv1alph1.AppSpec{},
			},
		},
	}
	err := client.Create(ctx, fakeCarvelPackage)
	Expect(err).NotTo(HaveOccurred())
}
