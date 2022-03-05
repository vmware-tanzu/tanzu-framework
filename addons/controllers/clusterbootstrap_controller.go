// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	errorsutil "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/discovery"
	cacheddiscovery "k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
	"k8s.io/utils/pointer"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	clusterapiutil "sigs.k8s.io/cluster-api/util"
	clusterapipatchutil "sigs.k8s.io/cluster-api/util/patch"
	clusterApiPredicates "sigs.k8s.io/cluster-api/util/predicates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	kappctrlv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	kapppkgiv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	kapppkgv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	versions "github.com/vmware-tanzu/carvel-vendir/pkg/vendir/versions/v1alpha1"
	addonconfig "github.com/vmware-tanzu/tanzu-framework/addons/pkg/config"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/types"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	"github.com/vmware-tanzu/tanzu-framework/addons/predicates"
	runtanzuv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

// ClusterBootstrapReconciler reconciles a ClusterBootstrap object
type ClusterBootstrapReconciler struct {
	client.Client
	Log     logr.Logger
	Scheme  *runtime.Scheme
	Config  *addonconfig.ClusterBootstrapControllerConfig
	context context.Context

	// internal properties
	controller    controller.Controller
	dynamicClient dynamic.Interface
	// on demand dynamic watches for provider refs
	providerWatches map[string]client.Object
	// discovery client for looking up api-resources and preferred versions
	cachedDiscoveryClient discovery.CachedDiscoveryInterface
	// cache for resolved api-resources so that look up is fast (cleared periodically)
	providerGVR map[schema.GroupKind]*schema.GroupVersionResource
	liveClient  client.Client
}

// NewClusterBootstrapReconciler returns a reconciler for ClusterBootstrap
func NewClusterBootstrapReconciler(c client.Client, log logr.Logger, scheme *runtime.Scheme, config *addonconfig.ClusterBootstrapControllerConfig) *ClusterBootstrapReconciler {
	return &ClusterBootstrapReconciler{
		Client: c,
		Log:    log,
		Scheme: scheme,
		Config: config,
	}
}

// ClusterBootstrapWatchInputs contains the inputs for Watches set in ClusterBootstrap
type ClusterBootstrapWatchInputs struct {
	src          source.Source
	eventHandler handler.EventHandler
}

// +kubebuilder:rbac:groups=run.tanzu.vmware.com,resources=clusterBootstraps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=run.tanzu.vmware.com,resources=clusterBootstraps/status,verbs=get;update;patch

// SetupWithManager performs the setup actions for an ClusterBootstrap controller, using the passed in mgr.
func (r *ClusterBootstrapReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	blder := ctrl.NewControllerManagedBy(mgr).For(&clusterapiv1beta1.Cluster{})

	// Set the Watches for resources watched by ClusterBootstrap
	for _, watchInputs := range r.watchesForClusterBootstrap() {
		blder.Watches(watchInputs.src, watchInputs.eventHandler, builder.WithPredicates(predicates.TKR(r.Log)))
	}

	ctrlr, err := blder.
		WithOptions(options).
		WithEventFilter(clusterApiPredicates.ResourceNotPaused(r.Log)).
		WithEventFilter(predicates.ClusterHasLabel(constants.TKRLabelClassyClusters, r.Log)).
		Build(r)
	if err != nil {
		r.Log.Error(err, "Error creating ClusterBootstrap controller")
		return err
	}

	r.controller = ctrlr
	r.context = ctx
	dynClient, err := dynamic.NewForConfig(mgr.GetConfig())
	if err != nil {
		r.Log.Error(err, "Error creating dynamic client")
		return err
	}
	r.dynamicClient = dynClient
	r.providerWatches = make(map[string]client.Object)

	r.providerGVR = make(map[schema.GroupKind]*schema.GroupVersionResource)
	clientset := kubernetes.NewForConfigOrDie(mgr.GetConfig())
	r.cachedDiscoveryClient = cacheddiscovery.NewMemCacheClient(clientset.Discovery())

	go r.periodicGVRCachesClean()

	r.liveClient, err = client.New(mgr.GetConfig(), client.Options{Scheme: mgr.GetScheme()})
	if err != nil {
		return err
	}

	return nil
}

// Reconcile performs the reconciliation action for the controller.
func (r *ClusterBootstrapReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues(constants.ClusterNamespaceLogKey, req.Namespace, constants.ClusterNameLogKey, req.Name)

	// get cluster object
	cluster := &clusterapiv1beta1.Cluster{}
	if err := r.Client.Get(ctx, req.NamespacedName, cluster); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Cluster not found")
			return ctrl.Result{}, nil
		}

		log.Error(err, "unable to fetch cluster")
		return ctrl.Result{}, err
	}

	// make sure the TKR object exists
	tkrName := cluster.Labels[constants.TKRLabelClassyClusters]
	tkr, err := util.GetTKRByName(r.context, r.Client, tkrName)
	if err != nil {
		log.Error(err, "unable to fetch TKR object", "name", tkrName)
		return ctrl.Result{}, err
	}

	// if tkr is not found, should not requeue for the reconciliation
	if tkr == nil {
		log.Info("TKR object not found", "name", tkrName)
		return ctrl.Result{}, nil
	}

	log.Info("Reconciling cluster")

	// if deletion timestamp is set, handle cluster deletion
	if !cluster.GetDeletionTimestamp().IsZero() {
		// TODO handle delete
		// https://github.com/vmware-tanzu/tanzu-framework/issues/1591
		return ctrl.Result{}, nil
	}
	return r.reconcileNormal(cluster, log)
}

