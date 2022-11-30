// Copyright 2022 VMware, Inc. All Rights Reserved.
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
	cliflag "k8s.io/component-base/cli/flag"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha2"
	"github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/util/buildinfo"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	utilruntime.Must(v1alpha2.AddToScheme(scheme))
	utilruntime.Must(v1alpha3.AddToScheme(scheme))
}

func main() {
	var webhookCertDir string
	var metricsAddr string
	var webhookServerPort int
	var tlsMinVersion string
	var tlsCipherSuites string
	flag.StringVar(&webhookCertDir, "webhook-cert-dir", "/tmp/k8s-webhook-server/serving-certs/", "Webhook cert directory.")
	flag.StringVar(&metricsAddr, "metrics-bind-addr", ":8080", "The address the metric endpoint binds to.")
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

	// Setup a Manager
	setupLog.Info("setting up manager")
	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		CertDir:            webhookCertDir,
		Port:               webhookServerPort,
	})
	if err != nil {
		setupLog.Error(err, "unable to set up controller manager")
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
	setupWebhooks(mgr)

	setupLog.Info("starting manager")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "unable to run manager")
		os.Exit(1)
	}
}

func setupWebhooks(mgr manager.Manager) {
	if err := (&v1alpha3.TanzuKubernetesRelease{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "TanzuKubernetesRelease")
		os.Exit(1)
	}
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
