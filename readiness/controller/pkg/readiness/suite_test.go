// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package readiness

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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
var timeout = 5 * time.Second
var interval = 100 * time.Millisecond
var setupLog = ctrl.Log.WithName("controllers").WithName("readiness")
var cancel context.CancelFunc

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
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

	err = (&ReadinessReconciler{
		Client: k8sManager.GetClient(),
		Scheme: k8sManager.GetScheme(),
		Log:    setupLog,
	}).SetupWithManager(k8sManager)
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
	It("Readiness with no checks should succeed", func() {
		readiness := getTestReadiness()
		err := k8sClient.Create(ctx, readiness)
		Expect(err).To(BeNil())

		Eventually(func() bool {
			ans := corev1alpha2.Readiness{}
			err := k8sClient.Get(ctx, types.NamespacedName{Name: readiness.Name}, &ans)
			return err == nil && ans.Status.Ready
		}, timeout, interval).Should(BeTrue())
	})

	It("Readiness with one check and no matching providers should not be ready", func() {
		provider := getTestReadinessProvider()
		provider.Spec.CheckRefs = []string{"check2"}
		provider.Status.State = corev1alpha2.ProviderSuccessState
		err := k8sClient.Create(ctx, provider)
		Expect(err).To(BeNil())

		readiness := getTestReadiness()
		readiness.Spec.Checks = append(readiness.Spec.Checks, corev1alpha2.Check{
			Name: "check1",
			Type: corev1alpha2.BasicReadinessCheck,
		})
		err = k8sClient.Create(ctx, readiness)
		Expect(err).To(BeNil())

		Eventually(func() bool {
			ans := corev1alpha2.Readiness{}
			err := k8sClient.Get(ctx, types.NamespacedName{Name: readiness.Name}, &ans)
			return err == nil && !ans.Status.Ready && len(ans.Status.CheckStatus) > 0 && len(ans.Status.CheckStatus[0].Providers) == 0
		}, timeout, interval).Should(BeTrue())
	})

	It("Readiness with one check and matching providers in inactive state; should not be ready", func() {
		provider := getTestReadinessProvider()
		provider.Spec.CheckRefs = []string{"check3"}
		provider.Status.State = corev1alpha2.ProviderFailureState
		err := k8sClient.Create(ctx, provider)
		Expect(err).To(BeNil())

		readiness := getTestReadiness()
		readiness.Spec.Checks = append(readiness.Spec.Checks, corev1alpha2.Check{
			Name: "check3",
			Type: corev1alpha2.BasicReadinessCheck,
		})
		err = k8sClient.Create(ctx, readiness)
		Expect(err).To(BeNil())

		Eventually(func() bool {
			ans := corev1alpha2.Readiness{}
			err := k8sClient.Get(ctx, types.NamespacedName{Name: readiness.Name}, &ans)
			return err == nil &&
				!ans.Status.Ready &&
				len(ans.Status.CheckStatus) == 1 &&
				len(ans.Status.CheckStatus[0].Providers) == 1
		}, timeout, interval).Should(BeTrue())
	})

	It("Readiness with one check and matching providers in active state; should be ready", func() {
		provider := getTestReadinessProvider()
		provider.Spec.CheckRefs = []string{"check4"}
		err := k8sClient.Create(ctx, provider)
		Expect(err).To(BeNil())

		provider.Status.State = corev1alpha2.ProviderSuccessState
		provider.Status.Conditions = []corev1alpha2.ReadinessConditionStatus{}
		err = k8sClient.Status().Update(ctx, provider)
		Expect(err).To(BeNil())

		readiness := getTestReadiness()
		readiness.Spec.Checks = append(readiness.Spec.Checks, corev1alpha2.Check{
			Name: "check4",
			Type: corev1alpha2.BasicReadinessCheck,
		})
		err = k8sClient.Create(ctx, readiness)
		Expect(err).To(BeNil())

		Eventually(func() bool {
			ans := corev1alpha2.Readiness{}
			err := k8sClient.Get(ctx, types.NamespacedName{Name: readiness.Name}, &ans)
			return err == nil &&
				ans.Status.Ready &&
				len(ans.Status.CheckStatus[0].Providers) == 1
		}, timeout, interval).Should(BeTrue())
	})

	It("Readiness with one check and two matching providers", func() {
		provider1 := getTestReadinessProvider()
		provider1.Spec.CheckRefs = []string{"check5"}
		err := k8sClient.Create(ctx, provider1)
		Expect(err).To(BeNil())

		provider1.Status.State = corev1alpha2.ProviderFailureState
		provider1.Status.Conditions = []corev1alpha2.ReadinessConditionStatus{}
		err = k8sClient.Status().Update(ctx, provider1)
		Expect(err).To(BeNil())

		provider2 := getTestReadinessProvider()
		provider2.Spec.CheckRefs = []string{"check5"}
		err = k8sClient.Create(ctx, provider2)
		Expect(err).To(BeNil())

		provider2.Status.State = corev1alpha2.ProviderFailureState
		provider2.Status.Conditions = []corev1alpha2.ReadinessConditionStatus{}
		err = k8sClient.Status().Update(ctx, provider2)
		Expect(err).To(BeNil())

		readiness := getTestReadiness()
		readiness.Spec.Checks = append(readiness.Spec.Checks, corev1alpha2.Check{
			Name: "check5",
			Type: corev1alpha2.BasicReadinessCheck,
		})
		err = k8sClient.Create(ctx, readiness)
		Expect(err).To(BeNil())

		// None of the providers is active
		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: readiness.Name}, readiness)
			return err == nil &&
				!readiness.Status.Ready &&
				len(readiness.Status.CheckStatus) == 1 &&
				len(readiness.Status.CheckStatus[0].Providers) == 2
		}, timeout, interval).Should(BeTrue())

		provider2.Status.State = corev1alpha2.ProviderSuccessState
		provider2.Status.Conditions = []corev1alpha2.ReadinessConditionStatus{}
		err = k8sClient.Status().Update(ctx, provider2)
		Expect(err).To(BeNil())

		// One of the providers is active
		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: readiness.Name}, readiness)
			return err == nil &&
				readiness.Status.Ready &&
				len(readiness.Status.CheckStatus[0].Providers) == 2
		}, timeout, interval).Should(BeTrue())

		provider1.Status.State = corev1alpha2.ProviderSuccessState
		provider1.Status.Conditions = []corev1alpha2.ReadinessConditionStatus{}
		err = k8sClient.Status().Update(ctx, provider1)
		Expect(err).To(BeNil())

		// Both proviers are active
		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: readiness.Name}, readiness)
			return err == nil &&
				readiness.Status.Ready &&
				len(readiness.Status.CheckStatus[0].Providers) == 2
		}, timeout, interval).Should(BeTrue())

	})

	It("Readiness with two checks", func() {
		readiness := getTestReadiness()
		readiness.Spec.Checks = append(readiness.Spec.Checks, corev1alpha2.Check{
			Name: "check6",
			Type: corev1alpha2.BasicReadinessCheck,
		})
		readiness.Spec.Checks = append(readiness.Spec.Checks, corev1alpha2.Check{
			Name: "check7",
			Type: corev1alpha2.BasicReadinessCheck,
		})
		err := k8sClient.Create(ctx, readiness)
		Expect(err).To(BeNil())

		// Two checks and no providers
		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: readiness.Name}, readiness)
			return err == nil &&
				!readiness.Status.Ready &&
				len(readiness.Status.CheckStatus) == 2 &&
				len(readiness.Status.CheckStatus[0].Providers) == 0 &&
				len(readiness.Status.CheckStatus[1].Providers) == 0
		}, timeout, interval).Should(BeTrue())

		provider1 := getTestReadinessProvider()
		provider1.Spec.CheckRefs = []string{"check6"}
		err = k8sClient.Create(ctx, provider1)
		Expect(err).To(BeNil())

		// Provider is available for one check and the provider is not successful
		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: readiness.Name}, readiness)
			return err == nil &&
				!readiness.Status.Ready &&
				len(readiness.Status.CheckStatus[0].Providers) == 1 &&
				len(readiness.Status.CheckStatus[1].Providers) == 0
		}, timeout, interval).Should(BeTrue())

		provider2 := getTestReadinessProvider()
		provider2.Spec.CheckRefs = []string{"check7"}
		err = k8sClient.Create(ctx, provider2)
		Expect(err).To(BeNil())

		// Provider is available for two checks and both providers are not successful
		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: readiness.Name}, readiness)
			return err == nil &&
				!readiness.Status.Ready &&
				len(readiness.Status.CheckStatus[0].Providers) == 1 &&
				len(readiness.Status.CheckStatus[1].Providers) == 1
		}, timeout, interval).Should(BeTrue())

		provider2.Status.State = corev1alpha2.ProviderSuccessState
		provider2.Status.Conditions = []corev1alpha2.ReadinessConditionStatus{}
		err = k8sClient.Status().Update(ctx, provider2)
		Expect(err).To(BeNil())

		// Provider is available for two checks and one of the providers is successful
		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: readiness.Name}, readiness)
			return err == nil &&
				!readiness.Status.Ready &&
				len(readiness.Status.CheckStatus[0].Providers) == 1 &&
				len(readiness.Status.CheckStatus[1].Providers) == 1 &&
				readiness.Status.CheckStatus[1].Ready &&
				!readiness.Status.CheckStatus[0].Ready
		}, timeout, interval).Should(BeTrue())

		provider1.Status.State = corev1alpha2.ProviderSuccessState
		provider1.Status.Conditions = []corev1alpha2.ReadinessConditionStatus{}
		err = k8sClient.Status().Update(ctx, provider1)
		Expect(err).To(BeNil())

		// Providers for all the checks are successful
		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: readiness.Name}, readiness)
			return err == nil &&
				readiness.Status.Ready &&
				len(readiness.Status.CheckStatus) == 2 &&
				len(readiness.Status.CheckStatus[0].Providers) == 1 &&
				len(readiness.Status.CheckStatus[1].Providers) == 1 &&
				readiness.Status.CheckStatus[1].Ready &&
				readiness.Status.CheckStatus[0].Ready
		}, timeout, interval).Should(BeTrue())

		readiness.Spec.Checks = []corev1alpha2.Check{readiness.Spec.Checks[0]}
		readiness.Spec.Checks = append(readiness.Spec.Checks, corev1alpha2.Check{
			Name: "check8",
			Type: corev1alpha2.BasicReadinessCheck,
		})
		err = k8sClient.Update(ctx, readiness)
		Expect(err).To(BeNil())

		// One check is removed and a new check is added
		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: readiness.Name}, readiness)
			return err == nil &&
				!readiness.Status.Ready &&
				len(readiness.Status.CheckStatus[0].Providers) == 1 &&
				len(readiness.Status.CheckStatus[1].Providers) == 0 &&
				readiness.Status.CheckStatus[0].Ready
		}, timeout, interval).Should(BeTrue())
	})

	It("Readiness with two checks and one provider", func() {
		readiness := getTestReadiness()
		readiness.Spec.Checks = append(readiness.Spec.Checks, corev1alpha2.Check{
			Name: "check8",
			Type: corev1alpha2.BasicReadinessCheck,
		})
		readiness.Spec.Checks = append(readiness.Spec.Checks, corev1alpha2.Check{
			Name: "check9",
			Type: corev1alpha2.BasicReadinessCheck,
		})
		err := k8sClient.Create(ctx, readiness)
		Expect(err).To(BeNil())

		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: readiness.Name}, readiness)
			return err == nil &&
				!readiness.Status.Ready &&
				len(readiness.Status.CheckStatus) == 2 &&
				len(readiness.Status.CheckStatus[0].Providers) == 0 &&
				len(readiness.Status.CheckStatus[1].Providers) == 0
		}, timeout, interval).Should(BeTrue())

		provider1 := getTestReadinessProvider()
		provider1.Spec.CheckRefs = []string{"check8", "check9"}
		err = k8sClient.Create(ctx, provider1)
		Expect(err).To(BeNil())

		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: readiness.Name}, readiness)
			return err == nil &&
				!readiness.Status.Ready &&
				len(readiness.Status.CheckStatus[0].Providers) == 1 &&
				len(readiness.Status.CheckStatus[1].Providers) == 1
		}, timeout, interval).Should(BeTrue())

		provider1.Status.State = corev1alpha2.ProviderSuccessState
		provider1.Status.Conditions = []corev1alpha2.ReadinessConditionStatus{}
		err = k8sClient.Status().Update(ctx, provider1)
		Expect(err).To(BeNil())

		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: readiness.Name}, readiness)
			return err == nil &&
				readiness.Status.Ready &&
				len(readiness.Status.CheckStatus[0].Providers) == 1 &&
				len(readiness.Status.CheckStatus[1].Providers) == 1
		}, timeout, interval).Should(BeTrue())
	})

})

func getTestReadiness() *corev1alpha2.Readiness {
	return &corev1alpha2.Readiness{
		ObjectMeta: v1.ObjectMeta{
			GenerateName: "test-readiness-",
		},
		Spec: corev1alpha2.ReadinessSpec{
			Checks: []corev1alpha2.Check{},
		},
	}
}

func getTestReadinessProvider() *corev1alpha2.ReadinessProvider {
	return &corev1alpha2.ReadinessProvider{
		ObjectMeta: v1.ObjectMeta{
			GenerateName: "test-readiness-provider-",
		},
		Spec: corev1alpha2.ReadinessProviderSpec{
			Conditions: []corev1alpha2.ReadinessProviderCondition{},
		},
	}
}
