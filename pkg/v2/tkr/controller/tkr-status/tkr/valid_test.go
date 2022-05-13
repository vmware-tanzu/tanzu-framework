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
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/uuid"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	kapppkgv1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/controller/tkr-status/tkr/reasons"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/resolver"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/resolver/data"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/testdata"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/util/version"
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

var _ = Describe("tkr-status/tkr.Reconciler", func() {
	BeforeEach(func() {
		ns = "tkg-system"
		osImages, tkrs, objects = genObjects(ns)
	})

	var (
		ctx context.Context
		tkr *runv1.TanzuKubernetesRelease
	)

	BeforeEach(func() {
		ctx = context.Background()
		tkr = testdata.ChooseTKR(tkrs)
	})

	JustBeforeEach(func() {
		tkrResolver := resolver.New()
		scheme := initScheme()
		for _, o := range objects {
			tkrResolver.Add(o)
		}
		c = uidSetter{fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()}
		r = &Reconciler{
			Client: c,
			Cache:  tkrResolver,
			Log:    logr.Discard(),
			Config: Config{Namespace: ns},
		}
	})

	Describe("checkTKRVersion()", func() {
		repeat(10, func() {
			It("should correctly determine if the TKR version and kubernetes.version match", func() {
				Expect(tkr.Spec.Kubernetes.Version).ToNot(Equal(tkr.Spec.Version))
				Expect(version.Prefixes(version.Label(tkr.Spec.Version))).To(HaveKey(version.Label(tkr.Spec.Kubernetes.Version)))
				Expect(checkTKRVersion(ctx, tkr)).To(Succeed(), "generated TKRs are expected to be valid")

				origK8sVersion := tkr.Spec.Kubernetes.Version
				k8sVersion := testdata.ChooseK8sVersionFromTKRs(tkrs)
				tkr.Spec.Kubernetes.Version = k8sVersion

				err := checkTKRVersion(ctx, tkr)
				Expect(r.setValidCondition(ctx, tkr)).To(Succeed())
				cond := conditions.Get(tkr, runv1.ConditionValid)

				if k8sVersion != origK8sVersion {
					err, hasReason := err.(reasons.HasReason)
					Expect(hasReason).To(BeTrue())
					expectedReason := reasons.TKRVersionMismatch("").Reason()
					Expect(err.Reason()).To(Equal(expectedReason))
					Expect(cond.Status).To(Equal(corev1.ConditionFalse))
					Expect(cond.Reason).To(Equal(expectedReason))
					return
				}
				Expect(cond.Status).To(Equal(corev1.ConditionTrue))
			})
		})
	})

	Describe("r.checkOSImages()", func() {
		When("no changes are made to generated TKRs", func() {
			repeat(10, func() {
				It("should not return errors", func() {
					Expect(r.checkOSImages(ctx, tkr)).To(Succeed(), "generated TKRs are expected to be valid")

					Expect(r.setValidCondition(ctx, tkr)).To(Succeed())
					cond := conditions.Get(tkr, runv1.ConditionValid)
					Expect(cond.Status).To(Equal(corev1.ConditionTrue))
				})
			})
		})

		When("an OSImage is missing", func() {
			BeforeEach(func() {
				tkr.Spec.OSImages[rand.Intn(len(tkr.Spec.OSImages))].Name += "-missing"
			})

			repeat(10, func() {
				It("should return an error indicating that the OSImage is missing", func() {
					err, hasReason := r.checkOSImages(ctx, tkr).(reasons.HasReason)
					Expect(hasReason).To(BeTrue(), "err should impl reasons.HasReason")
					Expect(err).To(HaveOccurred(), "tkr should be missing an OSImage")
					expectedReason := reasons.MissingOSImage("").Reason()
					Expect(err.Reason()).To(Equal(expectedReason))

					Expect(r.setValidCondition(ctx, tkr)).To(Succeed())
					cond := conditions.Get(tkr, runv1.ConditionValid)
					Expect(cond.Status).To(Equal(corev1.ConditionFalse))
					Expect(cond.Reason).To(Equal(expectedReason))
				})
			})
		})

		When("an OSImage has a different Kubernetes version", func() {
			BeforeEach(func() {
				osImageName := tkr.Spec.OSImages[rand.Intn(len(tkr.Spec.OSImages))].Name
				osImage := osImages[osImageName]
				osImage.Spec.KubernetesVersion += "-different"
			})

			repeat(10, func() {
				It("should return an error indicating that the OSImage is missing", func() {
					err, hasReason := r.checkOSImages(ctx, tkr).(reasons.HasReason)
					Expect(hasReason).To(BeTrue(), "err should impl reasons.HasReason")
					Expect(err).To(HaveOccurred(), "tkr should have an OSImage with mismatched Kubernetes version")
					expectedReason := reasons.OSImageVersionMismatch("").Reason()
					Expect(err.Reason()).To(Equal(expectedReason))

					Expect(r.setValidCondition(ctx, tkr)).To(Succeed())
					cond := conditions.Get(tkr, runv1.ConditionValid)
					Expect(cond.Status).To(Equal(corev1.ConditionFalse))
					Expect(cond.Reason).To(Equal(expectedReason))
				})
			})
		})
	})

	Describe("r.checkBootstrapPackages()", func() {
		When("no changes are made to generated TKRs", func() {
			repeat(100, func() {
				It("should not return errors", func() {
					Expect(r.checkBootstrapPackages(ctx, tkr)).To(Succeed(), "generated TKRs are expected to be valid")

					Expect(r.setValidCondition(ctx, tkr)).To(Succeed())
					cond := conditions.Get(tkr, runv1.ConditionValid)
					Expect(cond.Status).To(Equal(corev1.ConditionTrue))
				})
			})
		})

		When("an bootstrapPackage is missing", func() {
			BeforeEach(func() {
				tkr.Spec.BootstrapPackages = append(tkr.Spec.BootstrapPackages, corev1.LocalObjectReference{Name: "missing-package"})
			})

			repeat(100, func() {
				It("return an error indicating that the bootstrap Package is missing", func() {
					err, hasReason := r.checkBootstrapPackages(ctx, tkr).(reasons.HasReason)
					Expect(hasReason).To(BeTrue(), "err should impl reasons.HasReason")
					Expect(err).To(HaveOccurred(), "tkr should be missing a bootstrap Package")
					expectedReason := "MissingBootstrapPackage"
					Expect(err.Reason()).To(Equal(expectedReason))

					Expect(r.setValidCondition(ctx, tkr)).To(Succeed())
					cond := conditions.Get(tkr, runv1.ConditionValid)
					Expect(cond.Status).To(Equal(corev1.ConditionFalse))
					Expect(cond.Reason).To(Equal(expectedReason))
				})
			})
		})
	})

	Describe("r.checkClusterBootstrapTemplate()", func() {
		When("no changes are made to generated TKRs", func() {
			It("should not return errors", func() {
				Expect(r.checkClusterBootstrapTemplate(ctx, tkr)).To(Succeed(), "generated TKRs are expected to be valid")

				Expect(r.setValidCondition(ctx, tkr)).To(Succeed())
				cond := conditions.Get(tkr, runv1.ConditionValid)
				Expect(cond.Status).To(Equal(corev1.ConditionTrue))
			})
		})

		When("the CBT is missing", func() {
			JustBeforeEach(func() {
				cbt := &runv1.ClusterBootstrapTemplate{
					ObjectMeta: metav1.ObjectMeta{
						Name:      tkr.Name,
						Namespace: ns,
					},
				}
				Expect(r.Client.Delete(ctx, cbt)).To(Succeed())
			})

			When("the TKR is not a legacy TKR", func() {
				It("should return an error indicating that the CBT is missing", func() {
					err, hasReason := r.checkClusterBootstrapTemplate(ctx, tkr).(reasons.HasReason)
					Expect(hasReason).To(BeTrue(), "err should impl reasons.HasReason")
					Expect(err).To(HaveOccurred(), "tkr should be missing a CBT")
					expectedReason := reasons.MissingClusterBootstrapTemplate("").Reason()
					Expect(err.Reason()).To(Equal(expectedReason))

					Expect(r.setValidCondition(ctx, tkr)).To(Succeed())
					cond := conditions.Get(tkr, runv1.ConditionValid)
					Expect(cond.Status).To(Equal(corev1.ConditionFalse))
					Expect(cond.Reason).To(Equal(expectedReason))
				})
			})

			When("the TKR is a legacy TKR", func() {
				BeforeEach(func() {
					tkr.Labels = labels.Set{}
					tkr.Labels[runv1.LabelLegacyTKR] = ""
				})

				It("should not return an error", func() {
					Expect(r.checkBootstrapPackages(ctx, tkr)).To(Succeed(), "legacy TKR should not need a CBT")

					Expect(r.setValidCondition(ctx, tkr)).To(Succeed())
					cond := conditions.Get(tkr, runv1.ConditionValid)
					Expect(cond.Status).To(Equal(corev1.ConditionTrue))
				})
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
