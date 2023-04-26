//go:build envtest

// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package feature

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"math/big"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	corev1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/core/v1alpha2"
	testutil "github.com/vmware-tanzu/tanzu-framework/featuregates/controller/pkg/test"
)

func TestFeaturegate(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Featuregate Suite")
}

var (
	cfg       *rest.Config
	k8sClient client.Client
	testEnv   *envtest.Environment
	ctx       context.Context
	cancel    context.CancelFunc
	setupLog  = ctrl.Log.WithName("controllers").WithName("Features")

	timeout  = 5 * time.Second
	interval = 100 * time.Millisecond

	tmpDir                        string
	generatedWebhookManifestBytes []byte
)

func generateCertificateAndManifests() error {
	key, cert, err := testutil.GenerateSelfSignedCertsForTest("tanzu-featuregates-webhook-service.tkg-system.svc")
	if err != nil {
		return err
	}

	tmpDir, err = os.MkdirTemp("/tmp", "featuregate-test")
	if err != nil {
		return err
	}
	err = os.WriteFile(path.Join(tmpDir, "feature-webhook.crt"), cert, 0644)
	if err != nil {
		return err
	}

	err = os.WriteFile(path.Join(tmpDir, "feature-webhook.key"), key, 0644)
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

	ctx, cancel = context.WithCancel(context.TODO())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "..", "..", "apis", "core", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	testEnv.ControlPlane.APIServer.Configure().Append("admission-control", "ValidatingAdmissionWebhook")

	err = corev1alpha2.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

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
	k8sManager.GetWebhookServer().CertName = "feature-webhook.crt"
	k8sManager.GetWebhookServer().KeyName = "feature-webhook.key"
	k8sManager.GetWebhookServer().Port = 9443

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	dynamicClient, err := dynamic.NewForConfig(cfg)
	Expect(err).ToNot(HaveOccurred())
	Expect(dynamicClient).ToNot(BeNil())

	err = (&FeatureReconciler{
		Client: k8sManager.GetClient(),
		Scheme: k8sManager.GetScheme(),
		Log:    setupLog,
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "tkg-system"}}
	Expect(k8sClient.Create(ctx, ns)).To(Succeed())

	err = (&corev1alpha2.FeatureGate{}).SetupWebhookWithManager(k8sManager)
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

	err = os.RemoveAll(tmpDir)
	Expect(err).NotTo(HaveOccurred())
})

func getTestFeatureGate() *corev1alpha2.FeatureGate {
	randomNumber, err := rand.Int(rand.Reader, big.NewInt(100000))
	if err != nil {
		return nil
	}

	return &corev1alpha2.FeatureGate{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("featuregate-%v", randomNumber),
		},
		Spec: corev1alpha2.FeatureGateSpec{
			Features: []corev1alpha2.FeatureReference{},
		},
		Status: corev1alpha2.FeatureGateStatus{},
	}
}

func getTestFeature(stability corev1alpha2.StabilityLevel) *corev1alpha2.Feature {
	randomNumber, err := rand.Int(rand.Reader, big.NewInt(100000))
	if err != nil {
		return nil
	}

	return &corev1alpha2.Feature{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("feature-%v", randomNumber),
		},
		Spec: corev1alpha2.FeatureSpec{
			Stability: stability,
		},
	}
}

