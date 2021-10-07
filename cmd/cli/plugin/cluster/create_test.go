// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha4"

	runv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "create cluster test")
}

var _ = Describe("getLatestTKRVersionMatchingTKRPrefix", func() {
	var (
		tkrsWithPrefixMatch []runv1alpha1.TanzuKubernetesRelease
		tkrNamePrefix       string
		latestTKRVersion    string
		err                 error
	)
	const (
		TkrVersionPrefix_v1_17 = "v1.17" //nolint
	)

	JustBeforeEach(func() {
		latestTKRVersion, err = getLatestTKRVersionMatchingTKRPrefix(tkrNamePrefix, tkrsWithPrefixMatch)
	})

	Context("When the list of prefix matched TKRs has highest version TKR as incompatible", func() {
		BeforeEach(func() {
			tkr1 := getFakeTKR("v1.17.18---vmware.1-tkg.2", "v1.17.18+vmware.1", corev1.ConditionFalse, "")
			tkr2 := getFakeTKR("v1.17.8---vmware.1-tkg.1", "v1.17.8+vmware.1", corev1.ConditionTrue, "")
			tkr3 := getFakeTKR("v1.17.17---vmware.2-tkg.1", "v1.17.17---vmware.2", corev1.ConditionTrue, "")
			tkr4 := getFakeTKR("v1.17.14---vmware.1-tkg.1-rc.1", "v1.17.14---vmware.1", corev1.ConditionTrue, "")
			tkr5 := getFakeTKR("v1.17.17---vmware.1-tkg.2", "v1.17.17---vmware.1", corev1.ConditionTrue, "")

			tkrsWithPrefixMatch = []runv1alpha1.TanzuKubernetesRelease{tkr1, tkr4, tkr3, tkr2, tkr5}
			tkrNamePrefix = TkrVersionPrefix_v1_17
		})
		It("should return the next latest TKR version that is compatible", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(latestTKRVersion).To(Equal("v1.17.17+vmware.2-tkg.1"))
		})
	})
	Context("When the list of prefix matched TKRs has multiple latest TKRs", func() {
		BeforeEach(func() {
			tkr1 := getFakeTKR("v1.17.18---vmware.1-tkg.2", "v1.17.18---vmware.1", corev1.ConditionTrue, "")
			tkr2 := getFakeTKR("v1.17.18---vmware.2-tkg.1-rc.1", "v1.17.18---vmware.2-tkg.1", corev1.ConditionTrue, "")
			tkr3 := getFakeTKR("v1.17.15---vmware.1-tkg.1", "v1.17.15---vmware.1", corev1.ConditionTrue, "")
			tkr4 := getFakeTKR("v1.17.18---vmware.2-tkg.1-zlatest1", "1.17.18---vmware.2-tkg.1", corev1.ConditionTrue, "")
			tkrsWithPrefixMatch = []runv1alpha1.TanzuKubernetesRelease{tkr1, tkr2, tkr3, tkr4}
			tkrNamePrefix = TkrVersionPrefix_v1_17
		})
		It("should return error ", func() {
			Expect(err).To(HaveOccurred())
			errString := "found multiple TKrs [v1.17.18---vmware.2-tkg.1-zlatest1 v1.17.18---vmware.2-tkg.1-rc.1] matching the criteria"
			Expect(err.Error()).To(ContainSubstring(errString))
		})
	})
	Context("When the list of prefix matched TKRs has no compatible TKRs", func() {
		BeforeEach(func() {
			tkr1 := getFakeTKR("v1.17.18---vmware.1-tkg.2", "v1.17.18---vmware.1", corev1.ConditionFalse, "")
			tkr2 := getFakeTKR("v1.17.8---vmware.1-tkg.1", "v1.17.8---vmware.1", corev1.ConditionFalse, "")
			tkr3 := getFakeTKR("v1.17.17---vmware.2-tkg.1", "v1.17.17---vmware.2", corev1.ConditionFalse, "")
			tkrsWithPrefixMatch = []runv1alpha1.TanzuKubernetesRelease{tkr1, tkr3, tkr2}
			tkrNamePrefix = TkrVersionPrefix_v1_17
		})
		It("should return error as there is no single compatible TKR", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("could not find a matching compatible Tanzu Kubernetes release for name \"v1.17\""))
		})
	})
})

func getFakeTKR(tkrName, k8sversion string, compatibleStatus corev1.ConditionStatus, updatesAvailableMsg string) runv1alpha1.TanzuKubernetesRelease {
	tkr := runv1alpha1.TanzuKubernetesRelease{}
	tkr.Name = tkrName
	tkr.Spec.Version = strings.ReplaceAll(tkrName, "---", "+")
	tkr.Spec.KubernetesVersion = k8sversion
	tkr.Status.Conditions = []clusterv1.Condition{
		{
			Type:   clusterv1.ConditionType(runv1alpha1.ConditionCompatible),
			Status: compatibleStatus,
		},
		{
			Type:    clusterv1.ConditionType(runv1alpha1.ConditionUpgradeAvailable),
			Status:  corev1.ConditionTrue,
			Message: updatesAvailableMsg,
		},
	}
	return tkr
}
