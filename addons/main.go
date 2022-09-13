// Copyright 2020 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"flag"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
	capvv1beta1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/v1beta1"
	capvvmwarev1beta1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/vmware/v1beta1"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	capiremote "sigs.k8s.io/cluster-api/controllers/remote"
	controlplanev1beta1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	kapppkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	kappdatapkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/addons/controllers"
	antreacontroller "github.com/vmware-tanzu/tanzu-framework/addons/controllers/antrea"
	awsebscsicontroller "github.com/vmware-tanzu/tanzu-framework/addons/controllers/awsebscsi"
	calicocontroller "github.com/vmware-tanzu/tanzu-framework/addons/controllers/calico"
	cpicontroller "github.com/vmware-tanzu/tanzu-framework/addons/controllers/cpi"
	csicontroller "github.com/vmware-tanzu/tanzu-framework/addons/controllers/csi"
	kappcontroller "github.com/vmware-tanzu/tanzu-framework/addons/controllers/kapp-controller"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/buildinfo"
	addonconfig "github.com/vmware-tanzu/tanzu-framework/addons/pkg/config"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/crdwait"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/webhooks"
	addonwebhooks "github.com/vmware-tanzu/tanzu-framework/addons/webhooks"
	cniv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/addonconfigs/cni/v1alpha1"
	cpiv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/addonconfigs/cpi/v1alpha1"
	csiv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/addonconfigs/csi/v1alpha1"
	runtanzuv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
	runtanzuv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	vmoperatorv1alpha1 "github.com/vmware-tanzu/vm-operator-api/api/v1alpha1"
	topologyv1alpha1 "github.com/vmware-tanzu/vm-operator/external/tanzu-topology/api/v1alpha1"
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
	_ = capvv1beta1.AddToScheme(scheme)
	_ = capvvmwarev1beta1.AddToScheme(scheme)
	_ = vmoperatorv1alpha1.AddToScheme(scheme)
	_ = topologyv1alpha1.AddToScheme(scheme)

	// +kubebuilder:scaffold:scheme
}

type addonFlags struct {
	metricsAddr                     string
	enableLeaderElection            bool
	leaderElectionID                string
	leaderElectionLeaseDuration     time.Duration
	leaderElectionRenewDeadline     time.Duration
	leaderElectionRetryPeriod       time.Duration
	clusterConcurrency              int
	syncPeriod                      time.Duration
	appSyncPeriod                   time.Duration
	appWaitTimeout                  time.Duration
	clusterDeleteTimeout            time.Duration
	addonNamespace                  string
	addonServiceAccount             string
	addonClusterRole                string
	addonClusterRoleBinding         string
	addonImagePullPolicy            string
	corePackageRepoName             string
	webhookServerPort               int
	healthdAddr                     string
	ipFamilyClusterVarName          string
	featureGateClusterBootstrap     bool
	featureGatePackageInstallStatus bool
	enablePprof                     bool
	pprofBindAddress                string
}

