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
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake" // nolint:staticcheck

	. "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes"
	fakehelper "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/fakes/helper"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/region"
)

var _ = Describe("Unit tests for get cluster pinniped info", func() {
	var (
		err                               error
		regionalClusterClient             clusterclient.Client
		tkgClient                         *TkgClient
		crtClientFactory                  *fakes.CrtClientFactory
		mgmtClusterName                   string
		clusterClientOptions              clusterclient.Options
		fakeClientSet                     crtclient.Client
		searchNamespace                   string
		kubeconfig                        string
		endpoint                          string
		wlClusterEndpoint                 string
		tlsServer                         *ghttp.Server
		tlsServerWLCluster                *ghttp.Server
		issuer                            string
		issuerCA                          string
		apiGroupSuffix                    string
		apiGroupSuffixWLCluster           string
		conciergeIsClusterScoped          bool
		conciergeIsClusterScopedWLCluster bool
		servCert                          *x509.Certificate
		clusterPinnipedInfo               *ClusterPinnipedInfo
	)

	var (
		fakeIssuer         = "https://fakeissuer.com"
		fakeCAData         = "fakeCAData"
		fakeAPIGroupSuffix = "tuna.io"
	)

	BeforeEach(func() {
		tlsServer = ghttp.NewTLSServer()
		servCert = tlsServer.HTTPTestServer.Certificate()
		endpoint = tlsServer.URL()

		tlsServerWLCluster = ghttp.NewTLSServer()
		wlClusterEndpoint = tlsServerWLCluster.URL()

		tkgClient, err = CreateTKGClient("../fakes/config/config.yaml", testingDir, defaultTKGBoMFileForTesting, 2*time.Second)
		Expect(err).NotTo(HaveOccurred())
		crtClientFactory = &fakes.CrtClientFactory{}
		mgmtClusterName = fakeManagementClusterName
		searchNamespace = "tkg-system"
		clusterClientOptions = clusterclient.NewOptions(getFakePoller(), crtClientFactory, getFakeDiscoveryFactory(), nil)
	})
	AfterEach(func() {
		tlsServer.Close()
		tlsServerWLCluster.Close()
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
				fakeClientSet = fake.NewFakeClientWithScheme(scheme)
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cluster 'fake-mgmt-cluster' is not present in namespace"))
			})
		})
		Context("When cluster-info is not found in kube-public namespace", func() {
			BeforeEach(func() {
				// create a fake controller-runtime cluster with the []runtime.Object mentioned with createClusterOptions
				fakeClientSet = fake.NewFakeClientWithScheme(scheme,
					createFakeClusterRefObjects(mgmtClusterName, searchNamespace, endpoint)...)
				tlsServer.AppendHandlers(
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
		Context("When cluster-info is not found in kube-public namespace", func() {
			BeforeEach(func() {
				var clusterInfo string
				// create a fake controller-runtime cluster with the []runtime.Object mentioned with createClusterOptions
				fakeClientSet = fake.NewFakeClientWithScheme(scheme,
					createFakeClusterRefObjects(mgmtClusterName, searchNamespace, endpoint)...)
				clusterInfo = fakehelper.GetFakeClusterInfo(endpoint, servCert)
				tlsServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/cluster-info"),
						ghttp.RespondWith(http.StatusOK, clusterInfo),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/pinniped-info"),
						ghttp.RespondWith(http.StatusNotFound, "not found"),
					),
				)
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to get pinniped-info from cluster"))
			})
		})

		Context("When cluster-info and pinniped-info configmap are present in kube-public namespace", func() {
			BeforeEach(func() {
				var clusterInfo, pinnipedInfo string
				issuer = fakeIssuer
				issuerCA = fakeCAData
				apiGroupSuffix = fakeAPIGroupSuffix
				conciergeIsClusterScoped = false
				clusterInfo = fakehelper.GetFakeClusterInfo(endpoint, servCert)
				pinnipedInfo = fakehelper.GetFakePinnipedInfo(fakehelper.PinnipedInfo{
					ClusterName:              mgmtClusterName,
					Issuer:                   issuer,
					IssuerCABundleData:       issuerCA,
					ConciergeAPIGroupSuffix:  &apiGroupSuffix,
					ConciergeIsClusterScoped: conciergeIsClusterScoped,
				})
				// create a fake controller-runtime cluster with the []runtime.Object mentioned with createClusterOptions
				fakeClientSet = fake.NewFakeClientWithScheme(scheme,
					createFakeClusterRefObjects(mgmtClusterName, searchNamespace, endpoint)...)
				tlsServer.AppendHandlers(
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
				Expect(dereferenceStringPointer(clusterPinnipedInfo.PinnipedInfo.Data.ConciergeAPIGroupSuffix)).To(Equal(apiGroupSuffix))
				Expect(clusterPinnipedInfo.PinnipedInfo.Data.ConciergeIsClusterScoped).To(Equal(conciergeIsClusterScoped))
			})
		})
	})
	Describe("Get cluster pinniped info for workload cluster", func() {
		JustBeforeEach(func() {
			crtClientFactory.NewClientReturns(fakeClientSet, nil)
			// generate fake kubeconfig with testserver url for workload cluster
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
				fakeClientSet = fake.NewFakeClientWithScheme(scheme)
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
				fakeClientSet = fake.NewFakeClientWithScheme(scheme,
					createFakeClusterRefObjects("fake-workload-cluster", searchNamespace, endpoint)...)
				tlsServer.AppendHandlers(
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

		Context("When management cluster pinniped-info configmap is not present in kube-public namespace", func() {
			BeforeEach(func() {
				var managementClusterInfo, workloadClusterInfo string
				var clusterRefs []runtime.Object
				clusterRefs = append(clusterRefs, createFakeClusterRefObjects("fake-workload-cluster", "default", wlClusterEndpoint)[0])
				clusterRefs = append(clusterRefs, createFakeClusterRefObjects(mgmtClusterName, searchNamespace, endpoint)[0])
				managementClusterInfo = fakehelper.GetFakeClusterInfo(endpoint, servCert)
				workloadClusterInfo = fakehelper.GetFakeClusterInfo(tlsServerWLCluster.URL(), tlsServerWLCluster.HTTPTestServer.Certificate())
				searchNamespace = constants.DefaultNamespace
				// create a fake controller-runtime cluster with the []runtime.Object mentioned with createClusterOptions
				fakeClientSet = fake.NewFakeClientWithScheme(scheme, clusterRefs...)
				tlsServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/cluster-info"),
						ghttp.RespondWith(http.StatusOK, managementClusterInfo),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/pinniped-info"),
						ghttp.RespondWith(http.StatusNotFound, "not found"),
					),
				)

				tlsServerWLCluster.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/cluster-info"),
						ghttp.RespondWith(http.StatusOK, workloadClusterInfo),
					),
				)
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to get pinniped-info from management cluster"))
			})
		})

		Context("When WL cluster pinniped-info configmap is not present in kube-public namespace", func() {
			BeforeEach(func() {
				var managementClusterInfo, workloadClusterInfo, pinnipedInfo string
				var clusterRefs []runtime.Object
				clusterRefs = append(clusterRefs, createFakeClusterRefObjects("fake-workload-cluster", "default", wlClusterEndpoint)[0])
				clusterRefs = append(clusterRefs, createFakeClusterRefObjects(mgmtClusterName, searchNamespace, endpoint)[0])
				issuer = fakeIssuer
				issuerCA = fakeCAData
				apiGroupSuffix = fakeAPIGroupSuffix
				conciergeIsClusterScoped = false
				managementClusterInfo = fakehelper.GetFakeClusterInfo(endpoint, servCert)
				workloadClusterInfo = fakehelper.GetFakeClusterInfo(tlsServerWLCluster.URL(), tlsServerWLCluster.HTTPTestServer.Certificate())
				pinnipedInfo = fakehelper.GetFakePinnipedInfo(fakehelper.PinnipedInfo{
					ClusterName:              mgmtClusterName,
					Issuer:                   issuer,
					IssuerCABundleData:       issuerCA,
					ConciergeAPIGroupSuffix:  &apiGroupSuffix,
					ConciergeIsClusterScoped: conciergeIsClusterScoped,
				})
				searchNamespace = constants.DefaultNamespace
				// create a fake controller-runtime cluster with the []runtime.Object mentioned with createClusterOptions
				fakeClientSet = fake.NewFakeClientWithScheme(scheme, clusterRefs...)
				tlsServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/cluster-info"),
						ghttp.RespondWith(http.StatusOK, managementClusterInfo),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/pinniped-info"),
						ghttp.RespondWith(http.StatusOK, pinnipedInfo),
					),
				)

				tlsServerWLCluster.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/cluster-info"),
						ghttp.RespondWith(http.StatusOK, workloadClusterInfo),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/pinniped-info"),
						ghttp.RespondWith(http.StatusNotFound, "not found"),
					),
				)
			})
			It("should not return an error and utilize management cluster pinniped info", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(clusterPinnipedInfo.ClusterName).To(Equal("fake-workload-cluster"))
				Expect(clusterPinnipedInfo.ClusterInfo.Server).To(Equal(wlClusterEndpoint))
				Expect(clusterPinnipedInfo.PinnipedInfo.Data.Issuer).To(Equal(issuer))
				Expect(clusterPinnipedInfo.PinnipedInfo.Data.IssuerCABundle).To(Equal(issuerCA))
				Expect(dereferenceStringPointer(clusterPinnipedInfo.PinnipedInfo.Data.ConciergeAPIGroupSuffix)).To(Equal("pinniped.dev"))
				Expect(clusterPinnipedInfo.PinnipedInfo.Data.ConciergeIsClusterScoped).To(Equal(false))

			})
		})

		Context("When WL cluster pinniped-info configmap is present and malformed in kube-public namespace", func() {
			BeforeEach(func() {
				var managementClusterInfo, workloadClusterInfo, pinnipedInfo string
				var clusterRefs []runtime.Object
				clusterRefs = append(clusterRefs, createFakeClusterRefObjects("fake-workload-cluster", "default", wlClusterEndpoint)[0])
				clusterRefs = append(clusterRefs, createFakeClusterRefObjects(mgmtClusterName, searchNamespace, endpoint)[0])
				issuer = fakeIssuer
				issuerCA = fakeCAData
				apiGroupSuffix = fakeAPIGroupSuffix
				conciergeIsClusterScoped = false
				managementClusterInfo = fakehelper.GetFakeClusterInfo(endpoint, servCert)
				workloadClusterInfo = fakehelper.GetFakeClusterInfo(tlsServerWLCluster.URL(), tlsServerWLCluster.HTTPTestServer.Certificate())
				pinnipedInfo = fakehelper.GetFakePinnipedInfo(fakehelper.PinnipedInfo{
					ClusterName:              mgmtClusterName,
					Issuer:                   issuer,
					IssuerCABundleData:       issuerCA,
					ConciergeAPIGroupSuffix:  &apiGroupSuffix,
					ConciergeIsClusterScoped: conciergeIsClusterScoped,
				})
				searchNamespace = constants.DefaultNamespace
				// create a fake controller-runtime cluster with the []runtime.Object mentioned with createClusterOptions
				fakeClientSet = fake.NewFakeClientWithScheme(scheme, clusterRefs...)
				tlsServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/cluster-info"),
						ghttp.RespondWith(http.StatusOK, managementClusterInfo),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/pinniped-info"),
						ghttp.RespondWith(http.StatusOK, pinnipedInfo),
					),
				)

				tlsServerWLCluster.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/cluster-info"),
						ghttp.RespondWith(http.StatusOK, workloadClusterInfo),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/pinniped-info"),
						ghttp.RespondWith(http.StatusOK, "{incorrectJSON}"),
					),
				)
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("error parsing http response body"))
			})
		})

		Context("When cluster-info and pinniped-info configmap are present in kube-public namespace", func() {
			BeforeEach(func() {
				var managementClusterInfo, workloadClusterInfo, pinnipedInfo string
				var clusterRefs []runtime.Object
				clusterRefs = append(clusterRefs, createFakeClusterRefObjects("fake-workload-cluster", "default", wlClusterEndpoint)[0])
				clusterRefs = append(clusterRefs, createFakeClusterRefObjects(mgmtClusterName, searchNamespace, endpoint)[0])
				issuer = fakeIssuer
				issuerCA = fakeCAData
				apiGroupSuffix = fakeAPIGroupSuffix
				conciergeIsClusterScoped = false
				apiGroupSuffixWLCluster = "salmon.me"
				conciergeIsClusterScopedWLCluster = true
				managementClusterInfo = fakehelper.GetFakeClusterInfo(endpoint, servCert)
				workloadClusterInfo = fakehelper.GetFakeClusterInfo(tlsServerWLCluster.URL(), tlsServerWLCluster.HTTPTestServer.Certificate())
				pinnipedInfo = fakehelper.GetFakePinnipedInfo(fakehelper.PinnipedInfo{
					ClusterName:              mgmtClusterName,
					Issuer:                   issuer,
					IssuerCABundleData:       issuerCA,
					ConciergeAPIGroupSuffix:  &apiGroupSuffix,
					ConciergeIsClusterScoped: conciergeIsClusterScoped,
				})
				pinnipedInfoWorkloadCluster := fakehelper.GetFakePinnipedInfo(fakehelper.PinnipedInfo{
					ConciergeAPIGroupSuffix:  &apiGroupSuffixWLCluster,
					ConciergeIsClusterScoped: conciergeIsClusterScopedWLCluster,
				})
				searchNamespace = constants.DefaultNamespace
				// create a fake controller-runtime cluster with the []runtime.Object mentioned with createClusterOptions
				fakeClientSet = fake.NewFakeClientWithScheme(scheme, clusterRefs...)
				tlsServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/cluster-info"),
						ghttp.RespondWith(http.StatusOK, managementClusterInfo),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/pinniped-info"),
						ghttp.RespondWith(http.StatusOK, pinnipedInfo),
					),
				)

				tlsServerWLCluster.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/cluster-info"),
						ghttp.RespondWith(http.StatusOK, workloadClusterInfo),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/pinniped-info"),
						ghttp.RespondWith(http.StatusOK, pinnipedInfoWorkloadCluster),
					),
				)
			})
			It("should return the cluster pinniped information successfully", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(clusterPinnipedInfo.ClusterName).To(Equal("fake-workload-cluster"))
				Expect(clusterPinnipedInfo.ClusterInfo.Server).To(Equal(wlClusterEndpoint))
				Expect(clusterPinnipedInfo.PinnipedInfo.Data.Issuer).To(Equal(issuer))
				Expect(clusterPinnipedInfo.PinnipedInfo.Data.IssuerCABundle).To(Equal(issuerCA))
				Expect(dereferenceStringPointer(clusterPinnipedInfo.PinnipedInfo.Data.ConciergeAPIGroupSuffix)).To(Equal(apiGroupSuffixWLCluster))
				Expect(clusterPinnipedInfo.PinnipedInfo.Data.ConciergeIsClusterScoped).To(Equal(conciergeIsClusterScopedWLCluster))
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

func dereferenceStringPointer(pointer *string) string {
	if pointer != nil {
		return *pointer
	}

	return ""
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
