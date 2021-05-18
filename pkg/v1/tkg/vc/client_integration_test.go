// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package vc_test

import (
	"context"
	"net/url"
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/vc"
)

var _ = Describe("VC Client Integration Test", func() {
	var (
		vcHost     = os.Getenv(constants.ConfigVariableVsphereServer)
		vcPassword = os.Getenv(constants.ConfigVariableVspherePassword)
		vcUsername = os.Getenv(constants.ConfigVariableVsphereUsername)
	)

	BeforeEach(func() {
		if os.Getenv("RUN_VC_INTEG_TEST") != "1" {
			Skip("Skip integration tests for vc client ")
		}
	})

	Describe("New vc Client", func() {
		var (
			err        error
			thumbprint string
			vcClient   vc.Client
			insecure   bool
		)

		JustBeforeEach(func() {
			host := strings.TrimSpace(vcHost)
			if !strings.HasPrefix(host, "http") {
				host = "https://" + host
			}
			vcURL, perr := url.Parse(host)
			Expect(perr).ToNot(HaveOccurred())
			vcURL.Path = "/sdk"
			vcClient, err = vc.NewClient(vcURL, thumbprint, insecure)
		})

		Context("When the user sets VSPHERE_INSECURE environment variable", func() {
			BeforeEach(func() {
				insecure = true
			})
			It("should create the vc client and login to the vSphere successfully", func() {
				Expect(err).ToNot(HaveOccurred())
				_, err = vcClient.Login(context.TODO(), vcUsername, vcPassword)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("When the vc connection is the secured by wrong thumbprint", func() {
			BeforeEach(func() {
				insecure = false
				thumbprint = "wrong-thumbprint"
			})
			It("should fail to login to the vSphere", func() {
				Expect(err).To(HaveOccurred())
			})
		})

		Context("When the vc connection is the secured by correct thumbprint", func() {
			BeforeEach(func() {
				insecure = false
				thumbprint, err = vc.GetVCThumbprint(vcHost)
				Expect(err).ToNot(HaveOccurred())
			})
			It("should create the vc client and login to the vSphere successfully", func() {
				Expect(err).ToNot(HaveOccurred())
				_, err = vcClient.Login(context.TODO(), vcUsername, vcPassword)
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
