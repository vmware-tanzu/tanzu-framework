// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"flag"
	"os"

	runv1alpha1 "github.com/vmware-tanzu-private/core/apis/run/v1alpha1"
	tkrsourcectr "github.com/vmware-tanzu-private/core/pkg/v1/tkr/controllers/source"
	mgrcontext "github.com/vmware-tanzu-private/core/pkg/v1/tkr/pkg/context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = capi.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
	_ = runv1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var enableLeaderElection bool
	var bomImagePath string
	var bomMetadataImagePath string
	var caCertPath string
	var verifyCerts bool
	var initTKRDiscoveryFreq float64
	var continuousTKRDiscoverFreq float64
	flag.StringVar(&bomImagePath, "bom-image-path", "", "The BOM image path.")
	flag.StringVar(&bomMetadataImagePath, "bom-metadata-image-path", "", "The BOM compatibility metadata image path.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&verifyCerts, "registry-verify-certs", true, "Set whether to verify server's certificate chain and host name")
	flag.StringVar(&caCertPath, "registry-ca-cert-path", "", "Add CA certificates for registry API")
	flag.Float64Var(&initTKRDiscoveryFreq, "initial-discover-frequency", 60, "Initial TKR discovery frequency in seconds")
	flag.Float64Var(&continuousTKRDiscoverFreq, "continuous-discover-frequency", 600, "Continuous TKR discovery frequency in seconds")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:           scheme,
		Port:             9443,
		LeaderElection:   enableLeaderElection,
		LeaderElectionID: "abf9f9ab.tanzu.vmware.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	mgrContext := &mgrcontext.ControllerManagerContext{
		Client:               mgr.GetClient(),
		Context:              context.Background(),
		BOMImagePath:         bomImagePath,
		BOMMetadataImagePath: bomMetadataImagePath,
		Logger:               ctrllog.Log,
		Scheme:               mgr.GetScheme(),
		VerifyRegistryCert:   verifyCerts,
		RegistryCertPath:     caCertPath,
		TKRDiscoveryOption:   mgrcontext.NewTanzuKubernetesReleaseDiscoverOptions(initTKRDiscoveryFreq, continuousTKRDiscoverFreq),
	}

	if err := tkrsourcectr.AddToManager(mgrContext, mgr); err != nil {
		setupLog.Error(err, "error initialzing the tkr-source-controller")
		os.Exit(1)
	}

	/*if err := tkrlabelctr.AddToManager(mgrContext, mgr); err != nil {
		setupLog.Error(err, "error initialzing the tkr-labeling-controller")
		os.Exit(1)
	}*/

	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
