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
	coreFeatureController "github.com/vmware-tanzu/tanzu-framework/featuregates/controller/pkg/feature"
	configFeatureGateController "github.com/vmware-tanzu/tanzu-framework/featuregates/controller/pkg/featuregate"
	"github.com/vmware-tanzu/tanzu-framework/util/buildinfo"
	"github.com/vmware-tanzu/tanzu-framework/util/webhook/certs"
)

var (
	scheme                              = runtime.NewScheme()
	setupLog                            = ctrl.Log.WithName("setup")
	defaultWebhookConfigLabel           = "tanzu.vmware.com/featuregates-webhook-managed-certs=true"
	defaultWebhookServiceNamespace      = "default"
	defaultWebhookServiceName           = "tanzu-featuregates-webhook-service"
	defaultWebhookSecretNamespace       = "default"
	defaultWebhookSecretName            = "tanzu-featuregates-webhook-server-cert"
	defaultWebhookSecretVolumeMountPath = "/tmp/k8s-webhook-server/serving-certs"
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
	var (
		webhookServerPort            int
		tlsMinVersion                string
		tlsCipherSuites              string
		webhookConfigLabel           string
		webhookServiceNamespace      string
		webhookServiceName           string
		webhookSecretNamespace       string
		webhookSecretName            string
		webhookSecretVolumeMountPath string
	)

	flag.IntVar(&webhookServerPort, "webhook-server-port", 9443, "The port that the webhook server serves at.")
	flag.StringVar(&tlsMinVersion, "tls-min-version", "1.2", "The minimum TLS version to be used by the webhook server. Recommended values are \"1.2\" and \"1.3\".")
	flag.StringVar(&tlsCipherSuites, "tls-cipher-suites", "", "Comma-separated list of cipher suites for the server. If omitted, the default Go cipher suites will be used.\n"+fmt.Sprintf("Possible values are %s.", strings.Join(cliflag.TLSCipherPossibleValues(), ", ")))
	flag.StringVar(&webhookConfigLabel, "webhook-config-label", defaultWebhookConfigLabel, "The label used to select webhook configurations to update the certs for.")
	flag.StringVar(&webhookServiceNamespace, "webhook-service-namespace", defaultWebhookServiceNamespace, "The namespace in which webhook service is installed.")
	flag.StringVar(&webhookServiceName, "webhook-service-name", defaultWebhookServiceName, "The name of the webhook service.")
	flag.StringVar(&webhookSecretNamespace, "webhook-secret-namespace", defaultWebhookSecretNamespace, "The namespace in which webhook secret is installed.")
	flag.StringVar(&webhookSecretName, "webhook-secret-name", defaultWebhookSecretName, "The name of the webhook secret.")
	flag.StringVar(&webhookSecretVolumeMountPath, "webhook-secret-volume-mount-path", defaultWebhookSecretVolumeMountPath, "The filesystem path to which the webhook secret is mounted.")

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	setupLog.Info("Version", "version", buildinfo.Version, "buildDate", buildinfo.Date, "sha", buildinfo.SHA)

	var err error
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{Scheme: scheme, MetricsBindAddress: "0", Port: webhookServerPort, CertDir: webhookSecretVolumeMountPath})
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

	signalHandler := ctrl.SetupSignalHandler()

	// Start certificate manager
	setupLog.Info("Starting certificate manager")
	certManagerOpts := certs.Options{
		Client:                        mgr.GetClient(),
		Logger:                        ctrl.Log.WithName("featuregates-webhook-cert-manager"),
		CertDir:                       webhookSecretVolumeMountPath,
		WebhookConfigLabel:            webhookConfigLabel,
		RotationIntervalAnnotationKey: "tanzu.vmware.com/featuregates-webhook-rotation-interval",
		NextRotationAnnotationKey:     "tanzu.vmware.com/featuregates-webhook-next-rotation",
		RotationCountAnnotationKey:    "tanzu.vmware.com/featuregates-webhook-rotation-count",
		SecretName:                    webhookSecretName,
		SecretNamespace:               webhookSecretNamespace,
		ServiceName:                   webhookServiceName,
		ServiceNamespace:              webhookServiceNamespace,
	}

	certManager, err := certs.New(certManagerOpts)
	if err != nil {
		setupLog.Error(err, "failed to create certificate manager")
		os.Exit(1)
	}

	// Start cert manager.
	if err := certManager.Start(signalHandler); err != nil {
		setupLog.Error(err, "failed to start certificate manager")
		os.Exit(1)
	}

	// Wait for cert dir to be ready.
	if err := certManager.WaitForCertDirReady(); err != nil {
		setupLog.Error(err, "certificates not ready")
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(signalHandler); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
