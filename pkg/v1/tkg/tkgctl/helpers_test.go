// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
)

var _ = Describe("Cluster Class - IP Family Validation related test cases: ", func() {
	var cidrsIpv6Ipv4 string
	var cidrsIpv4 string
	var cidrsIpv4Ipv6 string
	BeforeEach(func() {
		cidrsIpv6Ipv4 = "2002::1234:abcd:ffff:c0a8:101/64,100.64.0.0/18"
		cidrsIpv4 = "100.96.0.0/12,100.64.0.0/16"
		cidrsIpv4Ipv6 = "100.64.0.0/18,2002::1234:abcd:ffff:c0a8:101/64"
	})
	Context("GetIPFamilyForGivenCIDRs related test cases", func() {
		var isIPV6Primary bool
		When("When isIPV6Primary is true ", func() {
			BeforeEach(func() {
				isIPV6Primary = true
			})
			It("first cidr do not have ipv6, should get error:", func() {
				_, err := GetIPFamilyForGivenCIDRs(strings.Split(cidrsIpv4, ","), isIPV6Primary)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf(errMessageIPv6EnabledCIDRHasNoIPv6, strings.Split(cidrsIpv4, ","))))
			})
			It("first cidr is ipv4 not ipv6, should get error:", func() {
				_, err := GetIPFamilyForGivenCIDRs(strings.Split(cidrsIpv4Ipv6, ","), isIPV6Primary)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf(errMessageIPv6EnabledCIDRHasNoIPv6, strings.Split(cidrsIpv4Ipv6, ","))))
			})
			It("first cidr is ipv6, should expect ip family as DualStackPrimaryIPv6Family", func() {
				family, err := GetIPFamilyForGivenCIDRs(strings.Split(cidrsIpv6Ipv4, ","), isIPV6Primary)
				Expect(err).To(BeNil())
				Expect(constants.DualStackPrimaryIPv6Family).To(Equal(family))
			})
			It("both cidr is ipv6, should expect IP Family as ipv6", func() {
				cidr := "2002::1234:abcd:ffff:c0a8:101/64,2002::1234:abcd:ffff:c0a8:101/24"
				family, err := GetIPFamilyForGivenCIDRs(strings.Split(cidr, ","), isIPV6Primary)
				Expect(err).To(BeNil())
				Expect(constants.IPv6Family).To(Equal(family))
			})
		})

		When("When isIPV6Primary is false", func() {
			BeforeEach(func() {
				isIPV6Primary = false
			})

			It("first cidr is ipv6, expect family as DualStackPrimaryIPv4Family", func() {
				family, err := GetIPFamilyForGivenCIDRs(strings.Split(cidrsIpv6Ipv4, ","), isIPV6Primary)
				Expect(err).To(BeNil())
				Expect(constants.DualStackPrimaryIPv4Family).To(Equal(family))
			})
			It("second cidr is ipv6, expect family as DualStackPrimaryIPv4Family", func() {
				family, err := GetIPFamilyForGivenCIDRs(strings.Split(cidrsIpv4Ipv6, ","), isIPV6Primary)
				Expect(err).To(BeNil())
				Expect(constants.DualStackPrimaryIPv4Family).To(Equal(family))
			})
			It("first ipv6 value, should return error", func() {
				cidrsIpv6Incorrect := "inCorrect2002::1234:abcd:ffff:c0a8:101/64,2002::1234:abcd:ffff:c0a8:101/24"
				_, err := GetIPFamilyForGivenCIDRs(strings.Split(cidrsIpv6Incorrect, ","), isIPV6Primary)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not parse CIDR"))
			})
			It("second ipv6 value, should return error", func() {
				cidrsIpv6Incorrect := "2002::1234:abcd:ffff:c0a8:101/64,IN-CORREC-2002::1234:abcd:ffff:c0a8:101/24"
				_, err := GetIPFamilyForGivenCIDRs(strings.Split(cidrsIpv6Incorrect, ","), isIPV6Primary)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not parse CIDR"))
			})
			It("incorrect cidr - first ipv4 value", func() {
				cidrsIncorrectFirstValue := "10055.64.0.0/18,100.64.0.0/18"
				_, err := GetIPFamilyForGivenCIDRs(strings.Split(cidrsIncorrectFirstValue, ","), isIPV6Primary)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not parse CIDR"))
			})
			It("When incorrect cidr - second ipv4 value, should return error", func() {
				cidrsIncorrectSecondValue := "100.64.0.0/18,10044.64.0.0/18"
				_, err := GetIPFamilyForGivenCIDRs(strings.Split(cidrsIncorrectSecondValue, ","), isIPV6Primary)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not parse CIDR"))
			})
			It("When correct cidr - both are ipv4, should get ipv4 as ip family", func() {
				family, err := GetIPFamilyForGivenCIDRs(strings.Split(cidrsIpv4, ","), isIPV6Primary)
				Expect(err).To(BeNil())
				Expect(constants.IPv4Family).To(Equal(family))
			})
			It("When correct cidr - only one ipv4 value, ipv4 as ip family", func() {
				cidrSingleValue := "100.64.0.0/18"
				family, err := GetIPFamilyForGivenCIDRs(strings.Split(cidrSingleValue, ","), isIPV6Primary)
				Expect(err).To(BeNil())
				Expect(constants.IPv4Family).To(Equal(family))
			})
			It("When correct cidrs - too many values - three ipv4 values, return error:", func() {
				cidrsThreeValues := "100.64.0.0/18,101.64.0.0/18,102.64.0.0/18"
				_, err := GetIPFamilyForGivenCIDRs(strings.Split(cidrsThreeValues, ","), isIPV6Primary)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("too many CIDRs specified"))
			})
		})
	})

	Context("stringArrayToStringWithCommaSeparatedElements test cases ", func() {
		It("When CIDR values as input, get array of CIDR values:", func() {
			cidr := "[100.96.0.0/12 100.64.0.0/16]"
			output := stringArrayToStringWithCommaSeparatedElements(cidr)
			Expect(output).To(Equal(cidrsIpv4))
		})
		It("When string values, get array of string values:", func() {
			cidr := "[str1 str2]"
			output := stringArrayToStringWithCommaSeparatedElements(cidr)
			Expect(output).To(Equal("str1,str2"))
		})
	})

	Context("Test GetIPFamilyForGivenClusterNetworkCIDRs for given clusterNetwork's pods CIDRs and service CIDRs", func() {
		var isIPV6Primary bool
		When("When isIPV6Primary is true:", func() {
			BeforeEach(func() {
				isIPV6Primary = true
			})

			It("all are ipv4, returns error ", func() {
				podCIDR := cidrsIpv4
				serviceCIDR := cidrsIpv4
				_, err := GetIPFamilyForGivenClusterNetworkCIDRs(podCIDR, serviceCIDR, isIPV6Primary)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("the isIPV6Primary: true, but the first value in CIDRs"))
			})
			It("ipv6 and ipv4, expect DualStackPrimaryIPv6Family", func() {
				podCIDR := cidrsIpv6Ipv4
				serviceCIDR := cidrsIpv6Ipv4
				family, err := GetIPFamilyForGivenClusterNetworkCIDRs(podCIDR, serviceCIDR, isIPV6Primary)
				Expect(err).To(BeNil())
				Expect(constants.DualStackPrimaryIPv6Family).To(Equal(family))
			})
		})

		When("When isIPV6Primary is false:", func() {
			BeforeEach(func() {
				isIPV6Primary = false
			})
			It("both ipv4 values, get ipv4 as IP family:", func() {
				podCIDR := cidrsIpv4
				serviceCIDR := cidrsIpv4
				family, err := GetIPFamilyForGivenClusterNetworkCIDRs(podCIDR, serviceCIDR, isIPV6Primary)
				Expect(err).To(BeNil())
				Expect(constants.IPv4Family).To(Equal(family))
			})
			It("both are ipv6,ipv4, should return ipv4,ipv6 ", func() {
				podCIDR := cidrsIpv6Ipv4
				serviceCIDR := cidrsIpv6Ipv4
				family, err := GetIPFamilyForGivenClusterNetworkCIDRs(podCIDR, serviceCIDR, isIPV6Primary)
				Expect(err).To(BeNil())
				Expect(constants.DualStackPrimaryIPv4Family).To(Equal(family))
			})
			It("different IP Families, podCIDR has ipv6 and ipv4, service cidr has only ipv4 type:  so return error", func() {
				podCIDR := cidrsIpv6Ipv4
				serviceCIDR := cidrsIpv4
				_, err := GetIPFamilyForGivenClusterNetworkCIDRs(podCIDR, serviceCIDR, isIPV6Primary)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("both are not same IP Families"))
			})
		})
	})

	Context("Test cases for getProviderNameFromTopologyClassName  - validate value of spec.topology.class attribute ", func() {
		When("When correct value is provided:", func() {
			It("aws provider, return aws as provider:", func() {
				cidr := "tkg-aws-default"
				provider, err := getProviderNameFromTopologyClassName(cidr)
				Expect(provider).To(Equal("aws"))
				Expect(err).To(BeNil())
			})
			It("azure provider, return azure as provider:", func() {
				cidr := "tkg-azure-default"
				provider, err := getProviderNameFromTopologyClassName(cidr)
				Expect(provider).To(Equal("azure"))
				Expect(err).To(BeNil())
			})
			It("Vsphere provider, return Vsphere as provider:", func() {
				cidr := "tkg-vsphere-default"
				provider, err := getProviderNameFromTopologyClassName(cidr)
				Expect(provider).To(Equal("vsphere"))
				Expect(err).To(BeNil())
			})
		})
		When("When incorrect or empty value provided:", func() {
			It("empty value, return error", func() {
				cidr := ""
				_, err := getProviderNameFromTopologyClassName(cidr)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(constants.TopologyClassIncorrectValueErrMsg))
			})
			It("incorrect value, return error", func() {
				cidr := "tkg-not-default"
				_, err := getProviderNameFromTopologyClassName(cidr)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(constants.TopologyClassIncorrectValueErrMsg))
			})
		})
	})
})
