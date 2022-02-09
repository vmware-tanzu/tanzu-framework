// Copyright 2020 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
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
	waitTimeout             = time.Second * 90
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
		Config: addonconfig.Config{
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
		Log:    setupLog,
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: 1})).To(Succeed())

	Expect((&antrea.AntreaConfigReconciler{
		Client: mgr.GetClient(),
		Log:    setupLog,
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: 1})).To(Succeed())

	Expect((&kappcontroller.KappControllerConfigReconciler{
		Client: mgr.GetClient(),
		Log:    setupLog,
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: 1})).To(Succeed())

	bootstrapReconciler := NewClusterBootstrapReconciler(mgr.GetClient(),
		ctrl.Log.WithName("controllers").WithName("ClusterBootstrap"),
		mgr.GetScheme(),
	)
	Expect(bootstrapReconciler.SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: 1})).To(Succeed())

	// pre-create namespace
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "tkr-system"}}
	Expect(k8sClient.Create(context.TODO(), ns)).To(Succeed())

	ns = &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "tkg-system"}}
	Expect(k8sClient.Create(context.TODO(), ns)).To(Succeed())

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
