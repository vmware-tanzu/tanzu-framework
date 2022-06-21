// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
package wcp_test

import (
	"crypto/tls"
	"errors"
	"net/http"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/auth/wcp"
)

func TestWCPAuth(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "WCP Auth Suite")
}

type errorRoundTripper struct{}

func (e *errorRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("an error occurred")
}

var _ = Describe("Tests for vSphere Supervisor discovery", func() {
	var endpoint string
	var httpClient *http.Client
	var tlsServer *ghttp.Server
	BeforeEach(func() {
		tlsServer = ghttp.NewTLSServer()
		endpoint = tlsServer.URL()
		httpClient = &http.Client{
			Transport: &http.Transport{
				// #nosec
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
	})
	Context("When the given endpoint exposes a login banner endpoint", func() {
		It("returns true", func() {
			tlsServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/wcp/loginbanner"),
					ghttp.RespondWith(http.StatusOK, "Hello World! This is a login banner."),
				),
			)
			result, err := wcp.IsVSphereSupervisor(endpoint, httpClient)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeTrue())
		})
	})
	Context("When the given endpoint does not expose a login banner endpoint", func() {
		It("returns false", func() {
			tlsServer.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/wcp/loginbanner"),
					ghttp.RespondWith(http.StatusNotFound, "I'm a 404"),
				),
			)
			result, err := wcp.IsVSphereSupervisor(endpoint, httpClient)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeFalse())
		})
	})
	Context("When an error occurs contacting the endpoint", func() {
		It("returns an error", func() {
			httpClient = &http.Client{
				Transport: &errorRoundTripper{},
			}
			_, err := wcp.IsVSphereSupervisor(endpoint, httpClient)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("an error occurred"))
		})
	})
})
