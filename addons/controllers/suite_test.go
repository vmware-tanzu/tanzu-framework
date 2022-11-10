// Copyright 2020-2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"bytes"
	"context"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	capvv1beta1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/v1beta1"
	capvvmwarev1beta1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/vmware/v1beta1"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	capiremote "sigs.k8s.io/cluster-api/controllers/remote"
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
	kapppkgv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	antrea "github.com/vmware-tanzu/tanzu-framework/addons/controllers/antrea"
	awsebscsi "github.com/vmware-tanzu/tanzu-framework/addons/controllers/awsebscsi"
	azurefilecsi "github.com/vmware-tanzu/tanzu-framework/addons/controllers/azurefilecsi"
	calico "github.com/vmware-tanzu/tanzu-framework/addons/controllers/calico"
	cpi "github.com/vmware-tanzu/tanzu-framework/addons/controllers/cpi"
	csi "github.com/vmware-tanzu/tanzu-framework/addons/controllers/csi"
	kappcontroller "github.com/vmware-tanzu/tanzu-framework/addons/controllers/kapp-controller"
	addonconfig "github.com/vmware-tanzu/tanzu-framework/addons/pkg/config"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/crdwait"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/webhooks"
	"github.com/vmware-tanzu/tanzu-framework/addons/test/testutil"
	addonwebhooks "github.com/vmware-tanzu/tanzu-framework/addons/webhooks"
	cniv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/addonconfigs/cni/v1alpha1"
	cpiv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/addonconfigs/cpi/v1alpha1"
	csiv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/addonconfigs/csi/v1alpha1"
	runtanzuv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
	runtanzuv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	vmoperatorv1alpha1 "github.com/vmware-tanzu/vm-operator-api/api/v1alpha1"
	topologyv1alpha1 "github.com/vmware-tanzu/vm-operator/external/tanzu-topology/api/v1alpha1"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

const (
	waitTimeout                         = time.Second * 90
	pollingInterval                     = time.Second * 2
	appSyncPeriod                       = 5 * time.Minute
	appWaitTimeout                      = 30 * time.Second
	addonNamespace                      = "tkg-system"
	addonServiceAccount                 = "tkg-addons-app-sa"
	addonClusterRole                    = "tkg-addons-app-cluster-role"
	addonClusterRoleBinding             = "tkg-addons-app-cluster-role-binding"
	addonImagePullPolicy                = "IfNotPresent"
	corePackageRepoName                 = "core"
	webhookServiceName                  = "tanzu-addons-manager-webhook-service"
	cniWebhookManifestFile              = "testdata/webhooks/test-antrea-calico-webhook-manifests.yaml"
	clusterbootstrapWebhookManifestFile = "testdata/webhooks/clusterbootstrap-webhook-manifests.yaml"
)

