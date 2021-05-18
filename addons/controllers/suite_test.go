// Copyright 2020 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	addonconfig "github.com/vmware-tanzu-private/core/addons/pkg/config"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	clusterapiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	controlplanev1alpha3 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1alpha3"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"

	runtanzuv1alpha1 "github.com/vmware-tanzu-private/core/apis/run/v1alpha1"
	pkgiv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	runtanzuv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	cfg           *rest.Config
	k8sClient     client.Client
	testEnv       *envtest.Environment
	ctx           = ctrl.SetupSignalHandler()
	scheme        = runtime.NewScheme()
	dynamicClient dynamic.Interface
	cancel        context.CancelFunc
	//clientset     *kubernetes.Clientset
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
	externalCRDPaths, err := getExternalCRDPaths()
	Expect(err).NotTo(HaveOccurred())
	Expect(externalCRDPaths).ToNot(BeEmpty())
	testEnv.CRDDirectoryPaths = externalCRDPaths
	testEnv.CRDDirectoryPaths = append(testEnv.CRDDirectoryPaths, filepath.Join("..", "..", "config", "crd", "bases"))
	testEnv.ErrorIfCRDPathMissing = true

	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	err = runtanzuv1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = clientgoscheme.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = kappctrl.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = clusterapiv1alpha3.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = controlplanev1alpha3.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = pkgiv1alpha1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient).ToNot(BeNil())

	//clientset, err := kubernetes.NewForConfig(cfg)
	//Expect(err).ToNot(HaveOccurred())
	//Expect(clientset).ToNot(BeNil())

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

	Expect((&AddonReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Addon"),
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

	// pre-create namespace
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "tkr-system"}}
	Expect(k8sClient.Create(context.TODO(), ns)).To(Succeed())

	ctx, cancel = context.WithCancel(ctx)

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

// get paths for external CRDs by introspecting versions of the go dependencies
func getExternalCRDPaths() ([]string, error) {
	externalDeps := map[string][]string{
		"sigs.k8s.io/cluster-api": {"config/crd/bases",
			"controlplane/kubeadm/config/crd/bases"},
		"github.com/vmware-tanzu/carvel-kapp-controller": {"config/crds.yml"},
	}

	var crdPaths []string
	gopath, err := exec.Command("go", "env", "GOPATH").Output()
	if err != nil {
		return crdPaths, err
	}
	for dep, crdDirs := range externalDeps {
		depPath, err := exec.Command("go", "list", "-m", "-f", "{{ .Path }}@{{ .Version }}", dep).Output()
		if err != nil {
			return crdPaths, err
		}
		for _, crdDir := range crdDirs {
			crdPaths = append(crdPaths, filepath.Join(strings.TrimSuffix(string(gopath), "\n"),
				"pkg", "mod", strings.TrimSuffix(string(depPath), "\n"), crdDir))
		}
	}

	logf.Log.Info("external CRD paths", "crdPaths", crdPaths)
	return crdPaths, nil
}
