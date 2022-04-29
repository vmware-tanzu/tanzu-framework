// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package clusterbootstrapclone

import (
	"context"
	"fmt"
	openapi_v2 "github.com/googleapis/gnostic/openapiv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	antreaconfigv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cni/v1alpha1"
	vspherecpiv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cpi/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/version"
	clientgodiscovery "k8s.io/client-go/discovery"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	controllreruntimefake "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Clusterbootstrap", func() {
	var (
		helper            *Helper
		fakeClient        client.Client
		fakeClientSet     *k8sfake.Clientset
		fakeDynamicClient *dynamicfake.FakeDynamicClient
		fakeDiscovery     *ClusterbootstrapFakeDiscovery
		scheme            *runtime.Scheme
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		_ = corev1.AddToScheme(scheme)
		_ = antreaconfigv1alpha1.AddToScheme(scheme)

		fakeClient = controllreruntimefake.NewClientBuilder().WithScheme(scheme).Build()

		fakeClientSet = k8sfake.NewSimpleClientset()
		fakeDiscovery = &ClusterbootstrapFakeDiscovery{
			fakeClientSet.Discovery(),
			[]*metav1.APIResourceList{
				{
					GroupVersion: corev1.SchemeGroupVersion.String(),
					APIResources: []metav1.APIResource{
						{Name: "secrets", Namespaced: true, Kind: "Secret"},
					},
				},
				{
					GroupVersion: antreaconfigv1alpha1.GroupVersion.String(),
					APIResources: []metav1.APIResource{
						{Name: "antreaconfigs", Namespaced: true, Kind: "AntreaConfig"},
					},
				},
				{
					GroupVersion: vspherecpiv1alpha1.GroupVersion.String(),
					APIResources: []metav1.APIResource{
						{Name: "vspherecpiconfigs", Namespaced: true, Kind: "VSphereCPIConfig"},
					},
				},
			},
		}
		helper = &Helper{
			Ctx:                         context.TODO(),
			K8sClient:                   fakeClient,
			AggregateAPIResourcesClient: fakeClient,
			DynamicClient:               fakeDynamicClient,
			DiscoveryClient:             fakeDiscovery,
			Logger:                      ctrl.Log.WithName("clusterbootstrap_test"),
		}
	})

	// Verify cloneEmbeddedLocalObjectRef
	Context("Verify EnsureOwnerRef()", func() {
		BeforeEach(func() {
			fakeDynamicClient = dynamicfake.NewSimpleDynamicClient(scheme, constructFakeAntreaConfig())
			helper.DynamicClient = fakeDynamicClient
		})
		It("should not fail", func() {
			clusterbootstrap := constructFakeClusterBootstrap()
			secrets := []*corev1.Secret{constructFakeSecret()}
			unstructuredObj := convertToUnstructured(constructFakeAntreaConfig())
			providers := []*unstructured.Unstructured{unstructuredObj}

			err := helper.EnsureOwnerRef(clusterbootstrap, secrets, providers)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("Verify cloneEmbeddedLocalObjectRef()", func() {
		BeforeEach(func() {
			fakeDynamicClient = dynamicfake.NewSimpleDynamicClient(scheme, constructFakeAntreaConfig(), constructFakeVSphereCPIConfig(), constructFakeSecret())
			helper.DynamicClient = fakeDynamicClient
		})
		It("should succeed when provider does not have embedded local object reference", func() {
			fakeCluster := constructFakeCluster()
			fakeProvider := convertToUnstructured(constructFakeAntreaConfig())
			err := helper.cloneEmbeddedLocalObjectRef(fakeCluster, fakeProvider)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should succeed when provider has embedded local object reference", func() {
			fakeCluster := constructFakeCluster()
			fakeProvider := convertToUnstructured(constructFakeVSphereCPIConfig())
			err := helper.cloneEmbeddedLocalObjectRef(fakeCluster, fakeProvider)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("Verify cloneProviderRef()", func() {
		BeforeEach(func() {
			fakeDynamicClient = dynamicfake.NewSimpleDynamicClient(scheme, constructFakeAntreaConfig())
			helper.DynamicClient = fakeDynamicClient
		})
		It("", func() {
			cluster := constructFakeCluster()
			antreaClusterbootstrapPackage := constructFakeAntreaClusterBootstrapPackage()
			unstructuredProvider, err := helper.cloneProviderRef(cluster, antreaClusterbootstrapPackage, "antrea.vmware.com", "fake-ns")
			Expect(err).NotTo(HaveOccurred())
			Expect(unstructuredProvider.GetName()).To(Equal(fmt.Sprintf("%s-%s-package", cluster.Name, "antrea.vmware.com")))
		})
	})
})

func convertToUnstructured(obj runtime.Object) *unstructured.Unstructured {
	//convert the runtime.Object to unstructured.Unstructured
	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	Expect(err).NotTo(HaveOccurred())
	return &unstructured.Unstructured{
		Object: unstructuredObj,
	}
}

func constructFakeAntreaClusterBootstrapPackage() *v1alpha3.ClusterBootstrapPackage {
	antreaAPIGroup := antreaconfigv1alpha1.GroupVersion.Group
	antreaConfig := constructFakeAntreaConfig()
	return &v1alpha3.ClusterBootstrapPackage{
		RefName: "fake-antrea-clusterbootstrarp-package",
		ValuesFrom: &v1alpha3.ValuesFrom{
			ProviderRef: &corev1.TypedLocalObjectReference{
				APIGroup: &antreaAPIGroup,
				Kind:     "AntreaConfig",
				Name:     antreaConfig.Name,
			},
		},
	}
}

func constructFakeCluster() *clusterapiv1beta1.Cluster {
	return &clusterapiv1beta1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake-cluster",
			Namespace: "fake-ns",
			UID:       "uid",
		},
		Spec: clusterapiv1beta1.ClusterSpec{},
	}
}

func constructFakeSecret() *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake-secret",
			Namespace: "fake-ns",
		},
		StringData: map[string]string{"key": "value"},
	}
}

func constructFakeClusterBootstrap() *v1alpha3.ClusterBootstrap {
	return &v1alpha3.ClusterBootstrap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake-clusterbootstrap",
			Namespace: "fake-ns",
			UID:       "uid",
		},
	}
}