var (
	cfg                     *rest.Config
	k8sClient               client.Client
	k8sConfig               *rest.Config
	testEnv                 *envtest.Environment
	ctx                     = ctrl.SetupSignalHandler()
	scheme                  = runtime.NewScheme()
	mgr                     manager.Manager
	dynamicClient           dynamic.Interface
	clientSet               *kubernetes.Clientset
	cancel                  context.CancelFunc
	certPath                string
	keyPath                 string
	tmpDir                  string
	webhookCertDetails      testutil.WebhookCertificatesDetails
	webhookSelectorString   string
	addonConfigCRDBasesPath = filepath.Join("..", "..", "apis", "addonconfigs", "config", "crd", "bases")
	runCRDPath              = filepath.Join("..", "..", "apis", "run", "config", "crd", "bases")
	localCRDPath            = filepath.Join("testdata", "internal-crds")
	setupLog                = ctrl.Log.WithName("controllers").WithName("Addon")
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

	// For each dependency, a real module containing go files should be passed in.
	externalDeps := map[string][]string{
		"sigs.k8s.io/cluster-api/api/v1beta1": {"config/crd/bases",
			"controlplane/kubeadm/config/crd/bases"},
		"github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1": {"config/crds.yml"},
		"sigs.k8s.io/cluster-api-provider-vsphere/apis/v1beta1":                      {"config/default/crd/bases", "config/supervisor/crd"},
	}

	externalCRDPaths, err := testutil.GetExternalCRDPaths(externalDeps)
	Expect(err).NotTo(HaveOccurred())
	Expect(externalCRDPaths).ToNot(BeEmpty())
	testEnv.CRDDirectoryPaths = externalCRDPaths

	// copy crds into a local directory "testdata/internal-crds", while excluding "tanzukubernetesreleases" CRD file, which its v1alpha3 version is already included in "testdata/dependency/crd" directory
	err = setupLocalCRDDirectory(localCRDPath)
	Expect(err).NotTo(HaveOccurred())

	copyCRDs := func(files []os.DirEntry, path string) {
		for _, f := range files {
			if !strings.Contains(f.Name(), "tanzukubernetesreleases") {
				data, err := os.ReadFile(filepath.Join(path, f.Name()))
				Expect(err).ToNot(HaveOccurred())

				destFilePath := filepath.Join(localCRDPath, f.Name())
				destFile, err := os.OpenFile(destFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
				Expect(err).NotTo(HaveOccurred())

				_, err = io.Copy(destFile, bytes.NewReader(data))
				Expect(err).NotTo(HaveOccurred())
				destFile.Close()
			}
		}
	}

	files, err := os.ReadDir(addonConfigCRDBasesPath)
	Expect(err).ToNot(HaveOccurred())
	copyCRDs(files, addonConfigCRDBasesPath)

	files, err = os.ReadDir(runCRDPath)
	Expect(err).ToNot(HaveOccurred())
	copyCRDs(files, runCRDPath)

	// If it is not possible to include the parent repo that contains the CRD yaml file, manually add the CRD definition file into testdata/dependency/crd
	// For example, virtualmachines CRD is in repo vm-operator, but introducing vm-operator would cause dependency conflict in go.mod, therefore the CRD file is manually ported in
	testEnv.CRDDirectoryPaths = append(testEnv.CRDDirectoryPaths,
		localCRDPath, filepath.Join("testdata"), filepath.Join("testdata", "dependency", "crd"))
	testEnv.ErrorIfCRDPathMissing = true

	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())
	testEnv.ControlPlane.APIServer.Configure().Append("admission-control", "MutatingAdmissionWebhook")
	testEnv.ControlPlane.APIServer.Configure().Append("admission-control", "ValidatingAdmissionWebhook")

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

	err = kapppkgv1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = cniv1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = cpiv1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = csiv1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = capvv1beta1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = capvvmwarev1beta1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = vmoperatorv1alpha1.AddToScheme(scheme)
	Expect(err).ToNot(HaveOccurred())

	err = topologyv1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient).ToNot(BeNil())

	dynamicClient, err = dynamic.NewForConfig(cfg)
	Expect(err).ToNot(HaveOccurred())
	Expect(dynamicClient).ToNot(BeNil())

	tmpDir, err = os.MkdirTemp("/tmp", "webhooktest")
	Expect(err).ToNot(HaveOccurred())
	certPath = path.Join(tmpDir, "tls.crt")
	keyPath = path.Join(tmpDir, "tls.key")

	options := manager.Options{
		Scheme:             scheme,
		MetricsBindAddress: "0",
		Host:               "127.0.0.1",
		Port:               9443,
		CertDir:            tmpDir,
	}
	mgr, err = ctrl.NewManager(testEnv.Config, options)
	Expect(err).ToNot(HaveOccurred())
	k8sConfig = mgr.GetConfig()

	clientSet, err = kubernetes.NewForConfig(cfg)
	Expect(err).ToNot(HaveOccurred())

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
		Config: addonconfig.CalicoConfigControllerConfig{
			ConfigControllerConfig: addonconfig.ConfigControllerConfig{SystemNamespace: constants.TKGSystemNS}},
	}).SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: 1})).To(Succeed())

	Expect((&cpi.VSphereCPIConfigReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("VSphereCPIConfig"),
		Scheme: mgr.GetScheme(),
		Config: addonconfig.VSphereCPIConfigControllerConfig{
			ConfigControllerConfig: addonconfig.ConfigControllerConfig{SystemNamespace: constants.TKGSystemNS}},
	}).SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: 1})).To(Succeed())

	Expect((&csi.VSphereCSIConfigReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("VSphereCSIConfig"),
		Scheme: mgr.GetScheme(),
		Config: addonconfig.VSphereCSIConfigControllerConfig{
			ConfigControllerConfig: addonconfig.ConfigControllerConfig{SystemNamespace: constants.TKGSystemNS}},
	}).SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: 1})).To(Succeed())

	Expect((&azurefilecsi.AzureFileCSIConfigReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("AzureFileCSIConfig"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: 1})).To(Succeed())

	Expect((&antrea.AntreaConfigReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("AntreaConfig"),
		Scheme: mgr.GetScheme(),
		Config: addonconfig.AntreaConfigControllerConfig{
			ConfigControllerConfig: addonconfig.ConfigControllerConfig{SystemNamespace: constants.TKGSystemNS}},
	}).SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: 1})).To(Succeed())

	Expect((&kappcontroller.KappControllerConfigReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("KappController"),
		Scheme: mgr.GetScheme(),
		Config: addonconfig.KappControllerConfigControllerConfig{
			ConfigControllerConfig: addonconfig.ConfigControllerConfig{SystemNamespace: constants.TKGSystemNS}},
	}).SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: 1})).To(Succeed())

	Expect((&MachineReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("MachineController"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: 1})).To(Succeed())

	Expect((&awsebscsi.AwsEbsCSIConfigReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("AwsEbsCSIConfig"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: 1})).To(Succeed())

	bootstrapReconciler := NewClusterBootstrapReconciler(
		mgr.GetClient(),
		ctrl.Log.WithName("controllers").WithName("ClusterBootstrap"),
		mgr.GetScheme(),
		&addonconfig.ClusterBootstrapControllerConfig{
			IPFamilyClusterClassVarName: constants.DefaultIPFamilyClusterClassVarName,
			SystemNamespace:             constants.TKGSystemNS,
			PkgiServiceAccount:          constants.PackageInstallServiceAccount,
			PkgiClusterRole:             constants.PackageInstallClusterRole,
			PkgiClusterRoleBinding:      constants.PackageInstallClusterRoleBinding,
			PkgiSyncPeriod:              constants.PackageInstallSyncPeriod,
			ClusterDeleteTimeout:        time.Second * 10,
		},
	)
	Expect(bootstrapReconciler.SetupWithManager(context.Background(), mgr, controller.Options{MaxConcurrentReconciles: 1})).To(Succeed())

	// set up a ClusterCacheTracker to provide to PackageInstallStatus controller which requires a connection to remote clusters
	l := ctrl.Log.WithName("remote").WithName("ClusterCacheTracker")
	tracker, err := capiremote.NewClusterCacheTracker(mgr, capiremote.ClusterCacheTrackerOptions{Log: &l})
	Expect(err).Should(BeNil())
	Expect(tracker).ShouldNot(BeNil())

	// set up CluterCacheReconciler to drops the accessor via deleteAccessor upon cluster deletion
	Expect((&capiremote.ClusterCacheReconciler{
		Client:  mgr.GetClient(),
		Log:     ctrl.Log.WithName("remote").WithName("ClusterCacheReconciler"),
		Tracker: tracker,
	}).SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: 1})).To(Succeed())

	Expect((NewPackageInstallStatusReconciler(
		mgr.GetClient(),
		ctrl.Log.WithName("controllers").WithName("PackageInstallStatus"),
		mgr.GetScheme(),
		&addonconfig.PackageInstallStatusControllerConfig{
			SystemNamespace: constants.TKGSystemNS,
		},
		tracker,
	)).SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: 1})).To(Succeed())

	// pre-create namespace
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "tkr-system"}}
	Expect(k8sClient.Create(context.TODO(), ns)).To(Succeed())

	ns = &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "tkg-system"}}
	Expect(k8sClient.Create(context.TODO(), ns)).To(Succeed())

	labelMatch, err := labels.NewRequirement(constants.AddonWebhookLabelKey, selection.Equals, []string{constants.AddonWebhookLabelValue})
	Expect(err).ToNot(HaveOccurred())
	webhookSelector := labels.NewSelector()
	webhookSelector = webhookSelector.Add(*labelMatch)
	webhookSelectorString = webhookSelector.String()
	_, err = webhooks.InstallNewCertificates(ctx, k8sConfig, certPath, keyPath, constants.WebhookScrtName, addonNamespace, webhookServiceName, webhookSelectorString)
	Expect(err).ToNot(HaveOccurred())

	// Set up the webhooks in the manager
	err = (&cniv1alpha1.AntreaConfig{}).SetupWebhookWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())
	err = (&cniv1alpha1.CalicoConfig{}).SetupWebhookWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())
	err = (&addonwebhooks.ClusterPause{Client: k8sClient}).SetupWebhookWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())
	clusterbootstrapWebhook := addonwebhooks.ClusterBootstrap{
		Client:          k8sClient,
		SystemNamespace: addonNamespace,
	}
	err = clusterbootstrapWebhook.SetupWebhookWithManager(ctx, mgr)
	Expect(err).ToNot(HaveOccurred())
	clusterbootstrapTemplateWebhook := addonwebhooks.ClusterBootstrapTemplate{
		SystemNamespace: addonNamespace,
	}
	err = clusterbootstrapTemplateWebhook.SetupWebhookWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		Expect(mgr.Start(ctx)).To(Succeed())
	}()

	// Setup the tkg-system namespace resources.
	// We do it here because  specs may be executed in parallel and
	// this is the only way to assure the tkg-system resources are ready before all specs start

	// Prepare clusterbootstrap webhooks webhooks
	f, err := os.Open(clusterbootstrapWebhookManifestFile)
	Expect(err).ToNot(HaveOccurred())
	err = testutil.CreateResources(f, cfg, dynamicClient)
	Expect(err).ToNot(HaveOccurred())
	f.Close()

	// set up the certificates and webhook before creating any objects
	By("Creating and installing new certificates for ClusterBootstrap Admission Webhooks")
	webhookCertDetails = testutil.WebhookCertificatesDetails{
		CertPath:           certPath,
		KeyPath:            keyPath,
		WebhookScrtName:    constants.WebhookScrtName,
		AddonNamespace:     addonNamespace,
		WebhookServiceName: webhookServiceName,
		LabelSelector:      webhookSelector,
	}
	err = testutil.SetupWebhookCertificates(ctx, k8sClient, k8sConfig, &webhookCertDetails)
	Expect(err).ToNot(HaveOccurred())

	// Create the rest of tkg-system resources
	f, err = os.Open("testdata/test-tkg-system-ns-resources.yaml")
	Expect(err).ToNot(HaveOccurred())
	defer f.Close()
	Expect(testutil.CreateResources(f, cfg, dynamicClient)).To(Succeed())

	close(done)
}, 60)

var _ = AfterSuite(func() {
	f, err := os.Open(clusterbootstrapWebhookManifestFile)
	Expect(err).ToNot(HaveOccurred())
	err = testutil.DeleteResources(f, cfg, dynamicClient, true)
	Expect(err).ToNot(HaveOccurred())
	f.Close()
	By("tearing down the test environment")
	cancel()
	err = testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())

	err = removeLocalCRDDirectory(localCRDPath)
	Expect(err).ToNot(HaveOccurred())
})

func setupLocalCRDDirectory(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(path, os.ModePerm); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func removeLocalCRDDirectory(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	return nil
}
