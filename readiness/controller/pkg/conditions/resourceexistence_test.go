// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package conditions

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	corev1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/core/v1alpha2"
	capabilitiesDiscovery "github.com/vmware-tanzu/tanzu-framework/capabilities/client/pkg/discovery"
	"github.com/vmware-tanzu/tanzu-framework/util/config"
	testutil "github.com/vmware-tanzu/tanzu-framework/util/test"
)

var (
	cfg          *rest.Config
	k8sClient    client.Client
	testEnv      *envtest.Environment
	ctx          context.Context
	cancel       context.CancelFunc
	queryClient  *capabilitiesDiscovery.ClusterQueryClient
	k8sClientset *kubernetes.Clientset
)

const defaultNamespace = "default"

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

	ctx, cancel = context.WithCancel(context.TODO())

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		MetricsBindAddress: "0",
		Host:               "127.0.0.1",
		Port:               9443,
	})
	Expect(err).ToNot(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	queryClient, err = capabilitiesDiscovery.NewClusterQueryClientForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())
	Expect(queryClient).NotTo(BeNil())

	// k8sClientset is package-scoped
	k8sClientset, err = kubernetes.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClientset).NotTo(BeNil())

	manifestBytes, err := os.ReadFile("testdata/rbac.yaml")
	Expect(err).ToNot(HaveOccurred())

	dynamicClient, err := dynamic.NewForConfig(cfg)
	Expect(err).NotTo(HaveOccurred())
	Expect(dynamicClient).NotTo(BeNil())

	err = testutil.CreateResourcesFromManifest(manifestBytes, cfg, dynamicClient)
	Expect(err).ToNot(HaveOccurred())

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
				Namespace: defaultNamespace,
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

		state, msg := NewResourceExistenceConditionFunc()(context.TODO(), queryClient, &corev1alpha2.ResourceExistenceCondition{
			APIVersion: "v1",
			Kind:       "Pod",
			Namespace:  &newPod.Namespace,
			Name:       newPod.Name,
		},
			"podCondition")

		Expect(state).To(Equal(corev1alpha2.ConditionSuccessState))
		Expect(msg).To(Equal("resource found"))
	})

	It("should fail when querying a non-existing namespaced resource", func() {
		state, msg := NewResourceExistenceConditionFunc()(context.TODO(), queryClient, &corev1alpha2.ResourceExistenceCondition{
			APIVersion: "v1",
			Kind:       "Pod",
			Namespace: func() *string {
				n := defaultNamespace
				return &n
			}(),
			Name: "somename",
		},
			"nonExistingCondition")

		Expect(state).To(Equal(corev1alpha2.ConditionFailureState))
		Expect(msg).To(Equal("resource not found"))
	})

	It("should succeed when querying an existing cluster scoped resource", func() {
		state, msg := NewResourceExistenceConditionFunc()(context.TODO(), queryClient, &corev1alpha2.ResourceExistenceCondition{
			APIVersion: "apiextensions.k8s.io/v1",
			Kind:       "CustomResourceDefinition",
			Name:       "readinesses.core.tanzu.vmware.com",
		},
			"crdCondition")

		Expect(state).To(Equal(corev1alpha2.ConditionSuccessState))
		Expect(msg).To(Equal("resource found"))
	})

	It("should fail when querying a non-existing cluster scoped resource", func() {
		state, msg := NewResourceExistenceConditionFunc()(context.TODO(), queryClient, &corev1alpha2.ResourceExistenceCondition{
			APIVersion: "apiextensions.k8s.io/v1",
			Kind:       "CustomResourceDefinition",
			Name:       "readinesses.config.tanzu.vmware.com",
		},
			"crdCondition")

		Expect(state).To(Equal(corev1alpha2.ConditionFailureState))
		Expect(msg).To(Equal("resource not found"))
	})

	It("should fail when resourceExistenceCondition is undefined", func() {
		state, msg := NewResourceExistenceConditionFunc()(context.TODO(), queryClient, nil, "undefinedCondition")

		Expect(state).To(Equal(corev1alpha2.ConditionFailureState))
		Expect(msg).To(Equal("resourceExistenceCondition is not defined"))
	})

	It("should succeed when query client has required permissions", func() {
		customQueryClient, err := getCustomQueryClient()
		Expect(err).To(BeNil())
		Expect(customQueryClient).ToNot(BeNil())

		newPod := v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: defaultNamespace,
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

		err = k8sClient.Create(context.TODO(), &newPod)
		Expect(err).To(BeNil())

		state, msg := NewResourceExistenceConditionFunc()(context.TODO(), customQueryClient, &corev1alpha2.ResourceExistenceCondition{
			APIVersion: "v1",
			Kind:       "Pod",
			Namespace:  &newPod.Namespace,
			Name:       newPod.Name,
		},
			"podCondition")

		Expect(state).To(Equal(corev1alpha2.ConditionSuccessState))
		Expect(msg).To(Equal("resource found"))
	})

	It("should fail when query client is missing required permissions", func() {
		customQueryClient, err := getCustomQueryClient()
		Expect(err).To(BeNil())
		Expect(customQueryClient).ToNot(BeNil())

		state, _ := NewResourceExistenceConditionFunc()(context.TODO(), customQueryClient, &corev1alpha2.ResourceExistenceCondition{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       "test-deploy",
			Namespace: func() *string {
				n := defaultNamespace
				return &n
			}(),
		},
			"deploymentCondition")

		Expect(state).To(Equal(corev1alpha2.ConditionFailureState))
	})
})

func getCustomQueryClient() (*capabilitiesDiscovery.ClusterQueryClient, error) {
	customCfg, err := config.GetConfigForServiceAccount(ctx, k8sClientset, cfg, defaultNamespace, "pod-sa")
	if err != nil {
		return nil, err
	}
	return capabilitiesDiscovery.NewClusterQueryClientForConfig(customCfg)
}
