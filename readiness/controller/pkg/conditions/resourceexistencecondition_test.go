package conditions

import (
	"context"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	corev1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/core/v1alpha2"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var ctx context.Context
var cancel context.CancelFunc
var dynamicClient *dynamic.DynamicClient
var discoveryClient *discovery.DiscoveryClient

func TestResourceExistenceCondition(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "ResourceExistenceCondition Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "..", "..", "apis", "core", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = corev1alpha2.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	ctx, cancel = context.WithCancel(context.TODO())

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme.Scheme,
		MetricsBindAddress: "0",
		Host:               "127.0.0.1",
		Port:               9443,
	})
	Expect(err).ToNot(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	dynamicClient, err = dynamic.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())

	discoveryClient, err = discovery.NewDiscoveryClientForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err = k8sManager.Start(ctx)
		Expect(err).ToNot(HaveOccurred(), "failed to run manager")
	}()

})

var _ = AfterSuite(func() {
	cancel()
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

var _ = Describe("Readiness controller", func() {
	It("should succeed when querying an existing namespaced resource", func() {
		newPod := v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "testpod",
				Namespace: "default",
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:  "test-container",
						Image: "test:tag",
					},
				},
			},
		}

		err := k8sClient.Create(context.TODO(), &newPod)
		Expect(err).To(BeNil())

		state, msg := NewResourceExistencConditionFunc(dynamicClient, discoveryClient)(context.TODO(), &corev1alpha2.ResourceExistenceCondition{
			APIVersion: "v1",
			Kind:       "Pod",
			Namespace:  &newPod.Namespace,
			Name:       newPod.Name,
		})

		Expect(state).To(Equal(corev1alpha2.ConditionSuccessState))
		Expect(msg).To(Equal(""))
	})

	It("should fail when querying a non-existing namespaced resource", func() {
		state, msg := NewResourceExistencConditionFunc(dynamicClient, discoveryClient)(context.TODO(), &corev1alpha2.ResourceExistenceCondition{
			APIVersion: "v1",
			Kind:       "Pod",
			Namespace: func() *string {
				n := "default"
				return &n
			}(),
			Name: "somename",
		})

		Expect(state).To(Equal(corev1alpha2.ConditionFailureState))
		Expect(msg).To(Equal("pods \"somename\" not found"))
	})

	It("should succeed when querying an existing cluster scoped resource", func() {
		state, msg := NewResourceExistencConditionFunc(dynamicClient, discoveryClient)(context.TODO(), &corev1alpha2.ResourceExistenceCondition{
			APIVersion: "apiextensions.k8s.io/v1",
			Kind:       "CustomResourceDefinition",
			Name:       "readinesses.core.tanzu.vmware.com",
		})

		Expect(state).To(Equal(corev1alpha2.ConditionSuccessState))
		Expect(msg).To(Equal(""))
	})

	It("should succeed when querying a non-existing cluster scoped resource", func() {
		state, msg := NewResourceExistencConditionFunc(dynamicClient, discoveryClient)(context.TODO(), &corev1alpha2.ResourceExistenceCondition{
			APIVersion: "apiextensions.k8s.io/v1",
			Kind:       "CustomResourceDefinition",
			Name:       "readinesses.config.tanzu.vmware.com",
		})

		Expect(state).To(Equal(corev1alpha2.ConditionFailureState))
		Expect(msg).To(Equal("customresourcedefinitions.apiextensions.k8s.io \"readinesses.config.tanzu.vmware.com\" not found"))
	})
})