// reconcileNormal reconciles the ClusterBootstrap object
func (r *ClusterBootstrapReconciler) reconcileNormal(cluster *clusterapiv1beta1.Cluster, log logr.Logger) (ctrl.Result, error) {
	// get or clone or patch from template
	clusterBootstrap, err := r.createOrPatchClusterBootstrapFromTemplate(cluster, log)
	if err != nil {
		return ctrl.Result{Requeue: true}, err
	}
	if clusterBootstrap == nil {
		return ctrl.Result{}, nil
	}

	// reconcile the proxy settings of the cluster
	if err := r.reconcileClusterProxyAndNetworkSettings(cluster, log); err != nil {
		return ctrl.Result{}, err
	}

	if cluster.Status.Phase != string(clusterapiv1beta1.ClusterPhaseProvisioned) {
		r.Log.Info(fmt.Sprintf("cluster %s/%s does not have status phase %s", cluster.Namespace, cluster.Name, clusterapiv1beta1.ClusterPhaseProvisioned))
		return ctrl.Result{}, nil
	}
	remoteClient, err := util.GetClusterClient(r.context, r.Client, r.Scheme, clusterapiutil.ObjectKey(cluster))
	if err != nil {
		log.Error(err, "Error getting remote cluster client")
		return ctrl.Result{Requeue: true}, err
	}

	// Create a PackageInstall CR under the cluster namespace for deploying the kapp-controller on the remote cluster.
	// We need kapp-controller to be deployed prior to CNI, CPI, CSI. This will be a no-op if the cluster object is mgmt
	// cluster.
	if err := r.createOrPatchKappPackageInstall(clusterBootstrap, cluster); err != nil {
		// Return error if kapp-controller fails to be deployed, let reconciler try again
		return ctrl.Result{}, err
	}

	if err := r.prepareRemoteCluster(cluster, remoteClient); err != nil {
		return ctrl.Result{}, err
	}

	// Create or patch the resources for CNI, CPI, CSI to be running on the remote cluster.
	// Those resources include Package CR, data value Secret, PackageInstall CR.
	var corePackages []*runtanzuv1alpha3.ClusterBootstrapPackage
	corePackages = append(corePackages, clusterBootstrap.Spec.CPI, clusterBootstrap.Spec.CSI)
	corePackages = append(corePackages, clusterBootstrap.Spec.CNIs...)

	// The following filtering out of nil items is not necessary in production
	// as we do not expect CNI, CPI, CSI to be nil and webhook should handle
	// the validations against those fields. This nil filter is mainly to allow
	// local envtest run when any above component is missing.
	corePackages = removeCorePackagesNils(corePackages)

	for _, corePackage := range corePackages {
		// There are different ways to have all the resources created or patched on remote cluster. Current solution is
		// to handle packages in sequence order. I.e., Create all resources for CNI first, and then CPI, CSI. It is also
		// possible to create all resources in a different order or in parallel. We will consider to use goroutines to create
		// all resources in parallel on remote cluster if there is performance issue from sequential ordering.
		if err := r.createOrPatchAddonResourcesOnRemote(cluster, corePackage, remoteClient); err != nil {
			// For core packages, we require all their creation or patching to succeed, so if error happens against any of the
			// packages, we return error and let the reconciler retry again.
			log.Error(err, fmt.Sprintf("unable to create or patch all the required resources for %s on cluster: %s/%s",
				corePackage.RefName, cluster.Namespace, cluster.Name))
			return ctrl.Result{RequeueAfter: time.Second * 10}, err
		}
	}

	// Create or patch the resources for additionalPackages
	for _, additionalPkg := range clusterBootstrap.Spec.AdditionalPackages {
		if err := r.createOrPatchAddonResourcesOnRemote(cluster, additionalPkg, remoteClient); err != nil {
			// Logging has been handled in createOrPatchAddonResourcesOnRemote()
			return ctrl.Result{Requeue: true}, err
		}
		// set watches on provider objects in additional packages if not already set
		if additionalPkg.ValuesFrom != nil && additionalPkg.ValuesFrom.ProviderRef != nil {
			if err := r.watchProvider(additionalPkg.ValuesFrom.ProviderRef, clusterBootstrap.Namespace, log); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

// createOrPatchClusterBootstrapFromTemplate will get, clone or update a ClusterBootstrap associated with a cluster
// all linked secret refs and object refs are cloned into the same namespace as ClusterBootstrap
func (r *ClusterBootstrapReconciler) createOrPatchClusterBootstrapFromTemplate(cluster *clusterapiv1beta1.Cluster,
	log logr.Logger) (*runtanzuv1alpha3.ClusterBootstrap, error) {

	tkrName := cluster.Labels[constants.TKRLabelClassyClusters]
	clusterBootstrapTemplate := &runtanzuv1alpha3.ClusterBootstrapTemplate{}
	key := client.ObjectKey{Namespace: r.Config.SystemNamespace, Name: tkrName}
	if err := r.Client.Get(r.context, key, clusterBootstrapTemplate); err != nil {
		log.Error(err, "unable to fetch ClusterBootstrapTemplate", "objectkey", key)
		return nil, err
	}

	clusterBootstrap := &runtanzuv1alpha3.ClusterBootstrap{}
	err := r.Client.Get(r.context, client.ObjectKeyFromObject(cluster), clusterBootstrap)
	// if found and resolved tkr is the same, return found object as the TKR is supposed to be immutable
	// also preserves any user changes
	if err == nil && tkrName == clusterBootstrap.Status.ResolvedTKR {
		return clusterBootstrap, nil
	}

	if !apierrors.IsNotFound(err) {
		return nil, err
	}
	if clusterBootstrap.UID == "" {
		log.Info("ClusterBootstrap for cluster does not exist, cloning from template")

		clusterBootstrap.Name = cluster.Name
		clusterBootstrap.Namespace = cluster.Namespace
		clusterBootstrap.Spec = clusterBootstrapTemplate.Spec.DeepCopy()
		// get selected CNI and populate clusterBootstrap.Spec.CNIs with it
		cniPackage, err := r.getCNIForClusterBootstrap(clusterBootstrapTemplate, cluster, log)
		if err != nil {
			return nil, err
		}
		clusterBootstrap.Spec.CNIs = []*runtanzuv1alpha3.ClusterBootstrapPackage{cniPackage}

		secrets, providers, err := r.cloneSecretsAndProviders(cluster, clusterBootstrap, clusterBootstrapTemplate.Namespace, log)
		if err != nil {
			r.Log.Error(err, "unable to clone secrets or providers")
			return nil, err
		}

		clusterBootstrap.OwnerReferences = []metav1.OwnerReference{
			{
				APIVersion:         clusterapiv1beta1.GroupVersion.String(),
				Kind:               cluster.Kind,
				Name:               cluster.Name,
				UID:                cluster.UID,
				Controller:         pointer.BoolPtr(true),
				BlockOwnerDeletion: pointer.BoolPtr(true),
			},
		}

		if err := r.Client.Create(r.context, clusterBootstrap); err != nil {
			return nil, err
		}
		// ensure ownerRef of clusterBootstrap on created secrets and providers, this can only be done after
		// clusterBootstrap is created
		ownerRef := metav1.OwnerReference{
			APIVersion:         runtanzuv1alpha3.GroupVersion.String(),
			Kind:               "ClusterBootstrap", // kind is empty after create
			Name:               clusterBootstrap.Name,
			UID:                clusterBootstrap.UID,
			Controller:         pointer.BoolPtr(true),
			BlockOwnerDeletion: pointer.BoolPtr(true),
		}
		if err := r.ensureOwnerRef(&ownerRef, secrets, providers); err != nil {
			r.Log.Error(err, fmt.Sprintf("unable to ensure ClusterBootstrap %s/%s as a ownerRef on created secrets and providers", clusterBootstrap.Namespace, clusterBootstrap.Name))
			return nil, err
		}

		clusterBootstrap.Status.ResolvedTKR = tkrName
		if err := r.Status().Update(r.context, clusterBootstrap); err != nil {
			r.Log.Error(err, fmt.Sprintf("unable to update the status of ClusterBootstrap %s/%s", clusterBootstrap.Namespace, clusterBootstrap.Name))
			return nil, err
		}
		r.Log.Info("cloned clusterBootstrap", "clusterBootstrap", clusterBootstrap)
		return clusterBootstrap, nil
	}

	// TODO upgrade needs patch (update versions of all packages, merge configs, add additional packages and remove packages that don't exist anymore)
	// https://github.com/vmware-tanzu/tanzu-framework/issues/1584
	if tkrName != clusterBootstrap.Status.ResolvedTKR {
		log.Info("TODO handle upgrade")
		return nil, nil
	}
	return nil, errors.New("should not happen")
}

// createOrPatchKappPackageInstall contains the logic that create/update PackageInstall CR for kapp-controller on
// mgmt cluster. The kapp-controller running on mgmt cluster reconciles the PackageInstall CR and creates kapp-controller resources
// on remote workload cluster. This is required for a workload cluster and its corresponding package installations to be functional.
func (r *ClusterBootstrapReconciler) createOrPatchKappPackageInstall(clusterBootstrap *runtanzuv1alpha3.ClusterBootstrap, cluster *clusterapiv1beta1.Cluster) error {
	// Skip if the cluster object represents the management cluster
	if _, exists := cluster.Labels[constants.ManagementClusterRoleLabel]; exists {
		r.Log.Info(fmt.Sprintf("cluster %s/%s is management cluster, skip creating or patching the PackageInstall CR for kapp-controller", cluster.Namespace, cluster.Name))
		return nil
	}

	// In order to create PackageInstall CR, we need to get the Package.Spec.RefName and Package.Spec.Version
	packageRefName, packageVersion, err := util.GetPackageMetadata(r.context, r.liveClient, clusterBootstrap.Spec.Kapp.RefName, cluster.Namespace)
	if packageRefName == "" || packageVersion == "" || err != nil {
		// Package.Spec.RefName and Package.Spec.Version are required fields for Package CR. We do not expect them to be
		// empty and error should not happen when fetching them from a Package CR.
		r.Log.Error(err, fmt.Sprintf("unable to fetch Package.Spec.RefName or Package.Spec.Version from Package %s/%s",
			cluster.Namespace, clusterBootstrap.Spec.Kapp.RefName))
		return err
	}

	pkgi := &kapppkgiv1alpha1.PackageInstall{
		ObjectMeta: metav1.ObjectMeta{
			// The legacy addons controller uses <cluster name>-<addon name> convention for naming the PackageInstall CR.
			// https://github.com/vmware-tanzu/tanzu-framework/blob/main/addons/controllers/package_reconciler.go#L195.
			// util.GeneratePackageInstallName() follows the same pattern.
			Name: util.GeneratePackageInstallName(cluster.Name, packageRefName),
			// kapp-controller PackageInstall CR is installed under the same namespace as tanzuClusterBootstrap. The namespace
			// is also the same as where the cluster belongs.
			Namespace: cluster.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: clusterapiv1beta1.GroupVersion.String(),
					Kind:       cluster.Kind,
					Name:       cluster.Name,
					UID:        cluster.UID,
				},
			},
		},
	}

	pkgiMutateFn := func() error {
		// TODO: Followup on the following fields, only add them if needed.
		// https://github.com/vmware-tanzu/tanzu-framework/issues/1677
		// if ipkg.ObjectMeta.Annotations == nil {
		//	 ipkg.ObjectMeta.Annotations = make(map[string]string)
		// }
		// ipkg.ObjectMeta.Annotations[addontypes.YttMarkerAnnotation] = ""
		pkgi.Spec.SyncPeriod = &metav1.Duration{Duration: r.Config.PkgiSyncPeriod}
		pkgi.Spec.PackageRef = &kapppkgiv1alpha1.PackageRef{
			// clusterBootstrap.Spec.Kapp.RefName is Package.Name. I.e., kapp-controller.tanzu.vmware.com.0.28.0+vmware.1-tkg.1-rc.1
			// PackageInstall.Spec.PackageRef looks for the Package.Spec.refName which is a short name of the full Package.Name
			// packageRefName and packageVersion are fetched from the Package CR.
			RefName: packageRefName,
			VersionSelection: &versions.VersionSelectionSemver{
				Constraints: packageVersion,
				Prereleases: &versions.VersionSelectionSemverPrereleases{},
			},
		}
		// Adding the cluster reference to PackageInstall spec to instruct kapp-controller where to deploy
		// the underlying resources
		clusterKubeconfigDetails := util.GetClusterKubeconfigSecretDetails(cluster)
		pkgi.Spec.Cluster = &kappctrlv1alpha1.AppCluster{
			KubeconfigSecretRef: &kappctrlv1alpha1.AppClusterKubeconfigSecretRef{
				Name: clusterKubeconfigDetails.Name,
				Key:  clusterKubeconfigDetails.Key,
			},
		}
		secretName, err := r.GetDataValueSecretNameFromBootstrapPackage(clusterBootstrap.Spec.Kapp, cluster.Namespace)
		if err != nil {
			return err
		}
		pkgi.Spec.Values = []kapppkgiv1alpha1.PackageInstallValues{
			{SecretRef: &kapppkgiv1alpha1.PackageInstallValuesSecretRef{
				Name: secretName},
			},
		}
		return nil
	}

	_, err = controllerutil.CreateOrPatch(r.context, r.Client, pkgi, pkgiMutateFn)
	if err != nil {
		r.Log.Error(err, fmt.Sprintf("unable to create or patch PackageInstall %s/%s for cluster: %s",
			pkgi.Namespace, pkgi.Name, cluster.Name))
		return err
	}

	r.Log.Info(fmt.Sprintf("created or patched the PackageInstall %s/%s for cluster %s", pkgi.Namespace, pkgi.Name, cluster.Name))
	return nil
}

// createOrPatchPackageOnRemote creates the Package CR on remote cluster. In order to install a package on remote cluster
// the Package CR needs to be present.
// createOrPatchPackageOnRemote returns a tuple: (<remote-package>, <error>)
func (r *ClusterBootstrapReconciler) createOrPatchPackageOnRemote(cluster *clusterapiv1beta1.Cluster,
	cbPkg *runtanzuv1alpha3.ClusterBootstrapPackage, clusterClient client.Client) (*kapppkgv1alpha1.Package, error) {

	var err error
	// Create or patch Package CR on remote cluster
	localPackage := &kapppkgv1alpha1.Package{}
	key := client.ObjectKey{Namespace: cluster.Namespace, Name: cbPkg.RefName}
	if err = r.liveClient.Get(r.context, key, localPackage); err != nil {
		// If there is an error to get the Carvel Package CR from local cluster, nothing needs to be created/cloned on remote.
		// Let the reconciler try again.
		r.Log.Error(err, fmt.Sprintf("unable to create or patch Package %s on cluster %s/%s. Error occurs when getting Package %s from the management cluster",
			key.String(), cluster.Namespace, cluster.Name, key.String()))
		return nil, err
	}
	remotePackage := &kapppkgv1alpha1.Package{}
	remotePackage.SetName(localPackage.Name)
	// The Package CR on remote cluster needs to be under configured system namespace
	remotePackage.SetNamespace(r.Config.SystemNamespace)
	_, err = controllerutil.CreateOrPatch(r.context, clusterClient, remotePackage, func() error {
		remotePackage.Spec = *localPackage.Spec.DeepCopy()
		// TODO: Follow up to see if we need to preserve all the other fields, like annotations
		// https://github.com/vmware-tanzu/tanzu-framework/issues/1678
		return nil
	})
	if err != nil {
		r.Log.Error(err, fmt.Sprintf("unable to create or patch Package resource %s/%s on cluster: %s/%s",
			remotePackage.Namespace, remotePackage.Name, cluster.Namespace, cluster.Name))
		return nil, err
	}
	return remotePackage, nil
}

// createOrPatchPackageInstallOnRemote creates or patches PackageInstall CR on remote cluster. The kapp-controller
// running on remote cluster will reconcile it and deploy resources.
func (r *ClusterBootstrapReconciler) createOrPatchPackageInstallOnRemote(cluster *clusterapiv1beta1.Cluster,
	cbPkg *runtanzuv1alpha3.ClusterBootstrapPackage, remoteSecret *corev1.Secret, clusterClient client.Client) (*kapppkgiv1alpha1.PackageInstall, error) {

	// In order to create PackageInstall CR, we need to get the Package.Spec.RefName and Package.Spec.Version
	remotePackageRefName, remotePackageVersion, err := util.GetPackageMetadata(r.context, clusterClient, cbPkg.RefName, r.Config.SystemNamespace)
	if remotePackageRefName == "" || remotePackageVersion == "" || err != nil {
		// Package.Spec.RefName and Package.Spec.Version are required fields for Package CR. We do not expect them to be
		// empty and error should not happen when fetching them from a Package CR.
		r.Log.Error(err, fmt.Sprintf("unable to fetch Package.Spec.RefName or Package.Spec.Version from Package %s/%s on cluster %s/%s",
			r.Config.SystemNamespace, cbPkg.RefName, cluster.Namespace, cluster.Name))
		return nil, err
	}

	// Create PackageInstall CRs on the remote workload cluster, kapp-controller will take care of reconciling them
	remotePkgi := &kapppkgiv1alpha1.PackageInstall{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GeneratePackageInstallName(cluster.Name, remotePackageRefName),
			Namespace: r.Config.SystemNamespace,
		},
	}

	_, err = controllerutil.CreateOrPatch(r.context, clusterClient, remotePkgi, func() error {
		remotePkgi.Spec.ServiceAccountName = r.Config.PkgiServiceAccount
		// remotePackageRefName and remotePackageVersion are fetched from the Package CR on remote cluster.
		remotePkgi.Spec.PackageRef = &kapppkgiv1alpha1.PackageRef{
			RefName: remotePackageRefName,
			VersionSelection: &versions.VersionSelectionSemver{
				Constraints: remotePackageVersion,
				Prereleases: &versions.VersionSelectionSemverPrereleases{},
			},
		}
		if remoteSecret != nil {
			// The nil remoteSecret means no data values for current ClusterBootstrapPackage are needed. And no remote secret
			// object gets created. The PackageInstall CR should be created without specifying the spec.Values.
			remotePkgi.Spec.Values = []kapppkgiv1alpha1.PackageInstallValues{
				{SecretRef: &kapppkgiv1alpha1.PackageInstallValuesSecretRef{
					Name: remoteSecret.Name},
				},
			}
		}
		return nil
	})
	if err != nil {
		r.Log.Error(err, fmt.Sprintf("unable to create or patch PackageInstall resource %s/%s on cluster: %s/%s",
			remotePkgi.Namespace, remotePkgi.Name, cluster.Namespace, cluster.Name))
		return nil, err
	}

	return remotePkgi, nil
}

