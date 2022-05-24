// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"os"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(runv1.AddToScheme(scheme))
	utilruntime.Must(clusterv1.AddToScheme(scheme))
}

func main() {
	var webhookCertDir string
	var metricsAddr string
	var webhookServerPort int
	flag.StringVar(&webhookCertDir, "webhook-cert-dir", "/tmp/k8s-webhook-server/serving-certs/", "Webhook cert directory.")
	flag.StringVar(&metricsAddr, "metrics-bind-addr", ":8080", "The address the metric endpoint binds to.")
	flag.IntVar(&webhookServerPort, "webhook-server-port", 9443, "The port that the webhook server serves at.")

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

	// Setup webhooks
	setupLog.Info("setting up webhook server")
	//hookServer := mgr.GetWebhookServer()
	//
	//setupLog.Info("registering webhooks to the webhook server")
	//hookServer.Register("/mutate-cluster", &webhook.Admission{
	//	Handler: &cluster.Webhook{
	//		Log:         mgr.GetLogger().WithName("handler.Cluster"),
	//		TKRResolver: tkrResolver,
	//		Client:      mgr.GetClient(),
	//	},
	//})

	setupLog.Info("registering webhooks to the webhook server")
	setupLog.Info("starting manager")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "unable to run manager")
		os.Exit(1)
	}
}
