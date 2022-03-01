// Copyright 2020 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/cert"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1beta1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	pkgiv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	antrea "github.com/vmware-tanzu/tanzu-framework/addons/controllers/antrea"
	calico "github.com/vmware-tanzu/tanzu-framework/addons/controllers/calico"
	kappcontroller "github.com/vmware-tanzu/tanzu-framework/addons/controllers/kapp-controller"
	addonconfig "github.com/vmware-tanzu/tanzu-framework/addons/pkg/config"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/crdwait"
	testutil "github.com/vmware-tanzu/tanzu-framework/addons/testutil"
	cniv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cni/v1alpha1"
	runtanzuv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
	runtanzuv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

const (
	waitTimeout             = time.Second * 60
	pollingInterval         = time.Second * 2
	appSyncPeriod           = 5 * time.Minute
	appWaitTimeout          = 30 * time.Second
	addonNamespace          = "tkg-system"
	addonServiceAccount     = "tkg-addons-app-sa"
	addonClusterRole        = "tkg-addons-app-cluster-role"
	addonClusterRoleBinding = "tkg-addons-app-cluster-role-binding"
	addonImagePullPolicy    = "IfNotPresent"
	corePackageRepoName     = "core"
)

var (
	cfg           *rest.Config
	k8sClient     client.Client
	testEnv       *envtest.Environment
	ctx           = ctrl.SetupSignalHandler()
	scheme        = runtime.NewScheme()
	dynamicClient dynamic.Interface
	cancel        context.CancelFunc
	certDir       string
	caPEM         *bytes.Buffer
	// clientset     *kubernetes.Clientset
)