func constructFakeVSphereCPIConfig() *vspherecpiv1alpha1.VSphereCPIConfig {
	mode := "vsphereCPI"
	emptyStr := ""
	return &vspherecpiv1alpha1.VSphereCPIConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VSphereCPIConfig",
			APIVersion: vspherecpiv1alpha1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake-vspherecpiconfig",
			Namespace: "fake-ns",
		},
		Spec: vspherecpiv1alpha1.VSphereCPIConfigSpec{
			VSphereCPI: vspherecpiv1alpha1.VSphereCPI{
				Mode: &mode,
				NonParavirtualConfig: &vspherecpiv1alpha1.NonParavirtualConfig{
					VSphereCredentialLocalObjRef: &corev1.TypedLocalObjectReference{
						APIGroup: &emptyStr,
						Kind:     "Secret",
						Name:     "fake-secret",
					},
				},
			},
		},
	}
}

func constructFakeAntreaConfig() *antreaconfigv1alpha1.AntreaConfig {
	return &antreaconfigv1alpha1.AntreaConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AntreaConfig",
			APIVersion: antreaconfigv1alpha1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake-antreaconfig",
			Namespace: "fake-ns",
		},
		Spec: antreaconfigv1alpha1.AntreaConfigSpec{
			Antrea: antreaconfigv1alpha1.Antrea{
				AntreaConfigDataValue: antreaconfigv1alpha1.AntreaConfigDataValue{TrafficEncapMode: "encap"},
			},
		},
	}
}

type ClusterbootstrapFakeDiscovery struct {
	discovery clientgodiscovery.DiscoveryInterface
	resources []*metav1.APIResourceList
}

func (c ClusterbootstrapFakeDiscovery) RESTClient() rest.Interface {
	return c.discovery.RESTClient()
}

func (c ClusterbootstrapFakeDiscovery) ServerGroups() (*metav1.APIGroupList, error) {
	return c.discovery.ServerGroups()
}

func (c ClusterbootstrapFakeDiscovery) ServerGroupsAndResources() ([]*metav1.APIGroup, []*metav1.APIResourceList, error) {
	return c.discovery.ServerGroupsAndResources()
}

func (c ClusterbootstrapFakeDiscovery) ServerVersion() (*version.Info, error) {
	return c.discovery.ServerVersion()
}

func (c ClusterbootstrapFakeDiscovery) OpenAPISchema() (*openapi_v2.Document, error) {
	return c.discovery.OpenAPISchema()
}

func (c ClusterbootstrapFakeDiscovery) setFakeServerPreferredResources(resources []*metav1.APIResourceList) {
	c.resources = resources
}

func (c ClusterbootstrapFakeDiscovery) getFakeServerPreferredResources() []*metav1.APIResourceList {
	return c.resources
}

func (c ClusterbootstrapFakeDiscovery) ServerResourcesForGroupVersion(groupVersion string) (*metav1.APIResourceList, error) {
	return c.discovery.ServerResourcesForGroupVersion(groupVersion)
}

func (c ClusterbootstrapFakeDiscovery) ServerResources() ([]*metav1.APIResourceList, error) {
	return c.discovery.ServerResources()
}

func (c ClusterbootstrapFakeDiscovery) ServerPreferredResources() ([]*metav1.APIResourceList, error) {
	return c.getFakeServerPreferredResources(), nil
}

func (c ClusterbootstrapFakeDiscovery) ServerPreferredNamespacedResources() ([]*metav1.APIResourceList, error) {
	return c.discovery.ServerPreferredNamespacedResources()
}