func (r *ClusterBootstrapReconciler) reconcileSystemNamespace(clusterClient client.Client) error {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: r.Config.SystemNamespace,
		},
	}

	result, err := controllerutil.CreateOrPatch(r.context, clusterClient, namespace, nil)
	if err != nil {
		r.Log.Error(err, "Error creating or patching system namespace")
		return err
	}
	if result != controllerutil.OperationResultNone {
		r.Log.Info("created namespace", "namespace", r.Config.SystemNamespace)
	}
	return nil
}

func (r *ClusterBootstrapReconciler) prepareRemoteCluster(cluster *clusterapiv1beta1.Cluster, clusterClient client.Client) error {
	if err := r.reconcileSystemNamespace(clusterClient); err != nil {
		return err
	}

	// Create the ServiceAccount on remote cluster, so it could be referenced in PackageInstall CR for kapp-controller
	// reconciliation.
	if _, err := r.createOrPatchAddonServiceAccountOnRemote(cluster, clusterClient); err != nil {
		return err
	}

	// Create the ClusterRole on remote cluster, and bind it to the ServiceAccount created in above. kapp-controller
	// reconciliation needs privileges.
	return r.createOrPatchAddonRBACOnRemote(cluster, clusterClient)
}

func removeCorePackagesNils(pkgs []*runtanzuv1alpha3.ClusterBootstrapPackage) []*runtanzuv1alpha3.ClusterBootstrapPackage {
	var filtered []*runtanzuv1alpha3.ClusterBootstrapPackage
	for _, pkg := range pkgs {
		if pkg != nil {
			filtered = append(filtered, pkg)
		}
	}
	return filtered
}

