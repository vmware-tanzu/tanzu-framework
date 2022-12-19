// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
//go:build !race

package crdwait

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/testing"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2/klogr"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1beta1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	kapppkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	runtanzuv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
)

var fakeClientSet *fake.Clientset
var fakeReaderWriter *fake.Clientset

var _ = Describe("WaitForCRDs", func() {

	var scheme *runtime.Scheme
	var crdWaiter CRDWaiter
	var fakeRecorder *record.FakeRecorder
	scheme = runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		}}

	BeforeEach(func() {
		fakeClientSet = fake.NewSimpleClientset(pod)
		// TODO @randomvariable: Fake APIReader

		crdWaiter = CRDWaiter{
			Ctx:         ctx,
			Client: 		fakeClientSet,
			APIReader: fakeReaderWriter,
			Logger:      klogr.New(),
			Scheme:      scheme,
		}
		fakeRecorder = record.NewFakeRecorder(50)
		crdWaiter.eventRecorder = fakeRecorder
	})

	Context("when resources do not exist ", func() {
		It("should fail", func() {
			var crds = map[schema.GroupVersion]*sets.String{}
			// cluster-api
			clusterapiv1alpha3Resources := sets.NewString("clusters")
			crds[clusterapiv1beta1.GroupVersion] = &clusterapiv1alpha3Resources

			// tkr
			runtanzuv1alpha1Resources := sets.NewString("tanzukubernetesreleases")
			crds[runtanzuv1alpha1.GroupVersion] = &runtanzuv1alpha1Resources

			crdWaiter.PollInterval = time.Second
			crdWaiter.PollTimeout = time.Second

			Expect(crdWaiter.WaitForCRDs(crds, pod, "foo")).To(HaveOccurred())

		})
	})

	Context("when resources exist ", func() {
		It("should not fail", func() {
			var crds = map[schema.GroupVersion]*sets.String{}
			// cluster-api
			clusterapiv1alpha3Resources := sets.NewString("clusters")
			crds[clusterapiv1beta1.GroupVersion] = &clusterapiv1alpha3Resources

			fakeClientSet.AddReactor("get", "resource", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
				return true, nil, nil
			})

			fakeClientSet.Resources = append(fakeClientSet.Resources,
				&metav1.APIResourceList{GroupVersion: clusterapiv1beta1.GroupVersion.String(),
					APIResources: []metav1.APIResource{
						{Name: "clusters", Namespaced: true, Kind: "Cluster"},
					},
				})
			crdWaiter.PollInterval = time.Second
			crdWaiter.PollTimeout = time.Second

			Expect(crdWaiter.WaitForCRDs(crds, pod, "foo")).NotTo(HaveOccurred())

		})
	})

	Context("when addon controller CRDs exist ", func() {
		It("should not fail", func() {
			fakeClientSet.AddReactor("get", "resource", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
				return true, nil, nil
			})

			fakeClientSet.Resources = append(fakeClientSet.Resources,
				&metav1.APIResourceList{GroupVersion: clusterapiv1beta1.GroupVersion.String(),
					APIResources: []metav1.APIResource{
						{Name: "clusters", Namespaced: true, Kind: "Cluster"},
					},
				},
				&metav1.APIResourceList{GroupVersion: controlplanev1beta1.GroupVersion.String(),
					APIResources: []metav1.APIResource{
						{Name: "kubeadmcontrolplanes", Namespaced: true, Kind: "KubeadmControlPlane"},
					},
				},
				&metav1.APIResourceList{GroupVersion: runtanzuv1alpha1.GroupVersion.String(),
					APIResources: []metav1.APIResource{
						{Name: "tanzukubernetesreleases", Namespaced: true, Kind: "TanzuKubernetesRelease"},
					},
				},
				&metav1.APIResourceList{GroupVersion: kappctrl.SchemeGroupVersion.String(),
					APIResources: []metav1.APIResource{
						{Name: "apps", Namespaced: true, Kind: "App"},
					},
				},
				&metav1.APIResourceList{GroupVersion: kapppkg.SchemeGroupVersion.String(),
					APIResources: []metav1.APIResource{
						{Name: "packageinstalls", Namespaced: true, Kind: "PackageInstalls"},
						{Name: "packagerepositories", Namespaced: true, Kind: "PackageRepositories"},
					},
				},
			)

			crdWaiter.PollInterval = time.Second
			crdWaiter.PollTimeout = time.Second

			Expect(crdWaiter.WaitForCRDs(getCRDs(), pod, "foo")).NotTo(HaveOccurred())
		})
	})

	Context("when resources eventually exist ", func() {
		It("should not fail", func() {
			var crds = map[schema.GroupVersion]*sets.String{}
			// cluster-api
			clusterapiv1alpha3Resources := sets.NewString("clusters")
			crds[clusterapiv1beta1.GroupVersion] = &clusterapiv1alpha3Resources

			fakeClientSet.AddReactor("get", "resource", func(action testing.Action) (handled bool, ret runtime.Object, err error) {
				return true, nil, nil
			})

			go func() {
				time.Sleep(time.Second * 3)
				fakeClientSet.Resources = append(fakeClientSet.Resources,
					&metav1.APIResourceList{GroupVersion: clusterapiv1beta1.GroupVersion.String(),
						APIResources: []metav1.APIResource{
							{Name: "clusters", Namespaced: true, Kind: "Cluster"},
						},
					})
			}()

			crdWaiter.PollInterval = time.Second
			crdWaiter.PollTimeout = time.Second * 10

			Expect(crdWaiter.WaitForCRDs(crds, pod, "foo")).NotTo(HaveOccurred())

		})
	})

	Context("when resources do not exist ", func() {
		It("should fail and emit events for missing GroupVersion", func() {
			var crds = map[schema.GroupVersion]*sets.String{}
			// cluster-api
			clusterapiv1alpha3Resources := sets.NewString("clusters")
			crds[clusterapiv1beta1.GroupVersion] = &clusterapiv1alpha3Resources

			crdWaiter.PollInterval = time.Second
			crdWaiter.PollTimeout = time.Second * 2

			Expect(crdWaiter.WaitForCRDs(crds, pod, "foo")).To(HaveOccurred())
			Expect(<-fakeRecorder.Events).To(ContainSubstring(fmt.Sprintf("The GroupVersion '%s' is not available yet", clusterapiv1beta1.GroupVersion.String())))

		})
	})

	Context("when GroupVersion exists but not all api-resources exist ", func() {
		It("should fail and emit events for missing api-resources", func() {
			var crds = map[schema.GroupVersion]*sets.String{}
			// cluster-api
			clusterapiv1alpha3Resources := sets.NewString("clusters")
			crds[clusterapiv1beta1.GroupVersion] = &clusterapiv1alpha3Resources

			fakeClientSet.Resources = append(fakeClientSet.Resources,
				&metav1.APIResourceList{GroupVersion: clusterapiv1beta1.GroupVersion.String(),
					APIResources: []metav1.APIResource{
						{Name: "foo", Namespaced: true, Kind: "Cluster"},
					},
				})
			crdWaiter.PollInterval = time.Second
			crdWaiter.PollTimeout = time.Second * 2

			Expect(crdWaiter.WaitForCRDs(crds, pod, "foo")).To(HaveOccurred())
			Expect(<-fakeRecorder.Events).To(ContainSubstring(fmt.Sprintf("The api-resources '[clusters]' in GroupVersion '%s' are not available yet", clusterapiv1beta1.GroupVersion.String())))

		})
	})
})

func getFakeClientSet() (kubernetes.Interface, error) {
	return fakeClientSet, nil
}

func getCRDs() map[schema.GroupVersion]*sets.String {
	var crds = map[schema.GroupVersion]*sets.String{}
	// cluster-api
	clusterapiv1alpha3Resources := sets.NewString("clusters")
	crds[clusterapiv1beta1.GroupVersion] = &clusterapiv1alpha3Resources

	controlplanev1alpha3Resources := sets.NewString("kubeadmcontrolplanes")
	crds[controlplanev1beta1.GroupVersion] = &controlplanev1alpha3Resources

	// tkr
	runtanzuv1alpha1Resources := sets.NewString("tanzukubernetesreleases")
	crds[runtanzuv1alpha1.GroupVersion] = &runtanzuv1alpha1Resources

	// kapp-controller APIs
	kappctrlv1alpha1Resources := sets.NewString("apps")
	crds[kappctrl.SchemeGroupVersion] = &kappctrlv1alpha1Resources

	kapppkgv1alpha1Resources := sets.NewString("packageinstalls", "packagerepositories")
	crds[kapppkg.SchemeGroupVersion] = &kapppkgv1alpha1Resources

	return crds
}
