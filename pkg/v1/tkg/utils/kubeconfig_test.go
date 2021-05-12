/*
Copyright 2020 The TKG Contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package utils_test

import (
	"crypto/x509"
	"encoding/pem"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	fakehelper "github.com/vmware-tanzu-private/core/pkg/v1/tkg/fakes/helper"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/utils"
)

var _ = Describe("Kubeconfig Tests", func() {
	var (
		err         error
		endpoint    string
		tlsserver   *ghttp.Server
		clustername string
		issuer      string
		issuerCA    string
		servCert    *x509.Certificate
	)

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
				_, err = utils.GetClusterInfoFromCluster(endpoint)
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
				_, err = utils.GetClusterInfoFromCluster(endpoint)
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
				cluster, err = utils.GetClusterInfoFromCluster(endpoint)
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
				_, err = utils.GetPinnipedInfoFromCluster(&cluster)
			})
			It("should return the error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("failed to get pinniped-info"))
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
				_, err = utils.GetPinnipedInfoFromCluster(&cluster)
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
				clustername = "fake-cluster"
				issuer = "https://fakeissuer.com"
				issuerCA = "fakeCAData"
				pinnipedInfo := fakehelper.GetFakePinnipedInfo(clustername, issuer, issuerCA)
				tlsserver.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/pinniped-info"),
						ghttp.RespondWith(http.StatusOK, pinnipedInfo),
					),
				)
				cluster.Server = endpoint
				certBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: servCert.Raw})
				cluster.CertificateAuthorityData = certBytes
				gotPinnipedInfo, err = utils.GetPinnipedInfoFromCluster(&cluster)
			})
			It("should return the pinniped-info successfully", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(gotPinnipedInfo.Data.ClusterName).Should(Equal(clustername))
				Expect(gotPinnipedInfo.Data.Issuer).Should(Equal(issuer))
				Expect(gotPinnipedInfo.Data.IssuerCABundle).Should(Equal(issuerCA))
			})
		})
	})
})
