/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
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

	reconciler := NewController(k8sClient)
	Expect(reconciler.SetupWithManager(mgr)).To(Succeed())

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
		"sigs.k8s.io/cluster-api": {"config/crd/bases",
			"controlplane/kubeadm/config/crd/bases"},
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
