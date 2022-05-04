// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package clusterbootstrapclone

import (
	"context"
	"fmt"

	openapiv2 "github.com/googleapis/gnostic/openapiv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	kapppkgv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	addontypes "github.com/vmware-tanzu/tanzu-framework/addons/pkg/types"
	antreaconfigv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cni/v1alpha1"
	vspherecpiv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cpi/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
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
		helper                        *Helper
		fakeClient                    client.Client
		fakeClientSet                 *k8sfake.Clientset
		fakeDynamicClient             *dynamicfake.FakeDynamicClient
		fakeDiscovery                 *ClusterbootstrapFakeDiscovery
		scheme                        *runtime.Scheme
		cluster                       *clusterapiv1beta1.Cluster
		antreaClusterbootstrapPackage *v1alpha3.ClusterBootstrapPackage
		fakeAntreaCarvelPkgRefName    = "antrea.vmware.com"
		fakeSourceNamespace           = "fake-ns"
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		_ = corev1.AddToScheme(scheme)
		_ = antreaconfigv1alpha1.AddToScheme(scheme)
		_ = kapppkgv1alpha1.AddToScheme(scheme)
		_ = v1alpha3.AddToScheme(scheme)

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

	Context("Verify EnsureOwnerRef()", func() {
		BeforeEach(func() {
			fakeDynamicClient = dynamicfake.NewSimpleDynamicClient(scheme, constructFakeAntreaConfig())
			helper.DynamicClient = fakeDynamicClient
		})
		It("should succeed to ensure owner references", func() {
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
			cluster = constructFakeCluster()
			antreaClusterbootstrapPackage = constructFakeAntreaClusterBootstrapPackageWithProviderRef()
		})
		It("should fail if provider dose not exist", func() {
			// reset dynamic client
			fakeDynamicClient = dynamicfake.NewSimpleDynamicClient(scheme)
			helper.DynamicClient = fakeDynamicClient

			createdOrUpdatedProvider, err := helper.cloneProviderRef(cluster, antreaClusterbootstrapPackage, fakeAntreaCarvelPkgRefName, fakeSourceNamespace)
			Expect(err).To(HaveOccurred())
			Expect(createdOrUpdatedProvider).To(BeNil())
		})
		It("should create the new provider successfully", func() {
			fakeDynamicClient = dynamicfake.NewSimpleDynamicClient(scheme, constructFakeAntreaConfig())
			helper.DynamicClient = fakeDynamicClient

			createdOrUpdatedProvider, err := helper.cloneProviderRef(cluster, antreaClusterbootstrapPackage, fakeAntreaCarvelPkgRefName, fakeSourceNamespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(createdOrUpdatedProvider.GetName()).To(Equal(fmt.Sprintf("%s-%s-package", cluster.Name, fakeAntreaCarvelPkgRefName)))
			Expect(createdOrUpdatedProvider.GetNamespace()).To(Equal(cluster.Namespace))
		})
	})

	Context("Verify cloneSecretRef()", func() {
		BeforeEach(func() {
			cluster = constructFakeCluster()
			antreaClusterbootstrapPackage = constructFakeAntreaClusterBootstrapPackageWithSecretRef()
		})
		It("should fail if secret dose not exist", func() {
			createdOrUpdatedProvider, err := helper.cloneSecretRef(cluster, antreaClusterbootstrapPackage, fakeAntreaCarvelPkgRefName, fakeSourceNamespace)
			Expect(err).To(HaveOccurred())
			Expect(createdOrUpdatedProvider).To(BeNil())
		})
		It("should success when the referenced secret exists", func() {
			// Create the fake secret
			initSecret := constructFakeSecret()
			err := fakeClient.Create(context.TODO(), initSecret)
			Expect(err).NotTo(HaveOccurred())

			createdOrUpdatedSecret, err := helper.cloneSecretRef(cluster, antreaClusterbootstrapPackage, fakeAntreaCarvelPkgRefName, fakeSourceNamespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(createdOrUpdatedSecret).NotTo(BeNil())
			Expect(createdOrUpdatedSecret.Namespace).To(Equal(cluster.Namespace))
		})
	})

	Context("Verify createSecretFromInline()", func() {
		BeforeEach(func() {
			cluster = constructFakeCluster()
			antreaClusterbootstrapPackage = constructFakeAntreaClusterBootstrapPackageWithInlineRef()
		})
		It("", func() {
			createdSecret, err := helper.createSecretFromInline(cluster, antreaClusterbootstrapPackage, fakeAntreaCarvelPkgRefName)
			Expect(err).NotTo(HaveOccurred())
			Expect(createdSecret).NotTo(BeNil())
			Expect(createdSecret.Namespace).To(Equal(cluster.Namespace))
			Expect(len(createdSecret.OwnerReferences)).NotTo(Equal(0))
			Expect(createdSecret.Labels[addontypes.ClusterNameLabel]).To(Equal(cluster.Name))
			Expect(createdSecret.Type).To(Equal(corev1.SecretType(constants.ClusterBootstrapManagedSecret)))
		})
	})

	Context("Verify cloneReferencedObjectsFromCBPackage()", func() {
		BeforeEach(func() {
			cluster = constructFakeCluster()
		})
		It("should return empty objects with no error when ValuesFrom is nil", func() {
			bootstrapPackage := &v1alpha3.ClusterBootstrapPackage{RefName: fakeAntreaCarvelPkgRefName}
			clonedSecret, clonedProvider, err := helper.cloneReferencedObjectsFromCBPackage(cluster, bootstrapPackage, fakeAntreaCarvelPkgRefName, fakeSourceNamespace)
			Expect(clonedSecret).To(BeNil())
			Expect(clonedProvider).To(BeNil())
			Expect(err).NotTo(HaveOccurred())
		})
		It("should return empty objects with no error when ValuesFrom is empty", func() {
			bootstrapPackage := &v1alpha3.ClusterBootstrapPackage{RefName: fakeAntreaCarvelPkgRefName, ValuesFrom: &v1alpha3.ValuesFrom{}}
			clonedSecret, clonedProvider, err := helper.cloneReferencedObjectsFromCBPackage(cluster, bootstrapPackage, fakeAntreaCarvelPkgRefName, fakeSourceNamespace)
			Expect(clonedSecret).To(BeNil())
			Expect(clonedProvider).To(BeNil())
			Expect(err).NotTo(HaveOccurred())
		})
		It("should success when ValuesFrom.Inline is not empty", func() {
			bootstrapPackage := constructFakeAntreaClusterBootstrapPackageWithInlineRef()
			clonedSecret, clonedProvider, err := helper.cloneReferencedObjectsFromCBPackage(cluster, bootstrapPackage, fakeAntreaCarvelPkgRefName, fakeSourceNamespace)
			Expect(clonedSecret).NotTo(BeNil())
			Expect(clonedProvider).To(BeNil())
			Expect(err).NotTo(HaveOccurred())
		})
		It("should success when ValuesFrom.ProviderRef is not empty", func() {
			bootstrapPackage := constructFakeAntreaClusterBootstrapPackageWithProviderRef()
			clonedSecret, clonedProvider, err := helper.cloneReferencedObjectsFromCBPackage(cluster, bootstrapPackage, fakeAntreaCarvelPkgRefName, fakeSourceNamespace)
			Expect(clonedSecret).To(BeNil())
			Expect(clonedProvider).NotTo(BeNil())
			Expect(err).NotTo(HaveOccurred())
		})
		It("should return error when ValuesFrom.SecretRef is not empty but dose not exist", func() {
			bootstrapPackage := constructFakeAntreaClusterBootstrapPackageWithSecretRef()
			clonedSecret, clonedProvider, err := helper.cloneReferencedObjectsFromCBPackage(cluster, bootstrapPackage, fakeAntreaCarvelPkgRefName, fakeSourceNamespace)
			Expect(clonedSecret).To(BeNil())
			Expect(clonedProvider).To(BeNil())
			Expect(apierrors.IsNotFound(err)).To(BeTrue())
		})
	})

	Context("Verify CloneReferencedObjectsFromCBPackages()", func() {
		BeforeEach(func() {
			cluster = constructFakeCluster()
		})
		It("should return empty objects with no error when CBPackages is empty", func() {
			clonedSecrets, clonedProviders, err := helper.CloneReferencedObjectsFromCBPackages(cluster, []*v1alpha3.ClusterBootstrapPackage{}, fakeSourceNamespace)
			Expect(clonedSecrets).To(BeNil())
			Expect(clonedProviders).To(BeNil())
			Expect(err).NotTo(HaveOccurred())
		})
		It("should return empty objects with error when no carvel package metadata is found", func() {
			clonedSecrets, clonedProviders, err := helper.CloneReferencedObjectsFromCBPackages(cluster,
				[]*v1alpha3.ClusterBootstrapPackage{constructFakeAntreaClusterBootstrapPackageWithProviderRef()},
				fakeSourceNamespace)
			Expect(clonedSecrets).To(BeNil())
			Expect(clonedProviders).To(BeNil())
			Expect(err).To(HaveOccurred())
		})
		It("should return nil clonedSecrets and non-nil clonedProviders when carvel package metadata is found and CB"+
			" Package has providerRef", func() {
			err := fakeClient.Create(context.TODO(), constructAntreaCarvelPackage(cluster.Namespace))
			Expect(err).To(BeNil())
			clonedSecrets, clonedProviders, err := helper.CloneReferencedObjectsFromCBPackages(cluster,
				[]*v1alpha3.ClusterBootstrapPackage{constructFakeAntreaClusterBootstrapPackageWithProviderRef()},
				fakeSourceNamespace)
			Expect(clonedSecrets).To(BeNil())
			Expect(len(clonedProviders)).To(Equal(1))
			Expect(err).NotTo(HaveOccurred())
		})
		It("should return nil clonedSecrets and non-nil clonedProviders when carvel package metadata is found and CB"+
			" Package has secretRef", func() {
			// Create a fake secret
			initSecret := constructFakeSecret()
			err := fakeClient.Create(context.TODO(), initSecret)
			Expect(err).NotTo(HaveOccurred())
			// Create a fake antrea carvel package
			err = fakeClient.Create(context.TODO(), constructAntreaCarvelPackage(cluster.Namespace))
			Expect(err).NotTo(HaveOccurred())

			clonedSecrets, clonedProviders, err := helper.CloneReferencedObjectsFromCBPackages(cluster,
				[]*v1alpha3.ClusterBootstrapPackage{constructFakeAntreaClusterBootstrapPackageWithSecretRef()},
				fakeSourceNamespace)
			Expect(len(clonedSecrets)).To(Equal(1))
			Expect(clonedProviders).To(BeNil())
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("Verify CreateClusterBootstrapFromTemplate()", func() {
		var (
			clusterbootstrapTemplate *v1alpha3.ClusterBootstrapTemplate
		)
		BeforeEach(func() {
			cluster = constructFakeCluster()
			clusterbootstrapTemplate = constructFakeClusterBootstrapTemplate()
		})
		It("should return clusterbootstrap without error", func() {
			// Create a fake antrea carvel package
			err := fakeClient.Create(context.TODO(), constructAntreaCarvelPackage(cluster.Namespace))
			Expect(err).NotTo(HaveOccurred())

			clusterbootstrap, err := helper.CreateClusterBootstrapFromTemplate(clusterbootstrapTemplate, cluster, "fake-tkr-name")
			Expect(err).NotTo(HaveOccurred())
			Expect(clusterbootstrap).NotTo(BeNil())
			Expect(len(clusterbootstrap.OwnerReferences)).To(Equal(1))
			Expect(clusterbootstrap.Status.ResolvedTKR).To(Equal("fake-tkr-name"))
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

func constructFakeClusterBootstrapTemplate() *v1alpha3.ClusterBootstrapTemplate {
	return &v1alpha3.ClusterBootstrapTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake-clusterbootstrap-template",
			Namespace: "fake-ns",
		},
		Spec: &v1alpha3.ClusterBootstrapTemplateSpec{
			CNI: constructFakeAntreaClusterBootstrapPackageWithProviderRef(),
		},
	}
}

var AntreaCBPackageRefName = "fake-antrea-clusterbootstrarp-package"

func constructAntreaCarvelPackage(namespace string) *kapppkgv1alpha1.Package {
	return &kapppkgv1alpha1.Package{
		ObjectMeta: metav1.ObjectMeta{
			Name:      AntreaCBPackageRefName,
			Namespace: namespace,
		},
		Spec: kapppkgv1alpha1.PackageSpec{
			RefName: "antrea.vmware.com",
		},
	}
}

func constructFakeAntreaClusterBootstrapPackageWithSecretRef() *v1alpha3.ClusterBootstrapPackage {
	return &v1alpha3.ClusterBootstrapPackage{
		RefName: AntreaCBPackageRefName,
		ValuesFrom: &v1alpha3.ValuesFrom{
			SecretRef: "fake-secret",
		},
	}
}

func constructFakeAntreaClusterBootstrapPackageWithInlineRef() *v1alpha3.ClusterBootstrapPackage {
	return &v1alpha3.ClusterBootstrapPackage{
		RefName: AntreaCBPackageRefName,
		ValuesFrom: &v1alpha3.ValuesFrom{
			Inline: map[string]interface{}{"foo": "bar"},
		},
	}
}

func constructFakeAntreaClusterBootstrapPackageWithProviderRef() *v1alpha3.ClusterBootstrapPackage {
	antreaAPIGroup := antreaconfigv1alpha1.GroupVersion.Group
	antreaConfig := constructFakeAntreaConfig()
	return &v1alpha3.ClusterBootstrapPackage{
		RefName: AntreaCBPackageRefName,
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
			Namespace: "fake-cluster-ns",
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
			Namespace: "fake-cluster-ns",
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

// ClusterbootstrapFakeDiscovery customize the behavior of fake client-go FakeDiscovery.ServerPreferredResources to return
// a customized APIResourceList.
// The client-go FakeDiscovery.ServerPreferredResources is hardcoded to return nil.
// https://github.com/kubernetes/client-go/blob/master/discovery/fake/discovery.go#L85
type ClusterbootstrapFakeDiscovery struct {
	fakeDiscovery clientgodiscovery.DiscoveryInterface
	resources     []*metav1.APIResourceList
}

func (c ClusterbootstrapFakeDiscovery) RESTClient() rest.Interface {
	return c.fakeDiscovery.RESTClient()
}

func (c ClusterbootstrapFakeDiscovery) ServerGroups() (*metav1.APIGroupList, error) {
	return c.fakeDiscovery.ServerGroups()
}

func (c ClusterbootstrapFakeDiscovery) ServerGroupsAndResources() ([]*metav1.APIGroup, []*metav1.APIResourceList, error) {
	return c.fakeDiscovery.ServerGroupsAndResources()
}

func (c ClusterbootstrapFakeDiscovery) ServerVersion() (*version.Info, error) {
	return c.fakeDiscovery.ServerVersion()
}

func (c ClusterbootstrapFakeDiscovery) OpenAPISchema() (*openapiv2.Document, error) {
	return c.fakeDiscovery.OpenAPISchema()
}

func (c ClusterbootstrapFakeDiscovery) getFakeServerPreferredResources() []*metav1.APIResourceList {
	return c.resources
}

func (c ClusterbootstrapFakeDiscovery) ServerResourcesForGroupVersion(groupVersion string) (*metav1.APIResourceList, error) {
	return c.fakeDiscovery.ServerResourcesForGroupVersion(groupVersion)
}

// Having nolint below to get rid of the complaining on the deprecation of ServerResources. We have to have the following
// function to customize the DiscoveryInterface
//nolint:staticcheck
func (c ClusterbootstrapFakeDiscovery) ServerResources() ([]*metav1.APIResourceList, error) {
	return c.fakeDiscovery.ServerResources()
}

func (c ClusterbootstrapFakeDiscovery) ServerPreferredResources() ([]*metav1.APIResourceList, error) {
	return c.getFakeServerPreferredResources(), nil
}

func (c ClusterbootstrapFakeDiscovery) ServerPreferredNamespacedResources() ([]*metav1.APIResourceList, error) {
	return c.fakeDiscovery.ServerPreferredNamespacedResources()
}
