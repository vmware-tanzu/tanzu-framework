// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package clusterbootstrapclone

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	controllreruntimefake "sigs.k8s.io/controller-runtime/pkg/client/fake"
	k8syaml "sigs.k8s.io/yaml"

	kapppkgv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	addontypes "github.com/vmware-tanzu/tanzu-framework/addons/pkg/types"
	"github.com/vmware-tanzu/tanzu-framework/addons/test/testutil"
	antreaconfigv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/addonconfigs/cni/v1alpha1"
	vspherecpiv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/addonconfigs/cpi/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

var (
	fakeAntreaCBPackageRefName        = "fake-antrea-clusterbootstrarp-package"
	fakeCSICBPackageRefName           = "fake-vsphere-csi-clusterbootstrarp-package"
	fakePinnipedCBPackageRefName      = "fake-pinniped-clusterbootstrarp-package"
	fakeMetricsServerCBPackageRefName = "fake-metrics-server-clusterbootstrarp-package"
)

var _ = Describe("ClusterbootstrapClone", func() {
	var (
		helper                        *Helper
		fakeClient                    client.Client
		fakeClientSet                 *k8sfake.Clientset
		fakeDynamicClient             *dynamicfake.FakeDynamicClient
		fakeDiscovery                 *testutil.FakeDiscovery
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
		fakeDiscovery = &testutil.FakeDiscovery{
			FakeDiscovery: fakeClientSet.Discovery(),
			Resources: []*metav1.APIResourceList{
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
		gvrHelper := &testutil.FakeGVRHelper{DiscoveryClient: fakeDiscovery}
		helper = &Helper{
			Ctx:                         context.TODO(),
			K8sClient:                   fakeClient,
			AggregateAPIResourcesClient: fakeClient,
			DynamicClient:               fakeDynamicClient,
			GVRHelper:                   gvrHelper,
			Logger:                      ctrl.Log.WithName("clusterbootstrap_test"),
		}
	})

	Context("Verify EnsureOwnerRef()", func() {
		BeforeEach(func() {
			fakeDynamicClient = dynamicfake.NewSimpleDynamicClient(scheme, constructFakeAntreaConfig())
			helper.DynamicClient = fakeDynamicClient
		})
		It("should succeed to ensure owner references", func() {
			clusterbootstrap := constructFakeEmptyClusterBootstrap()
			secrets := []*corev1.Secret{constructFakeSecret()}
			unstructuredObj := convertToUnstructured(constructFakeAntreaConfig())
			providers := []*unstructured.Unstructured{unstructuredObj}

			err := helper.EnsureOwnerRef(clusterbootstrap, secrets, providers)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("Verify AddClusterOwnerRef()", func() {
		BeforeEach(func() {
			fakeDynamicClient = dynamicfake.NewSimpleDynamicClient(scheme, constructFakeAntreaConfig(),
				constructFakeAntreaConfigWithClusterOwner("same-cluster-config", "same-cluster"),
			)
			helper.DynamicClient = fakeDynamicClient
		})
		It("should succeed on adding cluster as owner to orphan config", func() {
			fakeCluster := constructFakeCluster()
			fakeConfig := constructFakeAntreaConfig()
			fakeProvider := convertToUnstructured(fakeConfig)
			children := []*unstructured.Unstructured{fakeProvider}
			err := helper.AddClusterOwnerRef(fakeCluster, children, nil, nil)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should succeed on adding cluster as owner to config which already lists same cluster as owner", func() {
			fakeCluster := constructNamespacedFakeCluster("same-cluster", "fake-cluster-ns")
			fakeConfig := constructFakeAntreaConfigWithClusterOwner("same-cluster-config", fakeCluster.Name)
			fakeProvider := convertToUnstructured(fakeConfig)
			children := []*unstructured.Unstructured{fakeProvider}
			err := helper.AddClusterOwnerRef(fakeCluster, children, nil, nil)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should throw error if children has different cluster owner", func() {
			fakeCluster := constructNamespacedFakeCluster("same-cluster", "fake-cluster-ns")
			fakeConfig := constructFakeAntreaConfigWithClusterOwner("other-other-config", "other-cluster")
			fakeProvider := convertToUnstructured(fakeConfig)
			children := []*unstructured.Unstructured{fakeProvider}
			err := helper.AddClusterOwnerRef(fakeCluster, children, nil, nil)
			Expect(err).To(HaveOccurred())
		})
		It("should throw error if any of the children has different cluster owner", func() {
			fakeCluster := constructNamespacedFakeCluster("same-cluster", "fake-cluster-ns")
			fakeConfig := constructFakeAntreaConfigWithClusterOwner("same-cluster-config", fakeCluster.Name)
			fakeProvider := convertToUnstructured(fakeConfig)

			fakeConfig2 := constructFakeAntreaConfigWithClusterOwner("another-cluster-config", "some-other-cluster")
			fakeProvider2 := convertToUnstructured(fakeConfig2)

			children := []*unstructured.Unstructured{fakeProvider, fakeProvider2}
			err := helper.AddClusterOwnerRef(fakeCluster, children, nil, nil)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Verify AddclusterOwnerRefToExistingProviders()", func() {
		It("should succeed on adding cluster as owner to existing orphan config", func() {
			fakeCluster := constructFakeCluster()
			fakeBootstrap := constructFakeClusterBootstrap()
			fakeDynamicClient = dynamicfake.NewSimpleDynamicClient(scheme, constructFakeAntreaConfig())
			helper.DynamicClient = fakeDynamicClient

			err := helper.AddClusterOwnerRefToExistingProviders(fakeCluster, fakeBootstrap)
			Expect(err).ToNot(HaveOccurred())
		})
		It("should succeed on adding cluster as owner to config which already lists same cluster as owner", func() {
			fakeBootstrap := constructFakeClusterBootstrap()
			fakeAntreaConfig := constructFakeAntreaConfig()
			fakeCluster := constructNamespacedFakeCluster("same-cluster", "fake-ns")
			fakeAntreaConfigWithOwner := constructFakeAntreaConfigWithClusterOwner(fakeAntreaConfig.Name, fakeCluster.Name)
			fakeDynamicClient = dynamicfake.NewSimpleDynamicClient(scheme, fakeAntreaConfigWithOwner)
			helper.DynamicClient = fakeDynamicClient

			err := helper.AddClusterOwnerRefToExistingProviders(fakeCluster, fakeBootstrap)
			Expect(err).ToNot(HaveOccurred())
		})
		It("should throw error if config has different cluster owner", func() {
			fakeBootstrap := constructFakeClusterBootstrap()
			fakeAntreaConfig := constructFakeAntreaConfig()
			fakeCluster := constructNamespacedFakeCluster("same-cluster", "fake-ns")
			fakeAntreaConfigWithOwner := constructFakeAntreaConfigWithClusterOwner(fakeAntreaConfig.Name, "someothercluster")
			fakeDynamicClient = dynamicfake.NewSimpleDynamicClient(scheme, fakeAntreaConfigWithOwner)
			helper.DynamicClient = fakeDynamicClient

			err := helper.AddClusterOwnerRefToExistingProviders(fakeCluster, fakeBootstrap)
			Expect(err).To(HaveOccurred())
		})
		It("should handle non existing config ", func() {
			fakeBootstrap := constructFakeClusterBootstrap()
			fakeCluster := constructNamespacedFakeCluster("same-cluster", "fake-ns")
			fakeDynamicClient = dynamicfake.NewSimpleDynamicClient(scheme)
			helper.DynamicClient = fakeDynamicClient

			err := helper.AddClusterOwnerRefToExistingProviders(fakeCluster, fakeBootstrap)
			Expect(err).ToNot(HaveOccurred())
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
			antreaClusterbootstrapPackage = constructFakeClusterBootstrapPackageWithAntreaProviderRef()
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
			antreaConfig := constructFakeAntreaConfig()
			fakeDynamicClient = dynamicfake.NewSimpleDynamicClient(scheme, antreaConfig)
			helper.DynamicClient = fakeDynamicClient

			createdOrUpdatedProvider, err := helper.cloneProviderRef(cluster, antreaClusterbootstrapPackage, fakeAntreaCarvelPkgRefName, fakeSourceNamespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(createdOrUpdatedProvider.GetName()).To(Equal(fmt.Sprintf("%s-%s-package", cluster.Name, "antrea")))
			Expect(createdOrUpdatedProvider.GetNamespace()).To(Equal(cluster.Namespace))
			// previous annotations should be removed after clone
			Expect(antreaConfig.Annotations).Should(HaveKey(constants.TKGAnnotationTemplateConfig))
			Expect(createdOrUpdatedProvider.GetAnnotations()).ShouldNot(HaveKey(constants.TKGAnnotationTemplateConfig))
		})
	})

	Context("Verify cloneSecretRef()", func() {
		BeforeEach(func() {
			cluster = constructFakeCluster()
			antreaClusterbootstrapPackage = constructFakeClusterBootstrapPackageWithSecretRef()
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
			antreaClusterbootstrapPackage = constructFakeClusterBootstrapPackageWithInlineRef()
		})
		It("", func() {
			createdSecret, err := helper.CreateSecretFromInline(cluster, antreaClusterbootstrapPackage, fakeAntreaCarvelPkgRefName)
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
			bootstrapPackage := constructFakeClusterBootstrapPackageWithInlineRef()
			clonedSecret, clonedProvider, err := helper.cloneReferencedObjectsFromCBPackage(cluster, bootstrapPackage, fakeAntreaCarvelPkgRefName, fakeSourceNamespace)
			Expect(clonedSecret).NotTo(BeNil())
			Expect(clonedProvider).To(BeNil())
			Expect(err).NotTo(HaveOccurred())
		})
		It("should success when ValuesFrom.ProviderRef is not empty", func() {
			bootstrapPackage := constructFakeClusterBootstrapPackageWithAntreaProviderRef()
			clonedSecret, clonedProvider, err := helper.cloneReferencedObjectsFromCBPackage(cluster, bootstrapPackage, fakeAntreaCarvelPkgRefName, fakeSourceNamespace)
			Expect(clonedSecret).To(BeNil())
			Expect(clonedProvider).NotTo(BeNil())
			Expect(err).NotTo(HaveOccurred())
		})
		It("should return error when ValuesFrom.SecretRef is not empty but dose not exist", func() {
			bootstrapPackage := constructFakeClusterBootstrapPackageWithSecretRef()
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
				[]*v1alpha3.ClusterBootstrapPackage{constructFakeClusterBootstrapPackageWithAntreaProviderRef()},
				fakeSourceNamespace)
			Expect(clonedSecrets).To(BeNil())
			Expect(clonedProviders).To(BeNil())
			Expect(err).To(HaveOccurred())
		})
		It("should return nil clonedSecrets and non-nil clonedProviders when carvel package metadata is found and CB"+
			" Package has providerRef", func() {
			prepareCarvelPackages(fakeClient, cluster.Namespace)
			clonedSecrets, clonedProviders, err := helper.CloneReferencedObjectsFromCBPackages(cluster,
				[]*v1alpha3.ClusterBootstrapPackage{constructFakeClusterBootstrapPackageWithAntreaProviderRef()},
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
			prepareCarvelPackages(fakeClient, cluster.Namespace)

			clonedSecrets, clonedProviders, err := helper.CloneReferencedObjectsFromCBPackages(cluster,
				[]*v1alpha3.ClusterBootstrapPackage{constructFakeClusterBootstrapPackageWithSecretRef()},
				fakeSourceNamespace)
			Expect(len(clonedSecrets)).To(Equal(1))
			Expect(clonedProviders).To(BeNil())
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("Verify HandleExistingClusterBootstrap()", func() {
		var (
			clusterbootstrapTemplate *v1alpha3.ClusterBootstrapTemplate
		)
		BeforeEach(func() {
			cluster = constructFakeCluster()
			clusterbootstrapTemplate = constructFakeClusterBootstrapTemplate()
		})
		It("should return clusterbootstrap without error", func() {

			clusterBootstrap := constructFakeEmptyClusterBootstrap()
			clusterBootstrap.SetAnnotations(map[string]string{
				constants.AddCBMissingFieldsAnnotationKey: clusterbootstrapTemplate.Name,
			})
			clusterBootstrap.Spec = &v1alpha3.ClusterBootstrapTemplateSpec{
				CNI: &v1alpha3.ClusterBootstrapPackage{
					RefName: clusterbootstrapTemplate.Spec.CNI.RefName,
				},
				CSI: &v1alpha3.ClusterBootstrapPackage{
					RefName: clusterbootstrapTemplate.Spec.CSI.RefName,
				},
			}
			prepareCarvelPackages(fakeClient, cluster.Namespace)
			Expect(fakeClient.Create(context.TODO(), clusterbootstrapTemplate)).To(Succeed())
			Expect(fakeClient.Create(context.TODO(), clusterBootstrap)).To(Succeed())

			clusterBootstrap, err := helper.HandleExistingClusterBootstrap(clusterBootstrap, cluster, clusterbootstrapTemplate.Name, clusterbootstrapTemplate.Namespace)
			Expect(err).NotTo(HaveOccurred())
			Expect(clusterBootstrap).NotTo(BeNil())

			// OwnerReference is supposed to be set
			Expect(clusterBootstrap.OwnerReferences).NotTo(BeEmpty())
			// ValuesFrom is supposed to be added
			Expect(clusterBootstrap.Spec.CNI.ValuesFrom).NotTo(BeNil())
			Expect(clusterBootstrap.Spec.CSI.ValuesFrom).NotTo(BeNil())
			// Status.ResolvedTKR is supposed to be set
			Expect(clusterBootstrap.Status.ResolvedTKR).To(Equal(clusterbootstrapTemplate.Name))
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
			prepareCarvelPackages(fakeClient, cluster.Namespace)

			clusterbootstrap, err := helper.CreateClusterBootstrapFromTemplate(clusterbootstrapTemplate, cluster, "fake-tkr-name")
			Expect(err).NotTo(HaveOccurred())
			Expect(clusterbootstrap).NotTo(BeNil())

			Expect(clusterbootstrap.Spec.CNI.RefName).To(Equal(clusterbootstrapTemplate.Spec.CNI.RefName))
			Expect(clusterbootstrap.Spec.CNI.ValuesFrom).NotTo(BeNil())
			Expect(clusterbootstrap.Spec.CSI.RefName).To(Equal(clusterbootstrapTemplate.Spec.CSI.RefName))
			Expect(clusterbootstrap.Spec.CSI.ValuesFrom).NotTo(BeNil())

			Expect(len(clusterbootstrap.OwnerReferences)).To(Equal(1))
			Expect(clusterbootstrap.Status.ResolvedTKR).To(Equal("fake-tkr-name"))
		})

	})

	Context("Verify AddMissingSpecFieldsFromTemplate()", func() {
		var fakeClusterBootstrapTemplate *v1alpha3.ClusterBootstrapTemplate
		BeforeEach(func() {
			fakeClusterBootstrapTemplate = constructFakeClusterBootstrapTemplate()
		})
		It("should add what ClusterBootstrapTemplate has to the empty ClusterBootstrap", func() {
			clusterBootstrap := &v1alpha3.ClusterBootstrap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "fake-clusterbootstrap",
					Namespace: "fake-cluster-ns",
					UID:       "uid",
				},
				Spec: &v1alpha3.ClusterBootstrapTemplateSpec{},
			}

			err := helper.AddMissingSpecFieldsFromTemplate(fakeClusterBootstrapTemplate, clusterBootstrap, nil)
			Expect(err).NotTo(HaveOccurred())
			updatedClusterBootstrap := clusterBootstrap
			Expect(updatedClusterBootstrap.Spec).NotTo(BeNil())
			// Expect CNI to be added
			Expect(updatedClusterBootstrap.Spec.CNI).NotTo(BeNil())
			Expect(updatedClusterBootstrap.Spec.CNI.RefName).To(Equal(fakeClusterBootstrapTemplate.Spec.CNI.RefName))
			Expect(updatedClusterBootstrap.Spec.CNI.ValuesFrom.ProviderRef.Name).To(Equal(fakeClusterBootstrapTemplate.Spec.CNI.ValuesFrom.ProviderRef.Name))
			Expect(updatedClusterBootstrap.Spec.CNI.ValuesFrom.ProviderRef.Kind).To(Equal(fakeClusterBootstrapTemplate.Spec.CNI.ValuesFrom.ProviderRef.Kind))
			// Expect CSI to be added
			Expect(updatedClusterBootstrap.Spec.CSI).NotTo(BeNil())
			Expect(updatedClusterBootstrap.Spec.CSI.RefName).To(Equal(fakeClusterBootstrapTemplate.Spec.CSI.RefName))
			Expect(updatedClusterBootstrap.Spec.CSI.ValuesFrom).NotTo(BeNil())
			Expect(len(updatedClusterBootstrap.Spec.CSI.ValuesFrom.Inline)).To(Equal(len(fakeClusterBootstrapTemplate.Spec.CSI.ValuesFrom.Inline)))
			// Spec.Paused is not set in fakeClusterBootstrapTemplate, it should be false
			Expect(updatedClusterBootstrap.Spec.Paused).To(BeFalse())
			// The ClusterBootstrapPackage not set in fakeClusterBootstrapTemplate. They should not be copied
			Expect(updatedClusterBootstrap.Spec.CPI).To(BeNil())
			Expect(updatedClusterBootstrap.Spec.Kapp).To(BeNil())
			Expect(len(updatedClusterBootstrap.Spec.AdditionalPackages)).To(Equal(len(fakeClusterBootstrapTemplate.Spec.AdditionalPackages)))
		})

		It("should not overwrite the components which already exist", func() {
			antreaAPIGroup := antreaconfigv1alpha1.GroupVersion.Group
			fakeCPIClusterBootstrapPackage := constructFakeClusterBootstrapPackageWithSecretRef()
			fakeCSIClusterBootstrapPackage := constructFakeClusterBootstrapPackageWithInlineRef()
			// Update fakeClusterBootstrapTemplate by adding a fake CPI and CSI package
			fakeClusterBootstrapTemplate.Spec.CPI = fakeCPIClusterBootstrapPackage
			fakeClusterBootstrapTemplate.Spec.CSI = fakeCSIClusterBootstrapPackage

			clusterBootstrap := &v1alpha3.ClusterBootstrap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "fake-clusterbootstrap",
					Namespace: "fake-cluster-ns",
					UID:       "uid",
				},
				Spec: &v1alpha3.ClusterBootstrapTemplateSpec{
					// We do not expect this part to be overwritten to be what fakeClusterBootstrapTemplate has
					CNI: &v1alpha3.ClusterBootstrapPackage{
						RefName: "foo-antrea-clusterbootstrarp-package",
						ValuesFrom: &v1alpha3.ValuesFrom{
							ProviderRef: &corev1.TypedLocalObjectReference{
								APIGroup: &antreaAPIGroup,
								Kind:     "AntreaConfig",
								Name:     "fooAntreaConfig",
							},
						},
					},
					CSI: &v1alpha3.ClusterBootstrapPackage{
						RefName: "foo-vsphere-csi-clusterbootstrarp-package",
						ValuesFrom: &v1alpha3.ValuesFrom{
							Inline: map[string]interface{}{"should-not-be-updated": true},
						},
					},
					AdditionalPackages: []*v1alpha3.ClusterBootstrapPackage{
						{RefName: fakePinnipedCBPackageRefName, ValuesFrom: &v1alpha3.ValuesFrom{Inline: map[string]interface{}{"identity_management_type": "oidc"}}},
					},
				},
			}

			err := helper.AddMissingSpecFieldsFromTemplate(fakeClusterBootstrapTemplate, clusterBootstrap, nil)
			Expect(err).NotTo(HaveOccurred())
			updatedClusterBootstrap := clusterBootstrap
			Expect(updatedClusterBootstrap.Spec.CNI).NotTo(BeNil())
			// We do not expect the RefName and ValuesFrom gets overwritten if they already exist
			Expect(updatedClusterBootstrap.Spec.CNI.RefName).To(Equal("foo-antrea-clusterbootstrarp-package"))
			Expect(updatedClusterBootstrap.Spec.CNI.ValuesFrom.ProviderRef.Kind).To(Equal("AntreaConfig"))
			Expect(updatedClusterBootstrap.Spec.CNI.ValuesFrom.ProviderRef.Name).To(Equal("fooAntreaConfig"))
			Expect(updatedClusterBootstrap.Spec.CSI.RefName).To(Equal("foo-vsphere-csi-clusterbootstrarp-package"))
			Expect(len(updatedClusterBootstrap.Spec.CSI.ValuesFrom.Inline)).To(Equal(1))
			Expect(updatedClusterBootstrap.Spec.CSI.ValuesFrom.Inline["should-not-be-updated"]).To(BeTrue())
			// CPI should be added to updatedClusterBootstrap
			Expect(updatedClusterBootstrap.Spec.CPI).NotTo(BeNil())
			Expect(updatedClusterBootstrap.Spec.CPI.ValuesFrom.SecretRef).To(Equal(fakeCPIClusterBootstrapPackage.ValuesFrom.SecretRef))
			// The ClusterBootstrapPackage not set in fakeClusterBootstrapTemplate. They should not be copied
			Expect(updatedClusterBootstrap.Spec.Kapp).To(BeNil())
			Expect(len(updatedClusterBootstrap.Spec.AdditionalPackages)).To(Equal(len(fakeClusterBootstrapTemplate.Spec.AdditionalPackages)))
			for idx := range updatedClusterBootstrap.Spec.AdditionalPackages {
				Expect(updatedClusterBootstrap.Spec.AdditionalPackages[idx].RefName).To(Equal(fakeClusterBootstrapTemplate.Spec.AdditionalPackages[idx].RefName))
				Expect(updatedClusterBootstrap.Spec.AdditionalPackages[idx].ValuesFrom).To(Equal(fakeClusterBootstrapTemplate.Spec.AdditionalPackages[idx].ValuesFrom))
			}
		})

		It("should add valuesFrom back if not specified in ClusterBootstrap", func() {
			fakeCSIClusterBootstrapPackage := constructFakeClusterBootstrapPackageWithInlineRef()
			// Update fakeClusterBootstrapTemplate by adding a fake CSI package
			fakeClusterBootstrapTemplate.Spec.CSI = fakeCSIClusterBootstrapPackage

			clusterBootstrap := &v1alpha3.ClusterBootstrap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "fake-clusterbootstrap",
					Namespace: "fake-cluster-ns",
					UID:       "uid",
				},
				Spec: &v1alpha3.ClusterBootstrapTemplateSpec{
					CSI: &v1alpha3.ClusterBootstrapPackage{
						RefName: "foo-vsphere-csi-clusterbootstrarp-package",
					},
					AdditionalPackages: []*v1alpha3.ClusterBootstrapPackage{
						{RefName: fakePinnipedCBPackageRefName},
					},
				},
			}
			err := helper.AddMissingSpecFieldsFromTemplate(fakeClusterBootstrapTemplate, clusterBootstrap, nil)
			Expect(err).NotTo(HaveOccurred())
			updatedClusterBootstrap := clusterBootstrap

			// CNI exists in fakeClusterBootstrapTemplate, it should be added back
			Expect(updatedClusterBootstrap.Spec.CNI).NotTo(BeNil())
			// CSI in fakeClusterBootstrapTemplate has valuesFrom, so valuesFrom should be added back
			Expect(updatedClusterBootstrap.Spec.CSI.ValuesFrom).NotTo(BeNil())
			Expect(updatedClusterBootstrap.Spec.CSI.ValuesFrom.Inline).NotTo(BeNil())
			assertTwoMapsShouldEqual(updatedClusterBootstrap.Spec.CSI.ValuesFrom.Inline, fakeClusterBootstrapTemplate.Spec.CSI.ValuesFrom.Inline)

			Expect(len(updatedClusterBootstrap.Spec.AdditionalPackages)).To(Equal(len(fakeClusterBootstrapTemplate.Spec.AdditionalPackages)))
			for idx := range updatedClusterBootstrap.Spec.AdditionalPackages {
				Expect(updatedClusterBootstrap.Spec.AdditionalPackages[idx].RefName).To(Equal(fakeClusterBootstrapTemplate.Spec.AdditionalPackages[idx].RefName))
				Expect(updatedClusterBootstrap.Spec.AdditionalPackages[idx].ValuesFrom).To(Equal(fakeClusterBootstrapTemplate.Spec.AdditionalPackages[idx].ValuesFrom))
			}
		})

		It("should not add the fields which are meant to be skipped", func() {
			antreaAPIGroup := antreaconfigv1alpha1.GroupVersion.Group
			fakeCPIClusterBootstrapPackage := constructFakeClusterBootstrapPackageWithSecretRef()
			fakeCSIClusterBootstrapPackage := constructFakeClusterBootstrapPackageWithInlineRef()
			// Update fakeClusterBootstrapTemplate by adding a fake CPI and CSI package
			fakeClusterBootstrapTemplate.Spec.CPI = fakeCPIClusterBootstrapPackage
			fakeClusterBootstrapTemplate.Spec.CSI = fakeCSIClusterBootstrapPackage

			clusterBootstrap := &v1alpha3.ClusterBootstrap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "fake-clusterbootstrap",
					Namespace: "fake-cluster-ns",
					UID:       "uid",
				},
				Spec: &v1alpha3.ClusterBootstrapTemplateSpec{
					// We do not expect this part to be overwritten to be what fakeClusterBootstrapTemplate has
					CNI: &v1alpha3.ClusterBootstrapPackage{
						RefName: "foo-antrea-clusterbootstrarp-package",
						ValuesFrom: &v1alpha3.ValuesFrom{
							ProviderRef: &corev1.TypedLocalObjectReference{
								APIGroup: &antreaAPIGroup,
								Kind:     "AntreaConfig",
								Name:     "fooAntreaConfig",
							},
						},
					},
					CSI: &v1alpha3.ClusterBootstrapPackage{
						RefName: "foo-vsphere-csi-clusterbootstrarp-package",
						ValuesFrom: &v1alpha3.ValuesFrom{
							Inline: map[string]interface{}{"should-not-be-updated": true},
						},
					},
					AdditionalPackages: []*v1alpha3.ClusterBootstrapPackage{
						{RefName: fakePinnipedCBPackageRefName},
					},
				},
			}

			err := helper.AddMissingSpecFieldsFromTemplate(fakeClusterBootstrapTemplate, clusterBootstrap, map[string]interface{}{"valuesFrom": nil})
			Expect(err).NotTo(HaveOccurred())
			updatedClusterBootstrap := clusterBootstrap
			// CPI should be added to updatedClusterBootstrap
			Expect(updatedClusterBootstrap.Spec.CPI).NotTo(BeNil())
			// CPI's valuesFrom should be skipped
			Expect(updatedClusterBootstrap.Spec.CPI.ValuesFrom).To(BeNil())
			Expect(updatedClusterBootstrap.Spec.CPI.RefName).To(Equal(fakeClusterBootstrapTemplate.Spec.CPI.RefName))

			Expect(len(updatedClusterBootstrap.Spec.AdditionalPackages)).To(Equal(len(fakeClusterBootstrapTemplate.Spec.AdditionalPackages)))
			for idx := range updatedClusterBootstrap.Spec.AdditionalPackages {
				Expect(updatedClusterBootstrap.Spec.AdditionalPackages[idx].RefName).To(Equal(fakeClusterBootstrapTemplate.Spec.AdditionalPackages[idx].RefName))
				Expect(updatedClusterBootstrap.Spec.AdditionalPackages[idx].ValuesFrom).To(BeNil())
			}
		})

		It("should not add the fields which are meant to be skipped in additionalPackages", func() {
			antreaAPIGroup := antreaconfigv1alpha1.GroupVersion.Group
			fakeCPIClusterBootstrapPackage := constructFakeClusterBootstrapPackageWithSecretRef()
			fakeCSIClusterBootstrapPackage := constructFakeClusterBootstrapPackageWithInlineRef()
			// Update fakeClusterBootstrapTemplate by adding a fake CPI and CSI package
			fakeClusterBootstrapTemplate.Spec.CPI = fakeCPIClusterBootstrapPackage
			fakeClusterBootstrapTemplate.Spec.CSI = fakeCSIClusterBootstrapPackage

			clusterBootstrap := &v1alpha3.ClusterBootstrap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "fake-clusterbootstrap",
					Namespace: "fake-cluster-ns",
					UID:       "uid",
				},
				Spec: &v1alpha3.ClusterBootstrapTemplateSpec{
					// We do not expect this part to be overwritten to be what fakeClusterBootstrapTemplate has
					CNI: &v1alpha3.ClusterBootstrapPackage{
						RefName: "foo-antrea-clusterbootstrarp-package",
						ValuesFrom: &v1alpha3.ValuesFrom{
							ProviderRef: &corev1.TypedLocalObjectReference{
								APIGroup: &antreaAPIGroup,
								Kind:     "AntreaConfig",
								Name:     "fooAntreaConfig",
							},
						},
					},
				},
			}

			err := helper.AddMissingSpecFieldsFromTemplate(fakeClusterBootstrapTemplate, clusterBootstrap, map[string]interface{}{"valuesFrom": nil})
			Expect(err).NotTo(HaveOccurred())
			for _, additionalPackage := range clusterBootstrap.Spec.AdditionalPackages {
				Expect(additionalPackage.RefName).NotTo(BeEmpty())
				Expect(additionalPackage.ValuesFrom).To(BeNil())
			}
		})
	})

	Context("Verify CompleteCBPackageRefNamesFromTKR()", func() {
		It("should complete the partial filled RefName if there is a match", func() {
			clusterBootstrap := constructFakeEmptyClusterBootstrap()
			clusterBootstrap.Spec = &v1alpha3.ClusterBootstrapTemplateSpec{
				CNI: &v1alpha3.ClusterBootstrapPackage{
					RefName: "calico*",
					ValuesFrom: &v1alpha3.ValuesFrom{
						Inline: map[string]interface{}{"foo": "bar"},
					},
				},
				AdditionalPackages: []*v1alpha3.ClusterBootstrapPackage{
					{RefName: "pinniped*", ValuesFrom: &v1alpha3.ValuesFrom{Inline: map[string]interface{}{"identity_management_type": "oidc"}}},
				},
			}
			tanzuKubernetesRelease := constructFakeTanzuKubernetesRelease()
			err := helper.CompleteCBPackageRefNamesFromTKR(tanzuKubernetesRelease, clusterBootstrap)
			Expect(err).NotTo(HaveOccurred())
			Expect(clusterBootstrap.Spec.CNI.RefName).To(Equal("calico.tanzu.vmware.com.3.22.1+vmware.1-tkg.1-zshippable"))
			Expect(clusterBootstrap.Spec.CNI.ValuesFrom.Inline["foo"]).To(Equal("bar"))
			Expect(len(clusterBootstrap.Spec.AdditionalPackages)).To(Equal(1))
			Expect(clusterBootstrap.Spec.AdditionalPackages[0].RefName).To(Equal("pinniped.tanzu.vmware.com.0.12.1+vmware.1-tkg.1-zshippable"))
			Expect(clusterBootstrap.Spec.AdditionalPackages[0].ValuesFrom.Inline["identity_management_type"]).To(Equal("oidc"))
		})

		It("should return error if there is a no match", func() {
			clusterBootstrap := constructFakeEmptyClusterBootstrap()
			clusterBootstrap.Spec = &v1alpha3.ClusterBootstrapTemplateSpec{
				CNI: &v1alpha3.ClusterBootstrapPackage{
					RefName: "something-does-not-exist*",
					ValuesFrom: &v1alpha3.ValuesFrom{
						Inline: map[string]interface{}{"foo": "bar"},
					},
				},
			}
			tanzuKubernetesRelease := constructFakeTanzuKubernetesRelease()
			err := helper.CompleteCBPackageRefNamesFromTKR(tanzuKubernetesRelease, clusterBootstrap)
			Expect(err).To(HaveOccurred())
			// The original value should stay untouched
			Expect(clusterBootstrap.Spec.CNI.RefName).To(Equal("something-does-not-exist*"))
		})

		It("should return error if there is a multiple matches", func() {
			clusterBootstrap := constructFakeEmptyClusterBootstrap()
			clusterBootstrap.Spec = &v1alpha3.ClusterBootstrapTemplateSpec{
				CNI: &v1alpha3.ClusterBootstrapPackage{
					RefName: "v*", // v* matches multiple ClusterBootstrapPackage in tanzuKubernetesRelease
					ValuesFrom: &v1alpha3.ValuesFrom{
						Inline: map[string]interface{}{"foo": "bar"},
					},
				},
			}
			tanzuKubernetesRelease := constructFakeTanzuKubernetesRelease()
			err := helper.CompleteCBPackageRefNamesFromTKR(tanzuKubernetesRelease, clusterBootstrap)
			Expect(err).To(HaveOccurred())
		})

		It("should not touch the fully filled refName", func() {
			clusterBootstrap := constructFakeEmptyClusterBootstrap()
			clusterBootstrap.Spec = &v1alpha3.ClusterBootstrapTemplateSpec{
				CNI: &v1alpha3.ClusterBootstrapPackage{
					RefName: "calico*",
					ValuesFrom: &v1alpha3.ValuesFrom{
						Inline: map[string]interface{}{"foo": "bar"},
					},
				},
				CPI: &v1alpha3.ClusterBootstrapPackage{
					RefName: "fake-cpi",
					ValuesFrom: &v1alpha3.ValuesFrom{
						ProviderRef: &corev1.TypedLocalObjectReference{
							Name: "fake-secret",
							Kind: "secret",
						},
					},
				},
				AdditionalPackages: []*v1alpha3.ClusterBootstrapPackage{
					{RefName: "pinniped.tanzu.vmware.com.0.11.1", ValuesFrom: &v1alpha3.ValuesFrom{Inline: map[string]interface{}{"identity_management_type": "ldap"}}},
				},
			}
			tanzuKubernetesRelease := constructFakeTanzuKubernetesRelease()
			err := helper.CompleteCBPackageRefNamesFromTKR(tanzuKubernetesRelease, clusterBootstrap)
			Expect(err).NotTo(HaveOccurred())
			// The original value should stay untouched
			Expect(clusterBootstrap.Spec.CPI.RefName).To(Equal("fake-cpi"))
			Expect(clusterBootstrap.Spec.CPI.ValuesFrom.Inline).To(BeNil())
			Expect(clusterBootstrap.Spec.CPI.ValuesFrom.SecretRef).To(BeEmpty())
			Expect(clusterBootstrap.Spec.CPI.ValuesFrom.ProviderRef.Kind).To(Equal("secret"))
			Expect(clusterBootstrap.Spec.CPI.ValuesFrom.ProviderRef.Name).To(Equal("fake-secret"))
			Expect(len(clusterBootstrap.Spec.AdditionalPackages)).To(Equal(1))
			Expect(clusterBootstrap.Spec.AdditionalPackages[0].RefName).To(Equal("pinniped.tanzu.vmware.com.0.11.1"))
			Expect(clusterBootstrap.Spec.AdditionalPackages[0].ValuesFrom.Inline["identity_management_type"]).To(Equal("ldap"))
			// The partial filled refName should be updated
			Expect(clusterBootstrap.Spec.CNI.RefName).To(Equal("calico.tanzu.vmware.com.3.22.1+vmware.1-tkg.1-zshippable"))
			Expect(clusterBootstrap.Spec.CNI.ValuesFrom.Inline["foo"]).To(Equal("bar"))
		})
	})
})

func constructFakeTanzuKubernetesRelease() *v1alpha3.TanzuKubernetesRelease {
	tkrYAML := `
kind: TanzuKubernetesRelease
apiVersion: run.tanzu.vmware.com/v1alpha3
metadata:
  name: v1.23.5---vmware.1-tkg.1-zshippable
spec:
  version: v1.23.5+vmware.1-tkg.1-zshippable
  kubernetes:
    version: v1.23.5+vmware.1
    imageRepository: projects.registry.vmware.com/tkg
    etcd:
      imageTag: v3.5.2_vmware.4
    pause:
      imageTag: "3.6"
    coredns:
      imageTag: v1.8.6_vmware.5
  osImages:
  - name: v1.23.3---vmware.1-tkg.1-tkgs-ubuntu-2004
  - name: v1.23.3---vmware.1-tkg.1-tkgs-photon-3
  bootstrapPackages:
  - name: antrea.tanzu.vmware.com.1.2.3+vmware.4-tkg.2-advanced-zshippable
  - name: vsphere-pv-csi.tanzu.vmware.com.2.4.0+vmware.1-tkg.1-zshippable
  - name: vsphere-cpi.tanzu.vmware.com.1.22.6+vmware.1-tkg.1-zshippable
  - name: kapp-controller.tanzu.vmware.com.0.34.0+vmware.1-tkg.1-zshippable
  - name: guest-cluster-auth-service.tanzu.vmware.com.1.0.0+tkg.1-zshippable
  - name: metrics-server.tanzu.vmware.com.0.5.1+vmware.1-tkg.2-zshippable
  - name: secretgen-controller.tanzu.vmware.com.0.8.0+vmware.1-tkg.1-zshippable
  - name: pinniped.tanzu.vmware.com.0.12.1+vmware.1-tkg.1-zshippable
  - name: capabilities.tanzu.vmware.com.0.22.0-dev-43-g2dd1adc9+vmware.1
  - name: calico.tanzu.vmware.com.3.22.1+vmware.1-tkg.1-zshippable
`
	tkrJSONByte, err := k8syaml.YAMLToJSON([]byte(tkrYAML))
	Expect(err).NotTo(HaveOccurred())
	tkr := &v1alpha3.TanzuKubernetesRelease{}
	Expect(json.Unmarshal(tkrJSONByte, tkr)).To(Succeed())
	return tkr
}

func convertToUnstructured(obj runtime.Object) *unstructured.Unstructured {
	// convert the runtime.Object to unstructured.Unstructured
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
			CNI: constructFakeClusterBootstrapPackageWithAntreaProviderRef(),
			CSI: constructFakeClusterBootstrapPackageWithCSIInlineRef(),
			AdditionalPackages: []*v1alpha3.ClusterBootstrapPackage{
				{RefName: fakePinnipedCBPackageRefName, ValuesFrom: &v1alpha3.ValuesFrom{Inline: map[string]interface{}{"identity_management_type": "oidc"}}},
				{RefName: fakeMetricsServerCBPackageRefName},
			},
		},
	}
}

