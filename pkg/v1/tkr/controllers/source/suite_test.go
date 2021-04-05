// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package source

import (
	"context"
	"io/ioutil"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	runv1 "github.com/vmware-tanzu-private/core/apis/run/v1alpha1"
	runv1alpha1 "github.com/vmware-tanzu-private/core/apis/run/v1alpha1"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkr/fakes"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkr/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var bomContent17 []byte
var bomContent18 []byte
var bomContent193 []byte
var bomContent191 []byte
var badBom []byte
var metadataContent []byte

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TKR source controller test")
}

func addToScheme(scheme *runtime.Scheme) {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = capi.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = runv1alpha1.AddToScheme(scheme)
}

var _ = BeforeSuite(func() {
	bomContent17, _ = ioutil.ReadFile("../../fakes/boms/bom-v1.17.13+vmware.1.yaml")
	bomContent18, _ = ioutil.ReadFile("../../fakes/boms/bom-v1.18.10+vmware.1.yaml")
	bomContent193, _ = ioutil.ReadFile("../../fakes/boms/bom-v1.19.3+vmware.1.yaml")
	bomContent191, _ = ioutil.ReadFile("../../fakes/boms/bom-v1.19.1+vmware.1.yaml")
	badBom, _ = ioutil.ReadFile("../../fakes/boms/bad-bom.yaml")
	metadataContent, _ = ioutil.ReadFile("../../fakes/boms/metadata.yaml")
})

var _ = Describe("SyncRelease", func() {
	var (
		fakeClient   client.Client
		fakeRegistry *fakes.Registry
		scheme       *runtime.Scheme
		objects      []runtime.Object
		r            reconciler
		err          error
		added        []runv1alpha1.TanzuKubernetesRelease
		existing     []runv1alpha1.TanzuKubernetesRelease
	)

	JustBeforeEach(func() {
		scheme = runtime.NewScheme()
		addToScheme(scheme)
		fakeClient = fake.NewFakeClientWithScheme(scheme, objects...)
		r = reconciler{
			client:   fakeClient,
			log:      ctrllog.Log,
			scheme:   scheme,
			registry: fakeRegistry,
			bomImage: "my-registry.io/tkrs",
		}

		added, existing, err = r.SyncRelease(context.Background())
	})

	Context("When BOM images with proper content are published, and no TKR has been created", func() {

		BeforeEach(func() {
			fakeRegistry = &fakes.Registry{}
			fakeRegistry.ListImageTagsReturns([]string{"bom-v1.17.13+vmware.1", "bom-v1.18.10+vmware.1", "bom-v1.19.3+vmware.1"}, nil)

			fakeRegistry.GetFileReturnsOnCall(0, bomContent17, nil)
			fakeRegistry.GetFileReturnsOnCall(1, bomContent18, nil)
			fakeRegistry.GetFileReturnsOnCall(2, bomContent193, nil)
			objects = []runtime.Object{}

		})

		It("should create TKRs and BOM ConfigMap", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(len(added)).To(Equal(3))
			Expect(len(existing)).To(Equal(0))
			Expect(fakeRegistry.GetFileCallCount()).To(Equal(3))
		})

	})

	Context("When a new BOM images is released", func() {

		BeforeEach(func() {
			fakeRegistry = &fakes.Registry{}
			fakeRegistry.ListImageTagsReturns([]string{"bom-v1.17.13+vmware.1", "bom-v1.18.10+vmware.1", "bom-v1.19.3+vmware."}, nil)
			fakeRegistry.GetFileReturnsOnCall(0, bomContent193, nil)

			cm1 := newConfigMap("v1.17.13---vmware.1", map[string]string{constants.BomConfigMapTKRLabel: "v1.17.13---vmware.1"}, map[string]string{constants.BomConfigMapImageTagAnnotation: "bom-v1.17.13+vmware.1"}, bomContent17)
			cm2 := newConfigMap("v1.18.10---vmware.1", map[string]string{constants.BomConfigMapTKRLabel: "v1.18.10---vmware.1"}, map[string]string{constants.BomConfigMapImageTagAnnotation: "bom-v1.18.10+vmware.1"}, bomContent18)

			tkr1, _ := NewTkrFromBom("v1.17.13---vmware.1", bomContent17)
			tkr2, _ := NewTkrFromBom("v1.18.10---vmware.1", bomContent18)
			objects = []runtime.Object{cm1, cm2, &tkr1, &tkr2}
		})

		It("should create a new TKR based on the new BOM", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeRegistry.GetFileCallCount()).To(Equal(1))
			Expect(len(added)).To(Equal(1))
			Expect(len(existing)).To(Equal(2))
		})
	})

	Context("When the BOM ConfigMap exists, but the corresponding TKR is missing", func() {

		BeforeEach(func() {
			fakeRegistry = &fakes.Registry{}
			fakeRegistry.ListImageTagsReturns([]string{"bom-v1.17.13+vmware.1", "bom-v1.18.10+vmware.1"}, nil)
			cm1 := newConfigMap("v1.17.13---vmware.1", map[string]string{constants.BomConfigMapTKRLabel: "v1.17.13---vmware.1"}, map[string]string{constants.BomConfigMapImageTagAnnotation: "bom-v1.17.13+vmware.1"}, bomContent17)
			cm2 := newConfigMap("v1.18.10---vmware.1", map[string]string{constants.BomConfigMapTKRLabel: "v1.18.10---vmware.1"}, map[string]string{constants.BomConfigMapImageTagAnnotation: "bom-v1.18.10+vmware.1"}, bomContent18)
			tkr1, _ := NewTkrFromBom("v1.17.13---vmware.1", bomContent17)
			objects = []runtime.Object{cm1, cm2, &tkr1}
		})

		It("should generate the TKR from the BOM ConfigMap", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeRegistry.GetFileCallCount()).To(Equal(0))
			Expect(len(added)).To(Equal(1))
			Expect(len(existing)).To(Equal(1))

		})
	})

})

