// Copyright 2023 VMware, Inc. All Rights Reserved.
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
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	cliflag "k8s.io/component-base/cli/flag"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	corev1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/core/v1alpha2"
	capabilitiesDiscovery "github.com/vmware-tanzu/tanzu-framework/capabilities/client/pkg/discovery"
	"github.com/vmware-tanzu/tanzu-framework/readiness/controller/pkg/conditions"
	readinesscontroller "github.com/vmware-tanzu/tanzu-framework/readiness/controller/pkg/readiness"
	readinessprovidercontroller "github.com/vmware-tanzu/tanzu-framework/readiness/controller/pkg/readinessprovider"
	"github.com/vmware-tanzu/tanzu-framework/util/webhook/certs"
	//+kubebuilder:scaffold:imports
)

var (
	scheme                              = runtime.NewScheme()
	setupLog                            = ctrl.Log.WithName("setup")
	defaultWebhookConfigLabel           = "tanzu.vmware.com/readinessprovider-webhook-managed-certs=true"
	defaultWebhookServiceNamespace      = "default"
	defaultWebhookServiceName           = "tanzu-readinessprovider-webhook-service"
	defaultWebhookSecretNamespace       = "default"
	defaultWebhookSecretName            = "tanzu-readinessprovider-webhook-server-cert" //nolint:gosec
	defaultWebhookSecretVolumeMountPath = "/tmp/k8s-webhook-server/serving-certs"       //nolint:gosec
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

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

//nolint:funlen
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

	var metricsAddr string
	var probeAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	restConfig := ctrl.GetConfigOrDie()

	mgr, err := ctrl.NewManager(restConfig, ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   webhookServerPort,
		HealthProbeBindAddress: probeAddr,
		CertDir:                webhookSecretVolumeMountPath,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	k8sClientset := kubernetes.NewForConfigOrDie(restConfig)

	clusterQueryClient, err := capabilitiesDiscovery.NewClusterQueryClientForConfig(restConfig)
	if err != nil {
		setupLog.Error(err, "unable to create cluster query client")
		os.Exit(1)
	}

	cl, err := client.New(restConfig, client.Options{})
	if err != nil {
		setupLog.Error(err, "unable to create client")
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

	if err = (&readinesscontroller.ReadinessReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Log:    ctrl.Log.WithName("controllers").WithName("Readiness").WithValues("apigroup", "core"),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Readiness")
		os.Exit(1)
	}

	if err = (&readinessprovidercontroller.ReadinessProviderReconciler{
		Client:                     mgr.GetClient(),
		Clientset:                  k8sClientset,
		Log:                        ctrl.Log.WithName("controllers").WithName("ReadinessProvider").WithValues("apigroup", "core"),
		Scheme:                     mgr.GetScheme(),
		ResourceExistenceCondition: conditions.NewResourceExistenceConditionFunc(),
		RestConfig:                 restConfig,
		DefaultQueryClient:         clusterQueryClient,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ReadinessProvider")
		os.Exit(1)
	}

	if err = (&corev1alpha2.ReadinessProvider{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "ReadinessProvider", "apigroup", "config")
		os.Exit(1)
	}

	//+kubebuilder:scaffold:builder

	signalHandler := ctrl.SetupSignalHandler()

	setupLog.Info("Starting certificate manager")
	certManagerOpts := &certs.Options{
		Client:                        cl,
		Logger:                        ctrl.Log.WithName("readinessprovider-webhook-cert-manager"),
		CertDir:                       webhookSecretVolumeMountPath,
		WebhookConfigLabel:            webhookConfigLabel,
		RotationIntervalAnnotationKey: "tanzu.vmware.com/readinessprovider-webhook-rotation-interval",
		NextRotationAnnotationKey:     "tanzu.vmware.com/readinessprovider-webhook-next-rotation",
		RotationCountAnnotationKey:    "tanzu.vmware.com/readinessprovider-webhook-rotation-count",
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
