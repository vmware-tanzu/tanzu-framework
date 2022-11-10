// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"reflect"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	kapppkgiv1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	kapppkgv1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/tkr/controller/tkr-source/compatibility"
	"github.com/vmware-tanzu/tanzu-framework/tkr/controller/tkr-source/fetcher"
	"github.com/vmware-tanzu/tanzu-framework/tkr/controller/tkr-source/pkgcr"
	"github.com/vmware-tanzu/tanzu-framework/tkr/controller/tkr-source/registry"
	"github.com/vmware-tanzu/tanzu-framework/tkr/controller/tkr-source/tkr"
	"github.com/vmware-tanzu/tanzu-framework/util/buildinfo"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(runv1.AddToScheme(scheme))
	utilruntime.Must(clusterv1.AddToScheme(scheme))
	utilruntime.Must(kapppkgv1.AddToScheme(scheme))
	utilruntime.Must(kapppkgiv1.AddToScheme(scheme))
	utilruntime.Must(corev1.AddToScheme(scheme))
}

var (
	metricsAddr               string
	tkrNamespace              string
	legacyTKRNamespace        string
	tkrPkgServiceAccountName  string
	bomImagePath              string
	bomMetadataImagePath      string
	tkrRepoImagePath          string
	initTKRDiscoveryFreq      int
	continuousTKRDiscoverFreq int
	skipVerifyRegistryCerts   bool
)

func init() {
	flag.StringVar(&metricsAddr, "metrics-bind-addr", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&tkrNamespace, "namespace", "tkg-system", "Namespace for TKR related resources")
	flag.StringVar(&legacyTKRNamespace, "legacy-namespace", "", "Legacy namespace for TKR BOM ConfigMaps")
	flag.StringVar(&tkrPkgServiceAccountName, "sa-name", "tkr-source-controller-manager-sa", "ServiceAccount name used by TKR PackageInstalls")
	flag.StringVar(&bomImagePath, "bom-image-path", "", "The BOM image path.")
	flag.StringVar(&bomMetadataImagePath, "bom-metadata-image-path", "", "The BOM compatibility metadata image path.")
	flag.StringVar(&tkrRepoImagePath, "tkr-repo-image-path", "", "The TKR Package Repository image path.")
	flag.BoolVar(&skipVerifyRegistryCerts, "skip-verify-registry-cert", false, "Set whether to verify server's certificate chain and host name")
	flag.IntVar(&initTKRDiscoveryFreq, "initial-discover-frequency", 60, "Initial TKR discovery frequency in seconds")
	flag.IntVar(&continuousTKRDiscoverFreq, "continuous-discover-frequency", 600, "Continuous TKR discovery frequency in seconds")
	flag.Parse()

	setupLog.Info("Version", "version", buildinfo.Version, "buildDate", buildinfo.Date, "sha", buildinfo.SHA)

	registryConfig = registry.Config{
		TKRNamespace:       tkrNamespace,
		VerifyRegistryCert: !skipVerifyRegistryCerts,
	}
	fetcherConfig = fetcher.Config{
		TKRNamespace:         tkrNamespace,
		LegacyTKRNamespace:   legacyTKRNamespace,
		BOMImagePath:         bomImagePath,
		BOMMetadataImagePath: bomMetadataImagePath,
		TKRRepoImagePath:     tkrRepoImagePath,
		TKRDiscoveryOption: fetcher.TKRDiscoveryIntervals{
			InitialDiscoveryFrequency:    time.Duration(initTKRDiscoveryFreq) * time.Second,
			ContinuousDiscoveryFrequency: time.Duration(continuousTKRDiscoverFreq) * time.Second,
		},
	}
	pkgcrConfig = pkgcr.Config{
		ServiceAccountName: tkrPkgServiceAccountName,
	}
	compatibilityConfig = compatibility.Config{
		TKRNamespace: tkrNamespace,
	}
}

var (
	registryConfig      registry.Config
	fetcherConfig       fetcher.Config
	pkgcrConfig         pkgcr.Config
	compatibilityConfig compatibility.Config
)

func main() {
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr := createManager()
	ctx := signals.SetupSignalHandler()

	tkrCompatibility := &compatibility.Compatibility{
		Client: mgr.GetClient(),
		Config: compatibilityConfig,
		Log:    mgr.GetLogger().WithName("tkr-compatibility"),
	}
	registryInstance := registry.New(mgr.GetClient(), registryConfig)
	fetcherInstance := &fetcher.Fetcher{
		Log:           mgr.GetLogger().WithName("tkr-fetcher"),
		Client:        mgr.GetClient(),
		Config:        fetcherConfig,
		Registry:      registryInstance,
		Compatibility: tkrCompatibility,
	}
	pkgcrReconciler := &pkgcr.Reconciler{
		Log:      mgr.GetLogger().WithName("tkr-source"),
		Client:   mgr.GetClient(),
		Config:   pkgcrConfig,
		Registry: registryInstance,
	}
	compatibilityReconciler := &compatibility.Reconciler{
		Ctx:           ctx,
		Log:           mgr.GetLogger().WithName("tkr-compatibility"),
		Client:        mgr.GetClient(),
		Config:        compatibilityConfig,
		Compatibility: tkrCompatibility,
	}
	tkrReconciler := &tkr.Reconciler{
		Log:    mgr.GetLogger().WithName("tkr-compatibility"),
		Client: mgr.GetClient(),
	}

	setupWithManager(mgr, []managedComponent{
		registryInstance,
		fetcherInstance,
		pkgcrReconciler,
		compatibilityReconciler,
		tkrReconciler,
	})

	startManager(ctx, mgr)
}

func createManager() manager.Manager {
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