func constructFakeClusterBootstrap() *v1alpha3.ClusterBootstrap {
	return &v1alpha3.ClusterBootstrap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake-clusterbootstrap",
			Namespace: "fake-ns",
		},
		Spec: &v1alpha3.ClusterBootstrapTemplateSpec{
			CNI: constructFakeClusterBootstrapPackageWithAntreaProviderRef(),
			CSI: constructFakeClusterBootstrapPackageWithCSIInlineRef(),
			AdditionalPackages: []*v1alpha3.ClusterBootstrapPackage{
				{RefName: fakePinnipedCBPackageRefName, ValuesFrom: &v1alpha3.ValuesFrom{Inline: map[string]interface{}{"identity_management_type": "oidc"}}},
				{RefName: fakeMetricsServerCBPackageRefName},
			},
		},
	}
}

func prepareCarvelPackages(client client.Client, namespace string) {
	err := client.Create(context.TODO(), constructAntreaCarvelPackage(namespace))
	Expect(err).To(BeNil())
	err = client.Create(context.TODO(), constructCSICarvelPackage(namespace))
	Expect(err).To(BeNil())
	err = client.Create(context.TODO(), constructPinnipedCarvelPackage(namespace))
	Expect(err).To(BeNil())
	err = client.Create(context.TODO(), constructMetricsCarvelPackage(namespace))
	Expect(err).To(BeNil())
}