// createOrPatchAddonServiceAccountOnRemote creates or patches the addon ServiceAccount on remote cluster.
// The ServiceAccount will be referenced by the PackageInstall CR, so that kapp-controller on remote cluster could consume
// for PackageInstall reconciliation.
func (r *ClusterBootstrapReconciler) createOrPatchAddonServiceAccountOnRemote(cluster *clusterapiv1beta1.Cluster, clusterClient client.Client) (*corev1.ServiceAccount, error) {
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Config.PkgiServiceAccount,
			Namespace: r.Config.SystemNamespace,
		},
	}

	r.Log.Info(fmt.Sprintf("creating or patching ServiceAccount %s/%s on cluster %s/%s",
		serviceAccount.Namespace, serviceAccount.Name, cluster.Namespace, cluster.Name))

	_, err := controllerutil.CreateOrPatch(r.context, clusterClient, serviceAccount, nil)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			// If the error is IsAlreadyExists, we ignore and return nil
			return nil, nil
		}
		r.Log.Error(err, fmt.Sprintf("unable to create or patch ServiceAccount %s/%s on cluster %s/%s",
			serviceAccount.Namespace, serviceAccount.Name, cluster.Namespace, cluster.Name))
		return nil, err
	}

	return serviceAccount, nil
}

// createOrPatchAddonRBACOnRemote creates or patches the ClusterRole, ClusterRoleBinding on remote cluster.
// The ClusterRole is bound to the ServiceAccount which is referenced by PackageInstall CR, so that kapp-controller on remote
// cluster could have privileges to lifecycle manage package resources.
func (r *ClusterBootstrapReconciler) createOrPatchAddonRBACOnRemote(cluster *clusterapiv1beta1.Cluster, clusterClient client.Client) error {
	addonRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: r.Config.PkgiClusterRole,
		},
	}

	if _, err := controllerutil.CreateOrPatch(r.context, clusterClient, addonRole, func() error {
		addonRole.Rules = []rbacv1.PolicyRule{
			{
				APIGroups: []string{"*"},
				Verbs:     []string{"*"},
				Resources: []string{"*"},
			},
		}
		return nil
	}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			r.Log.Error(err,
				fmt.Sprintf("unable to create or patch ClusterRole %s/%s on cluster %s/%s",
					addonRole.Namespace, addonRole.Name, cluster.Namespace, cluster.Name))
			return err
		}
		r.Log.Info(fmt.Sprintf("ClusterRole %s already exists on cluster %s/%s. Nothing to create or patch.", addonRole.Name, cluster.Namespace, cluster.Name))
	}

	r.Log.Info(fmt.Sprintf("created or patched ClusterRole %s/%s on cluster %s/%s",
		addonRole.Namespace, addonRole.Name, cluster.Namespace, cluster.Name))

	addonRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: r.Config.PkgiClusterRoleBinding,
		},
	}
	if _, err := controllerutil.CreateOrPatch(r.context, clusterClient, addonRoleBinding, func() error {
		addonRoleBinding.Subjects = []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      r.Config.PkgiServiceAccount,
				Namespace: r.Config.SystemNamespace,
			},
		}
		addonRoleBinding.RoleRef = rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     r.Config.PkgiClusterRole,
		}
		return nil
	}); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			r.Log.Error(err, fmt.Sprintf("unable to create or patch ClusterRoleBinding %s/%s on cluster %s/%s",
				addonRoleBinding.Namespace, addonRoleBinding.Name, cluster.Namespace, cluster.ClusterName))
			return err
		}
		r.Log.Info(fmt.Sprintf("ClusterRoleBinding %s/%s already exists on cluster %s/%s. Nothing to create or patch.",
			addonRoleBinding.Namespace, addonRole.Name, cluster.Namespace, cluster.Name))
	}

	return nil
}

