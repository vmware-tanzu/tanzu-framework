// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	k8sscheme "k8s.io/client-go/kubernetes/scheme"
	cliflag "k8s.io/component-base/cli/flag"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	// +kubebuilder:scaffold:imports

	configv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	corev1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/core/v1alpha2"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/buildinfo"
	coreFeatureController "github.com/vmware-tanzu/tanzu-framework/featuregates/controller/pkg/feature"
	configFeatureGateController "github.com/vmware-tanzu/tanzu-framework/featuregates/controller/pkg/featuregate"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(k8sscheme.AddToScheme(scheme))
	utilruntime.Must(configv1alpha1.AddToScheme(scheme))
	utilruntime.Must(corev1alpha2.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func setCipherSuiteFunc(cipherSuiteString string) (func(cfg *tls.Config), error) {
	cipherSuites := strings.Split(cipherSuiteString, ",")
	suites, err := cliflag.TLSCipherSuites(cipherSuites)
	if err != nil {
		return nil, err
	}
	return func(cfg *tls.Config) {
		cfg.CipherSuites = suites
	}, nil
}

func main() {
	var webhookServerPort int
	var tlsMinVersion string
	var tlsCipherSuites string
	flag.IntVar(&webhookServerPort, "webhook-server-port", 9443, "The port that the webhook server serves at.")
	flag.StringVar(&tlsMinVersion, "tls-min-version", "1.2", "minimum TLS version in use by the webhook server. Recommended values are \"1.2\" and \"1.3\".")
	flag.StringVar(&tlsCipherSuites, "tls-cipher-suites", "", "Comma-separated list of cipher suites for the server. If omitted, the default Go cipher suites will be used.\n"+fmt.Sprintf("Possible values are %s.", strings.Join(cliflag.TLSCipherPossibleValues(), ", ")))

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	setupLog.Info("Version", "version", buildinfo.Version, "buildDate", buildinfo.Date, "sha", buildinfo.SHA)

	var err error
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{Scheme: scheme, MetricsBindAddress: "0", Port: webhookServerPort})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	mgr.GetWebhookServer().TLSMinVersion = tlsMinVersion
	if tlsCipherSuites != "" {
		cipherSuitesSetFunc, err := setCipherSuiteFunc(tlsCipherSuites)
		if err != nil {
			setupLog.Error(err, "unable to set TLS Cipher suites")
			os.Exit(1)
		}
		mgr.GetWebhookServer().TLSOpts = append(mgr.GetWebhookServer().TLSOpts, cipherSuitesSetFunc)
	}
	if err = (&configFeatureGateController.FeatureGateReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("FeatureGate").WithValues("apigroup", "config"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "FeatureGate", "apigroup", "config")
		os.Exit(1)
	}

	if err = (&coreFeatureController.FeatureReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Feature").WithValues("apigroup", "core"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Feature", "apigroup", "core")
		os.Exit(1)
	}

	if err = (&configv1alpha1.FeatureGate{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "FeatureGate", "apigroup", "config")
		os.Exit(1)
	}

	if err = (&corev1alpha2.FeatureGate{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "FeatureGate", "apigroup", "core")
		os.Exit(1)
	}

	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
