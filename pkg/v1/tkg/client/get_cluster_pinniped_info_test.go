// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"crypto/x509"
	"encoding/pem"
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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
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
		err                               error
		regionalClusterClient             clusterclient.Client
		tkgClient                         *TkgClient
		crtClientFactory                  *fakes.CrtClientFactory
		mgmtClusterName                   string
		clusterClientOptions              clusterclient.Options
		fakeClientSet                     crtclient.Client
		searchNamespace                   string
		kubeconfig                        string
		wlKubeconfig                      string
		endpoint                          string
		wlClusterEndpoint                 string
		tlsServer                         *ghttp.Server
		tlsServerWLCluster                *ghttp.Server
		issuer                            string
		issuerCA                          string
		conciergeIsClusterScoped          bool
		conciergeIsClusterScopedWLCluster bool
		servCert                          *x509.Certificate
		clusterPinnipedInfo               *ClusterPinnipedInfo
		isPacific                         bool
	)

	var (
		fakeIssuer = "https://fakeissuer.com"
		fakeCAData = "fakeCAData"
	)

	const (
		wlClusterName = "fake-workload-cluster"
	)

	BeforeEach(func() {
		tlsServer = ghttp.NewTLSServer()
		servCert = tlsServer.HTTPTestServer.Certificate()
		servCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: servCert.Raw})
		endpoint = tlsServer.URL()

		tlsServerWLCluster = ghttp.NewTLSServer()
		wlClusterEndpoint = tlsServerWLCluster.URL()
		wlServCert := tlsServer.HTTPTestServer.Certificate()
		wlServCertPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: wlServCert.Raw})

		tkgClient, err = CreateTKGClient("../fakes/config/config.yaml", testingDir, defaultTKGBoMFileForTesting, 2*time.Second)
		Expect(err).NotTo(HaveOccurred())
		crtClientFactory = &fakes.CrtClientFactory{}
		mgmtClusterName = fakeManagementClusterName
		searchNamespace = "tkg-system"
		clusterClientOptions = clusterclient.NewOptions(getFakePoller(), crtClientFactory, getFakeDiscoveryFactory(), nil)

		// generate fake kubeconfig with testserver url for workload cluster
		kubeconfig, err = getFakeKubeConfigFilePathWithServer(testingDir, endpoint, mgmtClusterName, servCertPEM)
		Expect(err).NotTo(HaveOccurred())
		wlKubeconfig, err = getFakeKubeConfigFilePathWithServer(testingDir, wlClusterEndpoint, wlClusterName, wlServCertPEM)
		Expect(err).NotTo(HaveOccurred())
	})
	AfterEach(func() {
		tlsServer.Close()
		tlsServerWLCluster.Close()
	})

	Describe("Get cluster pinniped info for management cluster", func() {
		JustBeforeEach(func() {
			crtClientFactory.NewClientReturns(fakeClientSet, nil)
			regionalClusterClient, err = clusterclient.NewClient(kubeconfig, "dummy-context", clusterClientOptions)
			Expect(err).NotTo(HaveOccurred())
			region := region.RegionContext{
				ContextName: "dummy-context",
			}
			options := GetClusterPinnipedInfoOptions{
				ClusterName: mgmtClusterName,
				Namespace:   searchNamespace,
			}
			clusterPinnipedInfo, err = tkgClient.GetMCClusterPinnipedInfo(regionalClusterClient, region, options)
		})

		Context("When cluster is not found", func() {
			BeforeEach(func() {
				// create a fake controller-runtime cluster with the []runtime.Object mentioned with createClusterOptions
				fakeClientSet = fake.NewClientBuilder().WithScheme(scheme).Build()
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(`failed to get cluster information: failed to get kubeconfig for cluster tkg-system/fake-mgmt-cluster: secrets "fake-mgmt-cluster-kubeconfig" not found`))
			})
		})
		Context("When cluster kubeconfig is invalid", func() {
			BeforeEach(func() {
				// create a fake controller-runtime cluster with the []runtime.Object mentioned with createClusterOptions
				fakeClientSet = fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(
					createFakeClusterRefObjects(mgmtClusterName, searchNamespace, "some-uid", endpoint, "invalid kubeconfig")...).Build()
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(HavePrefix("failed to get cluster information: failed to load the kubeconfig: couldn't get version/kind; json parse error"))
			})
		})
		Context("When pinniped-info is not found in kube-public namespace", func() {
			BeforeEach(func() {
				// create a fake controller-runtime cluster with the []runtime.Object mentioned with createClusterOptions
				fakeClientSet = fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(
					createFakeClusterRefObjects(mgmtClusterName, searchNamespace, "some-uid", endpoint, readFile(kubeconfig))...).Build()
				tlsServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/pinniped-info"),
						ghttp.RespondWith(http.StatusNotFound, "not found"),
					),
				)
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("failed to get pinniped-info from cluster"))
			})
		})

		Context("When cluster-info and pinniped-info configmap are present in kube-public namespace", func() {
			BeforeEach(func() {
				var pinnipedInfo string
				issuer = fakeIssuer
				issuerCA = fakeCAData
				conciergeIsClusterScoped = false
				pinnipedInfo = fakehelper.GetFakePinnipedInfo(fakehelper.PinnipedInfo{
					ClusterName:              mgmtClusterName,
					Issuer:                   issuer,
					IssuerCABundleData:       issuerCA,
					ConciergeIsClusterScoped: conciergeIsClusterScoped,
				})
				// create a fake controller-runtime cluster with the []runtime.Object mentioned with createClusterOptions
				fakeClientSet = fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(
					createFakeClusterRefObjects(mgmtClusterName, searchNamespace, "some-uid", endpoint, readFile(kubeconfig))...).Build()
				tlsServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/pinniped-info"),
						ghttp.RespondWith(http.StatusOK, pinnipedInfo),
					),
				)
			})
			It("should return the cluster pinniped information successfully", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(clusterPinnipedInfo.ClusterName).To(Equal(mgmtClusterName))
				Expect(clusterPinnipedInfo.ClusterAudience).To(BeNil())
				Expect(clusterPinnipedInfo.ClusterInfo.Server).To(Equal(endpoint))
				Expect(clusterPinnipedInfo.PinnipedInfo.Data.Issuer).To(Equal(issuer))
				Expect(clusterPinnipedInfo.PinnipedInfo.Data.IssuerCABundle).To(Equal(issuerCA))
				Expect(clusterPinnipedInfo.PinnipedInfo.Data.ConciergeIsClusterScoped).To(Equal(conciergeIsClusterScoped))
			})
		})
	})
	Describe("Get cluster pinniped info for workload cluster", func() {
		JustBeforeEach(func() {
			crtClientFactory.NewClientReturns(fakeClientSet, nil)

			regionalClusterClient, err = clusterclient.NewClient(kubeconfig, "dummy-context", clusterClientOptions)
			Expect(err).NotTo(HaveOccurred())

			region := region.RegionContext{
				SourceFilePath: kubeconfig,
				ContextName:    "dummy-context",
				ClusterName:    mgmtClusterName,
			}
			options := GetClusterPinnipedInfoOptions{
				ClusterName:         wlClusterName,
				Namespace:           constants.DefaultNamespace,
				IsManagementCluster: false,
			}
			clusterPinnipedInfo, err = tkgClient.GetWCClusterPinnipedInfo(regionalClusterClient, region, options, isPacific)
		})

		Context("When workload cluster is not found", func() {
			BeforeEach(func() {
				// create a fake controller-runtime cluster with the []runtime.Object mentioned with createClusterOptions
				fakeClientSet = fake.NewClientBuilder().WithScheme(scheme).Build()
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(`failed to get workload cluster information: failed to get kubeconfig for cluster default/fake-workload-cluster: secrets "fake-workload-cluster-kubeconfig" not found`))
			})
		})
		Context("When workload cluster kubeconfig is invalid", func() {
			BeforeEach(func() {
				searchNamespace = constants.DefaultNamespace
				// create a fake controller-runtime cluster with the []runtime.Object mentioned with createClusterOptions
				fakeClientSet = fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(
					createFakeClusterRefObjects("fake-workload-cluster", searchNamespace, "some-uid", endpoint, "invalid kubeconfig")...).Build()
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(HavePrefix("failed to get workload cluster information: failed to load the kubeconfig: couldn't get version/kind; json parse error"))
			})
		})

		Context("When management cluster pinniped-info configmap is not present in kube-public namespace", func() {
			BeforeEach(func() {
				var clusterRefs []runtime.Object
				clusterRefs = append(clusterRefs, createFakeClusterRefObjects(wlClusterName, "default", "some-uid", wlClusterEndpoint, readFile(wlKubeconfig))...)
				clusterRefs = append(clusterRefs, createFakeClusterRefObjects(mgmtClusterName, "some-uid", searchNamespace, endpoint, readFile(kubeconfig))...)
				searchNamespace = constants.DefaultNamespace
				// create a fake controller-runtime cluster with the []runtime.Object mentioned with createClusterOptions
				fakeClientSet = fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(clusterRefs...).Build()
				tlsServer.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/pinniped-info"),
						ghttp.RespondWith(http.StatusNotFound, "not found"),
					),
				)
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("failed to get pinniped-info from management cluster"))
			})
		})

		Context("When WL cluster pinniped-info configmap is not present in kube-public namespace", func() {
			BeforeEach(func() {
				var clusterRefs []runtime.Object
				clusterRefs = append(clusterRefs, createFakeClusterRefObjects(wlClusterName, "default", "some-uid", wlClusterEndpoint, readFile(wlKubeconfig))...)
				clusterRefs = append(clusterRefs, createFakeClusterRefObjects(mgmtClusterName, searchNamespace, "some-uid", endpoint, readFile(kubeconfig))...)
				issuer = fakeIssuer
				issuerCA = fakeCAData
				conciergeIsClusterScoped = false
				pinnipedInfoCM := fakehelper.PinnipedInfo{
					ClusterName:              mgmtClusterName,
					Issuer:                   issuer,
					IssuerCABundleData:       issuerCA,
					ConciergeIsClusterScoped: conciergeIsClusterScoped,
				}
				// Put fake pinniped-info in management cluster.
				pinnipedInfoConfigMap := getPinnipedInfoConfigMapObjectFromPinnipedInfo(pinnipedInfoCM)

				clusterRefs = append(clusterRefs, pinnipedInfoConfigMap)

				searchNamespace = constants.DefaultNamespace
				// create a fake controller-runtime cluster with the []runtime.Object mentioned with createClusterOptions
				fakeClientSet = fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(clusterRefs...).Build()

				tlsServerWLCluster.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/pinniped-info"),
						ghttp.RespondWith(http.StatusNotFound, "not found"),
					),
				)
			})
			It("should not return an error and utilize management cluster pinniped info", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(clusterPinnipedInfo.ClusterName).To(Equal(wlClusterName))
				Expect(clusterPinnipedInfo.ClusterAudience).To(BeNil())
				Expect(clusterPinnipedInfo.ClusterInfo.Server).To(Equal(wlClusterEndpoint))
				Expect(clusterPinnipedInfo.PinnipedInfo.Data.Issuer).To(Equal(issuer))
				Expect(clusterPinnipedInfo.PinnipedInfo.Data.IssuerCABundle).To(Equal(issuerCA))
				Expect(clusterPinnipedInfo.PinnipedInfo.Data.ConciergeIsClusterScoped).To(Equal(false))

			})
		})

		Context("When WL cluster pinniped-info configmap is present and malformed in kube-public namespace", func() {
			BeforeEach(func() {
				var clusterRefs []runtime.Object
				clusterRefs = append(clusterRefs, createFakeClusterRefObjects(wlClusterName, "default", "some-uid", wlClusterEndpoint, readFile(wlKubeconfig))...)
				clusterRefs = append(clusterRefs, createFakeClusterRefObjects(mgmtClusterName, searchNamespace, "some-uid", endpoint, readFile(kubeconfig))...)
				issuer = fakeIssuer
				issuerCA = fakeCAData
				conciergeIsClusterScoped = false
				pinnipedInfoCM := fakehelper.PinnipedInfo{
					ClusterName:              mgmtClusterName,
					Issuer:                   issuer,
					IssuerCABundleData:       issuerCA,
					ConciergeIsClusterScoped: conciergeIsClusterScoped,
				}
				clusterRefs = append(clusterRefs, getPinnipedInfoConfigMapObjectFromPinnipedInfo(pinnipedInfoCM))
				searchNamespace = constants.DefaultNamespace
				// create a fake controller-runtime cluster with the []runtime.Object mentioned with createClusterOptions
				fakeClientSet = fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(clusterRefs...).Build()

				tlsServerWLCluster.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/pinniped-info"),
						ghttp.RespondWith(http.StatusOK, "{incorrectJSON}"),
					),
				)
			})
			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(`failed to get pinniped-info from workload cluster: error parsing http response body: invalid character 'i' looking for beginning of object key string`))
			})
		})

		Context("When cluster-info and pinniped-info configmap are present in kube-public namespace", func() {
			BeforeEach(func() {
				var clusterRefs []runtime.Object
				clusterRefs = append(clusterRefs, createFakeClusterRefObjects(wlClusterName, "default", "some-uid", wlClusterEndpoint, readFile(wlKubeconfig))...)
				clusterRefs = append(clusterRefs, createFakeClusterRefObjects(mgmtClusterName, searchNamespace, "some-uid", endpoint, readFile(kubeconfig))...)
				issuer = fakeIssuer
				issuerCA = fakeCAData
				conciergeIsClusterScoped = false
				conciergeIsClusterScopedWLCluster = true
				pinnipedInfoCM := fakehelper.PinnipedInfo{
					ClusterName:              mgmtClusterName,
					Issuer:                   issuer,
					IssuerCABundleData:       issuerCA,
					ConciergeIsClusterScoped: conciergeIsClusterScoped,
				}
				clusterRefs = append(clusterRefs, getPinnipedInfoConfigMapObjectFromPinnipedInfo(pinnipedInfoCM))

				pinnipedInfoWorkloadCluster := fakehelper.GetFakePinnipedInfo(fakehelper.PinnipedInfo{
					ConciergeIsClusterScoped: conciergeIsClusterScopedWLCluster,
				})
				searchNamespace = constants.DefaultNamespace
				// create a fake controller-runtime cluster with the []runtime.Object mentioned with createClusterOptions
				fakeClientSet = fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(clusterRefs...).Build()

				tlsServerWLCluster.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/pinniped-info"),
						ghttp.RespondWith(http.StatusOK, pinnipedInfoWorkloadCluster),
					),
				)
			})
			It("should return the cluster pinniped information successfully", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(clusterPinnipedInfo.ClusterName).To(Equal(wlClusterName))
				Expect(clusterPinnipedInfo.ClusterAudience).To(BeNil())
				Expect(clusterPinnipedInfo.ClusterInfo.Server).To(Equal(wlClusterEndpoint))
				Expect(clusterPinnipedInfo.PinnipedInfo.Data.Issuer).To(Equal(issuer))
				Expect(clusterPinnipedInfo.PinnipedInfo.Data.IssuerCABundle).To(Equal(issuerCA))
				Expect(clusterPinnipedInfo.PinnipedInfo.Data.ConciergeIsClusterScoped).To(Equal(conciergeIsClusterScopedWLCluster))
			})
		})

		Context("When we are talking to a pacific cluster", func() {
			const wlClusterUID = "some-pacific-cluster-uid"

			BeforeEach(func() {
				isPacific = true

				var clusterRefs []runtime.Object
				clusterRefs = append(clusterRefs, createFakeClusterRefObjects(wlClusterName, "default", wlClusterUID, wlClusterEndpoint, readFile(wlKubeconfig))...)
				clusterRefs = append(clusterRefs, createFakeClusterRefObjects(mgmtClusterName, searchNamespace, "some-uid", endpoint, readFile(kubeconfig))...)
				issuer = fakeIssuer
				issuerCA = fakeCAData
				conciergeIsClusterScoped = false
				conciergeIsClusterScopedWLCluster = true
				pinnipedInfoCM := fakehelper.PinnipedInfo{
					ClusterName:              mgmtClusterName,
					Issuer:                   issuer,
					IssuerCABundleData:       issuerCA,
					ConciergeIsClusterScoped: conciergeIsClusterScoped,
				}
				clusterRefs = append(clusterRefs, getPinnipedInfoConfigMapObjectFromPinnipedInfo(pinnipedInfoCM))

				pinnipedInfoWorkloadCluster := fakehelper.GetFakePinnipedInfo(fakehelper.PinnipedInfo{
					ConciergeIsClusterScoped: conciergeIsClusterScopedWLCluster,
				})
				searchNamespace = constants.DefaultNamespace
				// create a fake controller-runtime cluster with the []runtime.Object mentioned with createClusterOptions
				fakeClientSet = fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(clusterRefs...).Build()

				tlsServerWLCluster.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/api/v1/namespaces/kube-public/configmaps/pinniped-info"),
						ghttp.RespondWith(http.StatusOK, pinnipedInfoWorkloadCluster),
					),
				)
			})
			It("should return the cluster pinniped information successfully with a special audience", func() {
				wantClusterAudience := fmt.Sprintf("%s-%s", wlClusterName, wlClusterUID)
				Expect(err).ToNot(HaveOccurred())
				Expect(clusterPinnipedInfo.ClusterName).To(Equal(wlClusterName))
				Expect(clusterPinnipedInfo.ClusterAudience).To(Equal(&wantClusterAudience))
				Expect(clusterPinnipedInfo.ClusterInfo.Server).To(Equal(wlClusterEndpoint))
				Expect(clusterPinnipedInfo.PinnipedInfo.Data.Issuer).To(Equal(issuer))
				Expect(clusterPinnipedInfo.PinnipedInfo.Data.IssuerCABundle).To(Equal(issuerCA))
				Expect(clusterPinnipedInfo.PinnipedInfo.Data.ConciergeIsClusterScoped).To(Equal(conciergeIsClusterScopedWLCluster))
			})
		})
	})
})

