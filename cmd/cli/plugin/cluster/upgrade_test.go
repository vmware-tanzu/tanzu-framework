// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes"
	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"

	corev1 "k8s.io/api/core/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	runv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
	runv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

var _ = Describe("getValidTKRVersionForUpgradeGivenTKRNamePrefix", func() {
	var (
		tkrs              []runv1alpha1.TanzuKubernetesRelease
		clusterName       string
		clusterK8sVersion string
		clusterLabels     map[string]string
		tkrNamePrefix     string
		namespace         string
		latestTKRVersion  string
		err               error
	)
	const (
		TkrVersionPrefix_v1_17    = "v1.17"    //nolint
		TkrVersionPrefix_v1_18    = "v1.18"    //nolint
		TkrVersionPrefix_v1_18_20 = "v1.18.20" //nolint
	)

	JustBeforeEach(func() {
		latestTKRVersion, err = getValidTKRVersionForUpgradeGivenTKRNamePrefix(clusterName, namespace, tkrNamePrefix,
			clusterK8sVersion, clusterLabels, tkrs)
	})

	Context("When user provides TKR name prefix and cluster had TKR version label", func() {
		BeforeEach(func() {
			clusterName = "fake-cluster-name"
			namespace = "fake-namespace"
			clusterLabels = map[string]string{
				"tanzuKubernetesRelease": "v1.17.18---vmware.1-tkg.2",
			}
		})

		Context("when TKR name prefix matches with more than one TKR version supported by cluster's TKR version for upgrade", func() {

			BeforeEach(func() {
				tkrNamePrefix = TkrVersionPrefix_v1_18

				availableUpgrades := fmt.Sprintf("TKR(s) with later version is available: %s,%s,%s,%s", "v1.18.8---vmware.1-tkg.1", "v1.18.17---vmware.2-tkg.1", "v1.18.14---vmware.1-tkg.1-rc.1", "v1.18.17---vmware.1-tkg.2")
				tkr1 := getFakeTKR("v1.17.18---vmware.1-tkg.2", "v1.17.18+vmware.1", corev1.ConditionTrue, availableUpgrades)
				tkr2 := getFakeTKR("v1.18.8---vmware.1-tkg.1", "v1.18.8+vmware.1", corev1.ConditionTrue, "")
				tkr3 := getFakeTKR("v1.18.17---vmware.2-tkg.1", "v1.18.17+vmware.2", corev1.ConditionTrue, "")
				tkr4 := getFakeTKR("v1.18.14---vmware.1-tkg.1-rc.1", "v1.18.14+vmware.1", corev1.ConditionTrue, "")
				tkr5 := getFakeTKR("v1.18.17---vmware.1-tkg.2", "v1.18.17+vmware.1", corev1.ConditionTrue, "")
				tkrs = []runv1alpha1.TanzuKubernetesRelease{tkr1, tkr4, tkr3, tkr2, tkr5}

			})
			It("should return the latest TKR version matching the prefix and upgrade supported by cluster's TKR version", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(latestTKRVersion).To(Equal("v1.18.17+vmware.2-tkg.1"))
			})
		})

		Context("when TKR name prefix doesn't matches with any TKR version supported by cluster's TKR version for upgrade", func() {

			BeforeEach(func() {
				tkrNamePrefix = TkrVersionPrefix_v1_18_20
				availableUpgrades := fmt.Sprintf("TKR(s) with later version is available: %s,%s,%s,%s", "v1.18.8---vmware.1-tkg.1", "v1.18.17---vmware.2-tkg.1", "v1.18.14---vmware.1-tkg.1-rc.1", "v1.18.17---vmware.1-tkg.2")
				tkr1 := getFakeTKR("v1.17.18---vmware.1-tkg.2", "v1.17.18+vmware.1", corev1.ConditionTrue, availableUpgrades)
				tkr2 := getFakeTKR("v1.18.8---vmware.1-tkg.1", "v1.18.8+vmware.1", corev1.ConditionTrue, "")
				tkr3 := getFakeTKR("v1.18.17---vmware.2-tkg.1", "v1.18.17+vmware.2", corev1.ConditionTrue, "")
				tkr4 := getFakeTKR("v1.18.14---vmware.1-tkg.1-rc.1", "v1.18.14+vmware.1", corev1.ConditionTrue, "")
				tkr5 := getFakeTKR("v1.18.17---vmware.1-tkg.2", "v1.18.17+vmware.1", corev1.ConditionTrue, "")
				tkrs = []runv1alpha1.TanzuKubernetesRelease{tkr1, tkr4, tkr3, tkr2, tkr5}

			})
			It("should return error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cluster cannot be upgraded, no compatible upgrades found matching the TKr prefix 'v1.18.20'"))
			})
		})
		Context("when there no available upgrades supported by cluster's TKR version for upgrade", func() {

			BeforeEach(func() {
				tkrNamePrefix = TkrVersionPrefix_v1_18_20
				availableUpgrades := ""
				tkr1 := getFakeTKR("v1.17.18---vmware.1-tkg.2", "v1.17.18+vmware.1", corev1.ConditionTrue, availableUpgrades)
				tkr2 := getFakeTKR("v1.18.8---vmware.1-tkg.1", "v1.18.8+vmware.1", corev1.ConditionTrue, "")
				tkr3 := getFakeTKR("v1.18.17---vmware.2-tkg.1", "v1.18.17+vmware.2", corev1.ConditionTrue, "")
				tkr4 := getFakeTKR("v1.18.14---vmware.1-tkg.1-rc.1", "v1.18.14+vmware.1", corev1.ConditionTrue, "")
				tkr5 := getFakeTKR("v1.18.17---vmware.1-tkg.2", "v1.18.17+vmware.1", corev1.ConditionTrue, "")
				tkrs = []runv1alpha1.TanzuKubernetesRelease{tkr1, tkr4, tkr3, tkr2, tkr5}

			})
			It("should return error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to get upgrade eligible TKrs"))
			})
		})
	})

	Context("When user provides TKR name prefix and cluster doesn't have TKR version label", func() {
		BeforeEach(func() {
			clusterLabels = map[string]string{}
		})

		Context("when TKR name prefix matches with more than one TKR version eligible for upgrade from cluster's current kubernetes version", func() {

			BeforeEach(func() {
				tkrNamePrefix = TkrVersionPrefix_v1_18
				clusterK8sVersion = "v1.17.16+vmware.1"
				availableUpgrades := ""
				tkr1 := getFakeTKR("v1.17.18---vmware.1-tkg.2", "v1.17.18+vmware.1", corev1.ConditionTrue, availableUpgrades)
				tkr2 := getFakeTKR("v1.18.8---vmware.1-tkg.1", "v1.18.8+vmware.1", corev1.ConditionTrue, "")
				tkr3 := getFakeTKR("v1.18.17---vmware.2-tkg.1-rc.3", "v1.18.17+vmware.2", corev1.ConditionTrue, "")
				tkr4 := getFakeTKR("v1.18.14---vmware.1-tkg.1-rc.1", "v1.18.14+vmware.1", corev1.ConditionTrue, "")
				tkr5 := getFakeTKR("v1.18.17---vmware.1-tkg.2", "v1.18.17+vmware.1", corev1.ConditionTrue, "")
				tkrs = []runv1alpha1.TanzuKubernetesRelease{tkr1, tkr4, tkr3, tkr2, tkr5}

			})
			It("should return the latest TKR version matching the prefix and upgrade supported by cluster's kubernetes version", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(latestTKRVersion).To(Equal("v1.18.17+vmware.2-tkg.1-rc.3"))
			})
		})

		Context("when there are no latest TKRs available for upgrade from cluster's current kubernetes version", func() {

			BeforeEach(func() {
				tkrNamePrefix = TkrVersionPrefix_v1_18
				clusterK8sVersion = "v1.17.16+vmware.2"
				availableUpgrades := ""
				tkr1 := getFakeTKR("v1.17.15---vmware.1-tkg.2", "v1.17.15+vmware.1", corev1.ConditionTrue, availableUpgrades)
				tkr2 := getFakeTKR("v1.17.16---vmware.1-tkg.1", "v1.17.16+vmware.1", corev1.ConditionTrue, "")
				tkr3 := getFakeTKR("v1.17.16---vmware.1-tkg.2-rc.3", "v1.17.16+vmware.1", corev1.ConditionTrue, "")
				tkr4 := getFakeTKR("v1.19.14---vmware.1-tkg.1-rc.1", "v1.19.14+vmware.1", corev1.ConditionTrue, "")
				tkrs = []runv1alpha1.TanzuKubernetesRelease{tkr1, tkr4, tkr3, tkr2}

			})
			It("should return the latest TKR version matching the prefix and upgrade supported by cluster's kubernetes version", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cluster cannot be upgraded as there are no available upgrades"))
			})
		})
		Context("when there multiple TKRs with same latest version available for upgrade from cluster's current kubernetes version", func() {

			BeforeEach(func() {
				tkrNamePrefix = TkrVersionPrefix_v1_17
				clusterK8sVersion = "v1.17.15+vmware.2"
				availableUpgrades := ""
				tkr1 := getFakeTKR("v1.17.15---vmware.1-tkg.2", "v1.17.15+vmware.1", corev1.ConditionTrue, availableUpgrades)
				tkr2 := getFakeTKR("v1.17.16---vmware.1-tkg.1", "v1.17.16+vmware.1", corev1.ConditionTrue, "")
				tkr3 := getFakeTKR("v1.17.16---vmware.1-tkg.1-rc.3", "v1.17.16+vmware.1", corev1.ConditionTrue, "")
				tkr4 := getFakeTKR("v1.19.14---vmware.1-tkg.1-rc.1", "v1.19.14+vmware.1", corev1.ConditionTrue, "")
				tkrs = []runv1alpha1.TanzuKubernetesRelease{tkr1, tkr2, tkr3, tkr4}

			})
			It("should return the latest TKR version matching the prefix and upgrade supported by cluster's kubernetes version", func() {
				Expect(err).To(HaveOccurred())
				errString := "found multiple TKrs [v1.17.16---vmware.1-tkg.1-rc.3 v1.17.16---vmware.1-tkg.1] matching the criteria"
				Expect(err.Error()).To(ContainSubstring(errString))
			})
		})

	})

})

