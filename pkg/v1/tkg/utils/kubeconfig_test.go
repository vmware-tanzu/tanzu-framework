// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

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