// createOrPatchAddonResourcesOnRemote creates or patches the resources for a cluster bootstrap package on remote workload
// cluster. The resources are [Package CR, Secret for PackageInstall, PackageInstall CR].
func (r *ClusterBootstrapReconciler) createOrPatchAddonResourcesOnRemote(cluster *clusterapiv1beta1.Cluster,
	cbPkg *runtanzuv1alpha3.ClusterBootstrapPackage, clusterClient client.Client) error {

	remotePackage, err := r.createOrPatchPackageOnRemote(cluster, cbPkg, clusterClient)
	if err != nil {
		return err
	}
	r.Log.Info(fmt.Sprintf("created the Package CR %s on cluster %s/%s", remotePackage.Name, cluster.Namespace,
		cluster.Name))

	// Create or patch the data value secret on remote cluster. The data value secret has been generated by each
	// addon config controller on local cluster.
	remoteSecret, err := r.createOrPatchPackageInstallSecretOnRemote(cluster, cbPkg, clusterClient)
	if err != nil {
		// We expect there is NO error to create or patch the secret used for PackageInstall in a cluster.
		// Logging has been handled by createOrPatchPackageInstallSecretOnRemote() already
		return err
	}
	if remoteSecret != nil {
		r.Log.Info(fmt.Sprintf("created or patched secret for package %s on cluster %s/%s", remotePackage.Name, cluster.Namespace,
			cluster.Name))
	}

	pkgi, err := r.createOrPatchPackageInstallOnRemote(cluster, cbPkg, remoteSecret, clusterClient)
	if err != nil {
		return err
	}
	r.Log.Info(fmt.Sprintf("created or patched the PackageInstall CR %s/%s on cluster %s/%s",
		pkgi.Namespace, pkgi.Name, cluster.Namespace, cluster.Name))

	return nil
}

// createOrPatchPackageInstallSecretOnRemote creates or patches the secret used for PackageInstall in a cluster
func (r *ClusterBootstrapReconciler) createOrPatchPackageInstallSecretOnRemote(cluster *clusterapiv1beta1.Cluster,
	cbpkg *runtanzuv1alpha3.ClusterBootstrapPackage, clusterClient client.Client) (*corev1.Secret, error) {

	secretName, err := r.GetDataValueSecretNameFromBootstrapPackage(cbpkg, cluster.Namespace)
	if err != nil {
		// logging has been handled in GetDataValueSecretNameFromBootstrapPackage()
		return nil, err
	}
	if secretName == "" {
		r.Log.Info(fmt.Sprintf("no data values secret is needed for ClusterBootstrapPackage: %s, nothing to be created or patched on cluster %s/%s",
			cbpkg.RefName, cluster.Namespace, cluster.Name))
		return nil, nil
	}

	localSecret := &corev1.Secret{}
	key := client.ObjectKey{Namespace: cluster.Namespace, Name: secretName}
	if err = r.Get(r.context, key, localSecret); err != nil {
		r.Log.Error(err, "unable to fetch secret", "objectKey", key)
		return nil, err
	}

	// TODO: This logic should be moved to cloneSecretsAndProviders()
	// https://github.com/vmware-tanzu/tanzu-framework/issues/1729
	// Add cluster and package labels to secrets if not already present
	// This helps us to track the secrets in the watch and trigger Reconcile requests when these secrets are updated
	patchedSecret := localSecret.DeepCopy()
	if patchSecretWithLabels(patchedSecret, util.ParseStringForLabel(cbpkg.RefName), cluster.Name) {
		if err := r.Patch(r.context, patchedSecret, client.MergeFrom(localSecret)); err != nil {
			r.Log.Error(err, "unable to patch secret labels for ", "secret", localSecret.Name)
			return nil, err
		}
		r.Log.Info(fmt.Sprintf("patched the secret %s/%s with package and cluster labels", localSecret.Namespace, localSecret.Name))
	}

	packageRefName, _, err := util.GetPackageMetadata(r.context, r.liveClient, cbpkg.RefName, cluster.Namespace)
	if err != nil {
		r.Log.Error(err, fmt.Sprintf("unable to get Package CR %s/%s for its metadata", cluster.Namespace, cbpkg.RefName))
		return nil, err
	}

	remoteSecret := &corev1.Secret{}
	remoteSecret.Name = util.GenerateDataValueSecretName(cluster.Name, packageRefName)
	// The secret will be created or patched under tkg-system namespace on remote cluster
	remoteSecret.Namespace = r.Config.SystemNamespace
	remoteSecret.Type = corev1.SecretTypeOpaque

	dataValuesSecretMutateFn := func() error {
		remoteSecret.Data = map[string][]byte{}
		for k, v := range patchedSecret.Data {
			remoteSecret.Data[k] = v
		}
		return nil
	}

	_, err = controllerutil.CreateOrPatch(r.context, clusterClient, remoteSecret, dataValuesSecretMutateFn)
	if err != nil {
		r.Log.Error(err, "error creating or patching addon data values secret")
		return nil, err
	}
	return remoteSecret, nil
}

// patchSecretWithLabels updates the secret by adding package and cluster labels
// Return true if a patch was required, false if the labels were already present
func patchSecretWithLabels(secret *corev1.Secret, pkgName, clusterName string) bool {
	updateLabels := false
	if secret.Labels == nil {
		secret.Labels = map[string]string{}
		updateLabels = true
	} else if secret.Labels[types.PackageNameLabel] != pkgName ||
		secret.Labels[types.ClusterNameLabel] != clusterName {
		updateLabels = true
	}
	if updateLabels {
		secret.Labels[types.PackageNameLabel] = pkgName
		secret.Labels[types.ClusterNameLabel] = clusterName
	}
	return updateLabels
}

// cloneSecretsAndProviders clones linked secrets and providers into the same namespace as clusterBootstrap
func (r *ClusterBootstrapReconciler) cloneSecretsAndProviders(cluster *clusterapiv1beta1.Cluster, bootstrap *runtanzuv1alpha3.ClusterBootstrap,
	templateNS string, log logr.Logger) ([]*corev1.Secret, []*unstructured.Unstructured, error) {

	var createdProviders []*unstructured.Unstructured
	var createdSecrets []*corev1.Secret

	packages := append([]*runtanzuv1alpha3.ClusterBootstrapPackage{
		bootstrap.Spec.CPI,
		bootstrap.Spec.CSI,
		bootstrap.Spec.Kapp,
	}, bootstrap.Spec.CNIs...)
	packages = append(packages, bootstrap.Spec.AdditionalPackages...)

	for _, pkg := range packages {
		if pkg == nil {
			continue
		}
		secret, provider, err := r.updateValues(cluster, bootstrap, pkg, templateNS, log)
		if err != nil {
			return nil, nil, err
		}
		if secret != nil {
			createdSecrets = append(createdSecrets, secret)
		}
		if provider != nil {
			createdProviders = append(createdProviders, provider)
		}
	}

	return createdSecrets, createdProviders, nil
}