var _ = Describe("getValidTKRVersionForUpgradeGivenFullTKRName", func() {
	var (
		tkrs             []runv1alpha1.TanzuKubernetesRelease
		clusterName      string
		clusterNamespace string
		clusterLabels    map[string]string
		tkrForUpgrade    runv1alpha1.TanzuKubernetesRelease
		latestTKRVersion string
		err              error
	)

	JustBeforeEach(func() {
		latestTKRVersion, err = getValidTKRVersionForUpgradeGivenFullTKRName(clusterName, clusterNamespace, clusterLabels, &tkrForUpgrade, tkrs)
	})

	Context("When cluster had TKR version label", func() {
		BeforeEach(func() {
			clusterNamespace = "fake-namespace1"
			clusterLabels = map[string]string{
				"tanzuKubernetesRelease": "v1.17.18---vmware.1-tkg.2",
			}
		})

		Context("when TKR associated with user provided TKR name is in the list of TKRs supported by cluster's TKR version for upgrade", func() {

			BeforeEach(func() {
				availableUpgrades := fmt.Sprintf("TKR(s) with later version is available: %s,%s,%s,%s", "v1.18.8---vmware.1-tkg.1", "v1.18.17---vmware.2-tkg.1", "v1.18.14---vmware.1-tkg.1-rc.1", "v1.18.17---vmware.1-tkg.2")
				tkr1 := getFakeTKR("v1.17.18---vmware.1-tkg.2", "v1.17.18+vmware.1", corev1.ConditionTrue, availableUpgrades)
				tkr2 := getFakeTKR("v1.18.8---vmware.1-tkg.1", "v1.18.8+vmware.1", corev1.ConditionTrue, "")
				tkr3 := getFakeTKR("v1.18.17---vmware.2-tkg.1", "v1.18.17+vmware.2", corev1.ConditionTrue, "")
				tkr4 := getFakeTKR("v1.18.14---vmware.1-tkg.1-rc.1", "v1.18.14+vmware.1", corev1.ConditionTrue, "")
				tkr5 := getFakeTKR("v1.18.17---vmware.1-tkg.2", "v1.18.17+vmware.1", corev1.ConditionTrue, "")
				tkrs = []runv1alpha1.TanzuKubernetesRelease{tkr1, tkr4, tkr3, tkr2, tkr5}

				tkrForUpgrade = tkr5
			})
			It("should return the version of TKR whose name matches with the user provided TKR name", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(latestTKRVersion).To(Equal("v1.18.17+vmware.1-tkg.2"))
			})
		})

		Context("when TKR associated with user provided TKR name is not in the list of TKRs supported by cluster's TKR version for upgrade", func() {

			BeforeEach(func() {
				availableUpgrades := fmt.Sprintf("TKR(s) with later version is available: %s,%s,%s", "v1.18.8---vmware.1-tkg.1", "v1.18.17---vmware.2-tkg.1", "v1.18.14---vmware.1-tkg.1-rc.1")
				tkr1 := getFakeTKR("v1.17.18---vmware.1-tkg.2", "v1.17.18+vmware.1", corev1.ConditionTrue, availableUpgrades)
				tkr2 := getFakeTKR("v1.18.8---vmware.1-tkg.1", "v1.18.8+vmware.1", corev1.ConditionTrue, "")
				tkr3 := getFakeTKR("v1.18.17---vmware.2-tkg.1", "v1.18.17+vmware.2", corev1.ConditionTrue, "")
				tkr4 := getFakeTKR("v1.18.14---vmware.1-tkg.1-rc.1", "v1.18.14+vmware.1", corev1.ConditionTrue, "")
				tkr5 := getFakeTKR("v1.18.17---vmware.1-tkg.2", "v1.18.17+vmware.1", corev1.ConditionTrue, "")
				tkrs = []runv1alpha1.TanzuKubernetesRelease{tkr1, tkr4, tkr3, tkr2, tkr5}

				tkrForUpgrade = tkr5
			})
			It("should return not return error", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(latestTKRVersion).To(Equal("v1.18.17+vmware.1-tkg.2"))
			})
		})

		Context("when existing TKR associated with cluster and user provided TKR is same and no new upgrades are available for the given TKR during upgrade, it should return provided TKR", func() {

			BeforeEach(func() {
				availableUpgrades := fmt.Sprintf("TKR(s) with later version is available: %s,%s,%s", "v1.18.8---vmware.1-tkg.1", "v1.18.17---vmware.2-tkg.1", "v1.18.14---vmware.1-tkg.1-rc.1")
				tkr1 := getFakeTKR("v1.17.18---vmware.1-tkg.2", "v1.17.18+vmware.1", corev1.ConditionTrue, availableUpgrades)
				tkr2 := getFakeTKR("v1.18.8---vmware.1-tkg.1", "v1.18.8+vmware.1", corev1.ConditionTrue, "")
				tkr3 := getFakeTKR("v1.18.17---vmware.2-tkg.1", "v1.18.17+vmware.2", corev1.ConditionTrue, "")
				tkr4 := getFakeTKR("v1.18.14---vmware.1-tkg.1-rc.1", "v1.18.14+vmware.1", corev1.ConditionTrue, "")
				tkr5 := getFakeTKR("v1.18.17---vmware.1-tkg.2", "v1.18.17+vmware.1", corev1.ConditionTrue, "")
				tkrs = []runv1alpha1.TanzuKubernetesRelease{tkr1, tkr4, tkr3, tkr2, tkr5}

				tkrForUpgrade = tkr5

				clusterLabels = map[string]string{
					"tanzuKubernetesRelease": "v1.18.17---vmware.1-tkg.2",
				}
			})
			It("should return not return error", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(latestTKRVersion).To(Equal("v1.18.17+vmware.1-tkg.2"))
			})
		})

	})

	Context("When cluster doesn't have TKR version label", func() {
		BeforeEach(func() {
			clusterLabels = map[string]string{}
		})

		Context("when TKR associated with user provided TKR name is in the list of TKRs supported by cluster's TKR version for upgrade", func() {

			BeforeEach(func() {
				availableUpgrades := fmt.Sprintf("TKR(s) with later version is available: %s,%s,%s,%s", "v1.18.8---vmware.1-tkg.1", "v1.18.17---vmware.2-tkg.1", "v1.18.14---vmware.1-tkg.1-rc.1", "v1.18.17---vmware.1-tkg.2")
				tkr1 := getFakeTKR("v1.17.18---vmware.1-tkg.2", "v1.17.18+vmware.1", corev1.ConditionTrue, availableUpgrades)
				tkr2 := getFakeTKR("v1.18.8---vmware.1-tkg.1", "v1.18.8+vmware.1", corev1.ConditionTrue, "")
				tkr3 := getFakeTKR("v1.18.17---vmware.2-tkg.1", "v1.18.17+vmware.2", corev1.ConditionTrue, "")
				tkr4 := getFakeTKR("v1.18.14---vmware.1-tkg.1-rc.1", "v1.18.14+vmware.1", corev1.ConditionTrue, "")
				tkr5 := getFakeTKR("v1.18.17---vmware.1-tkg.2", "v1.18.17+vmware.1", corev1.ConditionTrue, "")
				tkrs = []runv1alpha1.TanzuKubernetesRelease{tkr1, tkr4, tkr3, tkr2, tkr5}

				tkrForUpgrade = tkr5
			})
			It("should return the version of TKR whose name matches with the user provided TKR name", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(latestTKRVersion).To(Equal("v1.18.17+vmware.1-tkg.2"))
			})
		})
		Context("when TKR associated with user provided TKR name is not in the list of TKRs supported by cluster's TKR version for upgrade", func() {

			BeforeEach(func() {
				availableUpgrades := fmt.Sprintf("TKR(s) with later version is available: %s,%s,%s", "v1.18.8---vmware.1-tkg.1", "v1.18.17---vmware.2-tkg.1", "v1.18.14---vmware.1-tkg.1-rc.1")
				tkr1 := getFakeTKR("v1.17.18---vmware.1-tkg.2", "v1.17.18+vmware.1", corev1.ConditionTrue, availableUpgrades)
				tkr2 := getFakeTKR("v1.18.8---vmware.1-tkg.1", "v1.18.8+vmware.1", corev1.ConditionTrue, "")
				tkr3 := getFakeTKR("v1.18.17---vmware.2-tkg.1", "v1.18.17+vmware.2", corev1.ConditionTrue, "")
				tkr4 := getFakeTKR("v1.18.14---vmware.1-tkg.1-rc.1", "v1.18.14+vmware.1", corev1.ConditionTrue, "")
				tkr5 := getFakeTKR("v1.18.17---vmware.1-tkg.2", "v1.18.17+vmware.1", corev1.ConditionTrue, "")
				tkrs = []runv1alpha1.TanzuKubernetesRelease{tkr1, tkr4, tkr3, tkr2, tkr5}

				tkrForUpgrade = tkr5
			})
			It("should still return the version of TKR whose name matches with the user provided TKR name", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(latestTKRVersion).To(Equal("v1.18.17+vmware.1-tkg.2"))
			})
		})
	})

})