func constructAntreaCarvelPackage(namespace string) *kapppkgv1alpha1.Package {
	return &kapppkgv1alpha1.Package{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fakeAntreaCBPackageRefName,
			Namespace: namespace,
		},
		Spec: kapppkgv1alpha1.PackageSpec{
			RefName: "antrea.vmware.com",
		},
	}
}

func constructCSICarvelPackage(namespace string) *kapppkgv1alpha1.Package {
	return &kapppkgv1alpha1.Package{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fakeCSICBPackageRefName,
			Namespace: namespace,
		},
		Spec: kapppkgv1alpha1.PackageSpec{
			RefName: "vsphere-csi.vmware.com",
		},
	}
}

func constructPinnipedCarvelPackage(namespace string) *kapppkgv1alpha1.Package {
	return &kapppkgv1alpha1.Package{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fakePinnipedCBPackageRefName,
			Namespace: namespace,
		},
		Spec: kapppkgv1alpha1.PackageSpec{
			RefName: "pinniped.vmware.com",
		},
	}
}

func constructMetricsCarvelPackage(namespace string) *kapppkgv1alpha1.Package {
	return &kapppkgv1alpha1.Package{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fakeMetricsServerCBPackageRefName,
			Namespace: namespace,
		},
		Spec: kapppkgv1alpha1.PackageSpec{
			RefName: "metrics-server.vmware.com",
		},
	}
}

