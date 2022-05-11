// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/controller/tkr-status/clusterstatus"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/controller/tkr-status/tkr"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/resolver"
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
	var metricsAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-addr", ":8080", "The address the metric endpoint binds to.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	setupLog.Info("Version", "version", buildinfo.Version, "buildDate", buildinfo.Date, "sha", buildinfo.SHA)

	// Setup Manager
	setupLog.Info("setting up manager")
	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
	})
	if err != nil {
		setupLog.Error(err, "unable to set up controller manager")
		os.Exit(1)
	}

	ctx := signals.SetupSignalHandler()
	tkrResolver := resolver.New()

	if err := (&tkr.Reconciler{
		Log:    mgr.GetLogger().WithName("tkr.TKR"),
		Client: mgr.GetClient(),
		Cache:  tkrResolver,
		Object: &runv1.TanzuKubernetesRelease{},
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "TKR cache")
		os.Exit(1)
	}

	if err := (&tkr.Reconciler{
		Log:    mgr.GetLogger().WithName("tkr.OSImage"),
		Client: mgr.GetClient(),
		Cache:  tkrResolver,
		Object: &runv1.OSImage{},
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "OSImage cache")
		os.Exit(1)
	}

	if err := (&clusterstatus.Reconciler{
		Log:         mgr.GetLogger().WithName("cluster.UpdatesAvailable"),
		Client:      mgr.GetClient(),
		TKRResolver: tkrResolver,
		Context:     ctx,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Cluster UpdatesAvailable")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "unable to run manager")
		os.Exit(1)
	}
}