var _ = Describe("getValidTKRVersionFromClusterForUpgrade", func() {
	var (
		latestTKRVersion string
		err              error
		cluster          *clusterv1.Cluster
		tkrNamePrefix    string
	)
	const (
		TkrVersionPrefix_v1_18    = "v1.18"    //nolint
		TkrVersionPrefix_v1_18_20 = "v1.18.20" //nolint
	)
	JustBeforeEach(func() {
		latestTKRVersion, err = getValidTKRVersionFromClusterForUpgrade(cluster, tkrNamePrefix)
	})

	Context("when TKR associated with user provided TKR prefix name is not in the list of TKRs supported by cluster for upgrade", func() {

		BeforeEach(func() {
			availableUpdates := fmt.Sprintf("[%s %s %s]", "v1.18.18+vmware.1-tkg.2", "v1.18.8+vmware.1-tkg.1", "v1.18.14+vmware.1-tkg.1-rc.1")
			cluster = getFakeCluster("fake-cluster", availableUpdates)
			tkrNamePrefix = TkrVersionPrefix_v1_18_20
		})
		It("should return error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("no compatible upgrades found matching the TKR name/prefix '%s'", tkrNamePrefix)))
		})
	})

	Context("when existing TKR associated with cluster and user provided TKR is same and though no new upgrades are available for cluster to upgrade, it should return provided TKR", func() {

		BeforeEach(func() {
			clusterLabels := map[string]string{
				runv1alpha3.LabelTKR: "v1.18.17---vmware.1-tkg.2",
			}
			cluster = getFakeCluster("fake-cluster", "")
			cluster.Labels = clusterLabels
			tkrNamePrefix = "v1.18.17---vmware.1-tkg.2"
		})
		It("should return TKR name provided by user", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(latestTKRVersion).To(Equal(strings.ReplaceAll(tkrNamePrefix, "---", "+")))
		})
	})

	Context("when TKR associated with user provided TKR prefix name is in the list of TKRs supported by cluster for upgrade", func() {

		BeforeEach(func() {
			availableUpdates := fmt.Sprintf("[%s %s %s]", "v1.18.18+vmware.1-tkg.2", "v1.18.8+vmware.1-tkg.1", "v1.18.14+vmware.1-tkg.1-rc.1")
			cluster = getFakeCluster("fake-cluster", availableUpdates)
			tkrNamePrefix = TkrVersionPrefix_v1_18
		})
		It("should return the latest version of TKR whose name prefix matches with the user provided TKR prefix name", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(latestTKRVersion).To(Equal("v1.18.18+vmware.1-tkg.2"))
		})
	})
	Context("when TKR associated with user provided TKR name is in the list of TKRs supported by cluster for upgrade", func() {

		BeforeEach(func() {
			availableUpdates := fmt.Sprintf("[%s %s %s]", "v1.18.18+vmware.1-tkg.2", "v1.18.8+vmware.1-tkg.1", "v1.18.14+vmware.1-tkg.1-rc.1")
			cluster = getFakeCluster("fake-cluster", availableUpdates)
			tkrNamePrefix = "v1.18.14---vmware.1-tkg.1-rc.1"
		})
		It("should return the version of user provided TKR name", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(latestTKRVersion).To(Equal("v1.18.14+vmware.1-tkg.1-rc.1"))
		})
	})
	Context("when user provided TKR prefix name but the cluster doesn't have any updates available", func() {

		BeforeEach(func() {
			cluster = getFakeCluster("fake-cluster", "")
			cluster.Namespace = "fake-namespace"
			tkrNamePrefix = TkrVersionPrefix_v1_18
		})
		It("should return error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("no available upgrades for cluster '%s', namespace '%s'", cluster.Name, cluster.Namespace)))
		})
	})
})

