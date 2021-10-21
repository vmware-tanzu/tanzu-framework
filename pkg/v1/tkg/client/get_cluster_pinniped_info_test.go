// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake" // nolint:staticcheck,nolintlint

	. "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes"
	fakehelper "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes/helper"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/region"
)

var _ = Describe("Unit tests for get cluster pinniped info", func() {
	var (
		err                   error
		regionalClusterClient clusterclient.Client
		tkgClient             *TkgClient
		crtClientFactory      *fakes.CrtClientFactory
		mgmtClusterName       string
		clusterClientOptions  clusterclient.Options
		fakeClientSet         crtclient.Client
		searchNamespace       string
		kubeconfig            string
		endpoint              string
		tlsserver             *ghttp.Server
		issuer                string
		issuerCA              string
		servCert              *x509.Certificate
		clusterPinnipedInfo   *ClusterPinnipedInfo
	)

	var (
		fakeIssuer = "https://fakeissuer.com"
		fakeCAData = "fakeCAData"
	)

	BeforeEach(func() {
		tlsserver = ghttp.NewTLSServer()
		servCert = tlsserver.HTTPTestServer.Certificate()
		endpoint = tlsserver.URL()

		tkgClient, err = CreateTKGClient("../fakes/config/config.yaml", testingDir, defaultTKGBoMFileForTesting, 2*time.Second)
		Expect(err).NotTo(HaveOccurred())
		crtClientFactory = &fakes.CrtClientFactory{}
		mgmtClusterName = fakeManagementClusterName
		searchNamespace = "tkg-system"
		clusterClientOptions = clusterclient.NewOptions(getFakePoller(), crtClientFactory, getFakeDiscoveryFactory(), nil)
	})
	AfterEach(func() {
		tlsserver.Close()
	})

	Describe("Get cluster pinniped info for management cluster", func() {
		JustBeforeEach(func() {
			crtClientFactory.NewClientReturns(fakeClientSet, nil)
			kubeconfig, err = getFakeKubeConfigFilePathWithServer(testingDir, endpoint, mgmtClusterName)
			Expect(err).NotTo(HaveOccurred())
			regionalClusterClient, err = clusterclient.NewClient(kubeconfig, "dummy-context", clusterClientOptions)
			Expect(err).NotTo(HaveOccurred())
			region := region.RegionContext{
				ContextName: "dummy-context",
			}
			clusterPinnipedInfo, err = tkgClient.GetMCClusterPinnipedInfo(regionalClusterClient, region)
		})

		Context("When cluster is not found", func() {
			BeforeEach(func() {
				// create a fake controller-runtime cluster with the []runtime.Object mentioned with createClusterOptions
				fakeClientSet = fake.NewClientBuilder().WithScheme(scheme).Build()
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cluster 'fake-mgmt-cluster' is not present in namespace"))
			})
		})
		Context("When cluster-info is not found in kube-public namespace", func() {
			BeforeEach(func() {
				// create a fake controller-runtime cluster with the []runtime.Object mentioned with createClusterOptions
				fakeClientSet = fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(
					createFakeClusterRefObjects(mgmtClusterName, searchNamespace, endpoint)...).Build()
				tlsserver.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/cluster-info"),
						ghttp.RespondWith(http.StatusNotFound, "not found"),
					),
				)
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to get cluster-info from cluster"))
			})
		})

		Context("When cluster-info  and pinniped-info configmap are present kube-public namespace", func() {
			BeforeEach(func() {
				var clusterInfo, pinnipedInfo string
				issuer = fakeIssuer
				issuerCA = fakeCAData
				clusterInfo = fakehelper.GetFakeClusterInfo(endpoint, servCert)
				pinnipedInfo = fakehelper.GetFakePinnipedInfo(mgmtClusterName, issuer, issuerCA)
				// create a fake controller-runtime cluster with the []runtime.Object mentioned with createClusterOptions
				fakeClientSet = fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(
					createFakeClusterRefObjects(mgmtClusterName, searchNamespace, endpoint)...).Build()
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
			})
			It("should return the cluster pinniped information successfully", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(clusterPinnipedInfo.ClusterName).To(Equal(mgmtClusterName))
				Expect(clusterPinnipedInfo.ClusterInfo.Server).To(Equal(endpoint))
				Expect(clusterPinnipedInfo.PinnipedInfo.Data.Issuer).To(Equal(issuer))
				Expect(clusterPinnipedInfo.PinnipedInfo.Data.IssuerCABundle).To(Equal(issuerCA))
			})
		})
	})
	Describe("Get cluster pinniped info for workload cluster", func() {
		JustBeforeEach(func() {
			crtClientFactory.NewClientReturns(fakeClientSet, nil)
			// generate fake kubeconfig with testserver url for management cluster
			kubeconfig, err = getFakeKubeConfigFilePathWithServer(testingDir, endpoint, mgmtClusterName)
			Expect(err).NotTo(HaveOccurred())

			regionalClusterClient, err = clusterclient.NewClient(kubeconfig, "dummy-context", clusterClientOptions)
			Expect(err).NotTo(HaveOccurred())

			region := region.RegionContext{
				SourceFilePath: kubeconfig,
				ContextName:    "dummy-context",
			}
			options := GetClusterPinnipedInfoOptions{
				ClusterName:         "fake-workload-cluster",
				Namespace:           constants.DefaultNamespace,
				IsManagementCluster: false,
			}
			clusterPinnipedInfo, err = tkgClient.GetWCClusterPinnipedInfo(regionalClusterClient, region, options)
		})

		Context("When cluster is not found", func() {
			BeforeEach(func() {
				// create a fake controller-runtime cluster with the []runtime.Object mentioned with createClusterOptions
				fakeClientSet = fake.NewClientBuilder().WithScheme(scheme).Build()
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cluster 'fake-workload-cluster' is not present in namespace"))
			})
		})
		Context("When cluster-info is not found in kube-public namespace", func() {
			BeforeEach(func() {
				searchNamespace = constants.DefaultNamespace
				// create a fake controller-runtime cluster with the []runtime.Object mentioned with createClusterOptions
				fakeClientSet = fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(
					createFakeClusterRefObjects("fake-workload-cluster", searchNamespace, endpoint)...).Build()
				tlsserver.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/cluster-info"),
						ghttp.RespondWith(http.StatusNotFound, "not found"),
					),
				)
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to get cluster-info from cluster"))
			})
		})

		Context("When cluster-info  and pinniped-info configmap are present kube-public namespace", func() {
			BeforeEach(func() {
				var clusterInfo, pinnipedInfo string
				issuer = fakeIssuer
				issuerCA = fakeCAData
				clusterInfo = fakehelper.GetFakeClusterInfo(endpoint, servCert)
				pinnipedInfo = fakehelper.GetFakePinnipedInfo(mgmtClusterName, issuer, issuerCA)
				searchNamespace = constants.DefaultNamespace
				// create a fake controller-runtime cluster with the []runtime.Object mentioned with createClusterOptions
				fakeClientSet = fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(
					createFakeClusterRefObjects("fake-workload-cluster", searchNamespace, endpoint)...).Build()
				tlsserver.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/cluster-info"),
						ghttp.RespondWith(http.StatusOK, clusterInfo),
					),
					// Note:currently using the same testserver for Management cluster and workload-cluster while fetching cluster-info
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/cluster-info"),
						ghttp.RespondWith(http.StatusOK, clusterInfo),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/pinniped-info"),
						ghttp.RespondWith(http.StatusOK, pinnipedInfo),
					),
				)
			})
			It("should return the cluster pinniped information successfully", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(clusterPinnipedInfo.ClusterName).To(Equal("fake-workload-cluster"))
				Expect(clusterPinnipedInfo.ClusterInfo.Server).To(Equal(endpoint))
				Expect(clusterPinnipedInfo.PinnipedInfo.Data.Issuer).To(Equal(issuer))
				Expect(clusterPinnipedInfo.PinnipedInfo.Data.IssuerCABundle).To(Equal(issuerCA))
			})
		})
	})
})

