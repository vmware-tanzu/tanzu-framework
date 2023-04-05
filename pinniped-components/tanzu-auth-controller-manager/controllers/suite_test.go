// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/packages"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/kubernetes/scheme"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var k8sClient client.Client
var testEnv *envtest.Environment
var ctx = ctrl.SetupSignalHandler()
var cancel context.CancelFunc

const pinnipedNamespace = "pinniped-namespace"

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDInstallOptions: envtest.CRDInstallOptions{
			ErrorIfPathMissing: true,
			CleanUpAfterUse:    true,
		},
		ErrorIfCRDPathMissing: true,
	}
	var err error
	testEnv.CRDInstallOptions.Paths, err = getExternalCRDPaths()
	Expect(err).NotTo(HaveOccurred())

	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = corev1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = clusterapiv1beta1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	options := manager.Options{}
	mgr, err := ctrl.NewManager(testEnv.Config, options)
	Expect(err).NotTo(HaveOccurred())

	Expect(NewV1Controller(k8sClient).SetupWithManager(mgr)).To(Succeed())
	Expect(NewV3Controller(k8sClient).SetupWithManager(mgr)).To(Succeed())

	ctx, cancel = context.WithCancel(ctx)
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: pinnipedNamespace}}
	Expect(k8sClient.Create(ctx, ns)).To(Succeed())

	go func() {
		defer GinkgoRecover()
		Expect(mgr.Start(ctx)).To(Succeed())
	}()
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	// TODO: should we delete namespace here?
	cancel()
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

func getExternalCRDPaths() ([]string, error) {
	externalDeps := map[string][]string{
		"sigs.k8s.io/cluster-api/api/v1beta1": {"config/crd/bases",
			"controlplane/kubeadm/config/crd/bases"},
	}

	packageConfig := &packages.Config{
		Mode: packages.NeedModule,
	}

	var crdPaths []string
	for dep, crdDirs := range externalDeps {
		for _, crdDir := range crdDirs {
			pkgs, err := packages.Load(packageConfig, dep)
			if err != nil {
				return nil, err
			}
			pkg := pkgs[0]
			if pkg.Errors != nil {
				errs := []error{}
				for _, err := range pkg.Errors {
					errs = append(errs, err)
				}
				return nil, utilerrors.NewAggregate(errs)
			}
			crdPaths = append(crdPaths, filepath.Join(pkg.Module.Dir, crdDir))
		}
	}

	logf.Log.Info("external CRD paths", "crdPaths", crdPaths)
	return crdPaths, nil
}