var _ = Describe("getClusterResource", func() {
	var (
		err           error
		clusterName   string
		namespace     string
		clusterClient *fakes.ClusterClient
		cluster       *clusterv1.Cluster
	)
	BeforeEach(func() {
		clusterClient = &fakes.ClusterClient{}
		clusterName = "fake-cluster"
		namespace = "fake-namespace"
	})

	JustBeforeEach(func() {
		cluster, err = getClusterResource(clusterClient, clusterName, namespace)
	})

	Context("when resource is not found ", func() {

		BeforeEach(func() {
			namespace = ""
			clusterClient.GetResourceReturns(errors.New("fake list resource error"))
		})
		It("should return error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("unable to get cluster %q from namespace %q", clusterName, constants.DefaultNamespace)))
			Expect(cluster).To(BeNil())
		})
	})
	Context("when resource is found ", func() {

		BeforeEach(func() {
			namespace = "fake-namespace"
			clusterClient.GetResourceCalls(func(c interface{}, name, namespace string, pv clusterclient.PostVerifyrFunc, opt *clusterclient.PollOptions) error {
				cc := c.(*clusterv1.Cluster)
				fc := getFakeCluster(clusterName, "")
				*cc = *fc
				return nil
			})
		})
		It("should not return error", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(cluster).ToNot(BeNil())
		})
	})
})

func getFakeCluster(name, updatesAvailablesMsg string) *clusterv1.Cluster {
	cluster := &clusterv1.Cluster{}
	cluster.Name = name
	if len(updatesAvailablesMsg) == 0 {
		cluster.Status.Conditions = []clusterv1.Condition{
			{
				Type:   clusterv1.ConditionType(runv1alpha3.ConditionUpdatesAvailable),
				Status: corev1.ConditionFalse,
				Reason: "AlreadyUptoDate",
			},
		}
	} else {
		cluster.Status.Conditions = []clusterv1.Condition{
			{
				Type:    clusterv1.ConditionType(runv1alpha3.ConditionUpdatesAvailable),
				Status:  corev1.ConditionTrue,
				Message: updatesAvailablesMsg,
			},
		}
	}
	return cluster
}