func constructFakeClusterBootstrapPackageWithSecretRef() *v1alpha3.ClusterBootstrapPackage {
	return &v1alpha3.ClusterBootstrapPackage{
		RefName: fakeAntreaCBPackageRefName,
		ValuesFrom: &v1alpha3.ValuesFrom{
			SecretRef: "fake-secret",
		},
	}
}

func constructFakeClusterBootstrapPackageWithInlineRef() *v1alpha3.ClusterBootstrapPackage {
	return &v1alpha3.ClusterBootstrapPackage{
		RefName: fakeAntreaCBPackageRefName,
		ValuesFrom: &v1alpha3.ValuesFrom{
			Inline: map[string]interface{}{"foo": "bar"},
		},
	}
}

func constructFakeClusterBootstrapPackageWithAntreaProviderRef() *v1alpha3.ClusterBootstrapPackage {
	antreaAPIGroup := antreaconfigv1alpha1.GroupVersion.Group
	antreaConfig := constructFakeAntreaConfig()
	return &v1alpha3.ClusterBootstrapPackage{
		RefName: fakeAntreaCBPackageRefName,
		ValuesFrom: &v1alpha3.ValuesFrom{
			ProviderRef: &corev1.TypedLocalObjectReference{
				APIGroup: &antreaAPIGroup,
				Kind:     "AntreaConfig",
				Name:     antreaConfig.Name,
			},
		},
	}
}