// updateValues updates secretRef and/or providerRef
func (r *ClusterBootstrapReconciler) updateValues(cluster *clusterapiv1beta1.Cluster, bootstrap *runtanzuv1alpha3.ClusterBootstrap,
	cbPkg *runtanzuv1alpha3.ClusterBootstrapPackage, cbTemplateNamespace string, log logr.Logger) (*corev1.Secret, *unstructured.Unstructured, error) {

	packageRefName, _, err := util.GetPackageMetadata(r.context, r.liveClient, cbPkg.RefName, cluster.Namespace)
	if packageRefName == "" || err != nil {
		// Package.Spec.RefName and Package.Spec.Version are required fields for Package CR. We do not expect them to be
		// empty and error should not happen when fetching them from a Package CR.
		r.Log.Error(err, fmt.Sprintf("unable to fetch Package.Spec.RefName or Package.Spec.Version from Package %s/%s",
			cluster.Namespace, cbPkg.RefName))
		return nil, nil, err
	}

	if cbPkg.ValuesFrom == nil {
		return nil, nil, nil
	}
	if cbPkg.ValuesFrom.SecretRef != "" {
		secret, err := r.updateValuesFromSecret(cluster, bootstrap, cbPkg, cbTemplateNamespace, packageRefName, log)
		if err != nil {
			return nil, nil, err
		}
		return secret, nil, nil
	}

	if cbPkg.ValuesFrom.ProviderRef != nil {
		provider, err := r.updateValuesFromProvider(cluster, bootstrap, cbPkg, cbTemplateNamespace, packageRefName, log)
		if err != nil {
			return nil, nil, err
		}
		return nil, provider, nil
	}
	return nil, nil, nil
}

// ensureOwnerRef will ensure the provided OwnerReference onto the secrets and provider objects
func (r *ClusterBootstrapReconciler) ensureOwnerRef(ownerRef *metav1.OwnerReference, secrets []*corev1.Secret, providers []*unstructured.Unstructured) error {
	for _, secret := range secrets {
		ownerRefsMutateFn := func() error {
			secret.OwnerReferences = clusterapiutil.EnsureOwnerRef(secret.OwnerReferences, *ownerRef)
			return nil
		}
		_, err := controllerutil.CreateOrPatch(r.context, r.Client, secret, ownerRefsMutateFn)
		if err != nil {
			r.Log.Error(err, fmt.Sprintf("unable to create or patch the secret %s/%s with ownerRef", secret.Namespace, secret.Name))
			return err
		}
	}
	for _, provider := range providers {
		gvr, err := r.getGVR(provider.GroupVersionKind().GroupKind())
		if err != nil {
			r.Log.Error(err, fmt.Sprintf("unable to get GVR of provider %s/%s", provider.GetNamespace(), provider.GetName()))
			return err
		}
		err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
			// We need to get and update, otherwise there could have concurrency issue: ["the object has been modified; please
			// apply your changes to the latest version and try again"]
			newProvider, errGetProvider := r.dynamicClient.Resource(*gvr).Namespace(provider.GetNamespace()).Get(r.context, provider.GetName(), metav1.GetOptions{})
			if errGetProvider != nil {
				r.Log.Error(errGetProvider, fmt.Sprintf("unable to get provider %s/%s", provider.GetNamespace(), provider.GetName()))
				return errGetProvider
			}
			newProvider = newProvider.DeepCopy()
			newProvider.SetOwnerReferences(clusterapiutil.EnsureOwnerRef(provider.GetOwnerReferences(), *ownerRef))
			_, errUpdateProvider := r.dynamicClient.Resource(*gvr).Namespace(newProvider.GetNamespace()).Update(r.context, newProvider, metav1.UpdateOptions{})
			if errUpdateProvider != nil {
				r.Log.Error(errUpdateProvider, fmt.Sprintf("unable to update provider %s/%s", provider.GetNamespace(), provider.GetName()))
				return errUpdateProvider
			}
			return nil
		})
		if err != nil {
			r.Log.Error(err, fmt.Sprintf("unable to update the OwnerRefrences for provider %s/%s", provider.GetNamespace(), provider.GetName()))
			return err
		}
	}
	return nil
}

// getGVR returns a GroupVersionResource for a GroupKind
func (r *ClusterBootstrapReconciler) getGVR(gk schema.GroupKind) (*schema.GroupVersionResource, error) {
	if gvr, ok := r.providerGVR[gk]; ok {
		return gvr, nil
	}
	apiResourceList, err := r.cachedDiscoveryClient.ServerPreferredResources()
	if err != nil {
		return nil, err
	}
	for _, apiResource := range apiResourceList {
		gv, err := schema.ParseGroupVersion(apiResource.GroupVersion)
		if err != nil {
			return nil, err
		}
		if gv.Group == gk.Group {
			for i := 0; i < len(apiResource.APIResources); i++ {
				if apiResource.APIResources[i].Kind == gk.Kind {
					r.providerGVR[gk] = &schema.GroupVersionResource{Group: gv.Group, Resource: apiResource.APIResources[i].Name, Version: gv.Version}
					return r.providerGVR[gk], nil
				}
			}
		}
	}

	return nil, fmt.Errorf("unable to find server preferred resource %s/%s", gk.Group, gk.Kind)
}

// periodicGVRCachesClean invalidates caches used for GVR lookup
func (r *ClusterBootstrapReconciler) periodicGVRCachesClean() {
	ticker := time.NewTicker(constants.DiscoveryCacheInvalidateInterval)
	for {
		select {
		case <-r.context.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			r.cachedDiscoveryClient.Invalidate()
			r.providerGVR = make(map[schema.GroupKind]*schema.GroupVersionResource)
		}
	}
}

// updateValuesFromSecret updates secretRef in valuesFrom
func (r *ClusterBootstrapReconciler) updateValuesFromSecret(cluster *clusterapiv1beta1.Cluster, bootstrap *runtanzuv1alpha3.ClusterBootstrap,
	pkg *runtanzuv1alpha3.ClusterBootstrapPackage, templateNS, pkgRefName string, log logr.Logger) (*corev1.Secret, error) {

	var newSecret *corev1.Secret
	if pkg.ValuesFrom.SecretRef != "" {
		secret := &corev1.Secret{}
		key := client.ObjectKey{Namespace: templateNS, Name: pkg.ValuesFrom.SecretRef}
		if err := r.Get(r.context, key, secret); err != nil {
			log.Error(err, "unable to fetch secret", "objectkey", key)
			return nil, err
		}
		newSecret = secret.DeepCopy()
		newSecret.ObjectMeta.Reset()
		newSecret.Name = fmt.Sprintf("%s-%s-package", cluster.Name, pkgRefName)
		newSecret.Namespace = bootstrap.Namespace

		var createOrPatchErr error
		_, createOrPatchErr = controllerutil.CreateOrPatch(r.context, r.Client, newSecret, func() error {
			newSecret.OwnerReferences = []metav1.OwnerReference{
				{
					APIVersion: clusterapiv1beta1.GroupVersion.String(),
					Kind:       cluster.Kind,
					Name:       cluster.Name,
					UID:        cluster.UID,
				},
			}

			// Add cluster and package labels to cloned secrets
			if newSecret.Labels == nil {
				newSecret.Labels = map[string]string{}
			}
			newSecret.Labels[types.PackageNameLabel] = util.ParseStringForLabel(pkg.RefName)
			newSecret.Labels[types.ClusterNameLabel] = cluster.Name
			// Set secret.Type to ClusterBootstrapManagedSecret to enable us to Watch these secrets
			newSecret.Type = constants.ClusterBootstrapManagedSecret
			return nil
		})
		if createOrPatchErr != nil {
			return nil, createOrPatchErr
		}
		r.Log.Info(fmt.Sprintf("created or patched Secret %s/%s", newSecret.Namespace, newSecret.Name))
		pkg.ValuesFrom.SecretRef = newSecret.Name
	}
	return newSecret, nil
}

