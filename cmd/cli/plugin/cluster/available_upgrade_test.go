// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	runv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
)

var _ = Describe("availableUpgradesFromCluster", func() {
	var (
		tkrs    []runv1alpha1.TanzuKubernetesRelease
		cluster *clusterv1.Cluster
		err     error
		writer  bytes.Buffer
	)
	JustBeforeEach(func() {
		err = availableUpgradesFromCluster(cluster, tkrs, &writer)
	})

	Context("When cluster doesn't have available upgrades", func() {
		BeforeEach(func() {
			//availableUpdates := fmt.Sprintf("[%s %s %s]", "v1.18.18+vmware.1-tkg.2", "v1.18.8+vmware.1-tkg.1", "v1.18.14+vmware.1-tkg.1-rc.1")
			cluster = getFakeCluster("fake-cluster", "")

		})
		It("should not return error", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(writer.String()).To(BeEmpty())
		})
	})
	Context("When cluster have available upgrades", func() {
		BeforeEach(func() {
			tkr1 := getFakeTKR("v1.17.18---vmware.1-tkg.2", "v1.17.18+vmware.1", corev1.ConditionTrue, "")
			tkr2 := getFakeTKR("v1.18.8---vmware.1-tkg.1", "v1.18.8+vmware.1", corev1.ConditionTrue, "")
			tkr3 := getFakeTKR("v1.18.17---vmware.2-tkg.1", "v1.18.17+vmware.2", corev1.ConditionTrue, "")
			tkr4 := getFakeTKR("v1.18.14---vmware.1-tkg.1-rc.1", "v1.18.14+vmware.1", corev1.ConditionTrue, "")
			tkr5 := getFakeTKR("v1.18.18---vmware.1-tkg.2", "v1.18.18+vmware.1", corev1.ConditionTrue, "")
			tkrs = []runv1alpha1.TanzuKubernetesRelease{tkr1, tkr4, tkr3, tkr2, tkr5}

			availableUpdates := fmt.Sprintf("[%s %s %s]", "v1.18.18+vmware.1-tkg.2", "v1.18.17+vmware.2-tkg.1", "v1.18.14+vmware.1-tkg.1-rc.1")
			cluster = getFakeCluster("fake-cluster", availableUpdates)

		})
		It("should show the TKRs versions that are part of 'updatesAvailable' condition message value list ", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(writer.String()).To(ContainSubstring("v1.18.18+vmware.1-tkg.2"))
			Expect(writer.String()).To(ContainSubstring("1.18.17+vmware.2-tkg.1"))
			Expect(writer.String()).To(ContainSubstring("v1.18.14+vmware.1-tkg.1-rc.1"))
			Expect(writer.String()).ToNot(ContainSubstring("v1.18.8+vmware.1-tkg.1"))

		})
	})

})
