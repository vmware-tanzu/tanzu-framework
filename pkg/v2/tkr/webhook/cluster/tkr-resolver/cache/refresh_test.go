// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"context"
	"reflect"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/resolver/data"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/testdata"
)

const (
	k8s1_20_1 = "v1.20.1+vmware.1"
	k8s1_20_2 = "v1.20.2+vmware.1"
	k8s1_21_1 = "v1.21.1+vmware.1"
	k8s1_21_3 = "v1.21.3+vmware.1"
	k8s1_22_0 = "v1.22.0+vmware.1"
)

var k8sVersions = []string{k8s1_20_1, k8s1_20_2, k8s1_21_1, k8s1_21_3, k8s1_22_0}

func TestReconciler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TKR source controller test")
}

var _ = Describe("Reconciler", func() {
	var (
		scheme     *runtime.Scheme
		cache      *fakeCache
		osImages   data.OSImages
		tkrs       data.TKRs
		objects    []client.Object
		fakeClient client.Client

		tkrReconciler     *Reconciler
		osImageReconciler *Reconciler
	)

	BeforeEach(func() {
		scheme = addToScheme(runtime.NewScheme())
	})

	BeforeEach(func() {
		osImages, tkrs, objects = genObjects()
		fakeClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()
		cache = &fakeCache{}
	})

	JustBeforeEach(func() {
		tkrReconciler = &Reconciler{
			Client: fakeClient,
			Cache:  cache,
			Object: &runv1.TanzuKubernetesRelease{},
			Log:    logr.Discard(),
		}
		osImageReconciler = &Reconciler{
			Client: fakeClient,
			Cache:  cache,
			Object: &runv1.OSImage{},
			Log:    logr.Discard(),
		}
	})

	Describe("Reconcile()", func() {
		When("a TKR exists", func() {
			It("should add it to the TKRResolver cache", func() {
				for _, tkr := range testdata.RandNonEmptySubsetOfTKRs(tkrs) {
					result, err := tkrReconciler.Reconcile(context.Background(), ctrl.Request{
						NamespacedName: types.NamespacedName{
							Name: tkr.Name,
						},
					})
					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(Equal(ctrl.Result{}))

					Expect(cache.addedObjects).To(ContainElement(tkr))
				}
			})
		})

		When("an OSImage exists", func() {
			It("should add it to the TKRResolver cache", func() {
				for _, osImage := range testdata.RandNonEmptySubsetOfOSImages(osImages) {
					result, err := osImageReconciler.Reconcile(context.Background(), ctrl.Request{
						NamespacedName: types.NamespacedName{
							Name: osImage.Name,
						},
					})
					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(Equal(ctrl.Result{}))

					Expect(cache.addedObjects).To(ContainElement(osImage))
				}
			})
		})

		When("a TKR is not found", func() {
			BeforeEach(func() {
				tkrs = testdata.RandNonEmptySubsetOfTKRs(tkrs)
				for _, tkr := range tkrs {
					Expect(fakeClient.Delete(context.Background(), tkr)).To(Succeed())
				}
			})

			It("should remove it from the TKRResolver cache", func() {
				for _, tkr := range testdata.RandNonEmptySubsetOfTKRs(tkrs) {
					result, err := tkrReconciler.Reconcile(context.Background(), ctrl.Request{
						NamespacedName: types.NamespacedName{
							Name: tkr.Name,
						},
					})
					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(Equal(ctrl.Result{}))

					Expect(cache.removedObjects).To(ContainElement(HaveField("Name", tkr.Name)))
				}
			})
		})

		When("an OSImage is not found", func() {
			BeforeEach(func() {
				osImages = testdata.RandNonEmptySubsetOfOSImages(osImages)
				for _, osImage := range osImages {
					Expect(fakeClient.Delete(context.Background(), osImage)).To(Succeed())
				}
			})

			It("should remove it from the TKRResolver cache", func() {
				for _, osImage := range testdata.RandNonEmptySubsetOfOSImages(osImages) {
					result, err := osImageReconciler.Reconcile(context.Background(), ctrl.Request{
						NamespacedName: types.NamespacedName{
							Name: osImage.Name,
						},
					})
					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(Equal(ctrl.Result{}))

					Expect(cache.removedObjects).To(ContainElement(HaveField("Name", osImage.Name)))
				}
			})
		})

		When("the client returns an error", func() {
			var (
				expectedErr = errors.New("expected")
			)

			BeforeEach(func() {
				fakeClient = failingGet{err: expectedErr}
			})

			It("should return that error", func() {
				for _, tkr := range testdata.RandNonEmptySubsetOfTKRs(tkrs) {
					result, err := tkrReconciler.Reconcile(context.Background(), ctrl.Request{
						NamespacedName: types.NamespacedName{
							Name: tkr.Name,
						},
					})
					Expect(err).To(Equal(expectedErr))
					Expect(result).To(Equal(ctrl.Result{}))
				}
				for _, osImage := range testdata.RandNonEmptySubsetOfOSImages(osImages) {
					result, err := osImageReconciler.Reconcile(context.Background(), ctrl.Request{
						NamespacedName: types.NamespacedName{
							Name: osImage.Name,
						},
					})
					Expect(err).To(Equal(expectedErr))
					Expect(result).To(Equal(ctrl.Result{}))
				}
			})
		})
	})
})

func addToScheme(scheme *runtime.Scheme) *runtime.Scheme {
	utilruntime.Must(runv1.AddToScheme(scheme))
	return scheme
}

func genObjects() (data.OSImages, data.TKRs, []client.Object) {
	osImages := testdata.GenOSImages(k8sVersions, 10)
	tkrs := testdata.GenTKRs(5, testdata.SortOSImagesByK8sVersion(osImages))
	objects := make([]client.Object, 0, len(osImages)+len(tkrs))

	for _, osImage := range osImages {
		objects = append(objects, osImage)
	}
	for _, tkr := range tkrs {
		objects = append(objects, tkr)
	}
	return osImages, tkrs, objects
}

type fakeCache struct {
	addedObjects   map[string]interface{}
	removedObjects map[string]interface{}
}

func (f *fakeCache) Add(objects ...interface{}) {
	if f.addedObjects == nil {
		f.addedObjects = map[string]interface{}{}
	}
	for _, object := range objects {
		object := object.(client.Object)
		key := reflect.TypeOf(object).String() + "@" + object.GetName()
		f.addedObjects[key] = object
	}
}

func (f *fakeCache) Remove(objects ...interface{}) {
	if f.removedObjects == nil {
		f.removedObjects = map[string]interface{}{}
	}
	for _, object := range objects {
		object := object.(client.Object)
		key := reflect.TypeOf(object).String() + "@" + object.GetName()
		f.removedObjects[key] = object
	}
}

func (f *fakeCache) Get(string, interface{}) interface{} {
	panic("not supposed to be used yet")
}

type failingGet struct {
	client.Client
	err error
}

func (f failingGet) Get(context.Context, client.ObjectKey, client.Object) error {
	return f.err
}
