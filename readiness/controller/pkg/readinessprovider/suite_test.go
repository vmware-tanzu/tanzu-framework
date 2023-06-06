// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package readinessprovider

import (
	"context"
	"encoding/base64"
	"log"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/cert"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	corev1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/core/v1alpha2"
	testutil "github.com/vmware-tanzu/tanzu-framework/util/test"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var ctx context.Context
var cancel context.CancelFunc
var setupLog = ctrl.Log.WithName("controllers").WithName("readinessprovider")
var timeout = 10 * time.Second
var interval = 100 * time.Millisecond
var calls int
var tmpDir string
var generatedWebhookManifestBytes []byte

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

func generateCertificateAndManifests() error {
	cert, key, err := cert.GenerateSelfSignedCertKey("tanzu-readinessprovider-webhook-service.default.svc", []net.IP{}, []string{})
	if err != nil {
		return err
	}

	tmpDir, err = os.MkdirTemp("/tmp", "readinessprovider-test")
	if err != nil {
		return err
	}
	err = os.WriteFile(path.Join(tmpDir, "readinessprovider-webhook.crt"), cert, 0644)
	if err != nil {
		return err
	}

	err = os.WriteFile(path.Join(tmpDir, "readinessprovider-webhook.key"), key, 0644)
	if err != nil {
		return err
	}

	input, err := os.ReadFile("testdata/webhook.yaml")
	if err != nil {
		log.Fatalln(err)
	}

	lines := strings.Split(string(input), "\n")

	for i, line := range lines {
		if strings.Contains(line, "Cg==") {
			lines[i] = strings.Replace(lines[i], "Cg==", base64.StdEncoding.EncodeToString(cert), 1)
		}
	}
	generatedWebhookManifestBytes = []byte(strings.Join(lines, "\n"))
	return nil
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

	testEnv.ControlPlane.APIServer.Configure().Append("admission-control", "ValidatingAdmissionWebhook")

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

	err = generateCertificateAndManifests()
	Expect(err).ToNot(HaveOccurred())

	k8sManager.GetWebhookServer().TLSMinVersion = "1.2"
	k8sManager.GetWebhookServer().CertDir = tmpDir
	k8sManager.GetWebhookServer().CertName = "readinessprovider-webhook.crt"
	k8sManager.GetWebhookServer().KeyName = "readinessprovider-webhook.key"
	k8sManager.GetWebhookServer().Port = 9443

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	dynamicClient, err := dynamic.NewForConfig(cfg)
	Expect(err).ToNot(HaveOccurred())
	Expect(dynamicClient).ToNot(BeNil())

	err = (&ReadinessProviderReconciler{
		Client: k8sManager.GetClient(),
		Scheme: k8sManager.GetScheme(),
		Log:    setupLog,
		ResourceExistenceCondition: func(context context.Context, rec *corev1alpha2.ResourceExistenceCondition, conditionName string) (corev1alpha2.ReadinessConditionState, string) {
			if rec.Kind == "failurekind" {
				return corev1alpha2.ConditionFailureState, "TestFailure"
			}
			if rec.Kind == "inprogresskind" {
				return corev1alpha2.ConditionInProgressState, "TestInProgress"
			}
			if rec.Kind == "repeatkind" {
				calls++
			}

			return corev1alpha2.ConditionSuccessState, "TestSuccess"
		},
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	err = (&corev1alpha2.ReadinessProvider{}).SetupWebhookWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err = k8sManager.Start(ctx)
		Expect(err).ToNot(HaveOccurred(), "failed to run manager")
	}()

	err = testutil.CreateResourcesFromManifest(generatedWebhookManifestBytes, cfg, dynamicClient)
	Expect(err).ToNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	cancel()
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

var _ = Describe("Readiness Provider controller", func() {
	It("should fail when the condition type is empty", func() {
		readinessProvider := getTestReadinessProvider()
		readinessProvider.Spec.Conditions = append(readinessProvider.Spec.Conditions, corev1alpha2.ReadinessProviderCondition{})
		err := k8sClient.Create(ctx, readinessProvider)
		Expect(err).NotTo(BeNil())
	})

	It("should succeed when the provider has no conditions", func() {
		readinessProvider := getTestReadinessProvider()
		err := k8sClient.Create(ctx, readinessProvider)
		Expect(err).To(BeNil())

		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: readinessProvider.Name}, readinessProvider)
			return err == nil && readinessProvider.Status.State == corev1alpha2.ProviderSuccessState
		}, timeout, interval).Should(BeTrue())
	})

	It("should succeed when all the conditions satisfy", func() {
		readinessProvider := getTestReadinessProvider()
		readinessProvider.Spec.Conditions = append(readinessProvider.Spec.Conditions, corev1alpha2.ReadinessProviderCondition{
			Name:                       "cond1",
			ResourceExistenceCondition: &corev1alpha2.ResourceExistenceCondition{},
		})
		readinessProvider.Spec.Conditions = append(readinessProvider.Spec.Conditions, corev1alpha2.ReadinessProviderCondition{
			Name:                       "cond1",
			ResourceExistenceCondition: &corev1alpha2.ResourceExistenceCondition{},
		})
		err := k8sClient.Create(ctx, readinessProvider)
		Expect(err).To(BeNil())

		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: readinessProvider.Name}, readinessProvider)
			return err == nil &&
				readinessProvider.Status.State == corev1alpha2.ProviderSuccessState &&
				len(readinessProvider.Status.Conditions) == 2 &&
				readinessProvider.Status.Conditions[0].State == corev1alpha2.ConditionSuccessState &&
				readinessProvider.Status.Conditions[1].State == corev1alpha2.ConditionSuccessState
		}, timeout, interval).Should(BeTrue())
	})

	It("should fail when one of the conditions does not satisfy", func() {
		readinessProvider := getTestReadinessProvider()
		readinessProvider.Spec.Conditions = append(readinessProvider.Spec.Conditions, corev1alpha2.ReadinessProviderCondition{
			Name:                       "cond1",
			ResourceExistenceCondition: &corev1alpha2.ResourceExistenceCondition{},
		})
		readinessProvider.Spec.Conditions = append(readinessProvider.Spec.Conditions, corev1alpha2.ReadinessProviderCondition{
			Name: "cond2",
			ResourceExistenceCondition: &corev1alpha2.ResourceExistenceCondition{
				Kind: "failurekind",
			},
		})
		err := k8sClient.Create(ctx, readinessProvider)
		Expect(err).To(BeNil())

		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: readinessProvider.Name}, readinessProvider)
			return err == nil &&
				readinessProvider.Status.State == corev1alpha2.ProviderFailureState &&
				len(readinessProvider.Status.Conditions) == 2 &&
				readinessProvider.Status.Conditions[0].State == corev1alpha2.ConditionSuccessState &&
				readinessProvider.Status.Conditions[1].State == corev1alpha2.ConditionFailureState
		}, timeout, interval).Should(BeTrue())
	})

	It("should be in progress when one of the conditions is in progress and other succeed", func() {
		readinessProvider := getTestReadinessProvider()
		readinessProvider.Spec.Conditions = append(readinessProvider.Spec.Conditions, corev1alpha2.ReadinessProviderCondition{
			Name:                       "cond1",
			ResourceExistenceCondition: &corev1alpha2.ResourceExistenceCondition{},
		})
		readinessProvider.Spec.Conditions = append(readinessProvider.Spec.Conditions, corev1alpha2.ReadinessProviderCondition{
			Name: "cond2",
			ResourceExistenceCondition: &corev1alpha2.ResourceExistenceCondition{
				Kind: "inprogresskind",
			},
		})
		err := k8sClient.Create(ctx, readinessProvider)
		Expect(err).To(BeNil())

		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: readinessProvider.Name}, readinessProvider)
			return err == nil &&
				readinessProvider.Status.State == corev1alpha2.ProviderInProgressState &&
				len(readinessProvider.Status.Conditions) == 2 &&
				readinessProvider.Status.Conditions[0].State == corev1alpha2.ConditionSuccessState &&
				readinessProvider.Status.Conditions[1].State == corev1alpha2.ConditionInProgressState
		}, timeout, interval).Should(BeTrue())
	})

	It("should fail when one of the conditions does not satisfy and other is in progress", func() {
		readinessProvider := getTestReadinessProvider()
		readinessProvider.Spec.Conditions = append(readinessProvider.Spec.Conditions, corev1alpha2.ReadinessProviderCondition{
			Name: "cond1",
			ResourceExistenceCondition: &corev1alpha2.ResourceExistenceCondition{
				Kind: "inprogresskind",
			},
		})
		readinessProvider.Spec.Conditions = append(readinessProvider.Spec.Conditions, corev1alpha2.ReadinessProviderCondition{
			Name: "cond2",
			ResourceExistenceCondition: &corev1alpha2.ResourceExistenceCondition{
				Kind: "failurekind",
			},
		})
		err := k8sClient.Create(ctx, readinessProvider)
		Expect(err).To(BeNil())

		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: readinessProvider.Name}, readinessProvider)
			return err == nil &&
				readinessProvider.Status.State == corev1alpha2.ProviderFailureState &&
				len(readinessProvider.Status.Conditions) == 2 &&
				readinessProvider.Status.Conditions[0].State == corev1alpha2.ConditionInProgressState &&
				readinessProvider.Status.Conditions[1].State == corev1alpha2.ConditionFailureState
		}, timeout, interval).Should(BeTrue())
	})
})

func getTestReadinessProvider() *corev1alpha2.ReadinessProvider {
	return &corev1alpha2.ReadinessProvider{
		ObjectMeta: v1.ObjectMeta{
			GenerateName: "test-readiness-provider-",
		},
		Spec: corev1alpha2.ReadinessProviderSpec{
			Conditions: []corev1alpha2.ReadinessProviderCondition{},
			CheckRefs:  []string{},
		},
	}
}