var _ = Describe("UpdateTKRCompatibleCondition", func() {
	var (
		tkrs         []runv1.TanzuKubernetesRelease
		fakeClient   client.Client
		fakeRegistry *fakes.Registry
		scheme       *runtime.Scheme
		objects      []runtime.Object
		r            reconciler
		err          error
	)

	JustBeforeEach(func() {
		scheme = runtime.NewScheme()
		addToScheme(scheme)
		fakeClient = fake.NewFakeClientWithScheme(scheme, objects...)
		r = reconciler{
			client:   fakeClient,
			log:      ctrllog.Log,
			scheme:   scheme,
			registry: fakeRegistry,
			bomImage: "my-registry.io/tkrs",
		}
		err = r.UpdateTKRCompatibleCondition(context.Background(), tkrs)
	})

	Context("When reconcile the compatible condition of the TKRs", func() {
		BeforeEach(func() {
			fakeRegistry = &fakes.Registry{}
			fakeRegistry.ListImageTagsReturns([]string{"v0", "v1", "v2"}, nil)
			fakeRegistry.GetFileReturnsOnCall(0, nil, errors.New("cannot retrieve file from the image"))
			fakeRegistry.GetFileReturnsOnCall(1, metadataContent, nil)

			tkr1, _ := NewTkrFromBom("v1.17.13---vmware.1", bomContent17)
			tkr2, _ := NewTkrFromBom("v1.18.10---vmware.1", bomContent18)
			tkr3, _ := NewTkrFromBom("v1.19.3---vmware.1", bomContent193)
			tkr4, _ := NewTkrFromBom("v1.19.1---vmware.1", bomContent191)
			tkrs = []runv1.TanzuKubernetesRelease{tkr1, tkr4, tkr3, tkr2}

			mgmtcluster := newManagemntCluster("mgmt-cluster", map[string]string{constants.ManagememtClusterRoleLabel: ""}, map[string]string{constants.TKGVersionKey: "v1.1"})
			objects = []runtime.Object{mgmtcluster}
		})
		It("should update the TKRs' compatible condition", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(fakeRegistry.GetFileCallCount()).To(Equal(2))
			for _, tkr := range tkrs {
				if tkr.Name == "v1.19.3---vmware.1" {
					status, msg := getConditionStatusAndMessage(tkr.Status.Conditions, runv1.ConditionCompatible)
					Expect(string(status)).To(Equal("False"))
					Expect(msg).To(Equal(""))
				}

				if tkr.Name == "v1.18.10---vmware.1" {
					status, msg := getConditionStatusAndMessage(tkr.Status.Conditions, runv1.ConditionCompatible)
					Expect(string(status)).To(Equal("True"))
					Expect(msg).To(Equal(""))
				}

				if tkr.Name == "v1.19.1---vmware.1" {
					status, msg := getConditionStatusAndMessage(tkr.Status.Conditions, runv1.ConditionCompatible)
					Expect(string(status)).To(Equal("False"))
					Expect(msg).To(Equal(""))
				}

				if tkr.Name == "v1.17.13---vmware.1" {
					status, msg := getConditionStatusAndMessage(tkr.Status.Conditions, runv1.ConditionCompatible)
					Expect(string(status)).To(Equal("True"))
					Expect(msg).To(Equal(""))
				}
			}
		})
	})
})