func TestAddonController(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Addon Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {

	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{CRDInstallOptions: envtest.CRDInstallOptions{
		CleanUpAfterUse: true},
		ErrorIfCRDPathMissing: true,
	}

	externalDeps := map[string][]string{
		"sigs.k8s.io/cluster-api": {"config/crd/bases",
			"controlplane/kubeadm/config/crd/bases"},
		"github.com/vmware-tanzu/carvel-kapp-controller": {"config/crds.yml"},
	}
	externalCRDPaths, err := testutil.GetExternalCRDPaths(externalDeps)
	Expect(err).NotTo(HaveOccurred())
	Expect(externalCRDPaths).ToNot(BeEmpty())
	testEnv.CRDDirectoryPaths = externalCRDPaths
	testEnv.CRDDirectoryPaths = append(testEnv.CRDDirectoryPaths,
		filepath.Join("..", "..", "config", "crd", "bases"), filepath.Join("testdata"))
	testEnv.ErrorIfCRDPathMissing = true

	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())
	testEnv.ControlPlane.APIServer.Configure().Append("admission-control", "MutatingAdmissionWebhook")

	err = runtanzuv1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = runtanzuv1alpha3.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = clientgoscheme.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = kappctrl.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = clusterapiv1beta1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = controlplanev1beta1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = pkgiv1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = cniv1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient).ToNot(BeNil())

	err = runtanzuv1alpha3.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	dynamicClient, err = dynamic.NewForConfig(cfg)
	Expect(err).ToNot(HaveOccurred())
	Expect(dynamicClient).ToNot(BeNil())

	options := manager.Options{
		Scheme:             scheme,
		MetricsBindAddress: "0",
		Port:               9443,
	}
	mgr, err := ctrl.NewManager(testEnv.Config, options)
	Expect(err).ToNot(HaveOccurred())

	setupLog := ctrl.Log.WithName("controllers").WithName("Addon")

	ctx, cancel = context.WithCancel(ctx)
	crdwaiter := crdwait.CRDWaiter{
		Ctx: ctx,
		ClientSetFn: func() (kubernetes.Interface, error) {
			return kubernetes.NewForConfig(cfg)
		},
		Logger:       setupLog,
		Scheme:       scheme,
		PollInterval: constants.CRDWaitPollInterval,
		PollTimeout:  constants.CRDWaitPollTimeout,
	}

	if err := crdwaiter.WaitForCRDs(GetExternalCRDs(),
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "test-pod", Namespace: "default"}},
		constants.AddonControllerName,
	); err != nil {
		setupLog.Error(err, "unable to wait for CRDs")
		os.Exit(1)
	}

	Expect((&AddonReconciler{
		Client: mgr.GetClient(),
		Log:    setupLog,
		Scheme: mgr.GetScheme(),
		Config: addonconfig.AddonControllerConfig{
			AppSyncPeriod:           appSyncPeriod,
			AppWaitTimeout:          appWaitTimeout,
			AddonNamespace:          addonNamespace,
			AddonServiceAccount:     addonServiceAccount,
			AddonClusterRole:        addonClusterRole,
			AddonClusterRoleBinding: addonClusterRoleBinding,
			AddonImagePullPolicy:    addonImagePullPolicy,
			CorePackageRepoName:     corePackageRepoName,
		},
	}).SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: 1})).To(Succeed())

	Expect((&calico.CalicoConfigReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("CalicoConfig"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: 1})).To(Succeed())

	Expect((&antrea.AntreaConfigReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("AntreaConfig"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: 1})).To(Succeed())

	Expect((&kappcontroller.KappControllerConfigReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("KappController"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: 1})).To(Succeed())

	bootstrapReconciler := NewClusterBootstrapReconciler(mgr.GetClient(),
		ctrl.Log.WithName("controllers").WithName("ClusterBootstrap"),
		mgr.GetScheme(),
		&addonconfig.ClusterBootstrapControllerConfig{
			CNISelectionClusterVariableName: constants.DefaultCNISelectionClusterVariableName,
			HTTPProxyClusterClassVarName:    constants.DefaultHTTPProxyClusterClassVarName,
			HTTPSProxyClusterClassVarName:   constants.DefaultHTTPSProxyClusterClassVarName,
			NoProxyClusterClassVarName:      constants.DefaultNoProxyClusterClassVarName,
			ProxyCACertClusterClassVarName:  constants.DefaultProxyCaCertClusterClassVarName,
			IPFamilyClusterClassVarName:     constants.DefaultIPFamilyClusterClassVarName,
		},
	)
	Expect(bootstrapReconciler.SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: 1})).To(Succeed())

	// pre-create namespace
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "tkr-system"}}
	Expect(k8sClient.Create(context.TODO(), ns)).To(Succeed())

	ns = &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "tkg-system"}}
	Expect(k8sClient.Create(context.TODO(), ns)).To(Succeed())

	// setupForWebhooks sets up the certificates
	setupWebhooks(ctx, mgr, setupLog)

	go func() {
		defer GinkgoRecover()
		Expect(mgr.Start(ctx)).To(Succeed())
	}()

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	cancel()
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})

