// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package auth_test

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	clientauthenticationv1beta1 "k8s.io/client-go/pkg/apis/clientauthentication/v1beta1"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	tkgauth "github.com/vmware-tanzu/tanzu-framework/tkg/auth"
	"github.com/vmware-tanzu/tanzu-framework/tkg/fakes/helper"
)

var testingDir string

func TestTkgAuth(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tkg Auth Suite")
}

var _ = Describe("Unit tests for tkg auth", func() {
	var (
		err                      error
		endpoint                 string
		tlsserver                *ghttp.Server
		conciergeIsClusterScoped bool
		servCert                 *x509.Certificate
	)

	const (
		clustername = "fake-cluster"
		issuer      = "https://fakeissuer.com"
		issuerCA    = "fakeCAData"
	)

	Describe("Kubeconfig for Management cluster", func() {
		BeforeEach(func() {
			tlsserver = ghttp.NewTLSServer()
			servCert = tlsserver.HTTPTestServer.Certificate()
			endpoint = tlsserver.URL()
			err = createTempDirectory("kubeconfig-test")
			Expect(err).ToNot(HaveOccurred())
		})
		AfterEach(func() {
			tlsserver.Close()
			deleteTempDirectory()
		})
		Context("When the configMap 'cluster-info' is not present in kube-public namespace", func() {
			BeforeEach(func() {
				tlsserver.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/cluster-info"),
						ghttp.RespondWith(http.StatusNotFound, "not found"),
					),
				)
				_, _, err = tkgauth.KubeconfigWithPinnipedAuthLoginPlugin(endpoint, nil, tkgauth.DiscoveryStrategy{ClusterInfoConfigMap: tkgauth.DefaultClusterInfoConfigMap})
			})
			It("should return the error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("failed to get cluster-info"))
			})
		})
		Context("When the configMap 'pinniped-info' is not present in kube-public namespace", func() {
			BeforeEach(func() {
				clusterInfo := GetFakeClusterInfo(endpoint, servCert)
				tlsserver.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/cluster-info"),
						ghttp.RespondWith(http.StatusOK, clusterInfo),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/pinniped-info"),
						ghttp.RespondWith(http.StatusNotFound, "not found"),
					),
				)
				_, _, err = tkgauth.KubeconfigWithPinnipedAuthLoginPlugin(endpoint, nil, tkgauth.DiscoveryStrategy{ClusterInfoConfigMap: tkgauth.DefaultClusterInfoConfigMap})
			})
			It("should return the error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).Should(ContainSubstring("failed to get pinniped-info"))
			})
		})
		Context("When ConciergeIsClusterScoped is not set in 'pinniped-info' configMap", func() {
			var kubeConfigPath, kubeContext, kubeconfigMergeFilePath string
			BeforeEach(func() {
				var clusterInfo, pinnipedInfo string
				clusterInfo = GetFakeClusterInfo(endpoint, servCert)
				pinnipedInfo = helper.GetFakePinnipedInfo(
					helper.PinnipedInfo{
						ClusterName:        clustername,
						Issuer:             issuer,
						IssuerCABundleData: issuerCA,
					})
				tlsserver.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/cluster-info"),
						ghttp.RespondWith(http.StatusOK, clusterInfo),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/pinniped-info"),
						ghttp.RespondWith(http.StatusOK, pinnipedInfo),
					),
				)
				kubeconfigMergeFilePath = testingDir + "/config"
				options := &tkgauth.KubeConfigOptions{
					MergeFilePath: kubeconfigMergeFilePath,
				}
				kubeConfigPath, kubeContext, err = tkgauth.KubeconfigWithPinnipedAuthLoginPlugin(endpoint, options, tkgauth.DiscoveryStrategy{ClusterInfoConfigMap: tkgauth.DefaultClusterInfoConfigMap})
			})
			It("should set default values", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(kubeConfigPath).Should(Equal(kubeconfigMergeFilePath))
				Expect(len(kubeContext)).Should(Not(Equal(0)))
				config, err := clientcmd.LoadFromFile(kubeConfigPath)
				Expect(err).ToNot(HaveOccurred())
				gotClusterName := config.Contexts[kubeContext].Cluster
				cluster := config.Clusters[config.Contexts[kubeContext].Cluster]
				user := config.AuthInfos[config.Contexts[kubeContext].AuthInfo]
				Expect(cluster.Server).To(Equal(endpoint))
				Expect(gotClusterName).To(Equal(clustername))
				expectedExecConf := getExpectedExecConfig(endpoint, issuer, issuerCA, false, servCert)
				Expect(*user.Exec).To(Equal(*expectedExecConf))

			})
		})
		Context("When ConciergeIsClusterScoped is true in 'pinniped-info' configMap", func() {
			var kubeConfigPath, kubeContext, kubeconfigMergeFilePath string
			BeforeEach(func() {
				var clusterInfo, pinnipedInfo string
				clusterInfo = GetFakeClusterInfo(endpoint, servCert)
				pinnipedInfo = helper.GetFakePinnipedInfo(
					helper.PinnipedInfo{
						ClusterName:              clustername,
						Issuer:                   issuer,
						IssuerCABundleData:       issuerCA,
						ConciergeIsClusterScoped: true,
					})
				tlsserver.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/cluster-info"),
						ghttp.RespondWith(http.StatusOK, clusterInfo),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/pinniped-info"),
						ghttp.RespondWith(http.StatusOK, pinnipedInfo),
					),
				)
				kubeconfigMergeFilePath = testingDir + "/config"
				options := &tkgauth.KubeConfigOptions{
					MergeFilePath: kubeconfigMergeFilePath,
				}
				kubeConfigPath, kubeContext, err = tkgauth.KubeconfigWithPinnipedAuthLoginPlugin(endpoint, options, tkgauth.DiscoveryStrategy{ClusterInfoConfigMap: tkgauth.DefaultClusterInfoConfigMap})
			})
			It("should not include --concierge-namespace arg in kubeconfig", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(kubeConfigPath).Should(Equal(kubeconfigMergeFilePath))
				Expect(len(kubeContext)).Should(Not(Equal(0)))
				config, err := clientcmd.LoadFromFile(kubeConfigPath)
				Expect(err).ToNot(HaveOccurred())
				gotClusterName := config.Contexts[kubeContext].Cluster
				cluster := config.Clusters[config.Contexts[kubeContext].Cluster]
				user := config.AuthInfos[config.Contexts[kubeContext].AuthInfo]
				Expect(cluster.Server).To(Equal(endpoint))
				Expect(gotClusterName).To(Equal(clustername))
				expectedExecConf := getExpectedExecConfig(endpoint, issuer, issuerCA, true, servCert)
				Expect(*user.Exec).To(Equal(*expectedExecConf))

			})
		})
		Context("When the configMap 'pinniped-info' is present in kube-public namespace", func() {
			var kubeConfigPath, kubeContext, kubeconfigMergeFilePath string
			BeforeEach(func() {
				var clusterInfo, pinnipedInfo string
				conciergeIsClusterScoped = false
				clusterInfo = GetFakeClusterInfo(endpoint, servCert)
				pinnipedInfo = helper.GetFakePinnipedInfo(
					helper.PinnipedInfo{
						ClusterName:              clustername,
						Issuer:                   issuer,
						IssuerCABundleData:       issuerCA,
						ConciergeIsClusterScoped: conciergeIsClusterScoped,
					})
				tlsserver.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/cluster-info"),
						ghttp.RespondWith(http.StatusOK, clusterInfo),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/pinniped-info"),
						ghttp.RespondWith(http.StatusOK, pinnipedInfo),
					),
				)
				kubeconfigMergeFilePath = testingDir + "/config"
				options := &tkgauth.KubeConfigOptions{
					MergeFilePath: kubeconfigMergeFilePath,
				}
				kubeConfigPath, kubeContext, err = tkgauth.KubeconfigWithPinnipedAuthLoginPlugin(endpoint, options, tkgauth.DiscoveryStrategy{ClusterInfoConfigMap: tkgauth.DefaultClusterInfoConfigMap})
			})
			It("should generate the kubeconfig and merge the kubeconfig to given path", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(kubeConfigPath).Should(Equal(kubeconfigMergeFilePath))
				Expect(len(kubeContext)).Should(Not(Equal(0)))
				config, err := clientcmd.LoadFromFile(kubeConfigPath)
				Expect(err).ToNot(HaveOccurred())
				gotClusterName := config.Contexts[kubeContext].Cluster
				cluster := config.Clusters[config.Contexts[kubeContext].Cluster]
				user := config.AuthInfos[config.Contexts[kubeContext].AuthInfo]
				Expect(cluster.Server).To(Equal(endpoint))
				Expect(gotClusterName).To(Equal(clustername))
				expectedExecConf := getExpectedExecConfig(endpoint, issuer, issuerCA, conciergeIsClusterScoped, servCert)
				Expect(*user.Exec).To(Equal(*expectedExecConf))

			})
		})
		Describe("Get Tanzu local Kubeconfig path", func() {
			var localPath string
			Context("When TanzuLocalKubeConfigPath() is called", func() {
				BeforeEach(func() {
					localPath, err = tkgauth.TanzuLocalKubeConfigPath()
				})
				It("should return the tanzu local kubeconfig path", func() {
					Expect(err).ToNot(HaveOccurred())
					home, err := os.UserHomeDir()
					Expect(err).ToNot(HaveOccurred())
					Expect(localPath).Should(Equal(filepath.Join(home, tkgauth.TanzuLocalKubeDir, tkgauth.TanzuKubeconfigFile)))
				})
			})
		})

	})
})

