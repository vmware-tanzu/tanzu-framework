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

var _ = Describe("Cluster Class - IP Family Validation related test cases:", func() {
	var (
		cidrsIpv6Ipv4 = "2002::1234:abcd:ffff:c0a8:101/64,100.64.0.0/18"
		cidrsIpv4     = "100.96.0.0/12,100.64.0.0/16"
		cidrsIpv4Ipv6 = "100.64.0.0/18,2002::1234:abcd:ffff:c0a8:101/64"
	)
	Context("Cluster Class - IP Family related test cases", func() {
		Context("Test GetIPFamilyForGivenCIDRs for given CIDR ", func() {
			It("When isIPV6Primary is true but first cidr do not have ipv6, should get error:", func() {
				_, err := GetIPFamilyForGivenCIDRs(strings.Split(cidrsIpv4, ","), true)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf(errMessageIPv6EnabledCIDRHasNoIPv6, strings.Split(cidrsIpv4, ","))))
			})
			It("When isIPV6Primary is true but first cidr is ipv4 not ipv6, should get error:", func() {
				_, err := GetIPFamilyForGivenCIDRs(strings.Split(cidrsIpv4Ipv6, ","), true)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf(errMessageIPv6EnabledCIDRHasNoIPv6, strings.Split(cidrsIpv4Ipv6, ","))))
			})
			It("When isIPV6Primary is true, first cidr is ipv6, should expect ip family as DualStackPrimaryIPv6Family", func() {
				family, err := GetIPFamilyForGivenCIDRs(strings.Split(cidrsIpv6Ipv4, ","), true)
				Expect(err).To(BeNil())
				Expect(constants.DualStackPrimaryIPv6Family).To(Equal(family))
			})
			It("When isIPV6Primary is true, both cidr is ipv6, should expect IP Family as ipv6", func() {
				cidr := "2002::1234:abcd:ffff:c0a8:101/64,2002::1234:abcd:ffff:c0a8:101/24"
				family, err := GetIPFamilyForGivenCIDRs(strings.Split(cidr, ","), true)
				Expect(err).To(BeNil())
				Expect(constants.IPv6Family).To(Equal(family))
			})
			It("When isIPV6Primary is false, first cidr is ipv6, expect family as DualStackPrimaryIPv4Family", func() {
				family, err := GetIPFamilyForGivenCIDRs(strings.Split(cidrsIpv6Ipv4, ","), false)
				Expect(err).To(BeNil())
				Expect(constants.DualStackPrimaryIPv4Family).To(Equal(family))
			})
			It("When isIPV6Primary is false, second cidr is ipv6, expect family as DualStackPrimaryIPv4Family", func() {
				family, err := GetIPFamilyForGivenCIDRs(strings.Split(cidrsIpv4Ipv6, ","), false)
				Expect(err).To(BeNil())
				Expect(constants.DualStackPrimaryIPv4Family).To(Equal(family))
			})
			It("When incorrect cidr - first ipv6 value, should return error", func() {
				cidrsIpv6Incorrect := "inCorrect2002::1234:abcd:ffff:c0a8:101/64,2002::1234:abcd:ffff:c0a8:101/24"
				_, err := GetIPFamilyForGivenCIDRs(strings.Split(cidrsIpv6Incorrect, ","), false)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not parse CIDR"))
			})
			It("When incorrect cidr - second ipv6 value, should return error", func() {
				cidrsIpv6Incorrect := "2002::1234:abcd:ffff:c0a8:101/64,IN-CORREC-2002::1234:abcd:ffff:c0a8:101/24"
				_, err := GetIPFamilyForGivenCIDRs(strings.Split(cidrsIpv6Incorrect, ","), false)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not parse CIDR"))
			})
			It("incorrect cidr - first ipv4 value", func() {
				cidrsIncorrectFirstValue := "10055.64.0.0/18,100.64.0.0/18"
				_, err := GetIPFamilyForGivenCIDRs(strings.Split(cidrsIncorrectFirstValue, ","), false)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not parse CIDR"))
			})
			It("When incorrect cidr - second ipv4 value, should return error", func() {
				cidrsIncorrectSecondValue := "100.64.0.0/18,10044.64.0.0/18"
				_, err := GetIPFamilyForGivenCIDRs(strings.Split(cidrsIncorrectSecondValue, ","), false)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not parse CIDR"))
			})
			It("When correct cidr - both are ipv4, should get ipv4 as ip family", func() {
				family, err := GetIPFamilyForGivenCIDRs(strings.Split(cidrsIpv4, ","), false)
				Expect(err).To(BeNil())
				Expect(constants.IPv4Family).To(Equal(family))
			})
			It("When correct cidr - only one ipv4 value, ipv4 as ip family", func() {
				cidrSingleValue := "100.64.0.0/18"
				family, err := GetIPFamilyForGivenCIDRs(strings.Split(cidrSingleValue, ","), false)
				Expect(err).To(BeNil())
				Expect(constants.IPv4Family).To(Equal(family))
			})

			It("When correct cidrs - too many values - three ipv4 values, return error:", func() {
				cidrsThreeValues := "100.64.0.0/18,101.64.0.0/18,102.64.0.0/18"
				_, err := GetIPFamilyForGivenCIDRs(strings.Split(cidrsThreeValues, ","), false)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("too many CIDRs specified"))
			})
		})

		Context("String Array to String With Comma Separated Elements ", func() {
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
			It("When valid cidrs: both ipv4 values, get ipv4 as IP family:", func() {
				podcidr := cidrsIpv4
				servicecidr := cidrsIpv4
				isIPV6Primary := false
				family, err := GetIPFamilyForGivenClusterNetworkCIDRs(podcidr, servicecidr, isIPV6Primary)
				Expect(err).To(BeNil())
				Expect(constants.IPv4Family).To(Equal(family))
			})
			It("When valid cidrs: all are ipv4, isIPV6Primary is true, returns error ", func() {
				podcidr := cidrsIpv4
				servicecidr := cidrsIpv4
				isIPV6Primary := true
				_, err := GetIPFamilyForGivenClusterNetworkCIDRs(podcidr, servicecidr, isIPV6Primary)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("the isIPV6Primary: true, but the first value in CIDRs"))
			})
			It("When valid cidrs: ipv6 and ipv4,  isIPV6Primary is true, expect DualStackPrimaryIPv6Family", func() {
				podcidr := cidrsIpv6Ipv4
				servicecidr := cidrsIpv6Ipv4
				isIPV6Primary := true
				family, err := GetIPFamilyForGivenClusterNetworkCIDRs(podcidr, servicecidr, isIPV6Primary)
				Expect(err).To(BeNil())
				Expect(constants.DualStackPrimaryIPv6Family).To(Equal(family))
			})
			It("When valid cidrs: ipv6 and isIPV6Primary is false", func() {
				podcidr := cidrsIpv6Ipv4
				servicecidr := cidrsIpv6Ipv4
				isIPV6Primary := false
				family, err := GetIPFamilyForGivenClusterNetworkCIDRs(podcidr, servicecidr, isIPV6Primary)
				Expect(err).To(BeNil())
				Expect(constants.DualStackPrimaryIPv4Family).To(Equal(family))
			})
			It("When both are different cidrs: different familites, podCidr has ipv6 and ipv4, service cidr has only ipv4 type:  so return error", func() {
				podcidr := cidrsIpv6Ipv4
				servicecidr := cidrsIpv4
				isIPV6Primary := false
				_, err := GetIPFamilyForGivenClusterNetworkCIDRs(podcidr, servicecidr, isIPV6Primary)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("both are not same IP Families"))
			})
		})

		Context("Test cases for getProviderNameFromTopologyClassName  - validate value of spec.topology.class attribute ", func() {
			It("When valid value: aws provier, return aws as provider:", func() {
				cidr := "tkg-aws-default"
				provider, err := getProviderNameFromTopologyClassName(cidr)
				Expect(provider).To(Equal("aws"))
				Expect(err).To(BeNil())
			})
			It("When valid value: azure provier, return azure as provider:", func() {
				cidr := "tkg-azure-default"
				provider, err := getProviderNameFromTopologyClassName(cidr)
				Expect(provider).To(Equal("azure"))
				Expect(err).To(BeNil())
			})
			It("When valid value: Vsphere provier, return Vsphere as provider:", func() {
				cidr := "tkg-vsphere-default"
				provider, err := getProviderNameFromTopologyClassName(cidr)
				Expect(provider).To(Equal("vsphere"))
				Expect(err).To(BeNil())
			})
			It("When empty value, return error", func() {
				cidr := ""
				_, err := getProviderNameFromTopologyClassName(cidr)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(constants.TopologyClassIncorrectValueErrMsg))
			})
			It("When incorrect value, return error", func() {
				cidr := "tkg-not-default"
				_, err := getProviderNameFromTopologyClassName(cidr)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(constants.TopologyClassIncorrectValueErrMsg))
			})
		})
	})
})