func constructFakeClusterBootstrapPackageWithCSIInlineRef() *v1alpha3.ClusterBootstrapPackage {
	return &v1alpha3.ClusterBootstrapPackage{
		RefName: fakeCSICBPackageRefName,
		ValuesFrom: &v1alpha3.ValuesFrom{
			Inline: map[string]interface{}{"foo": "bar"},
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

func constructNamespacedFakeCluster(name, namespace string) *clusterapiv1beta1.Cluster {
	return &clusterapiv1beta1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
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

func constructFakeEmptyClusterBootstrap() *v1alpha3.ClusterBootstrap {
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
			Annotations: map[string]string{
				constants.TKGAnnotationTemplateConfig: "true",
			},
		},
		Spec: antreaconfigv1alpha1.AntreaConfigSpec{
			Antrea: antreaconfigv1alpha1.Antrea{
				AntreaConfigDataValue: antreaconfigv1alpha1.AntreaConfigDataValue{TrafficEncapMode: "encap"},
			},
		},
	}
}

func constructFakeAntreaConfigWithClusterOwner(configName, clusterName string) *antreaconfigv1alpha1.AntreaConfig {
	ownerRef := metav1.OwnerReference{
		APIVersion: "",
		Kind:       constants.ClusterKind,
		Name:       clusterName,
		UID:        "",
	}
	return &antreaconfigv1alpha1.AntreaConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AntreaConfig",
			APIVersion: antreaconfigv1alpha1.GroupVersion.String(),
		},

		ObjectMeta: metav1.ObjectMeta{
			Name:      configName,
			Namespace: "fake-ns",
			Annotations: map[string]string{
				constants.TKGAnnotationTemplateConfig: "true",
			},
			OwnerReferences: []metav1.OwnerReference{ownerRef},
		},
		Spec: antreaconfigv1alpha1.AntreaConfigSpec{
			Antrea: antreaconfigv1alpha1.Antrea{
				AntreaConfigDataValue: antreaconfigv1alpha1.AntreaConfigDataValue{TrafficEncapMode: "encap"},
			},
		},
	}
}

func assertTwoMapsShouldEqual(left, right map[string]interface{}) {
	for keyFromLeft, valueFromLeft := range left {
		valueFromRight, exist := right[keyFromLeft]
		Expect(exist).To(BeTrue())
		Expect(valueFromLeft).To(Equal(valueFromRight))
	}
}