var _ = Describe("UpdateTKRUpgradeAvailableCondition", func() {
	var (
		tkrs []runv1.TanzuKubernetesRelease
		r    reconciler
	)

	JustBeforeEach(func() {
		r = reconciler{}
		r.UpdateTKRUpgradeAvailableCondition(tkrs)
	})

	Context("When there are available upgrade for some of the TKRs", func() {
		BeforeEach(func() {
			tkr1, _ := NewTkrFromBom("v1.17.13---vmware.1", bomContent17)
			tkr2, _ := NewTkrFromBom("v1.18.10---vmware.1", bomContent18)
			tkr3, _ := NewTkrFromBom("v1.19.3---vmware.1", bomContent193)
			tkr4, _ := NewTkrFromBom("v1.19.1---vmware.1", bomContent191)
			tkrs = []runv1.TanzuKubernetesRelease{tkr1, tkr4, tkr3, tkr2}
		})
		It("should update the UpgradeAvailable Condition with proper message", func() {

			for _, tkr := range tkrs {
				if tkr.Name == "v1.19.3---vmware.1" {
					status, msg := getConditionStatusAndMessage(tkr.Status.Conditions, runv1.ConditionUpgradeAvailable)
					Expect(string(status)).To(Equal("False"))
					Expect(msg).To(Equal(""))
				}

				if tkr.Name == "v1.18.10---vmware.1" {
					status, msg := getConditionStatusAndMessage(tkr.Status.Conditions, runv1.ConditionUpgradeAvailable)
					Expect(string(status)).To(Equal("True"))
					Expect(msg).To(Equal("TKR(s) with later version is available: v1.19.1---vmware.1,v1.19.3---vmware.1"))
				}

				if tkr.Name == "v1.19.1---vmware.1" {
					status, msg := getConditionStatusAndMessage(tkr.Status.Conditions, runv1.ConditionUpgradeAvailable)
					Expect(string(status)).To(Equal("True"))
					Expect(msg).To(Equal("TKR(s) with later version is available: v1.19.3---vmware.1"))
				}

				if tkr.Name == "v1.17.13---vmware.1" {
					status, msg := getConditionStatusAndMessage(tkr.Status.Conditions, runv1.ConditionUpgradeAvailable)
					Expect(string(status)).To(Equal("True"))
					Expect(msg).To(Equal("TKR(s) with later version is available: v1.18.10---vmware.1"))
				}
			}
		})
	})
})