func GetFakeClusterInfo(server string, cert *x509.Certificate) string {
	clusterInfoJSON := `
	{
		"kind": "ConfigMap",
		"apiVersion": "v1",
    	"data": {
        "kubeconfig": "apiVersion: v1\nclusters:\n- cluster:\n    certificate-authority-data: %s\n    server: %s\n  name: \"\"\ncontexts: null\ncurrent-context: \"\"\nkind: Config\npreferences: {}\nusers: null\n"
    	},
		"metadata": {
		  "name": "cluster-info",
		  "namespace": "kube-public"
		}
	}`
	certBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
	clusterInfoJSON = fmt.Sprintf(clusterInfoJSON, base64.StdEncoding.EncodeToString(certBytes), server)

	return clusterInfoJSON
}

func getExpectedExecConfig(endpoint string, issuer string, issuerCA string, conciergeIsClusterScoped bool, servCert *x509.Certificate) *clientcmdapi.ExecConfig {
	certBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: servCert.Raw})
	args := []string{
		"pinniped-auth", "login",
		"--enable-concierge",
		"--concierge-authenticator-name=" + tkgauth.ConciergeAuthenticatorName,
		"--concierge-authenticator-type=" + tkgauth.ConciergeAuthenticatorType,
		"--concierge-is-cluster-scoped=" + strconv.FormatBool(conciergeIsClusterScoped),
		"--concierge-endpoint=" + endpoint,
		"--concierge-ca-bundle-data=" + base64.StdEncoding.EncodeToString(certBytes),
		"--issuer=" + issuer,
		"--scopes=" + tkgauth.PinnipedOIDCScopes,
		"--ca-bundle-data=" + issuerCA,
		"--request-audience=" + issuer,
	}

	if !conciergeIsClusterScoped {
		args = append(args, "--concierge-namespace="+tkgauth.ConciergeNamespace)
	}

	execConfig := &clientcmdapi.ExecConfig{
		APIVersion:      clientauthenticationv1beta1.SchemeGroupVersion.String(),
		Args:            args,
		Env:             []clientcmdapi.ExecEnvVar{},
		Command:         "tanzu",
		InteractiveMode: "IfAvailable",
	}
	return execConfig
}

func createTempDirectory(prefix string) error {
	var err error
	testingDir, err = os.MkdirTemp("", prefix)
	if err != nil {
		fmt.Println("Error TempDir: ", err.Error())
		return err
	}
	return nil
}
func deleteTempDirectory() {
	os.Remove(testingDir)
}