var _ = Describe("Featuregate controller", func() {
	It("Should not activate experimental features by default", func() {
		feature := getTestFeature(corev1alpha2.Experimental)
		Expect(k8sClient.Create(ctx, feature)).Should(Succeed())

		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: feature.Name}, feature)
			return err == nil
		}, timeout, interval).Should(BeTrue())

		featureGate := getTestFeatureGate()
		featureGate.Spec.Features = append(featureGate.Spec.Features, corev1alpha2.FeatureReference{
			Name:     feature.Name,
			Activate: false,
		})
		Expect(k8sClient.Create(ctx, featureGate)).Should(Succeed())

		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: featureGate.Name}, featureGate)
			return err == nil
		}, timeout, interval).Should(BeTrue())

		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: featureGate.Name}, featureGate)
			return err == nil && len(featureGate.Status.FeatureReferenceResults) == 1
		}, timeout, interval).Should(BeTrue())

		Expect(featureGate.Status.FeatureReferenceResults[0].Status).Should(Equal(corev1alpha2.FeatureReferenceStatus("Applied")))
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: feature.Name}, feature)).Should(BeNil())
		Expect(feature.Status.Activated).Should(Equal(false))

		featureGate.Spec.Features[0].Activate = true
		Expect(k8sClient.Update(ctx, featureGate)).ShouldNot(Succeed())

		Expect(k8sClient.Delete(ctx, feature)).Should(BeNil())
		Expect(k8sClient.Delete(ctx, featureGate)).Should(BeNil())
	})

	It("Should activate and de-activate experimental features when permanently voided all support guarantees", func() {
		feature := getTestFeature(corev1alpha2.Experimental)
		Expect(k8sClient.Create(ctx, feature)).Should(Succeed())

		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: feature.Name}, feature)
			return err == nil
		}, timeout, interval).Should(BeTrue())

		featureGate := getTestFeatureGate()
		featureGate.Spec.Features = append(featureGate.Spec.Features, corev1alpha2.FeatureReference{
			Name:                                feature.Name,
			Activate:                            false,
			PermanentlyVoidAllSupportGuarantees: true,
		})
		Expect(k8sClient.Create(ctx, featureGate)).Should(Succeed())

		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: featureGate.Name}, featureGate)
			return err == nil
		}, timeout, interval).Should(BeTrue())

		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: featureGate.Name}, featureGate)
			return err == nil && len(featureGate.Status.FeatureReferenceResults) == 1
		}, timeout, interval).Should(BeTrue())

		Expect(featureGate.Status.FeatureReferenceResults[0].Status).Should(Equal(corev1alpha2.FeatureReferenceStatus("Applied")))
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: feature.Name}, feature)).Should(BeNil())
		Expect(feature.Status.Activated).Should(Equal(false))

		featureGate.Spec.Features[0].Activate = true
		Expect(k8sClient.Update(ctx, featureGate)).Should(Succeed())

		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: feature.Name}, feature)
			return err == nil && feature.Status.Activated == true
		}, timeout, interval).Should(BeTrue())

		featureGate.Spec.Features[0].Activate = false
		Expect(k8sClient.Update(ctx, featureGate)).Should(Succeed())

		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: feature.Name}, feature)
			return err == nil && feature.Status.Activated == false
		}, timeout, interval).Should(BeTrue())

		// PermanentlyVoidAllSupportGuarantees once set to true, should not be set to false
		featureGate.Spec.Features[0].PermanentlyVoidAllSupportGuarantees = false
		Expect(k8sClient.Update(ctx, featureGate)).ShouldNot(Succeed())

		Expect(k8sClient.Delete(ctx, feature)).Should(BeNil())
		Expect(k8sClient.Delete(ctx, featureGate)).Should(BeNil())
	})

	It("Should activate preview features irrespective of permanently voiding all support guarantees", func() {
		feature := getTestFeature(corev1alpha2.TechnicalPreview)
		Expect(k8sClient.Create(ctx, feature)).Should(Succeed())

		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: feature.Name}, feature)
			return err == nil
		}, timeout, interval).Should(BeTrue())

		featureGate := getTestFeatureGate()
		featureGate.Spec.Features = append(featureGate.Spec.Features, corev1alpha2.FeatureReference{
			Name:     feature.Name,
			Activate: false,
		})
		Expect(k8sClient.Create(ctx, featureGate)).Should(Succeed())

		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: featureGate.Name}, featureGate)
			return err == nil
		}, timeout, interval).Should(BeTrue())

		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: featureGate.Name}, featureGate)
			return err == nil && len(featureGate.Status.FeatureReferenceResults) == 1
		}, timeout, interval).Should(BeTrue())

		Expect(featureGate.Status.FeatureReferenceResults[0].Status).Should(Equal(corev1alpha2.FeatureReferenceStatus("Applied")))
		Expect(k8sClient.Get(ctx, types.NamespacedName{Name: feature.Name}, feature)).Should(BeNil())
		Expect(feature.Status.Activated).Should(Equal(false))

		featureGate.Spec.Features[0].Activate = true
		Expect(k8sClient.Update(ctx, featureGate)).Should(Succeed())

		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: feature.Name}, feature)
			return err == nil && feature.Status.Activated == true
		}, timeout, interval).Should(BeTrue())

		featureGate.Spec.Features[0].Activate = false
		Expect(k8sClient.Update(ctx, featureGate)).Should(Succeed())

		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: feature.Name}, feature)
			return err == nil && feature.Status.Activated == false
		}, timeout, interval).Should(BeTrue())

		Expect(k8sClient.Delete(ctx, feature)).Should(BeNil())
		Expect(k8sClient.Delete(ctx, featureGate)).Should(BeNil())
	})

	It("Should activate stable features by default and stable features should be immutable", func() {
		feature := getTestFeature(corev1alpha2.Stable)
		Expect(k8sClient.Create(ctx, feature)).Should(Succeed())

		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: feature.Name}, feature)
			return err == nil
		}, timeout, interval).Should(BeTrue())

		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: feature.Name}, feature)
			return err == nil && feature.Status.Activated == true
		}, timeout, interval).Should(BeTrue())

		featureGate := getTestFeatureGate()
		featureGate.Spec.Features = append(featureGate.Spec.Features, corev1alpha2.FeatureReference{
			Name:     feature.Name,
			Activate: false,
		})
		Expect(k8sClient.Create(ctx, featureGate)).ShouldNot(Succeed())

		Expect(k8sClient.Delete(ctx, feature)).Should(BeNil())
	})

	It("Should activate deprecated features by default and deprected features should be mutable", func() {
		feature := getTestFeature(corev1alpha2.Deprecated)
		Expect(k8sClient.Create(ctx, feature)).Should(Succeed())

		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: feature.Name}, feature)
			return err == nil
		}, timeout, interval).Should(BeTrue())

		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: feature.Name}, feature)
			return err == nil && feature.Status.Activated == true
		}, timeout, interval).Should(BeTrue())

		featureGate := getTestFeatureGate()
		featureGate.Spec.Features = append(featureGate.Spec.Features, corev1alpha2.FeatureReference{
			Name:     feature.Name,
			Activate: false,
		})
		Expect(k8sClient.Create(ctx, featureGate)).Should(Succeed())

		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: featureGate.Name}, featureGate)
			return err == nil
		}, timeout, interval).Should(BeTrue())

		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: featureGate.Name}, featureGate)
			return err == nil && len(featureGate.Status.FeatureReferenceResults) == 1
		}, timeout, interval).Should(BeTrue())

		Expect(featureGate.Status.FeatureReferenceResults[0].Status).Should(Equal(corev1alpha2.FeatureReferenceStatus("Applied")))

		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: feature.Name}, feature)
			return err == nil && feature.Status.Activated == false
		}, timeout, interval).Should(BeTrue())

		Expect(k8sClient.Delete(ctx, feature)).Should(BeNil())
		Expect(k8sClient.Delete(ctx, featureGate)).Should(BeNil())
	})

	It("Should remove feature from featuregate status when reference is removed", func() {
		feature := getTestFeature(corev1alpha2.Deprecated)
		Expect(k8sClient.Create(ctx, feature)).Should(Succeed())

		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: feature.Name}, feature)
			return err == nil && feature.Status.Activated == true
		}, timeout, interval).Should(BeTrue())

		featureGate := getTestFeatureGate()
		featureGate.Spec.Features = append(featureGate.Spec.Features, corev1alpha2.FeatureReference{
			Name:     feature.Name,
			Activate: false,
		})
		Expect(k8sClient.Create(ctx, featureGate)).Should(Succeed())

		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: featureGate.Name}, featureGate)
			return err == nil
		}, timeout, interval).Should(BeTrue())

		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: featureGate.Name}, featureGate)
			return err == nil && len(featureGate.Status.FeatureReferenceResults) == 1
		}, timeout, interval).Should(BeTrue())

		featureGate.Spec.Features = []corev1alpha2.FeatureReference{}
		Expect(k8sClient.Update(ctx, featureGate)).Should(Succeed())

		Eventually(func() bool {
			err := k8sClient.Get(ctx, types.NamespacedName{Name: featureGate.Name}, featureGate)
			return err == nil && len(featureGate.Status.FeatureReferenceResults) == 0
		}, timeout, interval).Should(BeTrue())

		Expect(k8sClient.Delete(ctx, feature)).Should(BeNil())
		Expect(k8sClient.Delete(ctx, featureGate)).Should(BeNil())
	})
})