func getFakeKubeConfigFilePathWithServer(testingDir, endpoint, clustername string) (string, error) {
	kubeconfig := &clientcmdapi.Config{
		Kind:           "Config",
		APIVersion:     clientcmdapi.SchemeGroupVersion.Version,
		Clusters:       map[string]*clientcmdapi.Cluster{clustername: {Server: endpoint}}, //nolint:gofmt
		Contexts:       map[string]*clientcmdapi.Context{"dummy-context": {Cluster: clustername}},
		CurrentContext: "dummy-context",
	}
	f, err := os.CreateTemp(testingDir, "kube")
	if err != nil {
		fmt.Println("Error creating TempFile: ", err.Error())
		return "", err
	}
	err = clientcmd.WriteToFile(*kubeconfig, f.Name())
	if err != nil {
		return "", err
	}
	return f.Name(), nil
}

// TODO: Should be merged to pkg/fakes/helpers/fakeobjectcreator.go
func createFakeClusterRefObjects(name, namespace, endpoint string) []runtime.Object {
	u, _ := url.Parse(endpoint)
	host, port, _ := net.SplitHostPort(u.Host)
	portInt32, _ := strconv.ParseInt(port, 10, 32)

	// create a cluster object
	cluster := &capi.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Cluster",
			APIVersion: capi.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},

		Spec: capi.ClusterSpec{
			ControlPlaneEndpoint: capi.APIEndpoint{
				Host: host,
				Port: int32(portInt32),
			},
		},
	}

	runtimeObjects := []runtime.Object{}
	runtimeObjects = append(runtimeObjects, cluster)
	return runtimeObjects
}
