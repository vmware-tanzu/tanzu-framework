// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package avi

import (
	"errors"
	"testing"

	"github.com/avinetworks/sdk/go/models"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	avi_mocks "github.com/vmware-tanzu-private/core/pkg/v1/tkg/avi/mocks"
	avi_models "github.com/vmware-tanzu-private/core/pkg/v1/tkg/web/server/models"
)

var controllerClient client
var (
	ctrl                             *gomock.Controller
	mockMiniCloudClient              *avi_mocks.MockMiniCloudClient
	mockMiniServiceEngineGroupClient *avi_mocks.MockMiniServiceEngineGroupClient
	mockMiniNetworkClient            *avi_mocks.MockMiniNetworkClient
)

// TestControllerClient serves as the entry point for the test suite
func TestControllerClient(t *testing.T) {
	RegisterFailHandler(Fail)
	ctrl = gomock.NewController(t)
	defer ctrl.Finish()
	RunSpecs(t, "AVI Client Suite")
}

var _ = Describe("AVI Client", func() {
	Describe("APIs", func() {
		BeforeSuite(func() {
			cp := &avi_models.AviControllerParams{
				Username: "admin",
				Password: "Admin!23",
				Host:     "10.186.38.103",
				Tenant:   "admin",
			}

			mockMiniCloudClient = avi_mocks.NewMockMiniCloudClient(ctrl)
			mockMiniServiceEngineGroupClient = avi_mocks.NewMockMiniServiceEngineGroupClient(ctrl)
			mockMiniNetworkClient = avi_mocks.NewMockMiniNetworkClient(ctrl)
			controllerClient = client{
				ControllerParams:   cp,
				Cloud:              mockMiniCloudClient,
				ServiceEngineGroup: mockMiniServiceEngineGroupClient,
				Network:            mockMiniNetworkClient,
			}
		})

		It("should return a valid Client object", func() {
			Expect(controllerClient).ToNot(BeNil())
		})

		It("should return an error", func() {
			mockMiniCloudClient.EXPECT().GetAll(gomock.Any()).Return(nil, errors.New("failed")).Times(1)
			clouds, err := controllerClient.GetClouds()
			Expect(err).To(HaveOccurred())
			Expect(clouds).To(BeEmpty())
		})

		It("should return a collection of Cloud", func() {
			aviClouds := make([]*models.Cloud, 0)
			aviClouds = append(aviClouds, &models.Cloud{
				LastModified:              new(string),
				ApicConfiguration:         &models.APICConfiguration{},
				ApicMode:                  new(bool),
				AutoscalePollingInterval:  new(int32),
				AwsConfiguration:          &models.AwsConfiguration{},
				AzureConfiguration:        &models.AzureConfiguration{},
				CloudstackConfiguration:   &models.CloudStackConfiguration{},
				CustomTags:                []*models.CustomTag{},
				DhcpEnabled:               new(bool),
				DNSProviderRef:            new(string),
				DNSResolutionOnSe:         new(bool),
				DockerConfiguration:       &models.DockerConfiguration{},
				EastWestDNSProviderRef:    new(string),
				EastWestIPAMProviderRef:   new(string),
				EnableVipOnAllInterfaces:  new(bool),
				EnableVipStaticRoutes:     new(bool),
				GcpConfiguration:          &models.GCPConfiguration{},
				Ip6AutocfgEnabled:         new(bool),
				IPAMProviderRef:           new(string),
				LicenseTier:               new(string),
				LicenseType:               new(string),
				LinuxserverConfiguration:  &models.LinuxServerConfiguration{},
				MesosConfiguration:        &models.MesosConfiguration{},
				Mtu:                       new(int32),
				Name:                      new(string),
				NsxConfiguration:          &models.NsxConfiguration{},
				NsxtConfiguration:         &models.NsxtConfiguration{},
				ObjNamePrefix:             new(string),
				OpenstackConfiguration:    &models.OpenStackConfiguration{},
				Oshiftk8sConfiguration:    &models.OShiftK8SConfiguration{},
				PreferStaticRoutes:        new(bool),
				ProxyConfiguration:        &models.ProxyConfiguration{},
				RancherConfiguration:      &models.RancherConfiguration{},
				SeGroupTemplateRef:        new(string),
				StateBasedDNSRegistration: new(bool),
				TenantRef:                 new(string),
				URL:                       new(string),
				UUID:                      new(string),
				VcaConfiguration:          &models.VCloudAirConfiguration{},
				VcenterConfiguration:      &models.VCenterConfiguration{},
				Vtype:                     new(string),
			})
			mockMiniCloudClient.EXPECT().GetAll(gomock.Any()).Return(aviClouds, nil).AnyTimes()
			clouds, err := controllerClient.GetClouds()
			Expect(err).ToNot(HaveOccurred())
			Expect(len(clouds)).ToNot(Equal(0))
		})

		It("should return an error", func() {
			mockMiniServiceEngineGroupClient.EXPECT().GetAll(gomock.Any()).Return(nil, errors.New("failed")).Times(1)
			segs, err := controllerClient.GetServiceEngineGroups()
			Expect(err).To(HaveOccurred())
			Expect(segs).To(BeEmpty())
		})

		It("should return a collection of ServiceEngineGroup", func() {
			aviServiceGroups := make([]*models.ServiceEngineGroup, 0)
			aviServiceGroups = append(aviServiceGroups, &models.ServiceEngineGroup{
				LastModified:                      new(string),
				AcceleratedNetworking:             new(bool),
				ActiveStandby:                     new(bool),
				AdditionalConfigMemory:            new(int32),
				AdvertiseBackendNetworks:          new(bool),
				AggressiveFailureDetection:        new(bool),
				Algo:                              new(string),
				AllowBurst:                        new(bool),
				AppCachePercent:                   new(int32),
				AppCacheThreshold:                 new(int32),
				AppLearningMemoryPercent:          new(int32),
				ArchiveShmLimit:                   new(int32),
				AsyncSsl:                          new(bool),
				AsyncSslThreads:                   new(int32),
				AutoRebalance:                     new(bool),
				AutoRebalanceCapacityPerSe:        []int64{},
				AutoRebalanceCriteria:             []string{},
				AutoRebalanceInterval:             new(int32),
				AutoRedistributeActiveStandbyLoad: new(bool),
				AvailabilityZoneRefs:              []string{},
				BgpStateUpdateInterval:            new(int32),
				BufferSe:                          new(int32),
				CloudRef:                          new(string),
				CompressIPRulesForEachNsSubnet:    new(bool),
				ConfigDebugsOnAllCores:            new(bool),
				ConnectionMemoryPercentage:        new(int32),
				CoreShmAppCache:                   new(bool),
				CoreShmAppLearning:                new(bool),
				CPUReserve:                        new(bool),
				CPUSocketAffinity:                 new(bool),
				CustomSecuritygroupsData:          []string{},
				CustomSecuritygroupsMgmt:          []string{},
				CustomTag:                         []*models.CustomTag{},
				DataNetworkID:                     new(string),
				DatascriptTimeout:                 new(int64),
				DedicatedDispatcherCore:           new(bool),
				Description:                       new(string),
				DisableAviSecuritygroups:          new(bool),
				DisableCsumOffloads:               new(bool),
				DisableFlowProbes:                 new(bool),
				DisableGro:                        new(bool),
				DisableSeMemoryCheck:              new(bool),
				DisableTso:                        new(bool),
				DiskPerSe:                         new(int32),
				DistributeLoadActiveStandby:       new(bool),
				DistributeQueues:                  new(bool),
				DistributeVnics:                   new(bool),
				DpAggressiveHbFrequency:           new(int32),
				DpAggressiveHbTimeoutCount:        new(int32),
				DpHbFrequency:                     new(int32),
				DpHbTimeoutCount:                  new(int32),
				EnableGratarpPermanent:            new(bool),
				EnableHsmPriming:                  new(bool),
				EnableMultiLb:                     new(bool),
				EnablePcapTxRing:                  new(bool),
				EnableRouting:                     new(bool),
				EnableVipOnAllInterfaces:          new(bool),
				EnableVMAC:                        new(bool),
				EphemeralPortrangeEnd:             new(int32),
				EphemeralPortrangeStart:           new(int32),
				ExtraConfigMultiplier:             new(float64),
				ExtraSharedConfigMemory:           new(int32),
				FloatingIntfIP:                    []*models.IPAddr{},
				FloatingIntfIPSe2:                 []*models.IPAddr{},
				FlowTableNewSynMaxEntries:         new(int32),
				FreeListSize:                      new(int32),
				GcpConfig:                         &models.GCPSeGroupConfig{},
				GratarpPermanentPeriodicity:       new(int32),
				HaMode:                            new(string),
				HandlePerPktAttack:                new(bool),
				HardwaresecuritymodulegroupRef:    new(string),
				HeapMinimumConfigMemory:           new(int32),
				HmOnStandby:                       new(bool),
				HostAttributeKey:                  new(string),
				HostAttributeValue:                new(string),
				HostGatewayMonitor:                new(bool),
				Hypervisor:                        new(string),
				IgnoreRttThreshold:                new(int32),
				IngressAccessData:                 new(string),
				IngressAccessMgmt:                 new(string),
				InstanceFlavor:                    new(string),
				InstanceFlavorInfo:                &models.CloudFlavor{},
				Iptables:                          []*models.IptableRuleSet{},
				Labels:                            []*models.KeyValue{},
				LeastLoadCoreSelection:            new(bool),
				LicenseTier:                       new(string),
				LicenseType:                       new(string),
				LogDisksz:                         new(int32),
				LogMallocFailure:                  new(bool),
				MaxConcurrentExternalHm:           new(int32),
				MaxCPUUsage:                       new(int32),
				MaxMemoryPerMempool:               new(int32),
				MaxNumSeDps:                       new(int32),
				MaxPublicIpsPerLb:                 new(int32),
				MaxQueuesPerVnic:                  new(int32),
				MaxRulesPerLb:                     new(int32),
				MaxScaleoutPerVs:                  new(int32),
				MaxSe:                             new(int32),
				MaxVsPerSe:                        new(int32),
				MemReserve:                        new(bool),
				MemoryForConfigUpdate:             new(int32),
				MemoryPerSe:                       new(int32),
				MgmtNetworkRef:                    new(string),
				MgmtSubnet:                        &models.IPAddrPrefix{},
				MinCPUUsage:                       new(int32),
				MinScaleoutPerVs:                  new(int32),
				MinSe:                             new(int32),
				MinimumConnectionMemory:           new(int32),
				MinimumRequiredConfigMemory:       new(int32),
				NLogStreamingThreads:              new(int32),
				Name:                              new(string),
				NatFlowTCPClosedTimeout:           new(int32),
				NatFlowTCPEstablishedTimeout:      new(int32),
				NatFlowTCPHalfClosedTimeout:       new(int32),
				NatFlowTCPHandshakeTimeout:        new(int32),
				NatFlowUDPNoresponseTimeout:       new(int32),
				NatFlowUDPResponseTimeout:         new(int32),
				NetlinkPollerThreads:              new(int32),
				NetlinkSockBufSize:                new(int32),
				NonSignificantLogThrottle:         new(int32),
				NumDispatcherCores:                new(int32),
				NumFlowCoresSumChangesToIgnore:    new(int32),
				OpenstackAvailabilityZone:         new(string),
				OpenstackAvailabilityZones:        []string{},
				OpenstackMgmtNetworkName:          new(string),
				OpenstackMgmtNetworkUUID:          new(string),
				OsReservedMemory:                  new(int32),
				PcapTxMode:                        new(string),
				PcapTxRingRdBalancingFactor:       new(int32),
				PerApp:                            new(bool),
				PerVsAdmissionControl:             new(bool),
				PlacementMode:                     new(string),
				RealtimeSeMetrics:                 &models.MetricsRealTimeUpdate{},
				RebootOnPanic:                     new(bool),
				RebootOnStop:                      new(bool),
				ResyncTimeInterval:                new(int32),
				SeBandwidthType:                   new(string),
				SeDelayedFlowDelete:               new(bool),
				SeDeprovisionDelay:                new(int32),
				SeDosProfile:                      &models.DosThresholdProfile{},
				SeDpHmDrops:                       new(int32),
				SeDpMaxHbVersion:                  new(int32),
				SeDpVnicQueueStallEventSleep:      new(int32),
				SeDpVnicQueueStallThreshold:       new(int32),
				SeDpVnicQueueStallTimeout:         new(int32),
				SeDpVnicRestartOnQueueStallCount:  new(int32),
				SeDpVnicStallSeRestartWindow:      new(int32),
				SeDpdkPmd:                         new(int32),
				SeFlowProbeRetries:                new(int32),
				SeFlowProbeRetryTimer:             new(int32),
				SeFlowProbeTimer:                  new(int32),
				SeGroupAnalyticsPolicy:            &models.SeGroupAnalyticsPolicy{},
				SeHyperthreadedMode:               new(string),
				SeIPEncapIpc:                      new(int32),
				SeIpcUDPPort:                      new(int32),
				SeKniBurstFactor:                  new(int32),
				SeL3EncapIpc:                      new(int32),
				SeLro:                             new(bool),
				SeMpRingRetryCount:                new(int32),
				SeMtu:                             new(int32),
				SeNamePrefix:                      new(string),
				SePcapLookahead:                   new(bool),
				SePcapPktCount:                    new(int32),
				SePcapPktSz:                       new(int32),
				SePcapQdiscBypass:                 new(bool),
				SePcapReinitFrequency:             new(int32),
				SePcapReinitThreshold:             new(int32),
				SeProbePort:                       new(int32),
				SeRemotePuntUDPPort:               new(int32),
				SeRlProp:                          &models.RateLimiterProperties{},
				SeRouting:                         new(bool),
				SeRumSamplingNavInterval:          new(int32),
				SeRumSamplingNavPercent:           new(int32),
				SeRumSamplingResInterval:          new(int32),
				SeRumSamplingResPercent:           new(int32),
				SeSbDedicatedCore:                 new(bool),
				SeSbThreads:                       new(int32),
				SeThreadMultiplier:                new(int32),
				SeTracertPortRange:                &models.PortRange{},
				SeTunnelMode:                      new(int32),
				SeTunnelUDPPort:                   new(int32),
				SeTxBatchSize:                     new(int32),
				SeTxqThreshold:                    new(int32),
				SeUDPEncapIpc:                     new(int32),
				SeUseDpdk:                         new(int32),
				SeVnicTxSwQueueFlushFrequency:     new(int32),
				SeVnicTxSwQueueSize:               new(int32),
				SeVsHbMaxPktsInBatch:              new(int32),
				SeVsHbMaxVsInPkt:                  new(int32),
				SelfSeElection:                    new(bool),
				ServiceIp6Subnets:                 []*models.IPAddrPrefix{},
				ServiceIPSubnets:                  []*models.IPAddrPrefix{},
				ShmMinimumConfigMemory:            new(int32),
				SignificantLogThrottle:            new(int32),
				SslPreprocessSniHostname:          new(bool),
				TenantRef:                         new(string),
				TransientSharedMemoryMax:          new(int32),
				UdfLogThrottle:                    new(int32),
				URL:                               new(string),
				UseHyperthreadedCores:             new(bool),
				UseObjsync:                        new(bool),
				UseStandardAlb:                    new(bool),
				UUID:                              new(string),
				VcenterClusters:                   &models.VcenterClusters{},
				VcenterDatastoreMode:              new(string),
				VcenterDatastores:                 []*models.VcenterDatastore{},
				VcenterDatastoresInclude:          new(bool),
				VcenterFolder:                     new(string),
				VcenterHosts:                      &models.VcenterHosts{},
				Vcenters:                          []*models.PlacementScopeConfig{},
				VcpusPerSe:                        new(int32),
				VipAsg:                            &models.VipAutoscaleGroup{},
				VsHostRedundancy:                  new(bool),
				VsScaleinTimeout:                  new(int32),
				VsScaleinTimeoutForUpgrade:        new(int32),
				VsScaleoutTimeout:                 new(int32),
				VsSeScaleoutAdditionalWaitTime:    new(int32),
				VsSeScaleoutReadyTimeout:          new(int32),
				VsSwitchoverTimeout:               new(int32),
				VssPlacement:                      &models.VssPlacement{},
				VssPlacementEnabled:               new(bool),
				WafLearningInterval:               new(int32),
				WafLearningMemory:                 new(int32),
				WafMempool:                        new(bool),
				WafMempoolSize:                    new(int32),
			})
			mockMiniServiceEngineGroupClient.EXPECT().GetAll(gomock.Any()).Return(aviServiceGroups, nil).AnyTimes()
			segs, err := controllerClient.GetClouds()
			Expect(err).ToNot(HaveOccurred())
			Expect(len(segs)).ToNot(Equal(0))
		})
	})

	It("should return a collection of Networks", func() {
		aviNetworks := make([]*models.Network, 0)
		aviNetworks = append(aviNetworks, &models.Network{
			LastModified:             new(string),
			Attrs:                    []*models.KeyValue{},
			CloudRef:                 new(string),
			ConfiguredSubnets:        []*models.Subnet{},
			DhcpEnabled:              new(bool),
			ExcludeDiscoveredSubnets: new(bool),
			Ip6AutocfgEnabled:        new(bool),
			Labels:                   []*models.KeyValue{},
			Name:                     new(string),
			SyncedFromSe:             new(bool),
			TenantRef:                new(string),
			URL:                      new(string),
			UUID:                     new(string),
			VcenterDvs:               new(bool),
			VimgrnwRef:               new(string),
			VrfContextRef:            new(string),
		})
		mockMiniNetworkClient.EXPECT().GetAll(gomock.Any()).Return(aviNetworks, nil).AnyTimes()
		segs, err := controllerClient.GetVipNetworks()
		Expect(err).ToNot(HaveOccurred())
		Expect(len(segs)).ToNot(Equal(0))
	})
})
