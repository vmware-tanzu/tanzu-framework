// Copyright 2020 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"flag"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1beta1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	kapppkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	kappdatapkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/addons/controllers"
	antreacontroller "github.com/vmware-tanzu/tanzu-framework/addons/controllers/antrea"
	calicocontroller "github.com/vmware-tanzu/tanzu-framework/addons/controllers/calico"
	cpicontroller "github.com/vmware-tanzu/tanzu-framework/addons/controllers/cpi"
	csicontroller "github.com/vmware-tanzu/tanzu-framework/addons/controllers/csi"
	kappcontroller "github.com/vmware-tanzu/tanzu-framework/addons/controllers/kapp-controller"
	addonconfig "github.com/vmware-tanzu/tanzu-framework/addons/pkg/config"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/crdwait"
	cniv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cni/v1alpha1"
	cpiv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cpi/v1alpha1"
	csiv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/csi/v1alpha1"
	runtanzuv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
	runtanzuv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	klog.InitFlags(nil)

	_ = clientgoscheme.AddToScheme(scheme)
	_ = kappctrl.AddToScheme(scheme)
	_ = kapppkg.AddToScheme(scheme)
	_ = kappdatapkg.AddToScheme(scheme)
	_ = runtanzuv1alpha1.AddToScheme(scheme)
	_ = clusterapiv1beta1.AddToScheme(scheme)
	_ = controlplanev1beta1.AddToScheme(scheme)
	_ = runtanzuv1alpha3.AddToScheme(scheme)
	_ = cniv1alpha1.AddToScheme(scheme)
	_ = cpiv1alpha1.AddToScheme(scheme)
	_ = csiv1alpha1.AddToScheme(scheme)

	// +kubebuilder:scaffold:scheme
}

type addonFlags struct {
	metricsAddr                 string
	enableLeaderElection        bool
	clusterConcurrency          int
	syncPeriod                  time.Duration
	appSyncPeriod               time.Duration
	appWaitTimeout              time.Duration
	addonNamespace              string
	addonServiceAccount         string
	addonClusterRole            string
	addonClusterRoleBinding     string
	addonImagePullPolicy        string
	corePackageRepoName         string
	healthdAddr                 string
	httpProxyClusterVarName     string
	httpsProxyClusterVarName    string
	noProxyClusterVarName       string
	proxyCACertClusterVarName   string
	ipFamilyClusterVarName      string
	featureGateClusterBootstrap bool
}

func parseAddonFlags(addonFlags *addonFlags) {
	// controller configurations
	flag.StringVar(&addonFlags.metricsAddr, "metrics-bind-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&addonFlags.enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.IntVar(&addonFlags.clusterConcurrency, "cluster-concurrency", 10,
		"Number of clusters to process simultaneously")
	flag.DurationVar(&addonFlags.syncPeriod, "sync-period", 10*time.Minute,
		"The minimum interval at which watched resources are reconciled (e.g. 10m)")
	flag.DurationVar(&addonFlags.appSyncPeriod, "app-sync-period", 5*time.Minute, "Frequency of app reconciliation (e.g. 5m)")
	flag.DurationVar(&addonFlags.appWaitTimeout, "app-wait-timeout", 30*time.Second, "Maximum time to wait for app to be ready (e.g. 30s)")
	// resource configurations (optional)
	flag.StringVar(&addonFlags.addonNamespace, "addon-namespace", "tkg-system", "The namespace of addon resources")
	flag.StringVar(&addonFlags.addonServiceAccount, "addon-service-account-name", "tkg-addons-app-sa", "The name of addon service account")
	flag.StringVar(&addonFlags.addonClusterRole, "addon-cluster-role-name", "tkg-addons-app-cluster-role", "The name of addon clusterRole")
	flag.StringVar(&addonFlags.addonClusterRoleBinding, "addon-cluster-role-binding-name", "tkg-addons-app-cluster-role-binding", "The name of addon clusterRoleBinding")
	flag.StringVar(&addonFlags.addonImagePullPolicy, "addon-image-pull-policy", "IfNotPresent", "The addon image pull policy")
	flag.StringVar(&addonFlags.corePackageRepoName, "core-package-repo-name", "tanzu-core", "The name of core package repository")
	flag.StringVar(&addonFlags.healthdAddr, "health-addr", ":18316", "The address the health endpoint binds to.")
	flag.StringVar(&addonFlags.httpProxyClusterVarName, "http-proxy-cluster-var-name", constants.DefaultHTTPProxyClusterClassVarName, "HTTP proxy setting cluster variable name")
	flag.StringVar(&addonFlags.httpsProxyClusterVarName, "https-proxy-cluster-var-name", constants.DefaultHTTPSProxyClusterClassVarName, "HTTPS proxy setting cluster variable name")
	flag.StringVar(&addonFlags.noProxyClusterVarName, "no-proxy-cluster-var-name", constants.DefaultNoProxyClusterClassVarName, "No-proxy setting cluster variable name")
	flag.StringVar(&addonFlags.proxyCACertClusterVarName, "proxy-ca-cert-cluster-var-name", constants.DefaultProxyCaCertClusterClassVarName, "Proxy CA certificate cluster variable name")
	flag.StringVar(&addonFlags.ipFamilyClusterVarName, "ip-family-cluster-var-name", constants.DefaultIPFamilyClusterClassVarName, "IP family setting cluster variable name")
	flag.BoolVar(&addonFlags.featureGateClusterBootstrap, "feature-gate-cluster-bootstrap", false, "Feature gate to enable clusterbootstap and addonconfig controllers that rely on TKR v1alphav3")

	flag.Parse()
}

