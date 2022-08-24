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

	kappctrlv1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	kapppkgv1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/apis/run/util/version"
	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/controller/tkr-source/registry"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/testdata"
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
		reg     registry.Registry
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
			Registry: reg,
		}
	})

	When("a TKR Package hasn't been installed yet", func() {
		var isTKR int
		var tkrName string
		var tkrVersion string

		BeforeEach(func() {
			pkg = genPkg()
			isTKR = rand.Intn(2)
			if isTKR != 0 {
				pkg.Labels = map[string]string{
					LabelTKRPackage: "",
				}
			}
			objects = []client.Object{pkg}
			tkrVersion = pkg.Spec.Version
			tkrName = version.Label(tkrVersion)
			reg = fakeRegistry{
				imageParams: map[string]struct {
					tkrVersion string
					k8sVersion string
				}{
					pkg.Spec.Template.Spec.Fetch[0].ImgpkgBundle.Image: {
						tkrVersion: tkrVersion,
						k8sVersion: pkg.Spec.Version,
					},
				}}
		})

		repeat(100, func() {
			It("should install it", func() {
				if hasTKRPackageLabel(pkg) {
					_, err := r.Reconcile(ctx, testdata.Request(pkg))
					Expect(err).ToNot(HaveOccurred())
				}

				cm := &corev1.ConfigMap{}
				err := r.Client.Get(ctx,
					client.ObjectKey{Namespace: pkg.Namespace, Name: fmt.Sprintf("tkr-%s", version.Label(pkg.Spec.Version))},
					cm)

				switch hasTKRPackageLabel(pkg) {
				case true:
					Expect(err).ToNot(HaveOccurred())

					tkr := &runv1.TanzuKubernetesRelease{}
					Expect(r.Client.Get(ctx, installedObjectName(pkg, tkrName), tkr)).To(Succeed())

				case false:
					Expect(err).To(HaveOccurred())
					Expect(errors.IsNotFound(err)).To(BeTrue())
				}
			})
		})
	})
})

func installedObjectName(pkg *kapppkgv1.Package, name string) client.ObjectKey {
	return client.ObjectKey{Namespace: pkg.Namespace, Name: name}
}

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
			Template: kapppkgv1.AppTemplateSpec{
				Spec: &kappctrlv1.AppSpec{
					Fetch: []kappctrlv1.AppFetch{{
						ImgpkgBundle: &kappctrlv1.AppFetchImgpkgBundle{
							Image: fmt.Sprintf("example.org/%s", rand.String(13)),
						},
					}},
				},
			},
		},
	}
}

func initScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	utilruntime.Must(runv1.AddToScheme(scheme))
	utilruntime.Must(kapppkgv1.AddToScheme(scheme))
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

type fakeRegistry struct {
	registry.Registry

	imageParams map[string]struct {
		tkrVersion string
		k8sVersion string
	}
}

func (r fakeRegistry) GetFiles(image string) (map[string][]byte, error) {
	params := r.imageParams[image]
	return map[string][]byte{
		"config/tkr.yaml":     []byte(tkrStr(params.tkrVersion, params.k8sVersion)),
		"config/garbage.yaml": []byte(garbageStr),
	}, nil
}

func tkrStr(tkrVersion, k8sVersion string) string {
	return fmt.Sprintf(`
kind: TanzuKubernetesRelease
apiVersion: run.tanzu.vmware.com/v1alpha3
metadata:
  name: %s
spec:
  version: %s
  kubernetes:
    version: %s
    imageRepository: projects.registry.vmware.com/tkg
    etcd:
      imageTag: v3.5.2_vmware.4
    pause:
      imageTag: "3.6"
    coredns:
      imageTag: v1.8.6_vmware.5
  osImages:
  - name: ubuntu-2004-amd64-vmi-k8s-v1.23.5---vmware.1-tkg.1-zshippable
  - name: photon-3-amd64-vmi-k8s-v1.23.5---vmware.1-tkg.1-zshippable
  bootstrapPackages:
  - name: antrea.tanzu.vmware.com.1.2.3+vmware.4-tkg.2-advanced-zshippable
  - name: vsphere-pv-csi.tanzu.vmware.com.2.4.0+vmware.1-tkg.1-zshippable
  - name: vsphere-cpi.tanzu.vmware.com.1.22.6+vmware.1-tkg.1-zshippable
  - name: kapp-controller.tanzu.vmware.com.0.34.0+vmware.1-tkg.1-zshippable
  - name: guest-cluster-auth-service.tanzu.vmware.com.1.0.0+tkg.1-zshippable
  - name: metrics-server.tanzu.vmware.com.0.5.1+vmware.1-tkg.2-zshippable
  - name: secretgen-controller.tanzu.vmware.com.0.8.0+vmware.1-tkg.1-zshippable
  - name: pinniped.tanzu.vmware.com.0.12.1+vmware.1-tkg.1-zshippable
  - name: capabilities.tanzu.vmware.com.0.22.0-dev-57-gd9465b25+vmware.1
  - name: calico.tanzu.vmware.com.3.22.1+vmware.1-tkg.1-zshippable
`, version.Label(tkrVersion), tkrVersion, k8sVersion)
}

const garbageStr = `
(This (not even YAML))
`
