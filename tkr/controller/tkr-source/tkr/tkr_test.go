// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkr

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/uuid"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/tkr/resolver/data"
	"github.com/vmware-tanzu/tanzu-framework/tkr/util/testdata"
)

const (
	k8s1_20_1 = "v1.20.1+vmware.1"
	k8s1_20_2 = "v1.20.2+vmware.1"
	k8s1_21_1 = "v1.21.1+vmware.1"
	k8s1_21_3 = "v1.21.3+vmware.1"
	k8s1_22_0 = "v1.22.0+vmware.1"
)

var k8sVersions = []string{k8s1_20_1, k8s1_20_2, k8s1_21_1, k8s1_21_3, k8s1_22_0}

var (
	scheme = initScheme()
)

func TestReconciler(t *testing.T) {
	RegisterFailHandler(Fail)
	suiteConfig, _ := GinkgoConfiguration()
	suiteConfig.FailFast = true
	RunSpecs(t, "TKR Source Controller: Legacy TKR Labeler", suiteConfig)
}

var _ = Describe("Reconciler", func() {
	var (
		r   *Reconciler
		ctx context.Context

		tkrs    data.TKRs
		objects []client.Object
	)

	BeforeEach(func() {
		ctx = context.Background()
		_, tkrs, objects = genObjects()
	})

	JustBeforeEach(func() {
		r = &Reconciler{
			Log:    logr.Discard(),
			Client: uidSetter{fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()},
		}
	})

	It("should label TKRs without bootstrap packages as legacy", func() {
		for _, tkr0 := range tkrs {
			req := ctrl.Request{NamespacedName: types.NamespacedName{Name: tkr0.Name}}
			result, err := r.Reconcile(ctx, req)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))

			tkr := &runv1.TanzuKubernetesRelease{}
			Expect(r.Client.Get(ctx, client.ObjectKey{Name: tkr0.Name}, tkr)).To(Succeed())

			Expect(labels.Set(tkr.Labels).Has(runv1.LabelLegacyTKR)).To(Equal(tkr0.Spec.BootstrapPackages == nil))
		}
	})

	When("the reconciled TKR no longer exists", func() {
		repeat(10, func() {
			It("do nothing and succeed", func() {
				req := ctrl.Request{NamespacedName: types.NamespacedName{Name: rand.String(rand.IntnRange(3, 8))}}
				result, err := r.Reconcile(ctx, req)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal(ctrl.Result{}))
			})
		})
	})
})

func genObjects() (data.OSImages, data.TKRs, []client.Object) {
	osImages := testdata.GenOSImages(k8sVersions, 1000)
	tkrs := testdata.GenTKRs(50, testdata.SortOSImagesByK8sVersion(osImages))
	objects := make([]client.Object, 0, len(osImages)+len(tkrs))

	for _, osImage := range osImages {
		objects = append(objects, osImage)
	}
	for _, tkr := range tkrs {
		notLegacy := rand.Intn(2) == 1
		if notLegacy {
			tkr.Spec.BootstrapPackages = genBootstrapPackageRefs()
		}
		objects = append(objects, tkr)
	}
	return osImages, tkrs, objects
}

func genBootstrapPackageRefs() []corev1.LocalObjectReference {
	references := make([]corev1.LocalObjectReference, rand.IntnRange(1, 9))
	for i := range references {
		r := &references[i]
		r.Name = rand.String(rand.IntnRange(4, 10))
	}
	return references
}

func initScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	utilruntime.Must(runv1.AddToScheme(scheme))
	return scheme
}

func repeat(numTimes int, f func()) {
	for i := 0; i < numTimes; i++ {
		f()
	}
}

// uidSetter emulates real clusters' behavior of setting UIDs on objects being created
type uidSetter struct {
	client.Client
}

func (u uidSetter) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	obj.(metav1.Object).SetUID(uuid.NewUUID())
	return u.Client.Create(ctx, obj, opts...)
}
