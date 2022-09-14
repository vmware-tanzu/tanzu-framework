// Copyright 2021-2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgauth

import (
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/url"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	fakehelper "github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/fakes/helper"
)

const (
	fakeIssuer  = "https://fakeissuer.com"
	fakeCluster = "fake-cluster"
	fakeCAData  = "fakeCAData"
)

var _ = Describe("Kubeconfig Tests", func() {
	var (
		err                      error
		endpoint                 string
		tlsserver                *ghttp.Server
		clustername              string
		issuer                   string
		issuerCA                 string
		conciergeIsClusterScoped bool
		servCert                 *x509.Certificate
	)

	const kubeconfig1Path = "../fakes/config/kubeconfig/config1.yaml"
	Describe("Get cluster-info from the cluster", func() {
		BeforeEach(func() {
			tlsserver = ghttp.NewTLSServer()
			servCert = tlsserver.HTTPTestServer.Certificate()
			endpoint = tlsserver.URL()
		})
		AfterEach(func() {
			tlsserver.Close()
		})
		Context("When the configMap 'cluster-info' is not present in kube-public namespace", func() {
			BeforeEach(func() {
				tlsserver.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/cluster-info"),
						ghttp.RespondWith(http.StatusNotFound, "not found"),
					),
				)
				_, err = GetClusterInfoFromCluster(endpoint, "cluster-info")
			})
			It("should return the error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("failed to get cluster-info"))
			})
		})
		Context("When the configMap 'cluster-info' is present but the returned format is incorrect ", func() {
			BeforeEach(func() {
				tlsserver.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/cluster-info"),
						ghttp.RespondWith(http.StatusOK, "fake-format-value"),
					),
				)
				_, err = GetClusterInfoFromCluster(endpoint, "cluster-info")
			})
			It("should return the error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("error parsing http response body"))
			})
		})
		Context("When the configMap 'cluster-info' is present in kube-public namespace", func() {
			var cluster *clientcmdapi.Cluster
			BeforeEach(func() {
				clusterInfo := fakehelper.GetFakeClusterInfo(endpoint, servCert)
				tlsserver.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/cluster-info"),
						ghttp.RespondWith(http.StatusOK, clusterInfo),
					),
				)
				cluster, err = GetClusterInfoFromCluster(endpoint, "cluster-info")
			})
			It("should return the cluster information", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(cluster.Server).Should(Equal(endpoint))
			})
		})
		Context("When a different ConfigMap from the kube-public namespace is used for discovery", func() {
			var cluster *clientcmdapi.Cluster
			BeforeEach(func() {
				clusterInfo := fakehelper.GetFakeClusterInfo(endpoint, servCert)
				tlsserver.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/vip-cluster-info"),
						ghttp.RespondWith(http.StatusOK, clusterInfo),
					),
				)
				cluster, err = GetClusterInfoFromCluster(endpoint, "vip-cluster-info")
			})
			It("should return the cluster information", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(cluster.Server).Should(Equal(endpoint))
			})
		})
	})

	Describe("Get pinniped-info from the cluster", func() {
		BeforeEach(func() {
			tlsserver = ghttp.NewTLSServer()
			servCert = tlsserver.HTTPTestServer.Certificate()
			endpoint = tlsserver.URL()
		})
		AfterEach(func() {
			tlsserver.Close()
		})

		Context("When the configMap 'pinniped-info' is not present in kube-public namespace", func() {
			var cluster clientcmdapi.Cluster
			var gotPinnipedInfo *PinnipedConfigMapInfo
			BeforeEach(func() {
				tlsserver.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/pinniped-info"),
						ghttp.RespondWith(http.StatusNotFound, "not found"),
					),
				)
				cluster.Server = endpoint
				certBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: servCert.Raw})
				cluster.CertificateAuthorityData = certBytes
				gotPinnipedInfo, err = GetPinnipedInfoFromCluster(&cluster, nil)
			})
			It("should not return an error", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(gotPinnipedInfo).Should(BeNil())
			})
		})
		Context("When the configMap 'pinniped-info' is present but the returned format is incorrect", func() {
			var cluster clientcmdapi.Cluster
			BeforeEach(func() {
				tlsserver.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/pinniped-info"),
						ghttp.RespondWith(http.StatusOK, "fake-format-value"),
					),
				)
				cluster.Server = endpoint
				certBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: servCert.Raw})
				cluster.CertificateAuthorityData = certBytes
				_, err = GetPinnipedInfoFromCluster(&cluster, nil)
			})
			It("should return the pinniped-info successfully", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("error parsing http response body"))
			})
		})
		Context("When the configMap 'pinniped-info' is present in kube-public namespace", func() {
			var cluster clientcmdapi.Cluster
			var gotPinnipedInfo *PinnipedConfigMapInfo
			BeforeEach(func() {
				clustername = fakeCluster
				issuer = fakeIssuer
				issuerCA = fakeCAData
				conciergeIsClusterScoped = false
				pinnipedInfo := fakehelper.GetFakePinnipedInfo(fakehelper.PinnipedInfo{
					ClusterName:              clustername,
					Issuer:                   issuer,
					IssuerCABundleData:       issuerCA,
					ConciergeIsClusterScoped: conciergeIsClusterScoped})
				tlsserver.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/pinniped-info"),
						ghttp.RespondWith(http.StatusOK, pinnipedInfo),
					),
				)
				cluster.Server = endpoint
				certBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: servCert.Raw})
				cluster.CertificateAuthorityData = certBytes
				gotPinnipedInfo, err = GetPinnipedInfoFromCluster(&cluster, nil)
			})
			It("should return the pinniped-info successfully", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(gotPinnipedInfo.Data.ClusterName).Should(Equal(clustername))
				Expect(gotPinnipedInfo.Data.Issuer).Should(Equal(issuer))
				Expect(gotPinnipedInfo.Data.IssuerCABundle).Should(Equal(issuerCA))
				Expect(gotPinnipedInfo.Data.ConciergeIsClusterScoped).Should(Equal(conciergeIsClusterScoped))
			})
		})
		Context("When a different port is used for discovery of 'pinniped-info'", func() {
			var cluster clientcmdapi.Cluster
			var gotPinnipedInfo *PinnipedConfigMapInfo
			var discoveryTLSServer *ghttp.Server
			BeforeEach(func() {
				// The second TLS server mimics the different endpoints for
				// kube-apiserver and discovery.
				discoveryTLSServer = ghttp.NewTLSServer()
				discoveryEndpoint := discoveryTLSServer.URL()
				// URL is valid, ports are expected to fit in 16 bits, so we're
				// skipping a bunch of error handling.
				u, _ := url.Parse(discoveryEndpoint)
				discoveryPort64, _ := strconv.ParseInt(u.Port(), 10, 64)
				discoveryPort := int(discoveryPort64)

				clustername = fakeCluster
				issuer = fakeIssuer
				issuerCA = fakeCAData
				conciergeIsClusterScoped = false
				pinnipedInfo := fakehelper.GetFakePinnipedInfo(fakehelper.PinnipedInfo{
					ClusterName:              clustername,
					Issuer:                   issuer,
					IssuerCABundleData:       issuerCA,
					ConciergeIsClusterScoped: conciergeIsClusterScoped})
				discoveryTLSServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/pinniped-info"),
						ghttp.RespondWith(http.StatusOK, pinnipedInfo),
					),
				)
				cluster.Server = endpoint
				certBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: servCert.Raw})
				cluster.CertificateAuthorityData = certBytes
				gotPinnipedInfo, err = GetPinnipedInfoFromCluster(&cluster, &discoveryPort)
			})
			It("should return the pinniped-info successfully", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(gotPinnipedInfo.Data.ClusterName).Should(Equal(clustername))
				Expect(gotPinnipedInfo.Data.Issuer).Should(Equal(issuer))
				Expect(gotPinnipedInfo.Data.IssuerCABundle).Should(Equal(issuerCA))
				Expect(gotPinnipedInfo.Data.ConciergeIsClusterScoped).Should(Equal(conciergeIsClusterScoped))
			})
		})
		Context("When the concierge endpoint is distinct from the cluster endpoint", func() {
			var cluster clientcmdapi.Cluster
			var gotPinnipedInfo *PinnipedConfigMapInfo
			var conciergeEndpoint string
			BeforeEach(func() {
				clustername = fakeCluster
				issuer = fakeIssuer
				issuerCA = fakeCAData
				conciergeEndpoint = "my-favourite-concierge.com"
				conciergeIsClusterScoped = false
				pinnipedInfo := fakehelper.GetFakePinnipedInfo(fakehelper.PinnipedInfo{
					ClusterName:              clustername,
					Issuer:                   issuer,
					IssuerCABundleData:       issuerCA,
					ConciergeEndpoint:        conciergeEndpoint,
					ConciergeIsClusterScoped: conciergeIsClusterScoped})
				tlsserver.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/pinniped-info"),
						ghttp.RespondWith(http.StatusOK, pinnipedInfo),
					),
				)
				cluster.Server = endpoint
				certBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: servCert.Raw})
				cluster.CertificateAuthorityData = certBytes
				gotPinnipedInfo, err = GetPinnipedInfoFromCluster(&cluster, nil)
			})
			It("should return the pinniped-info successfully", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(gotPinnipedInfo.Data.ClusterName).Should(Equal(clustername))
				Expect(gotPinnipedInfo.Data.Issuer).Should(Equal(issuer))
				Expect(gotPinnipedInfo.Data.IssuerCABundle).Should(Equal(issuerCA))
				Expect(gotPinnipedInfo.Data.ConciergeIsClusterScoped).Should(Equal(conciergeIsClusterScoped))
				Expect(gotPinnipedInfo.Data.ConciergeEndpoint).Should(Equal(conciergeEndpoint))
			})
		})
	})
})
