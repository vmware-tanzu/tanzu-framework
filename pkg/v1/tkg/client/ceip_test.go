// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/vmware-tanzu-private/core/pkg/v1/tkg/client"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/fakes"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/region"
)

var _ = Describe("Unit tests for ceip", func() {
	var (
		err                   error
		regionalClusterClient *fakes.ClusterClient
		tkgClient             *TkgClient
		clusterName           string
		ceipStatus            ClusterCeipInfo
		context               region.RegionContext
	)

	BeforeEach(func() {
		context = region.RegionContext{
			ClusterName: "fake-cluster",
			ContextName: "fake-cluster-admin@fake-cluster",
		}
		regionalClusterClient = &fakes.ClusterClient{}
		tkgClient, err = CreateTKGClient("../fakes/config/config.yaml", testingDir, "../fakes/config/bom/tkg-bom-v1.3.1.yaml", 2*time.Second)
		Expect(err).NotTo(HaveOccurred())
		clusterName = "fake-cluster"
	})

	Describe("SetCEIP", func() {
		BeforeEach(func() {
			regionalClusterClient.RemoveCEIPTelemetryJobReturns(nil)
			regionalClusterClient.AddCEIPTelemetryJobReturns(nil)
			regionalClusterClient.GetRegionalClusterDefaultProviderNameReturns("aws:v0.5.5", nil)
		})

		Context("When opt-ing out of CEIP", func() {
			JustBeforeEach(func() {
				regionalClusterClient.IsPacificRegionalClusterReturns(false, nil)
				err = tkgClient.DoSetCEIPParticipation(regionalClusterClient, context, false, "true", "")
			})
			It("should not error", func() {
				Expect(err).To(Not(HaveOccurred()))
			})
			It("should have called RemoveCEIPTelemetryJob", func() {
				Expect(regionalClusterClient.RemoveCEIPTelemetryJobCallCount()).To(Equal(1))
			})
		})

		Context("When opt-ing in to CEIP on prod environment", func() {
			JustBeforeEach(func() {
				regionalClusterClient.IsPacificRegionalClusterReturns(false, nil)
				err = tkgClient.DoSetCEIPParticipation(regionalClusterClient, context, true, "true", "")
			})
			It("should not error", func() {
				Expect(err).To(Not(HaveOccurred()))
			})
			It("should have called AddCEIPTelemetryJob", func() {
				Expect(regionalClusterClient.AddCEIPTelemetryJobCallCount()).To(Equal(1))
			})
		})

		Context("When opt-ing in to CEIP on non-prod environment", func() {
			JustBeforeEach(func() {
				regionalClusterClient.IsPacificRegionalClusterReturns(false, nil)
				err = tkgClient.DoSetCEIPParticipation(regionalClusterClient, context, true, "true", "")
			})
			It("should not error", func() {
				Expect(err).To(Not(HaveOccurred()))
			})
			It("should have called AddCEIPTelemetryJob", func() {
				Expect(regionalClusterClient.AddCEIPTelemetryJobCallCount()).To(Equal(1))
			})
		})

		Context("When cluster is a pacific cluster", func() {
			JustBeforeEach(func() {
				regionalClusterClient.IsPacificRegionalClusterReturns(true, nil)
				err = tkgClient.DoSetCEIPParticipation(regionalClusterClient, context, true, "true", "")
			})
			It("should error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cannot change CEIP settings for a supervisor cluster which is on vSphere with Tanzu. Please change your CEIP settings within vSphere"))
			})
		})

		Context("When cluster is ipv6", func() {
			JustBeforeEach(func() {
				regionalClusterClient.IsPacificRegionalClusterReturns(false, nil)
				tkgClient.TKGConfigReaderWriter().Set(constants.ConfigVariableIPFamily, "ipv6")
				tkgClient.TKGConfigReaderWriter().Set(constants.TKGHTTPProxy, "fe80::")
				err = tkgClient.DoSetCEIPParticipation(regionalClusterClient, context, true, "true", "")
			})
			It("should add ::1 to noProxy", func() {
				_, _, _, _, _, _, _, noProxy := regionalClusterClient.AddCEIPTelemetryJobArgsForCall(0)
				Expect(noProxy).To(ContainSubstring("::1"))
			})
		})

		Context("When cluster is ipv4", func() {
			JustBeforeEach(func() {
				regionalClusterClient.IsPacificRegionalClusterReturns(false, nil)
				tkgClient.TKGConfigReaderWriter().Set(constants.ConfigVariableIPFamily, "ipv4")
				tkgClient.TKGConfigReaderWriter().Set(constants.TKGHTTPProxy, "1.2.3.4")
				err = tkgClient.DoSetCEIPParticipation(regionalClusterClient, context, true, "true", "")
			})
			It("should not add ::1 to noProxy", func() {
				_, _, _, _, _, _, _, noProxy := regionalClusterClient.AddCEIPTelemetryJobArgsForCall(0)
				Expect(noProxy).NotTo(ContainSubstring("::1"))
			})
		})
	})

	Describe("GetCEIP", func() {
		BeforeEach(func() {
		})

		Context("When cluster is opt'd in", func() {
			JustBeforeEach(func() {
				regionalClusterClient.IsPacificRegionalClusterReturns(false, nil)
				regionalClusterClient.HasCEIPTelemetryJobReturns(true, nil)
				ceipStatus, err = tkgClient.DoGetCEIPParticipation(regionalClusterClient, clusterName)
			})
			It("should not error", func() {
				Expect(err).To(Not(HaveOccurred()))
			})
			It("should have called HasCEIPTelemetryJob", func() {
				Expect(regionalClusterClient.HasCEIPTelemetryJobCallCount()).To(Equal(1))
			})
			It("should return correct status", func() {
				Expect(ceipStatus.CeipStatus).To(Equal(CeipOptInStatus))
				Expect(ceipStatus.ClusterName).To(Equal(clusterName))
			})
		})

		Context("When cluster is opt'd out", func() {
			JustBeforeEach(func() {
				regionalClusterClient.IsPacificRegionalClusterReturns(false, nil)
				regionalClusterClient.HasCEIPTelemetryJobReturns(false, nil)
				ceipStatus, err = tkgClient.DoGetCEIPParticipation(regionalClusterClient, clusterName)
			})
			It("should not error", func() {
				Expect(err).To(Not(HaveOccurred()))
			})
			It("should have called HasCEIPTelemetryJob", func() {
				Expect(regionalClusterClient.HasCEIPTelemetryJobCallCount()).To(Equal(1))
			})
			It("should return correct status", func() {
				Expect(ceipStatus.CeipStatus).To(Equal(CeipOptOutStatus))
				Expect(ceipStatus.ClusterName).To(Equal(clusterName))
			})
		})

		Context("When cluster is a pacific cluster", func() {
			JustBeforeEach(func() {
				regionalClusterClient.IsPacificRegionalClusterReturns(true, nil)
				regionalClusterClient.HasCEIPTelemetryJobReturns(false, nil)
				ceipStatus, err = tkgClient.DoGetCEIPParticipation(regionalClusterClient, clusterName)
			})
			It("should not error", func() {
				Expect(err).To(Not(HaveOccurred()))
			})
			It("should return correct status", func() {
				Expect(ceipStatus.CeipStatus).To(Equal(CeipPacificCluster))
				Expect(ceipStatus.ClusterName).To(Equal(clusterName))
			})
		})
	})
})
