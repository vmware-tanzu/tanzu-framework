// Copyright 2021-2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package utils_test

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

	fakehelper "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes/helper"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/utils"
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
				_, err = utils.GetClusterInfoFromCluster(endpoint, "cluster-info")
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
				_, err = utils.GetClusterInfoFromCluster(endpoint, "cluster-info")
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
				cluster, err = utils.GetClusterInfoFromCluster(endpoint, "cluster-info")
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
				cluster, err = utils.GetClusterInfoFromCluster(endpoint, "vip-cluster-info")
			})
			It("should return the cluster information", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(cluster.Server).Should(Equal(endpoint))
			})
		})
	})
	Describe("Get info from kubeconfig", func() {
		var (
			kubeconfigPath string
			context        string
			actual         string
			err            error
		)
		Context("When getting cluster name from kubeconfig", func() {
			JustBeforeEach(func() {
				actual, err = utils.GetClusterNameFromKubeconfigAndContext(kubeconfigPath, context)
			})
			AfterEach(func() {
				kubeconfigPath = ""
				context = ""
			})
			Context("When kubeconfig path is invalid", func() {
				BeforeEach(func() {
					kubeconfigPath = "../invalid1/config.yaml"
				})
				It("Should fail to find the kubeconfig", func() {
					Expect(err).To(HaveOccurred())
				})
			})
			Context("When context is missing", func() {
				BeforeEach(func() {
					kubeconfigPath = kubeconfig1Path
				})
				It("Should get the current context ", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(actual).To(Equal("horse-cluster"))
				})
			})
			Context("When context isn't found in kubeconfig", func() {
				BeforeEach(func() {
					kubeconfigPath = kubeconfig1Path
					context = "wrong"
				})
				It("Should return an error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("unable to find cluster name from kubeconfig file: \"../fakes/config/kubeconfig/config1.yaml\""))
				})
			})
		})
		Context("When getting server from kubeconfig", func() {
			JustBeforeEach(func() {
				actual, err = utils.GetClusterServerFromKubeconfigAndContext(kubeconfigPath, context)
			})
			Context("When kubeconfig path is invalid", func() {
				BeforeEach(func() {
					kubeconfigPath = "../invalid/config.yaml"
				})
				It("Should fail to open the kubeconfig", func() {
					Expect(err).To(HaveOccurred())
				})
			})
			Context("When context is empty", func() {
				BeforeEach(func() {
					kubeconfigPath = kubeconfig1Path
				})
				It("Should use the currennt context", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(actual).To(Equal("https://horse.org:4443"))
				})
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
			var gotPinnipedInfo *utils.PinnipedConfigMapInfo
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
				gotPinnipedInfo, err = utils.GetPinnipedInfoFromCluster(&cluster, nil)
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
				_, err = utils.GetPinnipedInfoFromCluster(&cluster, nil)
			})
			It("should return the pinniped-info successfully", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("error parsing http response body"))
			})
		})
		Context("When the configMap 'pinniped-info' is present in kube-public namespace", func() {
			var cluster clientcmdapi.Cluster
			var gotPinnipedInfo *utils.PinnipedConfigMapInfo
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
				gotPinnipedInfo, err = utils.GetPinnipedInfoFromCluster(&cluster, nil)
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
			var gotPinnipedInfo *utils.PinnipedConfigMapInfo
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
				gotPinnipedInfo, err = utils.GetPinnipedInfoFromCluster(&cluster, &discoveryPort)
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
			var gotPinnipedInfo *utils.PinnipedConfigMapInfo
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
				gotPinnipedInfo, err = utils.GetPinnipedInfoFromCluster(&cluster, nil)
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
