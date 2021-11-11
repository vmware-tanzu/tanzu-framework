// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"

	runv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
)

var _ = Describe("availableUpdatesInTKRs()", func() {
	var tkrs []runv1alpha1.TanzuKubernetesRelease

	tkr1 := getFakeTKR("v1.17.17---vmware.1-tkg.2", "v1.17.17+vmware.1", corev1.ConditionFalse, "[v1.17.18+vmware.1-tkg.1 v1.18.2+vmware.1-tkg.1]")
	tkr2 := getFakeTKR("v1.17.18---vmware.1-tkg.1", "v1.17.18+vmware.1", corev1.ConditionTrue, "")
	tkr3 := getFakeTKR("v1.18.1---vmware.1-tkg.2", "v1.18.1+vmware.1", corev1.ConditionFalse, "")
	tkr4 := getFakeTKR("v1.18.2---vmware.1-tkg.1", "v1.18.2+vmware.1", corev1.ConditionTrue, "")

	BeforeEach(func() {
		tkrs = []runv1alpha1.TanzuKubernetesRelease{tkr1, tkr2, tkr3, tkr4}
	})

	When("given a non-existent tkrName", func() {
		It("should return nil / empty list", func() {
			Expect(availableUpdatesInTKRs(tkrs, "blah")).To(BeEmpty())
		})
	})

	When("given a TKR with condition UpdatesAvailable=True", func() {
		It("should return only TKRs that are mentioned in the condition message", func() {
			Expect(availableUpdatesInTKRs(tkrs, "v1.17.17---vmware.1-tkg.2")).To(ContainElements(tkr2, tkr4))
			Expect(availableUpdatesInTKRs(tkrs, "v1.17.17---vmware.1-tkg.2")).ToNot(ContainElements(tkr3))
		})
	})

	When("given a TKR without condition UpdatesAvailable", func() {
		BeforeEach(func() {
			conditions.Delete(&tkrs[0], runv1alpha1.ConditionUpdatesAvailable)
		})

		When("the TKR does not have condition UpgradeAvailable=True", func() {
			It("should return only TKRs that are mentioned in the condition message", func() {
				Expect(availableUpdatesInTKRs(tkrs, "v1.17.17---vmware.1-tkg.2")).To(BeEmpty())
			})
		})

		When("the TKR has condition UpgradeAvailable=True", func() {
			BeforeEach(func() {
				conditions.Set(&tkrs[0], &clusterv1.Condition{
					Type:    runv1alpha1.ConditionUpgradeAvailable,
					Status:  corev1.ConditionTrue,
					Message: "Deprecated, TKR(s) with later version is available: v1.17.18---vmware.1-tkg.1,v1.18.2---vmware.1-tkg.1",
				})
			})

			It("should return only TKRs that are mentioned in the condition message", func() {
				Expect(availableUpdatesInTKRs(tkrs, "v1.17.17---vmware.1-tkg.2")).To(ContainElements(tkr2, tkr4))
				Expect(availableUpdatesInTKRs(tkrs, "v1.17.17---vmware.1-tkg.2")).ToNot(ContainElements(tkr3))
			})
		})
	})
})

var _ = Describe("filterTKRs()", func() {
	tkr1 := getFakeTKR("v1.17.17---vmware.1-tkg.2", "v1.17.17+vmware.1", corev1.ConditionFalse, "")
	tkr2 := getFakeTKR("v1.17.18---vmware.1-tkg.1", "v1.17.18+vmware.1", corev1.ConditionTrue, "")
	tkr3 := getFakeTKR("v1.18.1---vmware.1-tkg.2", "v1.18.1+vmware.1", corev1.ConditionFalse, "")
	tkr4 := getFakeTKR("v1.18.2---vmware.1-tkg.1", "v1.18.2+vmware.1", corev1.ConditionFalse, "")
	tkrs := []runv1alpha1.TanzuKubernetesRelease{tkr1, tkr2, tkr3, tkr4}

	It("should list only and all TKRs satisfying the predicate", func() {
		nameBeginsWith117 := func(tkr *runv1alpha1.TanzuKubernetesRelease) bool {
			return strings.HasPrefix(tkr.Name, "v1.17")
		}
		nameBeginsWith118 := func(tkr *runv1alpha1.TanzuKubernetesRelease) bool {
			return strings.HasPrefix(tkr.Name, "v1.18")
		}
		Expect(filterTKRs(tkrs, nameBeginsWith117)).To(ContainElements(tkr1, tkr2)) // all beginning with v1.17
		Expect(filterTKRs(tkrs, nameBeginsWith118)).To(ContainElements(tkr3, tkr4)) // all beginning with v1.18
		Expect(filterTKRs(tkrs, nameBeginsWith117)).ToNot(ContainElement(tkr3))     // begins with v1.18
		Expect(filterTKRs(tkrs, nameBeginsWith118)).ToNot(ContainElements(tkr1))    // begins with v1.17
	})
})