func getFakeKubeConfigFilePathWithServer(testingDir, endpoint, clustername string, caData []byte) (string, error) {
	kubeconfig := &clientcmdapi.Config{
		Kind:           "Config",
		APIVersion:     clientcmdapi.SchemeGroupVersion.Version,
		Clusters:       map[string]*clientcmdapi.Cluster{clustername: {Server: endpoint, CertificateAuthorityData: caData}}, //nolint:gofmt
		Contexts:       map[string]*clientcmdapi.Context{"dummy-context": {Cluster: clustername}},
		CurrentContext: "dummy-context",
	}
	f, err := os.CreateTemp(testingDir, "kube-*")
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
func createFakeClusterRefObjects(name, namespace, uid, endpoint, kubeconfig string) []runtime.Object {
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
			UID:       types.UID(uid),
		},

		Spec: capi.ClusterSpec{
			ControlPlaneEndpoint: capi.APIEndpoint{
				Host: host,
				Port: int32(portInt32),
			},
		},
	}

	kcp := &controlplanev1.KubeadmControlPlane{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KubeadmControlPlane",
			APIVersion: controlplanev1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    map[string]string{capi.ClusterLabelName: name},
		},
	}

	kubeconfigSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name + "-kubeconfig"},
		Data:       map[string][]byte{"value": []byte(kubeconfig)},
	}

	runtimeObjects := []runtime.Object{}
	runtimeObjects = append(runtimeObjects, cluster, kcp, kubeconfigSecret)
	return runtimeObjects
}

func readFile(path string) string {
	data, err := os.ReadFile(path)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	return string(data)
}

func getPinnipedInfoConfigMapObjectFromPinnipedInfo(pinnipedInfoCM fakehelper.PinnipedInfo) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "kube-public",
			Name:      "pinniped-info",
		},
		Data: map[string]string{
			"cluster_name":                pinnipedInfoCM.ClusterName,
			"issuer":                      pinnipedInfoCM.Issuer,
			"issuer_ca_bundle_data":       pinnipedInfoCM.IssuerCABundleData,
			"concierge_is_cluster_scoped": strconv.FormatBool(pinnipedInfoCM.ConciergeIsClusterScoped),
		},
	}
}
