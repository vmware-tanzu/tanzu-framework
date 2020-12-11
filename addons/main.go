package main

// Copyright (c) 2020 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

import (
	"flag"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	capiremote "sigs.k8s.io/cluster-api/controllers/remote"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports

	"github.com/vmware-tanzu-private/core/addons/controllers"

	runtanzuv1alpha1 "github.com/vmware-tanzu-private/core/apis/run/v1alpha1"
	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	clusterapiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	controlplanev1alpha3 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1alpha3"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = kappctrl.AddToScheme(scheme)
	_ = runtanzuv1alpha1.AddToScheme(scheme)
	_ = clusterapiv1alpha3.AddToScheme(scheme)
	_ = controlplanev1alpha3.AddToScheme(scheme)

	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var clusterConcurrency int
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.IntVar(&clusterConcurrency, "cluster-concurrency", 10,
		"Number of clusters to process simultaneously")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "5832a104.run.tanzu.addons",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Set up a ClusterCacheTracker and ClusterCacheReconciler to provide to controllers
	// requiring a connection to a remote cluster
	tracker, err := capiremote.NewClusterCacheTracker(
		ctrl.Log.WithName("remote").WithName("ClusterCacheTracker"),
		mgr,
	)
	if err != nil {
		setupLog.Error(err, "unable to create cluster cache tracker")
		os.Exit(1)
	}

	ctx := ctrl.SetupSignalHandler()

	if err = (&controllers.AddonReconciler{
		Client:  mgr.GetClient(),
		Log:     ctrl.Log.WithName("controllers").WithName("Addon"),
		Scheme:  mgr.GetScheme(),
		Tracker: tracker,
	}).SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: clusterConcurrency}); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Addon")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
