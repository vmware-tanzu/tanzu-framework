// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package predicates

import (
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	cniv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cni/v1alpha1"
)

var _ = Describe("Addon config annotation predicate", func() {
	Context("predicate: processIfConfigOfKindWithoutAnnotation()", func() {
		var (
			antreaConfigObj *cniv1alpha1.AntreaConfig
			configKind      string
			namespace       string
			logger          logr.Logger
			result          bool
		)

		BeforeEach(func() {
			namespace = "test-ns"
			logger = ctrl.Log.WithName("processIfConfigOfKindWithoutAnnotation")
			antreaConfigObj = &cniv1alpha1.AntreaConfig{
				TypeMeta: metav1.TypeMeta{Kind: constants.AntreaConfigKind},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-config-name",
					Namespace: namespace,
					Annotations: map[string]string{
						constants.TKGAnnotationTemplateConfig: "true",
					}},
			}
			configKind = constants.AntreaConfigKind
		})

		When("input config matches with specified Kind and has the annotation", func() {
			BeforeEach(func() {
				result = processIfConfigOfKindWithoutAnnotation(constants.TKGAnnotationTemplateConfig, configKind, namespace, antreaConfigObj, logger)
			})

			It("should return false", func() {
				Expect(result).To(BeFalse())
			})
		})

		When("input config does not have the specified annotation", func() {
			BeforeEach(func() {
				delete(antreaConfigObj.Annotations, constants.TKGAnnotationTemplateConfig)
				result = processIfConfigOfKindWithoutAnnotation(constants.TKGAnnotationTemplateConfig, configKind, namespace, antreaConfigObj, logger)
			})

			It("should return true", func() {
				Expect(result).To(BeTrue())
			})
		})

		When("input config's annotations is nil", func() {
			BeforeEach(func() {
				antreaConfigObj.Annotations = nil
				result = processIfConfigOfKindWithoutAnnotation(constants.TKGAnnotationTemplateConfig, configKind, namespace, antreaConfigObj, logger)
			})

			It("should return true", func() {
				Expect(result).To(BeTrue())
			})
		})

		When("input config does not match with the given Kind", func() {
			BeforeEach(func() {
				configKind = constants.CalicoConfigKind
				result = processIfConfigOfKindWithoutAnnotation(constants.TKGAnnotationTemplateConfig, configKind, namespace, antreaConfigObj, logger)
			})

			It("should return true", func() {
				Expect(result).To(BeTrue())
			})
		})

		When("input config is not in the given namespace", func() {
			BeforeEach(func() {
				result = processIfConfigOfKindWithoutAnnotation(constants.TKGAnnotationTemplateConfig, configKind, "another-ns", antreaConfigObj, logger)
			})

			It("should return true", func() {
				Expect(result).To(BeTrue())
			})
		})
	})
})
