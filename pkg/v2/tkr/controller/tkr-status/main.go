// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"reflect"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	kapppkgv1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/controller/tkr-status/clusterstatus"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/controller/tkr-status/osimage"
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
	utilruntime.Must(kapppkgv1.AddToScheme(scheme))
}

var (
	metricsAddr  string
	tkrNamespace string
)

func init() {
	flag.StringVar(&metricsAddr, "metrics-bind-addr", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&tkrNamespace, "namespace", "tkg-system", "Namespace for TKR related resources")
	flag.Parse()

	setupLog.Info("Version", "version", buildinfo.Version, "buildDate", buildinfo.Date, "sha", buildinfo.SHA)

	tkrConfig = tkr.Config{Namespace: tkrNamespace}
}

var (
	tkrConfig tkr.Config
)

func main() {
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	setupLog.Info("Version", "version", buildinfo.Version, "buildDate", buildinfo.Date, "sha", buildinfo.SHA)

	mgr := createManager()
	ctx := signals.SetupSignalHandler()

	tkrResolver := resolver.New()

	tkrReconciler := &tkr.Reconciler{
		Log:    mgr.GetLogger().WithName("tkr.TKR"),
		Client: mgr.GetClient(),
		Cache:  tkrResolver,
		Ctx:    ctx,
		Config: tkrConfig,
	}
	osImageReconciler := &osimage.Reconciler{
		Log:    mgr.GetLogger().WithName("osimage.OSImage"),
		Client: mgr.GetClient(),
		Cache:  tkrResolver,
	}
	clusterStatusReconciler := &clusterstatus.Reconciler{
		Log:         mgr.GetLogger().WithName("cluster.UpdatesAvailable"),
		Client:      mgr.GetClient(),
		TKRResolver: tkrResolver,
		Context:     ctx,
	}

	setupWithManager(mgr, []managedComponent{
		tkrReconciler,
		osImageReconciler,
		clusterStatusReconciler,
	})

	startManager(ctx, mgr)
}

func createManager() manager.Manager {
	setupLog.Info("setting up manager")
	mgr, err := manager.New(config.GetConfigOrDie(), manager.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
	})
	if err != nil {
		setupLog.Error(err, "unable to set up controller manager")
		os.Exit(1)
	}
	return mgr
}

func setupWithManager(mgr manager.Manager, managedComponents []managedComponent) {
	for _, c := range managedComponents {
		setupLog.Info("setting up component", "type", fullTypeName(c))
		if err := c.SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to setup component", "type", fullTypeName(c))
			os.Exit(1)
		}
	}
}

func fullTypeName(c managedComponent) string {
	cType := reflect.TypeOf(c)
	for cType.Kind() == reflect.Ptr {
		cType = cType.Elem()
	}
	return fmt.Sprintf("%s.%s", cType.PkgPath(), cType.Name())
}

type managedComponent interface {
	SetupWithManager(ctrl.Manager) error
}

func startManager(ctx context.Context, mgr manager.Manager) {
	setupLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "unable to run manager")
		os.Exit(1)
	}
}
