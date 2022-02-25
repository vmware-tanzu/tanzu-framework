// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"

	"github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
)

const (
	testReason  = "some reason"
	testMessage = "some message"
)

var _ = Describe("Summarize AppConditions", func() {
	Context("SummarizeAppConditions()", func() {
		var (
			conditions       []v1alpha1.AppCondition
			summaryCondition v1alpha1.AppCondition
		)

		When("there is any condition with 'Reconciling' type", func() {
			BeforeEach(func() {
				conditions = []v1alpha1.AppCondition{
					{
						Type:    v1alpha1.Reconciling,
						Status:  corev1.ConditionTrue,
						Reason:  testReason,
						Message: testMessage,
					},
					{
						Type:    v1alpha1.ReconcileSucceeded,
						Status:  corev1.ConditionTrue,
						Reason:  testReason,
						Message: testMessage,
					},
				}
				summaryCondition = SummarizeAppConditions(conditions)
			})

			It("the summarized condition's type should be 'Reconciling'", func() {
				Expect(summaryCondition.Type).Should(Equal(v1alpha1.Reconciling))
				Expect(summaryCondition.Status).Should(Equal(corev1.ConditionTrue))
			})
		})

		When("there is no condition with 'Reconciling' type && there is a condition with 'ReconcileFailed' type", func() {
			BeforeEach(func() {
				conditions = []v1alpha1.AppCondition{
					{
						Type:    v1alpha1.ReconcileSucceeded,
						Status:  corev1.ConditionTrue,
						Reason:  testReason,
						Message: testMessage,
					},
					{
						Type:    v1alpha1.ReconcileFailed,
						Status:  corev1.ConditionFalse,
						Reason:  testReason,
						Message: testMessage,
					},
				}
				summaryCondition = SummarizeAppConditions(conditions)
			})

			It("the summarized condition's type should be 'ReconcileFailed'", func() {
				Expect(summaryCondition.Type).Should(Equal(v1alpha1.ReconcileFailed))
				Expect(summaryCondition.Status).Should(Equal(corev1.ConditionFalse))
			})
		})

		When("there is no condition with 'Reconciling' or 'ReconcileFailed' types && there is a condition with 'ReconcileSucceeded' type", func() {
			BeforeEach(func() {
				conditions = []v1alpha1.AppCondition{
					{
						Type:    v1alpha1.ReconcileSucceeded,
						Status:  corev1.ConditionTrue,
						Reason:  testReason,
						Message: testMessage,
					},
					{
						Type:    v1alpha1.AppConditionType("Unknown"),
						Status:  corev1.ConditionFalse,
						Reason:  testReason,
						Message: testMessage,
					},
				}
				summaryCondition = SummarizeAppConditions(conditions)
			})

			It("the summarized condition's type should be 'ReconcileSucceeded'", func() {
				Expect(summaryCondition.Type).Should(Equal(v1alpha1.ReconcileSucceeded))
				Expect(summaryCondition.Status).Should(Equal(corev1.ConditionTrue))
			})
		})

		When("there is no condition with 'Reconciling' or 'ReconcileFailed' or 'ReconcileSucceeded' types", func() {
			BeforeEach(func() {
				conditions = []v1alpha1.AppCondition{
					{
						Type:    "SomeOtherCondition",
						Status:  corev1.ConditionTrue,
						Reason:  testReason,
						Message: testMessage,
					},
				}
				summaryCondition = SummarizeAppConditions(conditions)
			})

			It("the summarized condition's type should be ", func() {
				Expect(summaryCondition.Type).Should(Equal(UnknownCondition))
				Expect(summaryCondition.Status).Should(Equal(corev1.ConditionStatus("")))
			})
		})

		When("there are conditions with conflicting state types", func() {
			BeforeEach(func() {
				conditions = []v1alpha1.AppCondition{
					{
						Type:    v1alpha1.Reconciling,
						Status:  corev1.ConditionTrue,
						Reason:  testReason,
						Message: testMessage,
					},
					{
						Type:    v1alpha1.Reconciling,
						Status:  corev1.ConditionFalse,
						Reason:  testReason,
						Message: testMessage,
					},
					{
						Type:    v1alpha1.Reconciling,
						Status:  corev1.ConditionUnknown,
						Reason:  testReason,
						Message: testMessage,
					},
					{
						Type:    v1alpha1.Reconciling,
						Status:  "",
						Reason:  testReason,
						Message: testMessage,
					},
				}
				summaryCondition = SummarizeAppConditions(conditions)
			})

			It("the summarized condition's state would be 'True'", func() {
				Expect(summaryCondition.Type).Should(Equal(v1alpha1.Reconciling))
				Expect(summaryCondition.Status).Should(Equal(corev1.ConditionTrue))
			})
		})

	})
})