func parseAddonFlags(addonFlags *addonFlags) {
	// controller configurations

	// Bind metrics endpoint to localhost, following the pattern in cluster-api projects where they moved away from kube-rbac-proxy for just guarding metrics endpoint.
	// see https://github.com/kubernetes-sigs/cluster-api/issues/4679 and https://github.com/kubernetes-sigs/cluster-api/issues/4325
	// 18317 is an available port in Supervisor cluster
	flag.StringVar(&addonFlags.metricsAddr, "metrics-bind-addr", "localhost:18317", "The address the metric endpoint binds to.")
	flag.BoolVar(&addonFlags.enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&addonFlags.leaderElectionID, "leader-elect-id", constants.AddonControllerManagerLeaderElectionResourceName,
		"The name of the resource that leader election will use for holding the leader lock.")
	flag.DurationVar(&addonFlags.leaderElectionLeaseDuration, "leader-elect-lease-duration", 15*time.Second,
		"Interval at which non-leader candidates will wait to force acquire leadership")
	flag.DurationVar(&addonFlags.leaderElectionRenewDeadline, "leader-elect-renew-deadline", 10*time.Second,
		"Duration that the leading controller manager will retry refreshing leadership before giving up")
	flag.DurationVar(&addonFlags.leaderElectionRetryPeriod, "leader-elect-retry-period", 2*time.Second,
		"Duration the LeaderElector clients should wait between tries of actions")
	flag.IntVar(&addonFlags.clusterConcurrency, "cluster-concurrency", 10,
		"Number of clusters to process simultaneously")
	flag.DurationVar(&addonFlags.syncPeriod, "sync-period", 10*time.Minute,
		"The minimum interval at which watched resources are reconciled (e.g. 10m)")
	flag.DurationVar(&addonFlags.appSyncPeriod, "app-sync-period", 5*time.Minute, "Frequency of app reconciliation (e.g. 5m)")
	flag.DurationVar(&addonFlags.appWaitTimeout, "app-wait-timeout", 30*time.Second, "Maximum time to wait for app to be ready (e.g. 30s)")
	flag.DurationVar(&addonFlags.clusterDeleteTimeout, "cluster-delete-timeout", 10*time.Minute, "Maximum time to wait for addon resources to be deleted before allowing cluster deletion to proceed")

	// resource configurations (optional)
	flag.StringVar(&addonFlags.addonNamespace, "addon-namespace", "tkg-system", "The namespace of addon resources")
	flag.StringVar(&addonFlags.addonServiceAccount, "addon-service-account-name", "tkg-addons-app-sa", "The name of addon service account")
	flag.StringVar(&addonFlags.addonClusterRole, "addon-cluster-role-name", "tkg-addons-app-cluster-role", "The name of addon clusterRole")
	flag.StringVar(&addonFlags.addonClusterRoleBinding, "addon-cluster-role-binding-name", "tkg-addons-app-cluster-role-binding", "The name of addon clusterRoleBinding")
	flag.StringVar(&addonFlags.addonImagePullPolicy, "addon-image-pull-policy", "IfNotPresent", "The addon image pull policy")
	flag.StringVar(&addonFlags.corePackageRepoName, "core-package-repo-name", "tanzu-core", "The name of core package repository")
	flag.StringVar(&addonFlags.healthdAddr, "health-addr", ":18316", "The address the health endpoint binds to.")
	flag.IntVar(&addonFlags.webhookServerPort, "webhook-server-port", 9865, "The port that the webhook server serves at.")
	flag.StringVar(&addonFlags.ipFamilyClusterVarName, "ip-family-cluster-var-name", constants.DefaultIPFamilyClusterClassVarName, "IP family setting cluster variable name")
	flag.BoolVar(&addonFlags.featureGateClusterBootstrap, "feature-gate-cluster-bootstrap", false, "Feature gate to enable clusterbootstap and addonconfig controllers that rely on TKR v1alphav3")
	flag.BoolVar(&addonFlags.featureGatePackageInstallStatus, "feature-gate-package-install-status", false, "Feature gate to enable packageinstallstatus controller")
	flag.BoolVar(&addonFlags.enablePprof, "enable-pprof", false, "Enable pprof web server")
	flag.StringVar(&addonFlags.pprofBindAddress, "pprof-bind-addr", ":18318", "Bind address of pprof web server if enabled")

	flag.Parse()
}