// setupWebhooks configures certs/keys and sets up the webhooks with mgr
func setupWebhooks(ctx context.Context, mgr manager.Manager, setupLog logr.Logger) {
	var (
		f   *os.File
		err error
	)
	// Setup TLS certs and keys
	setupCertsAndKeysForWebhooks()

	// Apply the webhooks configuration
	webhooksFile := "webhooks/manifests.yaml"
	f, err = os.Open(webhooksFile)
	Expect(err).ToNot(HaveOccurred())
	defer f.Close()
	err = testutil.CreateResources(f, cfg, dynamicClient)
	Expect(err).ToNot(HaveOccurred())

	// Update the caBundle in validating and mutating webhooks
	k8sConfig := mgr.GetConfig()
	client, err := kubernetes.NewForConfig(k8sConfig)
	validatingWebhookConf, err := client.AdmissionregistrationV1().ValidatingWebhookConfigurations().
		Get(ctx, "validating-webhook-configuration", metav1.GetOptions{})
	Expect(err).ToNot(HaveOccurred())
	for idx := range validatingWebhookConf.Webhooks {
		validatingWebhookConf.Webhooks[idx].ClientConfig.CABundle = caPEM.Bytes()
	}
	_, err = client.AdmissionregistrationV1().ValidatingWebhookConfigurations().
		Update(ctx, validatingWebhookConf, metav1.UpdateOptions{})
	Expect(err).ToNot(HaveOccurred())

	mutatingWebhookConf, err := client.AdmissionregistrationV1().MutatingWebhookConfigurations().
		Get(ctx, "mutating-webhook-configuration", metav1.GetOptions{})
	Expect(err).ToNot(HaveOccurred())
	for idx := range mutatingWebhookConf.Webhooks {
		mutatingWebhookConf.Webhooks[idx].ClientConfig.CABundle = caPEM.Bytes()
	}
	_, err = client.AdmissionregistrationV1().MutatingWebhookConfigurations().
		Update(ctx, mutatingWebhookConf, metav1.UpdateOptions{})
	Expect(err).ToNot(HaveOccurred())

	// Configure and start the webhook server
	wh := mgr.GetWebhookServer()
	wh.CertDir = "/tmp/k8s-webhook-server/serving-certs"
	wh.Host = "127.0.0.1"
	wh.Port = 9443

	// Setup the webhooks in the manager
	if err = (&cniv1alpha1.AntreaConfig{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook for Antrea")
		os.Exit(1)
	}
	if err = (&cniv1alpha1.CalicoConfig{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook for Calico")
		os.Exit(1)
	}
}

// setupCertsAndKeysForWebhooks sets up the certificate for CA and the cert and keys for webhook server
func setupCertsAndKeysForWebhooks() {

	// Use the default directory and file names used by webhook server for TLS certs and keys
	certDir = "/tmp/k8s-webhook-server/serving-certs"
	webhookCertFile := filepath.Join(certDir, "tls.crt")
	webhookKeyFile := filepath.Join(certDir, "tls.key")

	Expect(os.MkdirAll(certDir, 0666)).To(Succeed())

	// CA private key
	caPrivateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	Expect(err).ToNot(HaveOccurred())

	caCert, err := cert.NewSelfSignedCACert(
		cert.Config{CommonName: "test-self-signed-certificate", Organization: []string{"TestOrg"}},
		caPrivateKey)
	Expect(err).ToNot(HaveOccurred())

	// encode the CA certificate
	caPEM = new(bytes.Buffer)
	_ = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caCert.Raw,
	})

	// generate the key for webhook server
	webhookKey, err := rsa.GenerateKey(rand.Reader, 4096)
	Expect(err).ToNot(HaveOccurred())

	// webhook server cert config
	webhookCert := &x509.Certificate{
		SerialNumber: big.NewInt(2022),
		Subject: pkix.Name{
			CommonName:   "localhost",
			Organization: []string{"TestOrg"},
		},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1)},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		SubjectKeyId: []byte{2, 0, 2, 2},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	// sign the webhook server certificate with the self-signed CA
	webhookCertBytes, err := x509.CreateCertificate(rand.Reader, webhookCert, caCert, &webhookKey.PublicKey, caPrivateKey)
	Expect(err).ToNot(HaveOccurred())

	// PEM encode the certificate and key for webhook server
	webhookCertPEM := new(bytes.Buffer)
	_ = pem.Encode(webhookCertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: webhookCertBytes,
	})
	webhookKeyPEM := new(bytes.Buffer)
	_ = pem.Encode(webhookKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(webhookKey),
	})
	// Write the cert and key for webhook server to the files
	Expect(ioutil.WriteFile(webhookCertFile, webhookCertPEM.Bytes(), 0600)).To(Succeed())
	Expect(ioutil.WriteFile(webhookKeyFile, webhookKeyPEM.Bytes(), 0644)).To(Succeed())

}
