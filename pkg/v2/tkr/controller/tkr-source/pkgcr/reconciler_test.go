// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package pkgcr

import (
	"context"
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/uuid"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	kapppkgiv1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	kapppkgv1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/testdata"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/version"
)

var (
	scheme = initScheme()
)

func TestReconciler(t *testing.T) {
	RegisterFailHandler(Fail)
	suiteConfig, _ := GinkgoConfiguration()
	suiteConfig.FailFast = true
	RunSpecs(t, "TKR Source Controller: Package Installer", suiteConfig)
}

const tkrServiceAccount = "tkr-service-account"

var _ = Describe("Reconciler", func() {
	var (
		r       *Reconciler
		objects []client.Object
		ctx     context.Context
		pkg     *kapppkgv1.Package
	)

	BeforeEach(func() {
		ctx = context.Background()
	})

	JustBeforeEach(func() {
		r = &Reconciler{
			Log:    logr.Discard(),
			Client: uidSetter{fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()},
			Config: Config{
				ServiceAccountName: tkrServiceAccount,
			},
		}
	})

	When("a TKR Package hasn't been installed yet", func() {
		var isTKR int

		BeforeEach(func() {
			pkg = genPkg()
			isTKR = rand.Intn(2)
			if isTKR != 0 {
				pkg.Labels = map[string]string{
					LabelTKRPackage: "",
				}
			}
			objects = []client.Object{pkg}
		})

		repeat(100, func() {
			It("should install it", func() {
				if hasTKRPackageLabel(pkg) {
					_, err := r.Reconcile(ctx, testdata.Request(pkg))
					Expect(err).ToNot(HaveOccurred())
				}

				pkgi := &kapppkgiv1.PackageInstall{}
				err := r.Client.Get(ctx,
					client.ObjectKey{Namespace: pkg.Namespace, Name: fmt.Sprintf("tkr-%s", version.Label(pkg.Spec.Version))},
					pkgi)

				switch hasTKRPackageLabel(pkg) {
				case true:
					Expect(err).ToNot(HaveOccurred())
					Expect(pkgi.Spec.ServiceAccountName).To(Equal(r.Config.ServiceAccountName))
					Expect(pkgi.Spec.PackageRef.RefName).To(Equal(pkg.Spec.RefName))
					Expect(pkgi.Spec.PackageRef.VersionSelection.Constraints).To(Equal(pkg.Spec.Version))
				case false:
					Expect(err).To(HaveOccurred())
					Expect(errors.IsNotFound(err)).To(BeTrue())
				}

				Expect(pkgi)
			})
		})
	})
})

func genPkg() *kapppkgv1.Package {
	name := rand.String(10)
	v := fmt.Sprintf("%v.%v.%v", rand.Intn(2), rand.Intn(10), rand.Intn(10))
	return &kapppkgv1.Package{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: rand.String(10),
			Name:      fmt.Sprintf("%s.%s", name, v),
		},
		Spec: kapppkgv1.PackageSpec{
			RefName: name,
			Version: v,
		},
	}
}

func initScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	utilruntime.Must(runv1.AddToScheme(scheme))
	utilruntime.Must(kapppkgv1.AddToScheme(scheme))
	utilruntime.Must(kapppkgiv1.AddToScheme(scheme))
	utilruntime.Must(corev1.AddToScheme(scheme))
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