func main() {
	// parse flags
	flags := &addonFlags{}
	parseAddonFlags(flags)

	ctrl.SetLogger(klogr.New())
	setupLog.Info("Version", "version", buildinfo.Version, "buildDate", buildinfo.Date, "sha", buildinfo.SHA)

	ctx := ctrl.SetupSignalHandler()

	crdwaiter := crdwait.CRDWaiter{
		Ctx: ctx,
		ClientSetFn: func() (kubernetes.Interface, error) {
			return kubernetes.NewForConfig(ctrl.GetConfigOrDie())
		},
		Logger:       setupLog,
		Scheme:       scheme,
		PollInterval: constants.CRDWaitPollInterval,
		PollTimeout:  constants.CRDWaitPollTimeout,
	}
	if err := crdwaiter.WaitForCRDs(controllers.GetExternalCRDs(),
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: os.Getenv("POD_NAME"), Namespace: os.Getenv("POD_NAMESPACE")}},
		constants.AddonControllerName,
	); err != nil {
		setupLog.Error(err, "unable to wait for CRDs")
		os.Exit(1)
	}
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     flags.metricsAddr,
		Port:                   9443,
		LeaderElection:         flags.enableLeaderElection,
		LeaderElectionID:       "5832a104.run.tanzu.addons",
		SyncPeriod:             &flags.syncPeriod,
		HealthProbeBindAddress: flags.healthdAddr,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	addonReconciler := &controllers.AddonReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("Addon"),
		Scheme: mgr.GetScheme(),
		Config: addonconfig.AddonControllerConfig{
			AppSyncPeriod:           flags.appSyncPeriod,
			AppWaitTimeout:          flags.appWaitTimeout,
			AddonNamespace:          flags.addonNamespace,
			AddonServiceAccount:     flags.addonServiceAccount,
			AddonClusterRole:        flags.addonClusterRole,
			AddonClusterRoleBinding: flags.addonClusterRoleBinding,
			AddonImagePullPolicy:    flags.addonImagePullPolicy,
			CorePackageRepoName:     flags.corePackageRepoName,
		},
	}
	if err = addonReconciler.SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: flags.clusterConcurrency}); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Addon")
		os.Exit(1)
	}
	if flags.featureGateClusterBootstrap {
		enableClusterBootstrapAndConfigControllers(ctx, mgr, flags)
	}

	setupChecks(mgr)
	setupLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func setupChecks(mgr ctrl.Manager) {
	if err := mgr.AddReadyzCheck("ping", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to create ready check")
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("ping", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to create health check")
		os.Exit(1)
	}
}

func enableClusterBootstrapAndConfigControllers(ctx context.Context, mgr ctrl.Manager, flags *addonFlags) {
	if err := (&calicocontroller.CalicoConfigReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("CalicoConfigController"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: 1}); err != nil {
		setupLog.Error(err, "unable to create CalicoConfigController", "controller", "calico")
		os.Exit(1)
	}

	if err := (&antreacontroller.AntreaConfigReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("AntreaConfigController"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: 1}); err != nil {
		setupLog.Error(err, "unable to create AntreaConfigController", "controller", "antrea")
		os.Exit(1)
	}
	if err := (&kappcontroller.KappControllerConfigReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("KappControllerConfig"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: 1}); err != nil {
		setupLog.Error(err, "unable to create KappControllerConfig", "controller", "kapp")
		os.Exit(1)
	}
	if err := (&cpicontroller.VSphereCPIConfigReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("VSphereCPIConfig"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: 1}); err != nil {
		setupLog.Error(err, "unable to create CPIConfigController", "controller", "vspherecpi")
		os.Exit(1)
	}
	if err := (&csicontroller.VSphereCSIConfigReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: 1}); err != nil {
		setupLog.Error(err, "unable to create VSphereCSIConfigController", "controller", "csi")
		os.Exit(1)
	}

	bootstrapReconciler := controllers.NewClusterBootstrapReconciler(
		mgr.GetClient(),
		ctrl.Log.WithName("ClusterBootstrapController"),
		mgr.GetScheme(),
		&addonconfig.ClusterBootstrapControllerConfig{
			HTTPProxyClusterClassVarName:   constants.DefaultHTTPProxyClusterClassVarName,
			HTTPSProxyClusterClassVarName:  constants.DefaultHTTPSProxyClusterClassVarName,
			NoProxyClusterClassVarName:     constants.DefaultNoProxyClusterClassVarName,
			ProxyCACertClusterClassVarName: constants.DefaultProxyCaCertClusterClassVarName,
			IPFamilyClusterClassVarName:    constants.DefaultIPFamilyClusterClassVarName,
			SystemNamespace:                flags.addonNamespace,
			PkgiServiceAccount:             constants.PackageInstallServiceAccount,
			PkgiClusterRole:                constants.PackageInstallClusterRole,
			PkgiClusterRoleBinding:         constants.PackageInstallClusterRoleBinding,
			PkgiSyncPeriod:                 flags.syncPeriod,
		},
	)
	if err := bootstrapReconciler.SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: 1}); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "clusterbootstrap")
		os.Exit(1)
	}
}