func main() {
	// parse flags
	flags := &addonFlags{}
	parseAddonFlags(flags)

	ctrl.SetLogger(klogr.New())
	setupLog.Info("Version", "version", buildinfo.Version, "buildDate", buildinfo.Date, "sha", buildinfo.SHA)

	ctx := ctrl.SetupSignalHandler()

	if flags.enablePprof {
		go func() {
			setupLog.Info("[Warn]: pprof web server enabled", "bindAddress", flags.pprofBindAddress)
			if err := http.ListenAndServe(flags.pprofBindAddress, nil); err != nil {
				setupLog.Error(err, "error binding to pprof-bind-addr")
				os.Exit(1)
			}
		}()
	}

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
	if flags.featureGateClusterBootstrap {
		if err := os.MkdirAll(constants.WebhookCertDir, 0755); err != nil {
			setupLog.Error(err, "unable to create directory for webhook certificates", "directory", constants.WebhookCertDir)
			os.Exit(1)
		}
	}
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     flags.metricsAddr,
		Port:                   flags.webhookServerPort,
		CertDir:                constants.WebhookCertDir,
		LeaderElection:         flags.enableLeaderElection,
		LeaderElectionID:       flags.leaderElectionID,
		LeaseDuration:          &flags.leaderElectionLeaseDuration,
		RenewDeadline:          &flags.leaderElectionRenewDeadline,
		RetryPeriod:            &flags.leaderElectionRetryPeriod,
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
		enableWebhooks(ctx, mgr, flags)
	}

	if flags.featureGatePackageInstallStatus {
		enablePackageInstallStatusController(ctx, mgr, flags)
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
		Config: addonconfig.CalicoConfigControllerConfig{
			ConfigControllerConfig: addonconfig.ConfigControllerConfig{SystemNamespace: flags.addonNamespace}},
	}).SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: 1}); err != nil {
		setupLog.Error(err, "unable to create CalicoConfigController", "controller", "calico")
		os.Exit(1)
	}

	if err := (&antreacontroller.AntreaConfigReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("AntreaConfigController"),
		Scheme: mgr.GetScheme(),
		Config: addonconfig.AntreaConfigControllerConfig{
			ConfigControllerConfig: addonconfig.ConfigControllerConfig{SystemNamespace: flags.addonNamespace}},
	}).SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: 1}); err != nil {
		setupLog.Error(err, "unable to create AntreaConfigController", "controller", "antrea")
		os.Exit(1)
	}
	if err := (&kappcontroller.KappControllerConfigReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("KappControllerConfig"),
		Scheme: mgr.GetScheme(),
		Config: addonconfig.KappControllerConfigControllerConfig{
			ConfigControllerConfig: addonconfig.ConfigControllerConfig{SystemNamespace: flags.addonNamespace}},
	}).SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: 1}); err != nil {
		setupLog.Error(err, "unable to create KappControllerConfig", "controller", "kapp")
		os.Exit(1)
	}
	if err := (&cpicontroller.VSphereCPIConfigReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("VSphereCPIConfig"),
		Scheme: mgr.GetScheme(),
		Config: addonconfig.VSphereCPIConfigControllerConfig{
			ConfigControllerConfig: addonconfig.ConfigControllerConfig{SystemNamespace: flags.addonNamespace}},
	}).SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: 1}); err != nil {
		setupLog.Error(err, "unable to create CPIConfigController", "controller", "vspherecpi")
		os.Exit(1)
	}

	if err := (&csicontroller.VSphereCSIConfigReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("VSphereCSIConfig"),
		Scheme: mgr.GetScheme(),
		Config: addonconfig.VSphereCSIConfigControllerConfig{
			ConfigControllerConfig: addonconfig.ConfigControllerConfig{SystemNamespace: flags.addonNamespace}},
	}).SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: 1}); err != nil {
		setupLog.Error(err, "unable to create CSIConfigController", "controller", "vspherecsi")
		os.Exit(1)
	}

	if err := (&awsebscsicontroller.AwsEbsCSIConfigReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("AwsEbsCSIConfig"),
		Scheme: mgr.GetScheme(),
		Config: addonconfig.AwsEbsCSIConfigControllerConfig{
			ConfigControllerConfig: addonconfig.ConfigControllerConfig{SystemNamespace: flags.addonNamespace}},
	}).SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: 1}); err != nil {
		setupLog.Error(err, "unable to create AwsEbsCSIConfigController", "controller", "awsebscsi")
		os.Exit(1)
	}

	if err := (&controllers.MachineReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("MachineController"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: 1}); err != nil {
		setupLog.Error(err, "unable to create addons MachineController", "controller", "machine")
		os.Exit(1)
	}

	bootstrapReconciler := controllers.NewClusterBootstrapReconciler(
		mgr.GetClient(),
		ctrl.Log.WithName("ClusterBootstrapController"),
		mgr.GetScheme(),
		&addonconfig.ClusterBootstrapControllerConfig{
			IPFamilyClusterClassVarName: constants.DefaultIPFamilyClusterClassVarName,
			SystemNamespace:             flags.addonNamespace,
			PkgiServiceAccount:          constants.PackageInstallServiceAccount,
			PkgiClusterRole:             constants.PackageInstallClusterRole,
			PkgiClusterRoleBinding:      constants.PackageInstallClusterRoleBinding,
			PkgiSyncPeriod:              flags.syncPeriod,
			ClusterDeleteTimeout:        flags.clusterDeleteTimeout,
		},
	)
	if err := bootstrapReconciler.SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: 1}); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "clusterbootstrap")
		os.Exit(1)
	}
}