var _ = Describe("initialReconcile", func() {
	var (
		fakeClient   client.Client
		fakeRegistry *fakes.Registry
		scheme       *runtime.Scheme
		objects      []runtime.Object
		r            reconciler
		stopChan     chan struct{}
	)

	JustBeforeEach(func() {
		scheme = runtime.NewScheme()
		addToScheme(scheme)
		fakeClient = fake.NewFakeClientWithScheme(scheme, objects...)
		r = reconciler{
			client:                     fakeClient,
			log:                        ctrllog.Log,
			scheme:                     scheme,
			registry:                   fakeRegistry,
			bomImage:                   "my-registry.io/tkrs",
			compatibilityMetadataImage: "",
		}
		initSyncDone := make(chan bool)
		ticker := time.NewTicker(time.Second)
		stopChan = make(chan struct{})
		go r.initialReconcile(ticker, initSyncDone, stopChan, 3)
		go func(stopChan chan struct{}) {
			time.Sleep(time.Second * 5)
			stopChan <- struct{}{}
		}(stopChan)

		<-initSyncDone
	})

	Context("When cluster is not ready,", func() {
		BeforeEach(func() {
			fakeRegistry = &fakes.Registry{}
			fakeRegistry.ListImageTagsReturns([]string{"bom-v1.17.13+vmware.1", "bom-v1.18.10+vmware.1", "bom-v1.19.3+vmware.1"}, nil)

			fakeRegistry.GetFileReturnsOnCall(0, bomContent17, nil)
			fakeRegistry.GetFileReturnsOnCall(1, bomContent18, nil)
			fakeRegistry.GetFileReturnsOnCall(2, bomContent193, nil)

			for i := 3; i < 10; i++ {
				fakeRegistry.GetFileReturnsOnCall(i, metadataContent, nil)
			}

			stopChan = make(chan struct{})

		})
		It("should keep the controller in intitial sync-up stage", func() {
			Expect(fakeRegistry.ListImageTagsCallCount()).Should(BeNumerically(">", 3))
		})
	})

	Context("When cluster is ready, but the bom content can not be retrieved", func() {
		BeforeEach(func() {
			fakeRegistry = &fakes.Registry{}
			fakeRegistry.ListImageTagsReturns([]string{"bom-v1.17.13+vmware.1", "bom-v1.18.10+vmware.1", "bom-v1.19.3+vmware.1"}, nil)
			for i := 0; i < 12; i += 3 {
				fakeRegistry.ListImageTagsReturnsOnCall(i, []string{"bom-v1.17.13+vmware.1", "bom-v1.18.10+vmware.1", "bom-v1.19.3+vmware.1"}, nil)
				fakeRegistry.ListImageTagsReturnsOnCall(i+1, []string{"v1"}, nil)
				fakeRegistry.ListImageTagsReturnsOnCall(i+2, []string{"bom-v1.17.13+vmware.1", "bom-v1.18.10+vmware.1", "bom-v1.19.3+vmware.1"}, nil)
			}
			fakeRegistry.GetFileReturnsOnCall(0, bomContent17, nil)
			fakeRegistry.GetFileReturnsOnCall(1, bomContent18, nil)
			fakeRegistry.GetFileReturnsOnCall(2, nil, errors.New("fake-error"))
			fakeRegistry.GetFileReturnsOnCall(3, metadataContent, nil)
			for i := 4; i < 16; i += 4 {
				fakeRegistry.GetFileReturnsOnCall(i+2, nil, errors.New("fake-error"))
				fakeRegistry.GetFileReturnsOnCall(i+3, metadataContent, nil)
			}
			mgmtcluster := newManagemntCluster("mgmt-cluster", map[string]string{constants.ManagememtClusterRoleLabel: ""}, map[string]string{constants.TKGVersionKey: "v1.1"})
			objects = []runtime.Object{mgmtcluster}
		})

		It("should exist from initial sync-up stage after 3 retries", func() {
			// 6 --> 3*(list_bom_tag + list_metadata_tag + initial-sync-up-check)
			Expect(fakeRegistry.ListImageTagsCallCount()).Should(BeNumerically("==", 9))
		})
	})

	Context("When cluster is ready, and bom content can be retrieved", func() {
		BeforeEach(func() {

			fakeRegistry = &fakes.Registry{}

			fakeRegistry.ListImageTagsReturnsOnCall(0, []string{"bom-v1.17.13+vmware.1", "bom-v1.18.10+vmware.1", "bom-v1.19.3+vmware.1"}, nil)
			fakeRegistry.ListImageTagsReturnsOnCall(1, []string{"v1"}, nil)
			fakeRegistry.ListImageTagsReturnsOnCall(2, []string{"bom-v1.17.13+vmware.1", "bom-v1.18.10+vmware.1", "bom-v1.19.3+vmware.1"}, nil)

			fakeRegistry.GetFileReturnsOnCall(0, bomContent17, nil)
			fakeRegistry.GetFileReturnsOnCall(1, bomContent18, nil)
			fakeRegistry.GetFileReturnsOnCall(2, bomContent193, nil)
			fakeRegistry.GetFileReturnsOnCall(3, metadataContent, nil)
			mgmtcluster := newManagemntCluster("mgmt-cluster", map[string]string{constants.ManagememtClusterRoleLabel: ""}, map[string]string{constants.TKGVersionKey: "v1.1"})
			objects = []runtime.Object{mgmtcluster}
		})
		It("should exist from initial sync-up stage after the first try", func() {
			// 3 --> list_bom_tag + list_metadata_tag + initial-sync-up-check
			Expect(fakeRegistry.ListImageTagsCallCount()).Should(BeNumerically("==", 3))
		})
	})
})

func getConditionStatusAndMessage(conditions []clusterv1.Condition, conditionType clusterv1.ConditionType) (corev1.ConditionStatus, string) {
	for _, condition := range conditions {
		if condition.Type == conditionType {
			return condition.Status, condition.Message
		}
	}
	return corev1.ConditionStatus(""), ""
}

func newConfigMap(name string, labels map[string]string, annotations map[string]string, content []byte) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   constants.TKRNamespace,
			Labels:      labels,
			Annotations: annotations,
		},
		BinaryData: map[string][]byte{constants.BomConfigMapContentKey: content},
	}
}

func newManagemntCluster(name string, labels map[string]string, annotations map[string]string) *clusterv1.Cluster {
	return &clusterv1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   constants.TKGNamespace,
			Labels:      labels,
			Annotations: annotations,
		},
	}
}
