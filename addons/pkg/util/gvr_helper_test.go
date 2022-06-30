package util_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	k8sfake "k8s.io/client-go/kubernetes/fake"

	kapppkgv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	"github.com/vmware-tanzu/tanzu-framework/addons/test/testutil"
	antreaconfigv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cni/v1alpha1"
	vspherecpiv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cpi/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

var _ = Describe("GVRHelper", func() {

	var (
		fakeClientSet *k8sfake.Clientset

		fakeDiscovery *testutil.FakeDiscovery
		scheme        *runtime.Scheme
		gvrHelper     util.GVRHelper
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		_ = corev1.AddToScheme(scheme)
		_ = antreaconfigv1alpha1.AddToScheme(scheme)
		_ = kapppkgv1alpha1.AddToScheme(scheme)
		_ = v1alpha3.AddToScheme(scheme)

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
			APIGroups: &metav1.APIGroupList{
				Groups: []metav1.APIGroup{
					{
						Name: "cni.tanzu.vmware.com",
						Versions: []metav1.GroupVersionForDiscovery{
							{
								GroupVersion: "cni.tanzu.vmware.com/v1alpha1",
								Version:      "v1alpha1",
							},
						},
						PreferredVersion: metav1.GroupVersionForDiscovery{
							GroupVersion: "cni.tanzu.vmware.com/v1alpha1",
							Version:      "v1alpha1",
						},
					},
				},
			},
		}
		gvrHelper = util.NewGVRHelper(context.TODO(), fakeDiscovery)

	})

	Context("when an existing API server resource is looked up", func() {

		It("should return a result with correct group version resource", func() {
			group := "cni.tanzu.vmware.com"
			version := "v1alpha1"
			resource := "antreaconfigs"
			antreaGVR := schema.GroupVersionResource{Group: group, Version: version, Resource: resource}

			found, err := gvrHelper.GetGVR(schema.GroupKind{Group: "cni.tanzu.vmware.com", Kind: "AntreaConfig"})
			Expect(err).ToNot(HaveOccurred())
			Expect(*found).To(Equal(antreaGVR))

		})
	})

	Context("when an API server resource that does not exist is looked up", func() {

		It("should return an error", func() {
			_, err := gvrHelper.GetGVR(schema.GroupKind{Group: "foo.tanzu.vmware.com", Kind: "FooBar"})
			Expect(err).To(HaveOccurred())
		})
	})

	Context("GetDiscoveryClient", func() {

		It("should not be nil", func() {
			Expect(gvrHelper.GetDiscoveryClient()).ShouldNot(BeNil())
		})
	})
})
