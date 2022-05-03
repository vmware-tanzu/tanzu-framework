// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	kapppkgiv1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	kapppkgv1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"

	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/controller/source/compatibility"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/controller/source/fetcher"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/controller/source/pkgcr"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(runv1.AddToScheme(scheme))
	utilruntime.Must(kapppkgv1.AddToScheme(scheme))
	utilruntime.Must(kapppkgiv1.AddToScheme(scheme))
	utilruntime.Must(corev1.AddToScheme(scheme))
}

func main() {
	var tkrNamespace string
	var tkrPkgServiceAccountName string
	var bomImagePath string
	var bomMetadataImagePath string
	var tkrRepoImagePath string
	var initTKRDiscoveryFreq int
	var continuousTKRDiscoverFreq int
	var skipVerifyRegistryCerts bool

	flag.StringVar(&tkrNamespace, "namespace", "tkr-system", "Namespace for TKR related resources")
	flag.StringVar(&tkrPkgServiceAccountName, "sa-name", "tkr-service-manager-sa", "ServiceAccount name used by TKR PackageInstalls")
	flag.StringVar(&bomImagePath, "bom-image-path", "", "The BOM image path.")
	flag.StringVar(&bomMetadataImagePath, "bom-metadata-image-path", "", "The BOM compatibility metadata image path.")
	flag.StringVar(&tkrRepoImagePath, "tkr-repo-image-path", "", "The TKR Package Repository image path.")
	flag.BoolVar(&skipVerifyRegistryCerts, "skip-verify-registry-cert", false, "Set whether to verify server's certificate chain and host name")
	flag.IntVar(&initTKRDiscoveryFreq, "initial-discover-frequency", 60, "Initial TKR discovery frequency in seconds")
	flag.IntVar(&continuousTKRDiscoverFreq, "continuous-discover-frequency", 600, "Continuous TKR discovery frequency in seconds")
	flag.Parse()

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
		Scheme: scheme,
	})
	if err != nil {
		setupLog.Error(err, "unable to set up controller manager")
		os.Exit(1)
	}

	ctx := signals.SetupSignalHandler()

	if err := mgr.Add(&fetcher.Fetcher{
		Log:    mgr.GetLogger().WithName("tkr-fetcher"),
		Client: mgr.GetClient(),
		Config: fetcher.Config{
			TKRNamespace:         tkrNamespace,
			BOMImagePath:         bomImagePath,
			BOMMetadataImagePath: bomMetadataImagePath,
			TKRRepoImagePath:     tkrRepoImagePath,
			VerifyRegistryCert:   !skipVerifyRegistryCerts,
			TKRDiscoveryOption: fetcher.TKRDiscoveryIntervals{
				InitialDiscoveryFrequency:    time.Duration(initTKRDiscoveryFreq) * time.Second,
				ContinuousDiscoveryFrequency: time.Duration(continuousTKRDiscoverFreq) * time.Second,
			},
		},
	}); err != nil {
		setupLog.Error(err, "unable to add fetcher to controller manager")
		os.Exit(1)
	}

	if err := (&pkgcr.Reconciler{
		Log:    mgr.GetLogger().WithName("tkr-source"),
		Client: mgr.GetClient(),
		Config: pkgcr.Config{
			ServiceAccountName: tkrPkgServiceAccountName,
		},
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "TKR Package")
		os.Exit(1)
	}

	if err := (&compatibility.Reconciler{
		Ctx:    ctx,
		Log:    mgr.GetLogger().WithName("tkr-compatibility"),
		Client: mgr.GetClient(),
		Config: compatibility.Config{
			TKRNamespace: tkrNamespace,
		},
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "TKR Compatibility")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "unable to run manager")
		os.Exit(1)
	}
}