func enableWebhooks(ctx context.Context, mgr ctrl.Manager, flags *addonFlags) {
	certPath := path.Join(constants.WebhookCertDir, "tls.crt")
	keyPath := path.Join(constants.WebhookCertDir, "tls.key")
	webhookTLS := webhooks.WebhookTLS{
		Ctx:           ctx,
		K8sConfig:     mgr.GetConfig(),
		CertPath:      certPath,
		KeyPath:       keyPath,
		Name:          constants.WebhookScrtName,
		ServiceName:   constants.WebhookServiceName,
		LabelSelector: constants.AddonWebhookLabelKey + "=" + constants.AddonWebhookLabelValue,
		Logger:        setupLog,
		Namespace:     flags.addonNamespace,
		RotationTime:  constants.WebhookCertLifeTime,
	}
	if err := webhookTLS.ManageCertificates(constants.WebhookCertManagementFrequency); err != nil {
		setupLog.Error(err, "Unable to start webhook tls certificate management")
		os.Exit(1)
	}
	// Set up the webhooks in the manager
	if err := (&cniv1alpha1.AntreaConfig{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to set up webhooks", "webhook", "antrea")
		os.Exit(1)
	}
	if err := (&cniv1alpha1.CalicoConfig{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to set up webhooks", "webhook", "calico")
		os.Exit(1)
	}
	clusterbootstrapWebhook := addonwebhooks.ClusterBootstrap{
		Client:          mgr.GetClient(),
		SystemNamespace: flags.addonNamespace,
	}
	if err := clusterbootstrapWebhook.SetupWebhookWithManager(ctx, mgr); err != nil {
		setupLog.Error(err, "unable to create ClusterBootstrap webhook", "webhook", "clusterbootstrap")
		os.Exit(1)
	}
	clusterbootstrapTemplateWebhook := addonwebhooks.ClusterBootstrapTemplate{
		SystemNamespace: flags.addonNamespace,
	}
	if err := clusterbootstrapTemplateWebhook.SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create clusterbootstrapTemplate webhook", "webhook", "clusterbootstraptemplate")
		os.Exit(1)
	}
	clusterPauseWebhook := addonwebhooks.ClusterPause{Client: mgr.GetClient()}
	if err := clusterPauseWebhook.SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to set up webhooks", "webhook", "clusterpause")
		os.Exit(1)
	}
}

func enablePackageInstallStatusController(ctx context.Context, mgr ctrl.Manager, flags *addonFlags) {
	// set up a ClusterCacheTracker to provide to PackageInstallStatus controller which requires a connection to remote clusters
	// the informers/caches are created only for objects accessed through Get/List in the code.
	// we only read PackageInstall resource through our cached client, by default the client excludes configmap and secret resources.
	l := ctrl.Log.WithName("remote").WithName("ClusterCacheTracker")
	tracker, err := capiremote.NewClusterCacheTracker(mgr, capiremote.ClusterCacheTrackerOptions{Log: &l})
	if err != nil {
		setupLog.Error(err, "unable to create cluster cache tracker")
		os.Exit(1)
	}

	// set up CluterCacheReconciler to drops the accessor via deleteAccessor upon cluster deletion.
	// after a Cluster is deleted (from the apiserver/etcd), the CluterCacheReconciler drops the accessor via deleteAccessor.
	// the healthchecking inside of the ClusterCacheTracker would remove the accessor at some point but the ClusterCacheReconciler does it more cleanly directly upon Cluster delete
	if err := (&capiremote.ClusterCacheReconciler{
		Client:  mgr.GetClient(),
		Log:     ctrl.Log.WithName("remote").WithName("ClusterCacheReconciler"),
		Tracker: tracker,
	}).SetupWithManager(ctx, mgr, controller.Options{MaxConcurrentReconciles: 1}); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ClusterCacheReconciler")
		os.Exit(1)
	}

	pkgiStatusReconciler := controllers.NewPackageInstallStatusReconciler(
		mgr.GetClient(),
		ctrl.Log.WithName("PackageInstallStatusController"),
		mgr.GetScheme(),
		&addonconfig.PackageInstallStatusControllerConfig{
			SystemNamespace: flags.addonNamespace,
		},
		tracker,
	)
	// Use a custom rater limiter to do bigger backoffs for failures because updating status shouldn't overwhelm resource usage
	if err := pkgiStatusReconciler.SetupWithManager(
		ctx,
		mgr,
		controller.Options{
			MaxConcurrentReconciles: 1,
			RateLimiter: workqueue.NewItemExponentialFailureRateLimiter(
				constants.PackageInstallStatusControllerRateLimitBaseDelay,
				constants.PackageInstallStatusControllerRateLimitMaxDelay),
		}); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "PackageInstallStatus")
		os.Exit(1)
	}
}
