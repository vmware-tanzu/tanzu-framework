// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"errors"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	cacheddiscovery "k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/pointer"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	clusterapiutil "sigs.k8s.io/cluster-api/util"
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

	"github.com/go-logr/logr"

	kapppkgiv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	kapppkgv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	versions "github.com/vmware-tanzu/carvel-vendir/pkg/vendir/versions/v1alpha1"
	addonconfig "github.com/vmware-tanzu/tanzu-framework/addons/pkg/config"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	addontypes "github.com/vmware-tanzu/tanzu-framework/addons/pkg/types"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	"github.com/vmware-tanzu/tanzu-framework/addons/predicates"
	runtanzuv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

// ClusterBootstrapReconciler reconciles a ClusterBootstrap object
type ClusterBootstrapReconciler struct {
	client.Client
	Log     logr.Logger
	Scheme  *runtime.Scheme
	Config  *addonconfig.Config
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
}

// NewClusterBootstrapReconciler returns a reconciler for ClusterBootstrap
func NewClusterBootstrapReconciler(c client.Client, log logr.Logger, scheme *runtime.Scheme, config *addonconfig.Config) *ClusterBootstrapReconciler {
	return &ClusterBootstrapReconciler{
		Client: c,
		Log:    log,
		Scheme: scheme,
		Config: config,
	}
}

// +kubebuilder:rbac:groups=run.tanzu.vmware.com,resources=clusterBootstraps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=run.tanzu.vmware.com,resources=clusterBootstraps/status,verbs=get;update;patch

