// Copyright 2020 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/klog"
	"k8s.io/klog/klogr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	// +kubebuilder:scaffold:imports

	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	pkgiv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	pkgv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	clusterapiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	controlplanev1alpha3 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1alpha3"

	"github.com/vmware-tanzu-private/core/addons/controllers"
	addonconfig "github.com/vmware-tanzu-private/core/addons/pkg/config"
	runtanzuv1alpha1 "github.com/vmware-tanzu-private/core/apis/run/v1alpha1"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	klog.InitFlags(nil)

	_ = clientgoscheme.AddToScheme(scheme)
	_ = kappctrl.AddToScheme(scheme)
	_ = pkgiv1alpha1.AddToScheme(scheme)
	_ = pkgv1alpha1.AddToScheme(scheme)
	_ = runtanzuv1alpha1.AddToScheme(scheme)
	_ = clusterapiv1alpha3.AddToScheme(scheme)
	_ = controlplanev1alpha3.AddToScheme(scheme)

	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var clusterConcurrency int
	var syncPeriod time.Duration
	var appSyncPeriod time.Duration
	var appWaitTimeout time.Duration
	var addonNamespace string
	var addonServiceAccount string
	var addonClusterRole string
	var addonClusterRoleBinding string
	var addonImagePullPolicy string
	var corePackageRepoName string

	// controller configurations
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.IntVar(&clusterConcurrency, "cluster-concurrency", 10,
		"Number of clusters to process simultaneously")
	flag.DurationVar(&syncPeriod, "sync-period", 10*time.Minute,
		"The minimum interval at which watched resources are reconciled (e.g. 10m)")
	flag.DurationVar(&appSyncPeriod, "app-sync-period", 5*time.Minute, "Frequency of app reconciliation (e.g. 5m)")
	flag.DurationVar(&appWaitTimeout, "app-wait-timeout", 30*time.Second, "Maximum time to wait for app to be ready (e.g. 30s)")
	// resource configurations (optional)
	flag.StringVar(&addonNamespace, "addon-namespace", "tkg-system", "The namespace of addon resources")
	flag.StringVar(&addonServiceAccount, "addon-service-account-name", "tkg-addons-app-sa", "The name of addon service account")
	flag.StringVar(&addonClusterRole, "addon-cluster-role-name", "tkg-addons-app-cluster-role", "The name of addon clusterRole")
	flag.StringVar(&addonClusterRoleBinding, "addon-cluster-role-binding-name", "tkg-addons-app-cluster-role-binding", "The name of addon clusterRoleBinding")
	flag.StringVar(&addonImagePullPolicy, "addon-image-pull-policy", "IfNotPresent", "The addon image pull policy")
	flag.StringVar(&corePackageRepoName, "core-package-repo-name", "core", "The name of core package repository")

	flag.Parse()

	ctrl.SetLogger(klogr.New())

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "5832a104.run.tanzu.addons",
		SyncPeriod:         &syncPeriod,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	ctx := ctrl.SetupSignalHandler()

	if err = (&controllers.AddonReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Addon"),
		Scheme: mgr.GetScheme(),
		Config: addonconfig.Config{
			AppSyncPeriod:           appSyncPeriod,
			AppWaitTimeout:          appWaitTimeout,
			AddonNamespace:          addonNamespace,
			AddonServiceAccount:     addonServiceAccount,
			AddonClusterRole:        addonClusterRole,
			AddonClusterRoleBinding: addonClusterRoleBinding,
			AddonImagePullPolicy:    addonImagePullPolicy,
			CorePackageRepoName:     corePackageRepoName,
		},
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