// updateValuesFromProvider updates providerRef in valuesFrom
func (r *ClusterBootstrapReconciler) updateValuesFromProvider(cluster *clusterapiv1beta1.Cluster, bootstrap *runtanzuv1alpha3.ClusterBootstrap,
	pkg *runtanzuv1alpha3.ClusterBootstrapPackage, cbTemplateNamespace, pkgRefName string, log logr.Logger) (*unstructured.Unstructured, error) {

	var newProvider *unstructured.Unstructured
	var createdOrUpdatedProvider *unstructured.Unstructured
	valuesFrom := pkg.ValuesFrom
	if valuesFrom.ProviderRef != nil {
		gvr, err := r.getGVR(schema.GroupKind{Group: *valuesFrom.ProviderRef.APIGroup, Kind: valuesFrom.ProviderRef.Kind})
		if err != nil {
			log.Error(err, "failed to getGVR")
			return nil, err
		}
		provider, err := r.dynamicClient.Resource(*gvr).Namespace(cbTemplateNamespace).Get(r.context, valuesFrom.ProviderRef.Name, metav1.GetOptions{})
		if err != nil {
			log.Error(err, fmt.Sprintf("unable to fetch provider %s/%s", cbTemplateNamespace, valuesFrom.ProviderRef.Name), "gvr", gvr)
			return nil, err
		}
		newProvider = provider.DeepCopy()
		unstructured.RemoveNestedField(newProvider.Object, "metadata")
		newProvider.SetOwnerReferences([]metav1.OwnerReference{
			{
				APIVersion: clusterapiv1beta1.GroupVersion.String(),
				Kind:       cluster.Kind,
				Name:       cluster.Name,
				UID:        cluster.UID,
			},
		})
		// Add cluster and package labels to cloned providers
		providerLabels := newProvider.GetLabels()
		if providerLabels == nil {
			newProvider.SetLabels(map[string]string{
				// A package refName could contain characters that K8S does not like as a label value.
				// For example, kapp-controller.tanzu.vmware.com.0.30.0+vmware.1-tkg.1 is a
				// valid package refName, but it contains "+" that K8S complains. We parse the refName by replacing
				// + to ---.
				types.PackageNameLabel: util.ParseStringForLabel(pkg.RefName),
				types.ClusterNameLabel: cluster.Name,
			})
		} else {
			providerLabels[types.PackageNameLabel] = util.ParseStringForLabel(pkg.RefName)
			providerLabels[types.ClusterNameLabel] = cluster.Name
		}

		newProvider.SetName(fmt.Sprintf("%s-%s-package", cluster.Name, pkgRefName))
		newProvider.SetNamespace(bootstrap.Namespace)
		log.Info(fmt.Sprintf("cloning provider %s/%s to namespace %s", cbTemplateNamespace, newProvider.GetName(), bootstrap.Namespace), "gvr", gvr)
		// newProvider and createdOrUpdatedProvider are different. The newProvider is the one we want apiserver to accept,
		// however createdOrUpdatedProvider is the actual object pointer that apiserver has already created with several managed fields,
		// The intent of this function is to return the actual object pointer that apiserver has created, so we should
		// return createdOrUpdatedProvider instead of newProvider. Otherwise, when the caller wants to make changes to the
		// created provider objects, there will be errors, i.e., [invalid: metadata.resourceVersion: Invalid value: 0x0: must be specified for an update]
		createdOrUpdatedProvider, err = r.dynamicClient.Resource(*gvr).Namespace(bootstrap.Namespace).Create(r.context, newProvider, metav1.CreateOptions{})
		if err != nil {
			// There are possibilities that current reconciliation loop fails due to various reasons, and during next reconciliation
			// loop, it is possible that the provider resource has been created. In this case, we want to run update/patch.
			if apierrors.IsAlreadyExists(err) {
				// Setting the resource version is because it's been removed at L984 with 'unstructured.RemoveNestedField(newProvider.Object, "metadata")'
				// apiserver requires that field to be present for concurrency control.
				newProvider.SetResourceVersion(provider.GetResourceVersion())
				createdOrUpdatedProvider, err = r.dynamicClient.Resource(*gvr).Namespace(bootstrap.Namespace).Update(r.context, newProvider, metav1.UpdateOptions{})
				if err != nil {
					log.Info(fmt.Sprintf("unable to updated provider %s/%s", newProvider.GetNamespace(), newProvider.GetName()), "gvr", gvr)
					return nil, err
				}
			} else {
				log.Error(err, fmt.Sprintf("unable to clone provider %s/%s", newProvider.GetNamespace(), newProvider.GetName()), "gvr", gvr)
				return nil, err
			}
		}

		valuesFrom.ProviderRef.Name = createdOrUpdatedProvider.GetName()
		log.Info(fmt.Sprintf("cloned provider %s/%s to namespace %s", createdOrUpdatedProvider.GetNamespace(), createdOrUpdatedProvider.GetName(), bootstrap.Namespace), "gvr", gvr)
	}

	return createdOrUpdatedProvider, nil
}

// watchProvider will set a watch on the Type indicated by providerRef if not already watching
func (r *ClusterBootstrapReconciler) watchProvider(providerRef *corev1.TypedLocalObjectReference, namespace string, log logr.Logger) error {
	if providerRef == nil {
		return nil
	}
	groupKind := fmt.Sprintf("%s/%s", *providerRef.APIGroup, providerRef.Kind)
	if _, ok := r.providerWatches[groupKind]; ok {
		// nothing to do, already watching
		return nil
	}

	gvr, err := r.getGVR(schema.GroupKind{Group: *providerRef.APIGroup, Kind: providerRef.Kind})
	if err != nil {
		log.Error(err, "failed to getGVR")
		return err
	}
	provider, err := r.dynamicClient.Resource(*gvr).Namespace(namespace).Get(r.context, providerRef.Name, metav1.GetOptions{})
	if err != nil {
		log.Error(err, "Error getting provider object", "provider", provider, "gvr", gvr)
		return err
	}
	r.providerWatches[groupKind] = provider

	log.Info("setting watch on provider", "provider", provider)
	// controller-runtime doesn't have an API to remove watches, would the controller panic if a CRD was deleted?
	return r.controller.Watch(&source.Kind{Type: provider},
		handler.EnqueueRequestsFromMapFunc(r.ProviderToClusters),
		predicate.Funcs{
			CreateFunc:  func(e event.CreateEvent) bool { return true },
			UpdateFunc:  func(e event.UpdateEvent) bool { return true },
			DeleteFunc:  func(e event.DeleteEvent) bool { return false },
			GenericFunc: func(e event.GenericEvent) bool { return true },
		},
	)
}

