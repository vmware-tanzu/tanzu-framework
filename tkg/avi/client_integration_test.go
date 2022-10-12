// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package avi

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	avi_models "github.com/vmware-tanzu/tanzu-framework/tkg/web/server/models"
)

var realClient Client

// TestClient tests the methods of a Client object
func TestRealClient(t *testing.T) {
	RegisterFailHandler(Fail)
	realClient = New()
	RunSpecs(t, "AVI Integration Test")
}

var _ = Describe("AVI Real Client", func() {

	BeforeEach(func() {
		Skip("Skip integration tests for AVI client ")
	})

	Describe("Verify credentials", func() {
		Context("When incorrect credentials provided", func() {
			It("should fail", func() {
				cp := &avi_models.AviControllerParams{
					Username: "admin",
					Password: "CHANGE_BEFORE_USE",
					Host:     "10.92.195.49",
					Tenant:   "admin",
				}

				authend, err := realClient.VerifyAccount(cp)
				Expect(err).To(HaveOccurred())
				Expect(authend).To(BeFalse())
			})
		})

		Context("When correct credentials provided", func() {
			It("should succeed", func() {
				cp := &avi_models.AviControllerParams{
					Username: "admin",
					Password: "CHANGE_BEFORE_USE",
					Host:     "10.92.195.49",
					Tenant:   "admin",
				}

				authend, err := realClient.VerifyAccount(cp)
				Expect(err).ToNot(HaveOccurred())
				Expect(authend).To(BeTrue())
			})
		})

		Context("When correct credentials inand correct CA data provided", func() {
			It("should fail", func() {
				cp := &avi_models.AviControllerParams{
					Username: "admin",
					Password: "CHANGE_BEFORE_USE",
					Host:     "10.92.195.49",
					Tenant:   "admin",
					CAData: `-----BEGIN CERTIFICATE-----
MIICxjCCAa6gAwIBAgITXtXf4r9uE4SvqdBVzxII1DF/0zANBgkqhkiG9w0BAQsF
ADATMREwDwYDVQQDDAhlMmUtdGVzdDAeFw0yMTAyMjUyMTIxMTRaFw0yMjAyMjUy
MTIxMTRaMBMxETAPBgNVBAMMCGUyZS10ZXN0MIIBIjANBgkqhkiG9w0BAQEFAAOC
AQ8AMIIBCgKCAQEAreSdOowJ6cNwhPGz1juzNQlTv6ky1bwpg0hrhOUBoPZpBrFp
XbfIVOXabXOAEHzmsM9QjEk1ly+IcgC/zlAz8ej3y2Ww5n6ApdpYgNIsM/27Z8Cs
l/3w1pS5fTYxQiSEIOm4f9C6GAgDYx3TNRA8K1DCqR5OlvkR47sTrP8wCGl9UloK
HuS3ooHq/EctfbmbgPBC5kOh7zmUlRij2dbHElwTaQo+M0rdLkyfsQV2age362Ud
a2NnlMA2TXarXt9HVIdptu88BEWx0hcFopG+6Wb7Rv7ual2lKP7puksSIe7BlyVI
mkNiaSEGNZjbPcZuvCvS+jKgcJCDmUnLMZ+YmwIDAQABoxMwETAPBgNVHREECDAG
hwQKwSerMA0GCSqGSIb3DQEBCwUAA4IBAQA1OC8V84j7bghh8VnJgAUdcScf7nEw
Mov7lBeTgoFGL5oMSCcwlYZXZidOfXfkxmaM0jGxu3DEGgp8XAk0MU8+YtKivG8e
FRAjDPLaOja0FBk52EvgGxDTCKafLnXOMB8/MHDe3gQw5NoJv+wEdU8wnmubgTxR
U2pTZ8EsCsv1l+neJ3oZiPsYhVgV4rw+uM7I5744JTOQCukZsBgoFVppfodjXWU1
SEEDY/GmvD0f1Q7i1Sewxoiuxv1e/ZI6oTQEQ6w+xSYlaV4wcCIiN3nJg77VuCKr
S7PlMtE24mmPhGhgoZ1kK8dCsXz66WcvPX7HMzDz041SgvbBQVKLQZ/p
-----END CERTIFICATE-----
`,
				}

				authend, err := realClient.VerifyAccount(cp)
				Expect(err).To(HaveOccurred())
				Expect(authend).To(BeFalse())
			})
		})

		Context("When correct credentials and CA data provided", func() {
			It("should succeed", func() {
				cp := &avi_models.AviControllerParams{
					Username: "admin",
					Password: "CHANGE_BEFORE_USE",
					Host:     "10.92.195.49",
					Tenant:   "admin",
					CAData: `-----BEGIN CERTIFICATE-----
MIICvzCCAaegAwIBAgIUENYz/IDqKKPkI6XWROAfUTv7lpgwDQYJKoZIhvcNAQEL
BQAwDzENMAsGA1UEAwwEdGVzdDAeFw0yMTA1MjEwMzUxMTJaFw0yMjA1MjEwMzUx
MTJaMA8xDTALBgNVBAMMBHRlc3QwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEK
AoIBAQDEZQGJjCqy6ISLiIkr8OmiRBLSzQ8+fXRfzs/0UFMy47m7Bsdc+H8H+XRU
AsPdawJ0SaTzNFFas5URqjKO1iN6GgB+L1ivn8BAJoOGHeqcf684rPFtoSOfIvDi
FXiX0uJOC8czkG/iYSxmy0WWic8AOTwyINt0Ospjg/wwK1KZRJVB75lgQfkd3zT9
iTeUMlG/rk+UWC2+OHImoL37sbg7c2hVH14QNoyb0T45TNpE1q2k6+8p104WHxbk
xbRDSIgfdupTPSjZIi+Mq5YbMCGsMCxm3uI12dh0kIAhNHC48/1IB69CB2Ks4qQb
kX1w2WoWINVl0hajFT2qZEwqRopJAgMBAAGjEzARMA8GA1UdEQQIMAaHBApcwzEw
DQYJKoZIhvcNAQELBQADggEBAMNj3piaZqawx7kBGv8Vew6LXFgVHO/zoX9d+/x0
sPgMqgX3rMs+RNtPT7bJvsYtrMF0KeVujv7Opx/bQfH38DCt7IhGRt+q+C+3mMV8
eao8j94vw3dkQ6476pJG6HaGqN0l0ADWZIpDpwWbgauzrnhMCUVDU5FAVy4xFt3u
dsDJ0l3mDTDDv1O3VFxB/CaQCfVT+gGmoK7xd+cFVRdJF3MCtqGK4UwaIjejHJk0
vejPF4Eq7ue7B2NJREhAJjyeM22AS7EYfUVHOJzDdLWEFF3HGFJaO2LX0TnkSAts
NZl6xLnQEOlZzdx0Eq2ElNgPwxMRfrysU8Q+MusiMCRxMeE=
-----END CERTIFICATE-----`,
				}

				authend, err := realClient.VerifyAccount(cp)
				Expect(err).ToNot(HaveOccurred())
				Expect(authend).To(BeTrue())
			})
		})
	})

	Describe("APIs", func() {
		BeforeEach(func() {
			cp := &avi_models.AviControllerParams{
				Username: "admin",
				Password: "CHANGE_BEFORE_USE",
				Host:     "10.92.195.49",
				Tenant:   "admin",
				CAData: `-----BEGIN CERTIFICATE-----
MIICvzCCAaegAwIBAgIUENYz/IDqKKPkI6XWROAfUTv7lpgwDQYJKoZIhvcNAQEL
BQAwDzENMAsGA1UEAwwEdGVzdDAeFw0yMTA1MjEwMzUxMTJaFw0yMjA1MjEwMzUx
MTJaMA8xDTALBgNVBAMMBHRlc3QwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEK
AoIBAQDEZQGJjCqy6ISLiIkr8OmiRBLSzQ8+fXRfzs/0UFMy47m7Bsdc+H8H+XRU
AsPdawJ0SaTzNFFas5URqjKO1iN6GgB+L1ivn8BAJoOGHeqcf684rPFtoSOfIvDi
FXiX0uJOC8czkG/iYSxmy0WWic8AOTwyINt0Ospjg/wwK1KZRJVB75lgQfkd3zT9
iTeUMlG/rk+UWC2+OHImoL37sbg7c2hVH14QNoyb0T45TNpE1q2k6+8p104WHxbk
xbRDSIgfdupTPSjZIi+Mq5YbMCGsMCxm3uI12dh0kIAhNHC48/1IB69CB2Ks4qQb
kX1w2WoWINVl0hajFT2qZEwqRopJAgMBAAGjEzARMA8GA1UdEQQIMAaHBApcwzEw
DQYJKoZIhvcNAQELBQADggEBAMNj3piaZqawx7kBGv8Vew6LXFgVHO/zoX9d+/x0
sPgMqgX3rMs+RNtPT7bJvsYtrMF0KeVujv7Opx/bQfH38DCt7IhGRt+q+C+3mMV8
eao8j94vw3dkQ6476pJG6HaGqN0l0ADWZIpDpwWbgauzrnhMCUVDU5FAVy4xFt3u
dsDJ0l3mDTDDv1O3VFxB/CaQCfVT+gGmoK7xd+cFVRdJF3MCtqGK4UwaIjejHJk0
vejPF4Eq7ue7B2NJREhAJjyeM22AS7EYfUVHOJzDdLWEFF3HGFJaO2LX0TnkSAts
NZl6xLnQEOlZzdx0Eq2ElNgPwxMRfrysU8Q+MusiMCRxMeE=
-----END CERTIFICATE-----`,
			}

			_, err := realClient.VerifyAccount(cp)

			if err != nil {
				panic(err)
			}
		})

		It("should return a valid Client object", func() {
			Expect(realClient).ToNot(BeNil())
		})

		It("should return a collection of Cloud", func() {
			clouds, err := realClient.GetClouds()
			Expect(err).ToNot(HaveOccurred())
			Expect(len(clouds)).ToNot(Equal(0))
		})

		It("should return a collection of Service Engine Group", func() {
			segs, err := realClient.GetServiceEngineGroups()
			Expect(err).ToNot(HaveOccurred())
			Expect(len(segs)).ToNot(Equal(0))
		})

		It("should return a collection of Vip networks", func() {
			vipNets, err := realClient.GetVipNetworks()
			Expect(err).ToNot(HaveOccurred())
			Expect(len(vipNets)).ToNot(Equal(0))
		})

	})

})