// SetupWithManager performs the setup actions for an ClusterBootstrap controller, using the passed in mgr.
func (r *ClusterBootstrapReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	ctrlr, err := ctrl.NewControllerManagedBy(mgr).
		For(&clusterapiv1beta1.Cluster{}).
		Watches(
			&source.Kind{Type: &runtanzuv1alpha3.TanzuKubernetesRelease{}},
			handler.EnqueueRequestsFromMapFunc(r.TKRToClusters),
			builder.WithPredicates(
				predicates.TKR(r.Log),
			),
		).
		Watches(
			&source.Kind{Type: &runtanzuv1alpha3.ClusterBootstrap{}},
			handler.EnqueueRequestsFromMapFunc(r.ClusterBootstrapToClusters),
			builder.WithPredicates(
				predicates.TKR(r.Log),
			),
		).
		WithOptions(options).
		WithEventFilter(clusterApiPredicates.ResourceNotPaused(r.Log)).
		Build(r)
	if err != nil {
		r.Log.Error(err, "Error creating an addon controller")
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

	tkrName := util.GetTKRNameForCluster(r.context, r.Client, cluster)
	if tkrName == "" {
		log.Info("cluster does not have an associated TKR")
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
		return ctrl.Result{}, err
	}
	if clusterBootstrap == nil {
		return ctrl.Result{}, nil
	}

	remoteClient, err := util.GetClusterClient(r.context, r.Client, r.Scheme, clusterapiutil.ObjectKey(cluster))
	if err != nil {
		log.Error(err, "Error getting remote cluster client")
		return ctrl.Result{}, err
	}

	// Create a PackageInstall CR under the cluster namespace for deploying the kapp-controller on the remote cluster.
	// We need kapp-controller to be deployed prior to CNI, CPI, CSI.
	if err := r.createOrPatchKappPackageInstall(clusterBootstrap, cluster); err != nil {
		// Return error if kapp-controller fails to be deployed, let reconciler try again
		return ctrl.Result{}, err
	}

	// Create the ServiceAccount on remote cluster, so it could be referenced in PackageInstall CR for kapp-controller
	// reconciliation.
	if _, err := r.createOrPatchAddonServiceAccountOnRemote(cluster, remoteClient); err != nil {
		return ctrl.Result{}, err
	}

	// Create the ClusterRole on remote cluster, and bind it to the ServiceAccount created in above. kapp-controller
	// reconciliation needs privileges.
	if err := r.createOrPatchAddonRoleOnRemote(cluster, remoteClient); err != nil {
		return ctrl.Result{}, err
	}

	// Create or patch the resources for CNI, CPI, CSI to be running on the remote cluster.
	// Those resources include Package CR, data value Secret, PackageInstall CR.
	var corePackages []*runtanzuv1alpha3.ClusterBootstrapPackage
	corePackages = append(corePackages, clusterBootstrap.Spec.CNI, clusterBootstrap.Spec.CPI, clusterBootstrap.Spec.CSI)
	for _, corePackage := range corePackages {
		// The following nil check is redundant, we do not expect CNI, CPI, CSI to be nil and webhook should handle
		// the validations against those fields. Having this nil check is mainly to allow local envtest could run when
		// any above component is missing.
		if corePackage == nil {
			continue
		}
		// There are different ways to have all the resources created or patched on remote cluster. Current solution is
		// to handle packages in sequence order. I.e., Create all resources for CNI first, and then CPI, CSI. It is also
		// possible to create all resources in a different order or in parallel. We will consider to use goroutines to create
		// all resources in parallel on remote cluster if there is performance issue from sequential ordering.
		if err := r.createOrPatchAddonResourcesOnRemote(cluster, corePackage, remoteClient); err != nil {
			// For core packages, we require all their creation or patching to succeed, so if error happens against any of the
			// package, we return error and let the reconciler retry again.
			log.Error(err, fmt.Sprintf("unable to create or patch all the required resources for %s on cluster: %s/%s",
				corePackage.RefName, cluster.Namespace, cluster.Name))
			return ctrl.Result{}, err
		}
	}

	// Create or patch the resources for additionalPackages
	for _, additionalPkg := range clusterBootstrap.Spec.AdditionalPackages {
		if err := r.createOrPatchAddonResourcesOnRemote(cluster, additionalPkg, remoteClient); err != nil {
			// Logging has been handled in createOrPatchAddonResourcesOnRemote()
			return ctrl.Result{}, err
		}
		// set watches on provider objects in additional packages if not already set
		if additionalPkg.ValuesFrom.ProviderRef != nil {
			if err := r.watchProvider(additionalPkg.ValuesFrom.ProviderRef, clusterBootstrap.Namespace, log); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

// createOrPatchTanzuClusterBootstrapFromTemplate will get, clone or update a TanzuClusterBootstrap associated with a cluster
// all linked secret refs and object refs are cloned into the same namespace as TanzuClusterBootstrap
func (r *ClusterBootstrapReconciler) createOrPatchClusterBootstrapFromTemplate(cluster *clusterapiv1beta1.Cluster,
	log logr.Logger) (*runtanzuv1alpha3.ClusterBootstrap, error) {

	tkrName := util.GetTKRNameForCluster(r.context, r.Client, cluster)
	if tkrName == "" {
		log.Info("cluster does not have an associated TKR")
		return nil, nil
	}

	clusterBootstrapTemplate := &runtanzuv1alpha3.ClusterBootstrapTemplate{}
	key := client.ObjectKey{Namespace: r.Config.AddonNamespace, Name: tkrName}
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
		secrets, providers, err := r.cloneSecretsAndProviders(cluster, clusterBootstrap, clusterBootstrapTemplate.Namespace, log)
		if err != nil {
			r.Log.Error(err, "unable to clone secrets, providers")
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
			r.Log.Error(err, "unable to ensure ownerref on created secrets and providers", "clusterBootstrap", clusterBootstrap)
			return nil, err
		}

		clusterBootstrap.Status.ResolvedTKR = tkrName
		if err := r.Status().Update(r.context, clusterBootstrap); err != nil {
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
// mgmt cluster. The kapp-controller runs on mgmt cluster reconciles the PackageInstall CR and creates kapp-controller resources
// on remote workload cluster. This is required for a workload cluster and its corresponding package installations to be functional.
func (r *ClusterBootstrapReconciler) createOrPatchKappPackageInstall(clusterBootstrap *runtanzuv1alpha3.ClusterBootstrap, cluster *clusterapiv1beta1.Cluster) error {
	pkgi := &kapppkgiv1alpha1.PackageInstall{
		ObjectMeta: metav1.ObjectMeta{
			// The legacy addons controller uses <cluster name>-<addon name> convention for naming the PackageInstall CR.
			// https://github.com/vmware-tanzu/tanzu-framework/blob/main/addons/controllers/package_reconciler.go#L195.
			// util.GeneratePackageInstallName() follows the same pattern.
			Name: util.GeneratePackageInstallName(cluster.Name, util.GetPackageShortName(clusterBootstrap.Spec.Kapp.RefName)),
			// kapp-controller PackageInstall CR is installed under the same namespace as tanzuClusterBootstrap. The namespace
			// is also the same as where the cluster belongs.
			Namespace: clusterBootstrap.Namespace,
		},
	}

	pkgiMutateFn := func() error {
		// TODO: Followup on the following fields, only add them if needed.
		// https://github.com/vmware-tanzu/tanzu-framework/issues/1677
		// if ipkg.ObjectMeta.Annotations == nil {
		//	 ipkg.ObjectMeta.Annotations = make(map[string]string)
		// }
		// ipkg.ObjectMeta.Annotations[addontypes.YttMarkerAnnotation] = ""
		// ipkg.Spec.SyncPeriod = &metav1.Duration{Duration: r.Config.AppSyncPeriod}
		// ipkg.Spec.ServiceAccountName = r.Config.AddonServiceAccount
		pkgi.Spec.PackageRef = &kapppkgiv1alpha1.PackageRef{
			RefName: clusterBootstrap.Spec.Kapp.RefName,
			VersionSelection: &versions.VersionSelectionSemver{
				Prereleases: &versions.VersionSelectionSemverPrereleases{},
			},
		}
		pkgi.Spec.Values = []kapppkgiv1alpha1.PackageInstallValues{
			{SecretRef: &kapppkgiv1alpha1.PackageInstallValuesSecretRef{
				// The secret name could also be fetched from kappConfig.Spec.Status.SecretRef. However, the naming convention
				// of that field is the same as the follow. To simplify the implementation and reduce the call to r.Client.Get()
				// on kappConfig CR, use the util function to construct the secret name.
				// TODO: Follow up with reviewer to see if this is feasible
				Name: util.GenerateDataValueSecretName(cluster.Name, constants.KappControllerAddonName)},
			},
		}

		return nil
	}

	_, err := controllerutil.CreateOrPatch(r.context, r.Client, pkgi, pkgiMutateFn)
	if err != nil {
		r.Log.Error(err, fmt.Sprintf("unable to create or patch %s PackageInstall resource for cluster: %s",
			constants.KappControllerAddonName, cluster.Name))
		return err
	}

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
	if err = r.Client.Get(r.context, key, localPackage); err != nil {
		// If there is an error to get the Carvel Package CR from local cluster, nothing needs to be created/cloned on remote.
		// Let the reconciler try again.
		r.Log.Error(err, fmt.Sprintf("unable to get package %s, nothing needs to be created on cluster %s/%s",
			key.String(), cluster.Namespace, cluster.Name))
		return nil, err
	}
	remotePackage := &kapppkgv1alpha1.Package{}
	remotePackage.SetName(localPackage.Name)
	// The Package CR on remote cluster needs to be under tkg-system namespace
	remotePackage.SetNamespace(r.Config.AddonNamespace)
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
	cbPkg *runtanzuv1alpha3.ClusterBootstrapPackage, clusterClient client.Client) (*kapppkgiv1alpha1.PackageInstall, error) {

	// Create PackageInstall CRs on the remote workload cluster, kapp-controller will take care of reconciling them
	remotePkgi := &kapppkgiv1alpha1.PackageInstall{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GeneratePackageInstallName(cluster.Name, util.GetPackageShortName(cbPkg.RefName)),
			Namespace: r.Config.AddonNamespace,
		},
	}
	remotePackageName := cbPkg.RefName
	_, err := controllerutil.CreateOrPatch(r.context, clusterClient, remotePkgi, func() error {
		remotePkgi.Spec.ServiceAccountName = r.Config.AddonServiceAccount
		remotePkgi.Spec.PackageRef = &kapppkgiv1alpha1.PackageRef{
			RefName: remotePackageName,
			VersionSelection: &versions.VersionSelectionSemver{
				Prereleases: &versions.VersionSelectionSemverPrereleases{},
			},
		}
		remotePkgi.Spec.Values = []kapppkgiv1alpha1.PackageInstallValues{
			{SecretRef: &kapppkgiv1alpha1.PackageInstallValuesSecretRef{
				Name: remotePackageName},
			},
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

// createOrPatchAddonServiceAccountOnRemote creates or patches the addon ServiceAccount on remote cluster.
// The ServiceAccount will be referenced by the PackageInstall CR, so that kapp-controller on remote cluster could consume
// for PackageInstall reconciliation.
func (r *ClusterBootstrapReconciler) createOrPatchAddonServiceAccountOnRemote(cluster *clusterapiv1beta1.Cluster, clusterClient client.Client) (*corev1.ServiceAccount, error) {
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Config.AddonServiceAccount,
			Namespace: r.Config.AddonNamespace,
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

// createOrPatchAddonRoleOnRemote creates or patches the ClusterRole, ClusterRoleBinding on remote cluster.
// The ClusterRole is bound to the ServiceAccount which is referenced by PackageInstall CR, so that kapp-controller on remote
// cluster could have privileges to lifecycle manage package resources.
func (r *ClusterBootstrapReconciler) createOrPatchAddonRoleOnRemote(cluster *clusterapiv1beta1.Cluster, clusterClient client.Client) error {
	addonRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: r.Config.AddonClusterRole,
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
			Name: r.Config.AddonClusterRoleBinding,
		},
	}
	if _, err := controllerutil.CreateOrPatch(r.context, clusterClient, addonRoleBinding, func() error {
		addonRoleBinding.Subjects = []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      r.Config.AddonServiceAccount,
				Namespace: r.Config.AddonNamespace,
			},
		}
		addonRoleBinding.RoleRef = rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     r.Config.AddonClusterRole,
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
	remoteSecret, err := r.createOrPatchPackageInstallSecretOnRemote(cluster, cbPkg, clusterClient, r.Log)
	// We expect there is NO error to create or patch the secret used for PackageInstall in a cluster.
	if err != nil {
		// Logging has been handled by createOrPatchPackageInstallSecret() already
		return err
	}
	if remoteSecret == nil {
		// The nil secret happens when tcbPkg.ValuesFrom is a ProviderRef and no data value secret for that provider is found
		// on local cluster. This error occurs when handling additionalPackages. But it does not hurt to check nil here,
		// and return error which let the reconciler handles the retry.
		// The detailed logging has been handled by createOrPatchPackageInstallSecret().
		return fmt.Errorf("unable to create or patch the data value secret on cluster: %s/%s", cluster.Namespace, cluster.Name)
	}
	r.Log.Info(fmt.Sprintf("created secret for package %s on cluster %s/%s", remotePackage.Name, cluster.Namespace,
		cluster.Name))

	pkgi, err := r.createOrPatchPackageInstallOnRemote(cluster, cbPkg, clusterClient)
	if err != nil {
		return err
	}
	r.Log.Info(fmt.Sprintf("created the PackageInstall CR %s/%s on cluster %s/%s",
		pkgi.Namespace, pkgi.Name, cluster.Namespace, cluster.Name))

	return nil
}

// createOrPatchPackageInstallSecret creates or patches or the secret used for PackageInstall in a cluster
func (r *ClusterBootstrapReconciler) createOrPatchPackageInstallSecretOnRemote(cluster *clusterapiv1beta1.Cluster,
	pkg *runtanzuv1alpha3.ClusterBootstrapPackage, clusterClient client.Client, log logr.Logger) (*corev1.Secret, error) {

	secret := &corev1.Secret{}

	if pkg.ValuesFrom.SecretRef != "" {
		key := client.ObjectKey{Namespace: cluster.Namespace, Name: pkg.ValuesFrom.SecretRef}
		if err := r.Get(r.context, key, secret); err != nil {
			log.Error(err, "unable to fetch secret", "objectkey", key)
			return nil, err
		}
	}

	if pkg.ValuesFrom.ProviderRef != nil {
		gvr, err := r.getGVR(schema.GroupKind{Group: *pkg.ValuesFrom.ProviderRef.APIGroup, Kind: pkg.ValuesFrom.ProviderRef.Kind})
		if err != nil {
			log.Error(err, "failed to getGVR")
			return nil, err
		}
		provider, err := r.dynamicClient.Resource(*gvr).Namespace(cluster.Namespace).Get(r.context, pkg.ValuesFrom.ProviderRef.Name, metav1.GetOptions{}, "status")

		if err != nil {
			log.Error(err, "unable to fetch provider", "provider", pkg.ValuesFrom.ProviderRef, "gvr", gvr)
			return nil, err
		}
		secretName, found, err := unstructured.NestedString(provider.UnstructuredContent(), "status", "secretRef")
		if err != nil {
			log.Error(err, "unable to fetch secretRef in provider", "provider", provider)
			return nil, err
		}
		if !found {
			log.Info("provider status does not have secretRef", "provider", provider)
			return nil, nil
		}
		key := client.ObjectKey{Namespace: cluster.Namespace, Name: secretName}
		if err := r.Get(r.context, key, secret); err != nil {
			log.Error(err, "unable to fetch secret", "objectkey", key)
			return nil, err
		}
	}

	dataValuesSecret := &corev1.Secret{}
	dataValuesSecret.Name = util.GenerateDataValueSecretName(cluster.Name, util.GetPackageShortName(pkg.RefName))
	// The secret will be created or patched under tkg-system namespace on remote cluster
	dataValuesSecret.Namespace = r.Config.AddonNamespace
	dataValuesSecret.Type = corev1.SecretTypeOpaque

	dataValuesSecretMutateFn := func() error {
		dataValuesSecret.Data = map[string][]byte{}
		for k, v := range secret.Data {
			dataValuesSecret.Data[k] = v
		}
		return nil
	}

	_, err := controllerutil.CreateOrPatch(r.context, clusterClient, dataValuesSecret, dataValuesSecretMutateFn)
	if err != nil {
		log.Error(err, "Error creating or patching addon data values secret")
		return nil, err
	}
	return dataValuesSecret, nil
}

// cloneSecretsAndProviders clones linked secrets and providers into the same namespace as clusterBootstrap
func (r *ClusterBootstrapReconciler) cloneSecretsAndProviders(cluster *clusterapiv1beta1.Cluster, bootstrap *runtanzuv1alpha3.ClusterBootstrap,
	templateNS string, log logr.Logger) ([]*corev1.Secret, []*unstructured.Unstructured, error) {

	var createdProviders []*unstructured.Unstructured
	var createdSecrets []*corev1.Secret

	packages := append([]*runtanzuv1alpha3.ClusterBootstrapPackage{
		bootstrap.Spec.CNI,
		bootstrap.Spec.CPI,
		bootstrap.Spec.CSI,
		bootstrap.Spec.Kapp,
	}, bootstrap.Spec.AdditionalPackages...)

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
	pkg *runtanzuv1alpha3.ClusterBootstrapPackage, templateNS string, log logr.Logger) (*corev1.Secret, *unstructured.Unstructured, error) {

	if pkg.ValuesFrom == nil {
		return nil, nil, nil
	}
	if pkg.ValuesFrom.SecretRef != "" {
		secret, err := r.updateValuesFromSecret(cluster, bootstrap, pkg, templateNS, log)
		if err != nil {
			return nil, nil, err
		}
		return secret, nil, nil
	}

	if pkg.ValuesFrom.ProviderRef != nil {
		provider, err := r.updateValuesFromProvider(cluster, bootstrap, pkg, templateNS, log)
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
			return err
		}
	}
	for _, provider := range providers {
		provider.SetOwnerReferences(clusterapiutil.EnsureOwnerRef(provider.GetOwnerReferences(), *ownerRef))
		gvr, err := r.getGVR(provider.GroupVersionKind().GroupKind())
		if err != nil {
			return err
		}

		_, err = r.dynamicClient.Resource(*gvr).Namespace(provider.GetNamespace()).Update(r.context, provider, metav1.UpdateOptions{})
		if err != nil {
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
	pkg *runtanzuv1alpha3.ClusterBootstrapPackage, templateNS string, log logr.Logger) (*corev1.Secret, error) {

	var newSecret *corev1.Secret
	if pkg.ValuesFrom.SecretRef != "" {
		secret := &corev1.Secret{}
		key := client.ObjectKey{Namespace: templateNS, Name: pkg.ValuesFrom.SecretRef}
		if err := r.Get(r.context, key, secret); err != nil {
			log.Error(err, "unable to fetch secret", "objectkey", key)
			return nil, err
		}
		newSecret = secret.DeepCopy()
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

		newSecret.Labels[addontypes.PackageNameLabel] = util.ParseStringForLabel(pkg.RefName)
		newSecret.Labels[addontypes.ClusterNameLabel] = cluster.Name

		newSecret.Name = fmt.Sprintf("%s-%s-package", cluster.Name, util.GetPackageShortName(pkg.RefName))
		newSecret.Namespace = bootstrap.Namespace
		if err := r.Create(r.context, newSecret); err != nil {
			return nil, err
		}
		pkg.ValuesFrom.SecretRef = newSecret.Name
	}

	return newSecret, nil
}

// updateValuesFromProvider updates providerRef in valuesFrom
func (r *ClusterBootstrapReconciler) updateValuesFromProvider(cluster *clusterapiv1beta1.Cluster, bootstrap *runtanzuv1alpha3.ClusterBootstrap,
	pkg *runtanzuv1alpha3.ClusterBootstrapPackage, templateNS string, log logr.Logger) (*unstructured.Unstructured, error) {

	var newProvider *unstructured.Unstructured
	valuesFrom := pkg.ValuesFrom
	if valuesFrom.ProviderRef != nil {
		gvr, err := r.getGVR(schema.GroupKind{Group: *valuesFrom.ProviderRef.APIGroup, Kind: valuesFrom.ProviderRef.Kind})
		if err != nil {
			log.Error(err, "failed to getGVR")
			return nil, err
		}
		provider, err := r.dynamicClient.Resource(*gvr).Namespace(templateNS).Get(r.context, valuesFrom.ProviderRef.Name, metav1.GetOptions{})
		if err != nil {
			log.Error(err, fmt.Sprintf("unable to fetch provider %s/%s", templateNS, valuesFrom.ProviderRef.Name), "gvr", gvr)
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
				addontypes.PackageNameLabel: util.ParseStringForLabel(pkg.RefName),
				addontypes.ClusterNameLabel: cluster.Name,
			})
		} else {
			providerLabels[addontypes.PackageNameLabel] = util.ParseStringForLabel(pkg.RefName)
			providerLabels[addontypes.ClusterNameLabel] = cluster.Name
		}

		newProvider.SetName(fmt.Sprintf("%s-%s-package", cluster.Name, util.GetPackageShortName(pkg.RefName)))
		log.Info(fmt.Sprintf("cloning provider %s/%s to namespace %s", newProvider.GetNamespace(), newProvider.GetName(), bootstrap.Namespace), "gvr", gvr)
		newProvider, err = r.dynamicClient.Resource(*gvr).Namespace(bootstrap.Namespace).Create(r.context, newProvider, metav1.CreateOptions{})
		if err != nil {
			// There are possibilities that current reconciliation loop fails due to various reasons, and during next reconciliation
			// loop, it is possible that the provider resource has been created. In this case, we want to run update/patch.
			if apierrors.IsAlreadyExists(err) {
				newProvider, err = r.dynamicClient.Resource(*gvr).Namespace(bootstrap.Namespace).Update(r.context, newProvider, metav1.UpdateOptions{})
				if err != nil {
					log.Info(fmt.Sprintf("updated provider %s/%s in namespace %s", newProvider.GetNamespace(), newProvider.GetName(), bootstrap.Namespace), "gvr", gvr)
					return newProvider, nil
				}
			}
			log.Error(err, "unable to clone provider", "gvr", gvr)
			return nil, err
		}

		valuesFrom.ProviderRef.Name = newProvider.GetName()
		log.Info(fmt.Sprintf("cloned provider %s/%s to namespace %s", newProvider.GetNamespace(), newProvider.GetName(), bootstrap.Namespace), "gvr", gvr)
	}

	return newProvider, nil
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