// GetDataValueSecretNameFromBootstrapPackage attempts to get the data value secret name associated with a ClusterBootstrapPackage.
// Users have three ways to provide the data values for a ClusterBootstrapPackage: [Inline, SecretRef, ProviderRef], or
// leave ClusterBootstrapPackage.ValuesFrom field as nil. If data values are provided by ProviderRef, the corresponding
// controller needs to generate the secret object.
//
// Returns:
// - string: The secret name which references to the Secret CR on mgmt cluster under a particular cluster namespace.
// - error: whether there is error when getting the secret name.
func (r *ClusterBootstrapReconciler) GetDataValueSecretNameFromBootstrapPackage(cbPkg *runtanzuv1alpha3.ClusterBootstrapPackage, clusterNamespace string) (string, error) {
	// When valuesFrom is nil, we interpret it as no data values are needed for the package installation.
	if cbPkg.ValuesFrom == nil {
		r.Log.Info(fmt.Sprintf("no data values are provided to the ClusterBootstrapPackage.ValuesFrom field. ClusterBootstrapPackage.RefName: %s", cbPkg.RefName))
		return "", nil
	}

	// TODO: Handle inline valueFrom. https://github.com/vmware-tanzu/tanzu-framework/issues/1694

	if cbPkg.ValuesFrom.SecretRef != "" {
		return cbPkg.ValuesFrom.SecretRef, nil
	}

	if cbPkg.ValuesFrom.ProviderRef != nil {
		gvr, err := r.getGVR(schema.GroupKind{Group: *cbPkg.ValuesFrom.ProviderRef.APIGroup, Kind: cbPkg.ValuesFrom.ProviderRef.Kind})
		if err != nil {
			r.Log.Error(err, "unable to get GVR")
			return "", err
		}
		provider, err := r.dynamicClient.Resource(*gvr).Namespace(clusterNamespace).Get(r.context, cbPkg.ValuesFrom.ProviderRef.Name, metav1.GetOptions{}, "status")
		if err != nil {
			r.Log.Error(err, "unable to fetch provider", "GVR", gvr)
			return "", err
		}
		secretName, found, err := unstructured.NestedString(provider.UnstructuredContent(), "status", "secretRef")
		if err != nil {
			r.Log.Error(err, "unable to fetch secretRef in provider", "GVR", gvr)
			return "", err
		}
		if !found {
			// In this case, we expect the secretRef to be present under status subresource and its value gets updated by
			// the corresponding controller. However, the config controller might not create the secret in time.
			r.Log.Info("provider status does not have secretRef", "GVR", gvr)
			return "", nil
		}
		return secretName, nil
	}

	// When valuesFrom is not nil, but either valuesFrom.Inline, valuesFrom.SecretRef, or valuesFrom.providerRef is empty or nil,
	// we interpret it as the data value secret for that package has not been available yet. One of those three fields needs
	// to be provided either by the user or the controller.
	err := fmt.Errorf("unable to get the data value secret name from the ClusterBootstrapPackage.ValuesFrom field. "+
		"ClusterBootstrapPackage.RefName: %s. One of the fields under ClusterBootstrapPackage.ValuesFrom is empty or nil",
		cbPkg.RefName)
	// The message in err object has sufficient information
	r.Log.Error(err, "")
	return "", err
}

func (r *ClusterBootstrapReconciler) getCNIForClusterBootstrap(
	clusterBootstrapTemplate *runtanzuv1alpha3.ClusterBootstrapTemplate,
	cluster *clusterapiv1beta1.Cluster,
	log logr.Logger) (*runtanzuv1alpha3.ClusterBootstrapPackage, error) {

	var clusterBootstrapPackage *runtanzuv1alpha3.ClusterBootstrapPackage

	selectedCNI, err := util.ParseClusterVariableString(cluster, r.Config.CNISelectionClusterVariableName)
	if err != nil {
		log.Error(err, "Error parsing cluster variable value for the CNI selection")
		return nil, err
	}
	if selectedCNI != "" {
		var getPkgMetadataErrs []error
		foundCNI := false
		for _, cni := range clusterBootstrapTemplate.Spec.CNIs {
			// Package should be available in cluster namespace
			pkgRefName, _, getPkgMetadataErr := util.GetPackageMetadata(r.context, r.liveClient, cni.RefName, cluster.Namespace)
			if getPkgMetadataErr != nil {
				getPkgMetadataErrs = append(getPkgMetadataErrs, getPkgMetadataErr)
			}
			// selectedCNI string from cluster.Topology.Variables could be any arbitrary string. I.e., antrea or antrea.tanzu.vmware.com
			// A Carvel package refName could follow a different naming convention.
			// When comparing the selectedCNI string with
			if strings.HasPrefix(pkgRefName, selectedCNI) {
				clusterBootstrapPackage = cni
				foundCNI = true
				break
			}
		}
		if len(getPkgMetadataErrs) != 0 && !foundCNI {
			return nil, errorsutil.NewAggregate(getPkgMetadataErrs)
		}
	} else {
		if len(clusterBootstrapTemplate.Spec.CNIs) > 0 {
			clusterBootstrapPackage = clusterBootstrapTemplate.Spec.CNIs[0]
		} else {
			return nil, errors.New("no CNI was specified in the ClusterClass or in ClusterBootstrap.Spec.CNIs")
		}
	}
	return clusterBootstrapPackage, nil
}

func (r *ClusterBootstrapReconciler) watchesForClusterBootstrap() []ClusterBootstrapWatchInputs {
	return []ClusterBootstrapWatchInputs{
		{
			&source.Kind{Type: &runtanzuv1alpha3.TanzuKubernetesRelease{}},
			handler.EnqueueRequestsFromMapFunc(r.TKRToClusters),
		},
		{
			&source.Kind{Type: &corev1.Secret{}},
			handler.EnqueueRequestsFromMapFunc(r.SecretsToClusters),
		},
	}
}

func (r *ClusterBootstrapReconciler) reconcileClusterProxyAndNetworkSettings(cluster *clusterapiv1beta1.Cluster,
	log logr.Logger) error {

	// use patchHelper to auto detect if there is diff in Cluster CR when performing update
	patchHelper, err := clusterapipatchutil.NewHelper(cluster, r.Client)
	if err != nil {
		return err
	}

	// We want the reconciliation to continue even if there are errors in getting proxy settings
	// Log an error and proceed with defaulting to empty string
	// Individual config controllers are responsible for validating the info provided
	HTTPProxy, err := util.ParseClusterVariableString(cluster, r.Config.HTTPProxyClusterClassVarName)
	if err != nil {
		log.Error(err, "unable to fetch cluster HTTP proxy setting, defaulting to empty")
	}
	HTTPSProxy, err := util.ParseClusterVariableString(cluster, r.Config.HTTPSProxyClusterClassVarName)
	if err != nil {
		log.Error(err, "unable to fetch cluster HTTPS proxy setting, defaulting to empty")
	}
	NoProxy, err := util.ParseClusterVariableString(cluster, r.Config.NoProxyClusterClassVarName)
	if err != nil {
		log.Error(err, "unable to fetch cluster no-proxy setting, defaulting to empty")
	}
	ProxyCACert, err := util.ParseClusterVariableString(cluster, r.Config.ProxyCACertClusterClassVarName)
	if err != nil {
		log.Error(err, "unable to fetch cluster proxy CA certificate, defaulting to empty")
	}
	IPFamily, err := util.ParseClusterVariableString(cluster, r.Config.IPFamilyClusterClassVarName)
	if err != nil {
		log.Error(err, "unable to fetch cluster IP family, defaulting to empty")
	}

	if cluster.Annotations == nil {
		cluster.Annotations = map[string]string{}
	}
	cluster.Annotations[types.HTTPProxyConfigAnnotation] = HTTPProxy
	cluster.Annotations[types.HTTPSProxyConfigAnnotation] = HTTPSProxy
	cluster.Annotations[types.NoProxyConfigAnnotation] = NoProxy
	cluster.Annotations[types.ProxyCACertConfigAnnotation] = ProxyCACert
	cluster.Annotations[types.IPFamilyConfigAnnotation] = IPFamily

	log.Info("setting proxy and network configurations in Cluster annotation", types.HTTPProxyConfigAnnotation, HTTPProxy, types.HTTPSProxyConfigAnnotation, HTTPSProxy, types.NoProxyConfigAnnotation, NoProxy, types.ProxyCACertConfigAnnotation, ProxyCACert, types.IPFamilyConfigAnnotation, IPFamily)

	if err := patchHelper.Patch(r.context, cluster); err != nil {
		log.Error(err, "unable to patch Cluster Annotation")
		return err
	}

	return nil
}
