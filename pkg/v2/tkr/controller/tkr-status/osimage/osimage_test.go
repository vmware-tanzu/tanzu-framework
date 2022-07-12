// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package osimage

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	kapppkgv1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/resolver"
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
	suiteConfig, _ := GinkgoConfiguration()
	suiteConfig.FailFast = true
	RunSpecs(t, "TKR Status Reconciler test", suiteConfig)
}

var (
	r        *Reconciler
	c        client.Client
	osImages data.OSImages
	tkrs     data.TKRs
	objects  []client.Object
	ns       string
)

var _ = Describe("tkr-status/osimage.Reconciler", func() {
	BeforeEach(func() {
		ns = "tkg-system"
		osImages, tkrs, objects = genObjects(ns)
	})

	var (
		ctx     context.Context
		tkr     *runv1.TanzuKubernetesRelease
		osImage *runv1.OSImage
	)

	BeforeEach(func() {
		ctx = context.Background()
		tkr = testdata.ChooseTKR(tkrs)
		osImage = testdata.ChooseOSImage(tkr, osImages)
	})

	JustBeforeEach(func() {
		tkrResolver := resolver.New()
		scheme := initScheme()
		c = uidSetter{fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()}
		r = &Reconciler{
			Client: c,
			Cache:  tkrResolver,
			Log:    logr.Discard(),
		}
	})

	Describe("r.Reconcile()", func() {
		repeat(10, func() {
			It("should complete without errors", func() {
				result, err := r.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Name: osImage.Name}})
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal(ctrl.Result{}))
			})
		})
	})
})

func initScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = kapppkgv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = runv1.AddToScheme(scheme)
	return scheme
}

func genObjects(ns string) (data.OSImages, data.TKRs, []client.Object) {
	osImages := testdata.GenOSImages(k8sVersions, 80)
	tkrs := testdata.GenTKRs(10, testdata.SortOSImagesByK8sVersion(osImages))
	cbts := testdata.GenCBTs(ns, tkrs)
	objects := make([]client.Object, 0, len(osImages)+len(tkrs)+len(cbts))

	for _, osImage := range osImages {
		osImage.SetUID(uuid.NewUUID())
		objects = append(objects, osImage)
	}
	for _, tkr := range tkrs {
		tkr.SetUID(uuid.NewUUID())
		objects = append(objects, tkr)
	}
	for _, cbt := range cbts {
		cbt.SetUID(uuid.NewUUID())
		objects = append(objects, cbt)
	}
	return osImages, tkrs, objects
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
