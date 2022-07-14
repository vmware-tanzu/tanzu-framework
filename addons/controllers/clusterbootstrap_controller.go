// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/utils/pointer"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	clusterapiutil "sigs.k8s.io/cluster-api/util"
	clusterapipatchutil "sigs.k8s.io/cluster-api/util/patch"
	clusterApiPredicates "sigs.k8s.io/cluster-api/util/predicates"
	secretutil "sigs.k8s.io/cluster-api/util/secret"
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
	addontypes "github.com/vmware-tanzu/tanzu-framework/addons/pkg/types"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util/clusterbootstrapclone"
	"github.com/vmware-tanzu/tanzu-framework/addons/predicates"
	runtanzuv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	tkgconstants "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	tkrconstants "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/constants"
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
	providerWatches              map[string]client.Object
	aggregatedAPIResourcesClient client.Client
	// helper for looking up api-resources and getting preferred versions
	gvrHelper util.GVRHelper
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

	clientset := kubernetes.NewForConfigOrDie(mgr.GetConfig())
	r.gvrHelper = util.NewGVRHelper(ctx, clientset.DiscoveryClient)

	r.aggregatedAPIResourcesClient, err = client.New(mgr.GetConfig(), client.Options{Scheme: mgr.GetScheme()})
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
	tkrName := util.GetClusterLabel(cluster.Labels, constants.TKRLabelClassyClusters)
	if tkrName == "" {
		return ctrl.Result{}, nil
	}

	tkr, err := util.GetTKRByNameV1Alpha3(r.context, r.Client, tkrName)
	if err != nil {
		log.Error(err, "unable to fetch TKR object", "name", tkrName)
		return ctrl.Result{}, err
	}

	// if tkr is not found, should not requeue for the reconciliation
	if tkr == nil {
		log.Info("TKR object not found", "name", tkrName)
		return ctrl.Result{}, nil
	}

	if _, labelFound := tkr.Labels[constants.TKRLabelLegacyClusters]; labelFound {
		log.Info("Skipping reconciling due to tkr label", "name", tkrName, "label", constants.TKRLabelLegacyClusters)
		return ctrl.Result{}, nil
	}

	log.Info("Reconciling cluster")

	// if deletion timestamp is set, handle cluster deletion
	if !cluster.GetDeletionTimestamp().IsZero() {
		return r.reconcileDelete(cluster, log)
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

	// Create a PackageInstall CR under the cluster namespace for deploying the kapp-controller on the remote cluster.
	// We need kapp-controller to be deployed prior to CNI, CPI, CSI. This will be a no-op if the cluster object is mgmt
	// cluster.
	if err := r.createOrPatchKappPackageInstall(clusterBootstrap, cluster); err != nil {
		// Return error if kapp-controller fails to be deployed, let reconciler try again
		return ctrl.Result{}, err
	}

	remoteClient, err := util.GetClusterClient(r.context, r.Client, r.Scheme, clusterapiutil.ObjectKey(cluster))
	if err != nil {
		return ctrl.Result{RequeueAfter: constants.RequeueAfterDuration}, fmt.Errorf("failed to get remote cluster client: %w", err)
	}

	if err := r.prepareRemoteCluster(cluster, remoteClient); err != nil {
		return ctrl.Result{}, err
	}

	_, err = r.createOrPatchResourcesForCorePackages(cluster, clusterBootstrap, remoteClient, log)
	if err != nil {
		return ctrl.Result{RequeueAfter: constants.RequeueAfterDuration}, err
	}

	_, err = r.createOrPatchResourcesForAdditionalPackages(cluster, clusterBootstrap, remoteClient, log)
	if err != nil {
		return ctrl.Result{RequeueAfter: constants.RequeueAfterDuration}, err
	}

	return ctrl.Result{}, nil
}

func (r *ClusterBootstrapReconciler) createOrPatchResourcesForCorePackages(cluster *clusterapiv1beta1.Cluster,
	clusterBootstrap *runtanzuv1alpha3.ClusterBootstrap,
	remoteClient client.Client,
	log logr.Logger) (ctrl.Result, error) {

	// Create or patch the resources for CNI, CPI, CSI to be running on the remote cluster.
	// Those resources include Package CR, data value Secret, PackageInstall CR.
	var corePackages []*runtanzuv1alpha3.ClusterBootstrapPackage
	corePackages = append(corePackages, clusterBootstrap.Spec.CNI, clusterBootstrap.Spec.CPI, clusterBootstrap.Spec.CSI)

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
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func (r *ClusterBootstrapReconciler) createOrPatchResourcesForAdditionalPackages(cluster *clusterapiv1beta1.Cluster,
	clusterBootstrap *runtanzuv1alpha3.ClusterBootstrap,
	remoteClient client.Client,
	log logr.Logger) (ctrl.Result, error) {

	for _, additionalPkg := range clusterBootstrap.Spec.AdditionalPackages {
		if err := r.createOrPatchAddonResourcesOnRemote(cluster, additionalPkg, remoteClient); err != nil {
			// Logging has been handled in createOrPatchAddonResourcesOnRemote()
			return ctrl.Result{}, err
		}
		// set watches on provider objects in additional packages if not already set
		if additionalPkg.ValuesFrom != nil && additionalPkg.ValuesFrom.ProviderRef != nil {
			if err := r.watchProvider(additionalPkg.ValuesFrom.ProviderRef, clusterBootstrap.Namespace, log); err != nil {
				return ctrl.Result{}, err
			}
		}
	}
	if len(clusterBootstrap.Spec.AdditionalPackages) > 0 { // If we reach this and there are at least one additional package, we need to add finalizer unless it is a management cluster
		_, isManagmentCluster := cluster.Labels[tkrconstants.ManagementClusterRoleLabel]
		if !isManagmentCluster {
			err := r.addFinalizersToClusterResources(cluster, log)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	if err := r.handleClusterUnpause(cluster, clusterBootstrap, log); err != nil {
		log.Error(err, fmt.Sprintf("unable to unpause the cluster: %s/%s", cluster.Namespace, cluster.Name))
		// Need to requeue if unpause is unsuccessful
		return ctrl.Result{Requeue: true}, err
	}

	return ctrl.Result{}, nil
}

func (r *ClusterBootstrapReconciler) addFinalizersToClusterResources(cluster *clusterapiv1beta1.Cluster, log logr.Logger) error {
	err := r.addFinalizer(cluster, cluster.DeepCopy())
	if err != nil {
		log.Error(err, "failed to add finalizer to cluster ")
		return err
	}

	clusterKubeConfigSecret := &corev1.Secret{}
	key := client.ObjectKey{Namespace: cluster.Namespace, Name: secretutil.Name(cluster.Name, secretutil.Kubeconfig)}
	err = r.Client.Get(r.context, key, clusterKubeConfigSecret)
	if err != nil {
		return err
	}
	err = r.addFinalizer(clusterKubeConfigSecret, clusterKubeConfigSecret.DeepCopy())
	if err != nil {
		log.Error(err, "failed to add finalizer to cluster kubeconfig secret")
		return err
	}

	clusterBootstrap := &runtanzuv1alpha3.ClusterBootstrap{}
	err = r.Client.Get(r.context, client.ObjectKeyFromObject(cluster), clusterBootstrap)
	if err != nil {
		return err
	}
	err = r.addFinalizer(clusterBootstrap, clusterBootstrap.DeepCopy())
	if err != nil {
		log.Error(err, "failed to add finalizer to clusterboostrap")
		return err
	}

	return nil
}

func (r *ClusterBootstrapReconciler) addFinalizer(o client.Object, deepCopy client.Object) error {
	controllerutil.AddFinalizer(deepCopy, addontypes.AddonFinalizer)
	return r.Client.Patch(r.context, deepCopy, client.MergeFrom(o))

}

// handleClusterUnpause unpauses the cluster if the cluster pause annotation is set by cluster pause webhook (cluster has "tkg.tanzu.vmware.com/paused" annotation)
func (r *ClusterBootstrapReconciler) handleClusterUnpause(cluster *clusterapiv1beta1.Cluster, clusterBootstrap *runtanzuv1alpha3.ClusterBootstrap, log logr.Logger) error {
	if cluster.Spec.Paused && cluster.Annotations != nil {
		if value, ok := cluster.Annotations[tkgconstants.ClusterPauseLabel]; ok && value == clusterBootstrap.Status.ResolvedTKR {
			patchedCluster := cluster.DeepCopy()
			delete(patchedCluster.Annotations, tkgconstants.ClusterPauseLabel)
			patchedCluster.Spec.Paused = false
			err := r.Client.Patch(r.context, patchedCluster, client.MergeFrom(cluster))
			if err != nil {
				return err
			}
			log.Info(fmt.Sprintf("successfully unpaused the cluster %s/%s after ClusterBootstrap reconciliation", cluster.Namespace, cluster.Name))
		}
	}
	return nil
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
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, err
	}

	if clusterBootstrap.Spec != nil && clusterBootstrap.Spec.Paused {
		// Skip reconcile if ClusterBootstrap is paused
		log.Info("ClusterBootstrap is paused, blocking further processing")
		return nil, nil
	}
	// if found and resolved tkr is the same, return found object as the TKR is supposed to be immutable
	// also preserves any user changes
	if tkrName == clusterBootstrap.Status.ResolvedTKR {
		return clusterBootstrap, nil
	}

	clusterBootstrapHelper := clusterbootstrapclone.NewHelper(
		r.context, r.Client, r.aggregatedAPIResourcesClient, r.dynamicClient, r.gvrHelper, r.Log)
	if clusterBootstrap.UID == "" {
		// When ClusterBootstrap.UID is empty, that means this is the ClusterBootstrap CR about to be created by clusterbootstrap_controller.
		// And clusterBootstrap.Status.ResolvedTKR will be updated accordingly.
		log.Info(fmt.Sprintf("ClusterBootstrap for cluster %s/%s does not exist, creating from template %s/%s",
			cluster.Namespace, cluster.Name, clusterBootstrapTemplate.Namespace, clusterBootstrapTemplate.Name))
		return clusterBootstrapHelper.CreateClusterBootstrapFromTemplate(clusterBootstrapTemplate, cluster, tkrName)
	} else if clusterBootstrap.Status.ResolvedTKR == "" {
		// Possible cases fall into this block:
		// 1. ClusterBootstrap CR has been created by clusterbootstrap_controller in first reconciliation but errored out before clusterBootstrap.Status was set. The clusterbootstrap_controller reconciles again.
		// 2. ClusterBootstrap CR is created by third party(e.g. Tanzu CLI). The clusterbootstrap_controller catches the event and reconciles.
		log.Info(fmt.Sprintf("Handling existing ClusterBootstrap %s/%s", clusterBootstrap.Namespace, clusterBootstrap.Name))
		return clusterBootstrapHelper.HandleExistingClusterBootstrap(clusterBootstrap, cluster, tkrName, r.Config.SystemNamespace)
	}
	// Handle ClusterBootstrap update when TKR version of the cluster is upgraded
	if tkrName != clusterBootstrap.Status.ResolvedTKR {
		log.Info(fmt.Sprintf("Upgrading ClusterBootstrap from TKR %s to TKR %s", clusterBootstrap.Status.ResolvedTKR, tkrName))
		return r.patchClusterBootstrapFromTemplate(cluster, clusterBootstrap, clusterBootstrapTemplate, clusterBootstrapHelper, tkrName, log)
	}
	return nil, errors.New("should not happen")
}

// patchClusterBootstrapFromTemplate will patch ClusterBootstrap associated with a cluster in case of TKR upgrade
func (r *ClusterBootstrapReconciler) patchClusterBootstrapFromTemplate(
	cluster *clusterapiv1beta1.Cluster,
	clusterBootstrap *runtanzuv1alpha3.ClusterBootstrap,
	clusterBootstrapTemplate *runtanzuv1alpha3.ClusterBootstrapTemplate,
	clusterBootstrapHelper *clusterbootstrapclone.Helper,
	tkrName string,
	log logr.Logger) (*runtanzuv1alpha3.ClusterBootstrap, error) {

	// Will update ClusterBootstrap based on new clusterBootstrapTemplate
	updatedClusterBootstrap := clusterBootstrap.DeepCopy()
	patchHelper, err := clusterapipatchutil.NewHelper(updatedClusterBootstrap, r.Client)
	if err != nil {
		return nil, err
	}
	if clusterBootstrapTemplate.Spec == nil || updatedClusterBootstrap.Spec == nil {
		return nil, errors.New("ClusterBootstrap and ClusterBootstrapTemplate spec can't be nil")
	}

	packages, err := r.mergeClusterBootstrapPackagesWithTemplate(cluster, updatedClusterBootstrap, clusterBootstrapTemplate, log)
	if err != nil {
		return nil, err
	}

	// Handle newly added package values
	secrets, providers, err := clusterBootstrapHelper.CloneReferencedObjectsFromCBPackages(cluster, packages, clusterBootstrapTemplate.Namespace)
	if err != nil {
		r.Log.Error(err, "unable to clone secrets or providers")
		return nil, err
	}

	// No need to update ClusterBootstrap ownerRef
	// Patch the Spec and update the Resolved TKR
	updatedClusterBootstrap.Status.ResolvedTKR = tkrName
	if err := patchHelper.Patch(r.context, updatedClusterBootstrap); err != nil {
		log.Error(err, "failed to updated clusterBootstrap")
		return nil, err
	}
	// ensure ownerRef of clusterBootstrap on created secrets and providers, this can only be done after
	// clusterBootstrap is updated
	if err := clusterBootstrapHelper.EnsureOwnerRef(updatedClusterBootstrap, secrets, providers); err != nil {
		r.Log.Error(err, fmt.Sprintf("unable to ensure ClusterBootstrap %s/%s as a ownerRef on created secrets and providers", clusterBootstrap.Namespace, clusterBootstrap.Name))
		return nil, err
	}

	r.Log.Info("updated clusterBootstrap", "clusterBootstrap", updatedClusterBootstrap)
	return updatedClusterBootstrap, nil
}

// mergeClusterBootstrapPackagesWithTemplate will merge all the packageRefs according to the new ClusterBootstrapTemplate
func (r *ClusterBootstrapReconciler) mergeClusterBootstrapPackagesWithTemplate(
	cluster *clusterapiv1beta1.Cluster,
	updatedClusterBootstrap *runtanzuv1alpha3.ClusterBootstrap,
	clusterBootstrapTemplate *runtanzuv1alpha3.ClusterBootstrapTemplate,
	log logr.Logger) ([]*runtanzuv1alpha3.ClusterBootstrapPackage, error) {

	// Upgrade the refName of all the core packages
	// Package updates keep the users' customization in valuesFrom
	// We assume the following enforced by our build and also webhook:
	//    1. ClusterBootstrapTemplate will always have Kapp and CNI package available
	//    2. ClusterBootstrapTemplate will always have consistent core packages refNames in different TKR versions (same name, different version)
	//    3. The Group and Kind for default core package providers will not change across different TKR versions
	//    4. All packages, including additional packages, can't be deleted (meaning the package refName can't be changed, only allow version bump)
	//    5. We will keep users' customization on valuesFrom of each package, users are responsible for the correctness of the content they put in will work with the next version.
	packages := make([]*runtanzuv1alpha3.ClusterBootstrapPackage, 0)
	if updatedClusterBootstrap.Spec.CNI == nil {
		log.Info("no CNI package specified in ClusterBootstarp, should not happen. Continue with CNI in ClusterBootstrapTemplate of new TKR")
		updatedClusterBootstrap.Spec.CNI = clusterBootstrapTemplate.Spec.CNI.DeepCopy()
	} else {
		// We don't allow change to the CNI selection once it starts running, however we allow version bump
		//TODO: check correctness of the following statement, as we still allow version bump
		// ClusterBootstrap webhook will make sure the package RefName always match the original CNI
		updatedCNI, cniNamePrefix, err := util.GetBootstrapPackageNameFromTKR(r.context, r.Client, updatedClusterBootstrap.Spec.CNI.RefName, cluster)
		if err != nil {
			errorMsg := fmt.Sprintf("unable to find any CNI bootstrap package prefixed with '%s' for ClusterBootstrap %s/%s in TKR", cniNamePrefix, cluster.Name, cluster.Namespace)
			return nil, errors.Wrap(err, errorMsg)
		}
		updatedClusterBootstrap.Spec.CNI.RefName = updatedCNI
	}

	if updatedClusterBootstrap.Spec.Kapp == nil {
		log.Info("no Kapp-Controller package specified in ClusterBootstarp, should not happen. Continue with Kapp-Controller in ClusterBootstrapTemplate of new TKR")
		updatedClusterBootstrap.Spec.Kapp = clusterBootstrapTemplate.Spec.Kapp.DeepCopy()
	} else {
		updatedClusterBootstrap.Spec.Kapp.RefName = clusterBootstrapTemplate.Spec.Kapp.RefName
	}

	// CSI and CPI can be nil, only update if it's present
	// According to assumption 2, no need to do nil check on template
	if updatedClusterBootstrap.Spec.CSI == nil {
		newCSIPkg := clusterBootstrapTemplate.Spec.CSI.DeepCopy()
		updatedClusterBootstrap.Spec.CSI = newCSIPkg
		packages = append(packages, newCSIPkg)
	} else {
		updatedClusterBootstrap.Spec.CSI.RefName = clusterBootstrapTemplate.Spec.CSI.RefName
	}

	if updatedClusterBootstrap.Spec.CPI == nil {
		newCPIPkg := clusterBootstrapTemplate.Spec.CPI.DeepCopy()
		updatedClusterBootstrap.Spec.CPI = newCPIPkg
		packages = append(packages, newCPIPkg)
	} else {
		updatedClusterBootstrap.Spec.CPI.RefName = clusterBootstrapTemplate.Spec.CPI.RefName
	}

	// Since we don't allow users to delete additional packages in our webhook
	// Meaning the users will not be able to customize the packageRefName
	// Find all the corresponding pairs in ClusterBootstrap and new ClusterBootstrapTemplate to update
	// Add the additional package if it's only present in the new ClusterBootstrapTemplate
	// Leave the package as it is if it's only present in ClusterBootstrap but not in the new Template
	additionalPackageMap := map[string]*runtanzuv1alpha3.ClusterBootstrapPackage{}

	for _, pkg := range updatedClusterBootstrap.Spec.AdditionalPackages {
		packageRefName, _, err := util.GetPackageMetadata(r.context, r.aggregatedAPIResourcesClient, pkg.RefName, cluster.Namespace)
		if err != nil || packageRefName == "" {
			errorMsg := fmt.Sprintf("unable to fetch Package.Spec.RefName or Package.Spec.Version from Package %s/%s", cluster.Namespace, pkg.RefName)
			r.Log.Error(err, errorMsg)
			return nil, errors.Wrap(err, errorMsg)
		}
		additionalPackageMap[packageRefName] = pkg
	}

	for _, templatePkg := range clusterBootstrapTemplate.Spec.AdditionalPackages {
		// use the refName in package CR, since the package CR hasn't been cloned at this point, use SystemNamespace to fetch packageCR
		packageRefName, _, err := util.GetPackageMetadata(r.context, r.aggregatedAPIResourcesClient, templatePkg.RefName, r.Config.SystemNamespace)
		if err != nil || packageRefName == "" {
			errorMsg := fmt.Sprintf("unable to fetch Package.Spec.RefName or Package.Spec.Version from Package %s/%s", cluster.Namespace, templatePkg.RefName)
			r.Log.Error(err, errorMsg)
			return nil, errors.Wrap(err, errorMsg)
		}

		// Find the one to one match for additional package in new ClusterBootstrapTemplate and old ClusterBootstrap and update
		if pkg, ok := additionalPackageMap[packageRefName]; ok {
			pkg.RefName = templatePkg.RefName
		} else {
			// If new additional package is added in ClusterBootstrapTemplate, just add it to updated ClusterBootstrap
			newPkg := templatePkg.DeepCopy()
			updatedClusterBootstrap.Spec.AdditionalPackages = append(updatedClusterBootstrap.Spec.AdditionalPackages, newPkg)
			packages = append(packages, newPkg)
		}
	}

	return packages, nil
}

// createOrPatchKappPackageInstall contains the logic that create/update PackageInstall CR for kapp-controller on
// mgmt cluster. The kapp-controller running on mgmt cluster reconciles the PackageInstall CR and creates kapp-controller resources
// on remote workload cluster. This is required for a workload cluster and its corresponding package installations to be functional.
func (r *ClusterBootstrapReconciler) createOrPatchKappPackageInstall(clusterBootstrap *runtanzuv1alpha3.ClusterBootstrap, cluster *clusterapiv1beta1.Cluster) error {
	// Skip if the cluster object represents the management cluster
	if _, exists := cluster.Labels[tkrconstants.ManagementClusterRoleLabel]; exists {
		r.Log.Info(fmt.Sprintf("cluster %s/%s is management cluster, skip creating or patching the PackageInstall CR for kapp-controller", cluster.Namespace, cluster.Name))
		return nil
	}

	// In order to create PackageInstall CR, we need to get the Package.Spec.RefName and Package.Spec.Version
	packageRefName, packageVersion, err := util.GetPackageMetadata(r.context, r.aggregatedAPIResourcesClient, clusterBootstrap.Spec.Kapp.RefName, cluster.Namespace)
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
			Annotations: map[string]string{addontypes.ClusterNameAnnotation: cluster.Name, addontypes.ClusterNamespaceAnnotation: cluster.Namespace},
		},
	}

	pkgiMutateFn := func() error {
		// TODO: Followup on the following fields, only add them if needed.
		// https://github.com/vmware-tanzu/tanzu-framework/issues/1677
		// if ipkg.ObjectMeta.Annotations == nil {
		//	 ipkg.ObjectMeta.Annotations = make(map[string]string)
		// }
		// ipkg.ObjectMeta.Annotations[addontypes.YttMarkerAnnotation] = ""
		pkgi.Spec.NoopDelete = true
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
		// Copy the kapp-controller secret in mgmt cluster to include the TKR nodeSelector info
		secret, err := r.createOrPatchPackageInstallSecretForKapp(cluster, clusterBootstrap.Spec.Kapp, r.Client)
		if err != nil {
			return err
		}
		if secret != nil {
			pkgi.Spec.Values = []kapppkgiv1alpha1.PackageInstallValues{
				{SecretRef: &kapppkgiv1alpha1.PackageInstallValuesSecretRef{
					Name: secret.Name},
				},
			}
		} else {
			r.Log.Info("[Warning]: Empty secret for kapp-controller package. Either KappControllerConfig controller has not reconciled yet or "+
				"ClusterBootstrap is mis-configured with an incorrect clusterBootstrap.Spec.Kapp.ValuesFrom",
				"clusterBootstrap", clusterBootstrap.Name, "namespace", clusterBootstrap.Namespace)
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
	if err = r.aggregatedAPIResourcesClient.Get(r.context, key, localPackage); err != nil {
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
			Name:        util.GeneratePackageInstallName(cluster.Name, remotePackageRefName),
			Namespace:   r.Config.SystemNamespace,
			Annotations: map[string]string{addontypes.ClusterNameAnnotation: cluster.Name, addontypes.ClusterNamespaceAnnotation: cluster.Namespace},
		},
	}

	_, err = controllerutil.CreateOrPatch(r.context, clusterClient, remotePkgi, func() error {
		remotePkgi.Spec.ServiceAccountName = r.Config.PkgiServiceAccount
		remotePkgi.Spec.SyncPeriod = &metav1.Duration{Duration: r.Config.PkgiSyncPeriod}
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

// reconcileSystemNamespace creates system namespace on remote workload cluster. This is because the system namespace
// might not have been created yet when this controller reconciles remote cluster.
func (r *ClusterBootstrapReconciler) reconcileSystemNamespace(clusterClient client.Client) error {
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: r.Config.SystemNamespace,
		},
	}

	result, err := controllerutil.CreateOrPatch(r.context, clusterClient, namespace, nil)
	if err != nil {
		r.Log.Error(err, "unable to create or patch system namespace")
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

	// Create or patch the data value secret on a cluster. The data value secret has been generated by each
	// addon config controller on local cluster.
	remoteSecret, err := r.createOrPatchPackageInstallSecret(cluster, cbPkg, clusterClient)
	if err != nil {
		// We expect there is NO error to create or patch the secret used for PackageInstall in a cluster.
		// Logging has been handled by createOrPatchPackageInstallSecretOnRemote() already
		return err
	}
	if remoteSecret != nil {
		r.Log.Info(fmt.Sprintf("created or patched secret %s/%s for package %s on cluster %s/%s", remoteSecret.Namespace, remoteSecret.Name, remotePackage.Name, cluster.Namespace,
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

func (r *ClusterBootstrapReconciler) patchSecretWithTKGSDataValues(cluster *clusterapiv1beta1.Cluster, secret *corev1.Secret) error {
	// Add TKR NodeSelector info if it's a TKGS cluster
	infraRef, err := util.GetInfraProvider(cluster)
	if err != nil {
		return err
	}
	if infraRef == tkgconstants.InfrastructureProviderVSphere {
		ok, err := util.IsTKGSCluster(r.context, r.dynamicClient, r.gvrHelper.GetDiscoveryClient(), cluster)
		if err != nil {
			return err
		}
		if ok {
			upgradeDataValues := addontypes.TKGSDataValues{
				NodeSelector: addontypes.NodeSelector{
					TanzuKubernetesRelease: cluster.Labels[constants.TKRLabelClassyClusters],
				},
				Deployment: addontypes.DeploymentUpdateInfo{
					UpdateStrategy: constants.TKGSDeploymentUpdateStrategy,
					RollingUpdate: &addontypes.RollingUpdateInfo{
						MaxSurge:       constants.TKGSDeploymentUpdateMaxSurge,
						MaxUnavailable: constants.TKGSDeploymentUpdateMaxUnavailable,
					},
				},
				Daemonset: addontypes.DaemonsetUpdateInfo{
					UpdateStrategy: constants.TKGSDaemonsetUpdateStrategy,
				},
			}
			TKRDataValueYamlBytes, err := yaml.Marshal(upgradeDataValues)
			if err != nil {
				return err
			}
			if secret.StringData == nil {
				secret.StringData = make(map[string]string)
			}
			secret.StringData[constants.TKGSDataValueFileName] = string(TKRDataValueYamlBytes)

			r.Log.Info(fmt.Sprintf("added TKGS data values to secret %s/%s", secret.Namespace, secret.Name))
		} else {
			r.Log.Info(fmt.Sprintf("skip adding TKGS data values to secret %s/%s because %s/%s is not a TKGS cluster", secret.Namespace, secret.Name, cluster.Namespace, cluster.Name))
		}
	}

	return nil
}

func (r *ClusterBootstrapReconciler) getDataValueSecretFromBootstrapPackage(cluster *clusterapiv1beta1.Cluster, cbpkg *runtanzuv1alpha3.ClusterBootstrapPackage) (*corev1.Secret, error) {
	secretName, err := r.GetDataValueSecretNameFromBootstrapPackage(cbpkg, cluster)
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
			return nil, fmt.Errorf("unable to patch secret labels for secret '%s/%s': %w", localSecret.Namespace, localSecret.Name, err)
		}
		r.Log.Info(fmt.Sprintf("patched the secret %s/%s with package and cluster labels", localSecret.Namespace, localSecret.Name))
	}

	return patchedSecret, nil
}

func (r *ClusterBootstrapReconciler) createOrPatchPackageInstallSecretForKapp(cluster *clusterapiv1beta1.Cluster,
	cbpkg *runtanzuv1alpha3.ClusterBootstrapPackage, clusterClient client.Client) (*corev1.Secret, error) {

	localSecret, err := r.getDataValueSecretFromBootstrapPackage(cluster, cbpkg)
	if err != nil {
		return nil, err
	}
	//controller hasn't finished reconciling
	if localSecret == nil {
		return nil, nil
	}

	dataValuesSecretMutateFn := func() error {
		if err := r.patchSecretWithTKGSDataValues(cluster, localSecret); err != nil {
			return err
		}
		return nil
	}

	_, err = controllerutil.CreateOrPatch(r.context, clusterClient, localSecret, dataValuesSecretMutateFn)
	if err != nil {
		r.Log.Error(err, "error creating or patching addon data values secret")
		return nil, err
	}
	return localSecret, nil
}

// createOrPatchPackageInstallSecret creates or patches the secret used for PackageInstall in a cluster
func (r *ClusterBootstrapReconciler) createOrPatchPackageInstallSecret(cluster *clusterapiv1beta1.Cluster,
	cbpkg *runtanzuv1alpha3.ClusterBootstrapPackage, clusterClient client.Client) (*corev1.Secret, error) {

	localSecret, err := r.getDataValueSecretFromBootstrapPackage(cluster, cbpkg)
	if err != nil {
		return nil, err
	}
	if localSecret == nil {
		return nil, nil
	}

	packageRefName, _, err := util.GetPackageMetadata(r.context, r.aggregatedAPIResourcesClient, cbpkg.RefName, cluster.Namespace)
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
		remoteSecret.StringData = make(map[string]string)
		for k, v := range localSecret.Data {
			remoteSecret.StringData[k] = string(v)
		}

		if err := r.patchSecretWithTKGSDataValues(cluster, remoteSecret); err != nil {
			return err
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
	} else if secret.Labels[addontypes.PackageNameLabel] != pkgName ||
		secret.Labels[addontypes.ClusterNameLabel] != clusterName {
		updateLabels = true
	}
	if updateLabels {
		secret.Labels[addontypes.PackageNameLabel] = pkgName
		secret.Labels[addontypes.ClusterNameLabel] = clusterName
	}
	return updateLabels
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

	gvr, err := r.gvrHelper.GetGVR(schema.GroupKind{Group: *providerRef.APIGroup, Kind: providerRef.Kind})
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
func (r *ClusterBootstrapReconciler) GetDataValueSecretNameFromBootstrapPackage(cbPkg *runtanzuv1alpha3.ClusterBootstrapPackage, cluster *clusterapiv1beta1.Cluster) (string, error) {
	var (
		packageRefName string
		err            error
	)

	packageRefName, _, err = util.GetPackageMetadata(r.context, r.aggregatedAPIResourcesClient, cbPkg.RefName, cluster.Namespace)
	if packageRefName == "" || err != nil {
		// Package.Spec.RefName and Package.Spec.Version are required fields for Package CR. We do not expect them to be
		// empty and error should not happen when fetching them from a Package CR.
		r.Log.Error(err, fmt.Sprintf("unable to fetch Package.Spec.RefName or Package.Spec.Version from Package %s/%s",
			cluster.Namespace, cbPkg.RefName))
		return "", err
	}

	if cbPkg.ValuesFrom != nil {
		if cbPkg.ValuesFrom.Inline != nil {
			packageSecretName := util.GeneratePackageSecretName(cluster.Name, packageRefName)
			secret := &corev1.Secret{}
			key := client.ObjectKey{Namespace: cluster.Namespace, Name: packageSecretName}
			if err = r.Get(r.context, key, secret); err != nil {
				r.Log.Error(err, "unable to fetch secret for package with inline config", "objectkey", key)
				return "", err
			}
			return packageSecretName, nil
		}

		if cbPkg.ValuesFrom.SecretRef != "" {
			return cbPkg.ValuesFrom.SecretRef, nil
		}

		if cbPkg.ValuesFrom.ProviderRef != nil {
			gvr, err := r.gvrHelper.GetGVR(schema.GroupKind{Group: *cbPkg.ValuesFrom.ProviderRef.APIGroup, Kind: cbPkg.ValuesFrom.ProviderRef.Kind})
			if err != nil {
				r.Log.Error(err, "unable to get GVR")
				return "", err
			}
			provider, err := r.dynamicClient.Resource(*gvr).Namespace(cluster.Namespace).Get(r.context, cbPkg.ValuesFrom.ProviderRef.Name, metav1.GetOptions{}, "status")
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
	} else { // if cbPkg.ValuesFrom == nil
		// When valuesFrom is nil, we still need to create data values secret for vsphere infrastructure in TKGs
		infraRef, err := util.GetInfraProvider(cluster)
		if err != nil {
			return "", err
		}
		if infraRef == tkgconstants.InfrastructureProviderVSphere {
			ok, err := util.IsTKGSCluster(r.context, r.dynamicClient, r.gvrHelper.GetDiscoveryClient(), cluster)
			if err != nil {
				return "", err
			}
			if ok {
				packageSecretName, err := r.generateSecretForPackagesWithEmptyValuesFrom(cbPkg, cluster, packageRefName)
				if err != nil {
					return "", err
				}
				return packageSecretName, nil
			}
		}
		// cbPkg.ValuesFrom is nil and not TKGS
		return "", nil
	}

	// When valuesFrom is not nil, but either valuesFrom.Inline, valuesFrom.SecretRef, or valuesFrom.providerRef is empty or nil,
	// we interpret it as the data value secret for that package has not been available yet. One of those three fields needs
	// to be provided either by the user or the controller.
	err = fmt.Errorf("unable to get the data value secret name from the ClusterBootstrapPackage.ValuesFrom field. "+
		"ClusterBootstrapPackage.RefName: %s. One of the fields under ClusterBootstrapPackage.ValuesFrom is empty or nil",
		cbPkg.RefName)
	// The message in err object has sufficient information
	r.Log.Error(err, "")
	return "", err
}

func (r *ClusterBootstrapReconciler) watchesForClusterBootstrap() []ClusterBootstrapWatchInputs {
	return []ClusterBootstrapWatchInputs{
		{
			&source.Kind{Type: &runtanzuv1alpha3.TanzuKubernetesRelease{}},
			handler.EnqueueRequestsFromMapFunc(r.TKRToClusters),
		},
		{
			&source.Kind{Type: &runtanzuv1alpha3.ClusterBootstrap{}},
			handler.EnqueueRequestsFromMapFunc(r.ClusterBootstrapToClusters),
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
	HTTPProxy, err := util.ParseClusterVariableInterface(cluster, "proxy", "httpProxy")
	if err != nil {
		log.Error(err, "unable to fetch cluster HTTP proxy setting, defaulting to empty")
	}
	HTTPSProxy, err := util.ParseClusterVariableInterface(cluster, "proxy", "httpsProxy")
	if err != nil {
		log.Error(err, "unable to fetch cluster HTTPS proxy setting, defaulting to empty")
	}
	NoProxy, err := util.ParseClusterVariableInterface(cluster, "proxy", "noProxy")
	if err != nil {
		log.Error(err, "unable to fetch cluster no-proxy proxy setting, defaulting to empty")
	}
	ProxyCACert, err := util.ParseClusterVariableCert(cluster, "trust", "additionalTrustedCAs", "data")
	if err != nil {
		log.Error(err, "unable to fetch cluster proxy CA certificate, defaulting to empty")
	}
	IPFamily, err := util.ParseClusterVariableString(cluster, r.Config.IPFamilyClusterClassVarName)
	if err != nil {
		log.Error(err, "unable to fetch cluster IP family, defaulting to empty")
	}
	SkipTLSVerify, err := util.ParseClusterVariableList(cluster, "skipTLSVerify")
	if err != nil {
		log.Error(err, "unable to fetch cluster IP family, defaulting to empty")
	}
	if cluster.Annotations == nil {
		cluster.Annotations = map[string]string{}
	}

	cluster.Annotations[addontypes.HTTPProxyConfigAnnotation] = HTTPProxy
	cluster.Annotations[addontypes.HTTPSProxyConfigAnnotation] = HTTPSProxy
	cluster.Annotations[addontypes.NoProxyConfigAnnotation] = NoProxy
	cluster.Annotations[addontypes.ProxyCACertConfigAnnotation] = ProxyCACert
	cluster.Annotations[addontypes.IPFamilyConfigAnnotation] = IPFamily
	cluster.Annotations[addontypes.SkipTLSVerifyConfigAnnotation] = SkipTLSVerify

	log.Info("setting proxy and network configurations in Cluster annotation", addontypes.HTTPProxyConfigAnnotation, HTTPProxy, addontypes.HTTPSProxyConfigAnnotation, HTTPSProxy, addontypes.NoProxyConfigAnnotation, NoProxy, addontypes.ProxyCACertConfigAnnotation, ProxyCACert, addontypes.IPFamilyConfigAnnotation, IPFamily, addontypes.SkipTLSVerifyConfigAnnotation, SkipTLSVerify)

	if err := patchHelper.Patch(r.context, cluster); err != nil {
		log.Error(err, "unable to patch Cluster Annotation")
		return err
	}

	return nil
}

func (r *ClusterBootstrapReconciler) makeClusterIsReadyForDeletion(cluster *clusterapiv1beta1.Cluster, remoteClient client.Client, log logr.Logger) (bool, error) {
	log.Info("preparing cluster for deletion")
	clusterBootstrap := &runtanzuv1alpha3.ClusterBootstrap{}
	err := r.Client.Get(r.context, client.ObjectKeyFromObject(cluster), clusterBootstrap)
	if err != nil && !apierrors.IsNotFound(err) {
		log.Error(err, "failed to lookup clusterbootstrap")
		return false, err
	}
	if apierrors.IsNotFound(err) { // if there is no clusterbootstrap then the cluster is ready for deletion
		return true, nil
	}

	if hasPackageInstalls(r.context, remoteClient, cluster, r.Config.SystemNamespace, clusterBootstrap.Spec.AdditionalPackages, log) {
		log.Info("cluster has additional packageInstalls that need to be deleted")
		err = r.removeAdditionalPackageInstalls(remoteClient, cluster, clusterBootstrap, log)
		if err != nil {
			return false, err
		}
		return false, nil
	}
	return true, nil
}

func (r *ClusterBootstrapReconciler) removeFinalizersFromClusterResources(cluster *clusterapiv1beta1.Cluster, log logr.Logger) error {
	clusterBootstrap := &runtanzuv1alpha3.ClusterBootstrap{}
	err := r.Client.Get(r.context, client.ObjectKeyFromObject(cluster), clusterBootstrap)
	if err == nil {
		log.Info("removing finalizer for clusterbootstrap")
		err = r.removeFinalizer(clusterBootstrap, clusterBootstrap.DeepCopy())
		if err != nil {
			return err
		}
	} else if err != nil && !apierrors.IsNotFound(err) {
		log.Error(err, "failed to lookup clusterbootstrap")
		return err
	}

	clusterKubeConfigSecret := &corev1.Secret{}
	key := client.ObjectKey{Namespace: cluster.Namespace, Name: secretutil.Name(cluster.Name, secretutil.Kubeconfig)}
	err = r.Client.Get(r.context, key, clusterKubeConfigSecret)
	if err == nil {
		log.Info("removing finalizer for kubeconfig secret")
		err = r.removeFinalizer(clusterKubeConfigSecret, clusterKubeConfigSecret.DeepCopy())
		if err != nil {
			return err
		}
	} else if !apierrors.IsNotFound(err) {
		return err
	}
	log.Info("removing finalizer for cluster")
	err = r.removeFinalizer(cluster, cluster.DeepCopy())
	if err != nil {
		return err
	}

	return nil
}

func (r *ClusterBootstrapReconciler) reconcileDelete(cluster *clusterapiv1beta1.Cluster, log logr.Logger) (ctrl.Result, error) {
	okToRemoveFinalizers := false
	var err error

	log.Info("reconciling cluster delete")
	existingCluster := &clusterapiv1beta1.Cluster{}
	err = r.Client.Get(r.context, client.ObjectKeyFromObject(cluster), existingCluster)
	if apierrors.IsNotFound(err) {
		log.Info("cluster not found. Skipping reconciling")
		return ctrl.Result{}, nil
	}
	if err != nil {
		log.Error(err, "failed to lookup cluster")
		return ctrl.Result{RequeueAfter: constants.RequeueAfterDuration}, err
	}

	timeOutReached := time.Now().After(cluster.GetDeletionTimestamp().Add(r.Config.ClusterDeleteTimeout))
	if timeOutReached {
		log.Info("cluster delete reconcile timeout reached. Proceeding with cluster deletion")
		okToRemoveFinalizers = true
	} else {
		remoteClient, err := util.GetClusterClient(r.context, r.Client, r.Scheme, clusterapiutil.ObjectKey(cluster))
		if err != nil {
			return ctrl.Result{RequeueAfter: constants.RequeueAfterDuration}, fmt.Errorf("failed to get remote cluster client: %w", err)
		}
		okToRemoveFinalizers, err = r.makeClusterIsReadyForDeletion(cluster, remoteClient, log)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	if !okToRemoveFinalizers {
		log.Info("cluster is not ready for deletion")
		return ctrl.Result{RequeueAfter: constants.RequeueAfterDuration}, nil
	}

	log.Info("cluster ready for deletion. Removing finalizers")
	if err = r.removeFinalizersFromClusterResources(cluster, log); err != nil {
		log.Error(err, "unable to remove finalizers")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ClusterBootstrapReconciler) removeFinalizer(o client.Object, deepCopy client.Object) error {
	if controllerutil.ContainsFinalizer(deepCopy, addontypes.AddonFinalizer) {
		controllerutil.RemoveFinalizer(deepCopy, addontypes.AddonFinalizer)
		return r.Client.Patch(r.context, deepCopy, client.MergeFrom(o))
	}
	return nil

}

func hasPackageInstalls(ctx context.Context, remoteClient client.Client,
	cluster *clusterapiv1beta1.Cluster, namespace string, packages []*runtanzuv1alpha3.ClusterBootstrapPackage,
	log logr.Logger) bool {

	for _, pkg := range packages {
		pkgInstallName := util.GeneratePackageInstallName(cluster.Name, pkg.RefName)
		if packageInstallExistsAndCanBeDeleted(ctx, pkgInstallName, namespace, remoteClient, log) {
			log.Info("found " + pkgInstallName + " packageInstall on cluster")
			return true
		}
	}
	return false
}

func packageInstallExistsAndCanBeDeleted(ctx context.Context, pkgInstallName, pkgInstallNamespace string,
	remoteClient client.Client, log logr.Logger) bool {

	pkgInstall := &kapppkgiv1alpha1.PackageInstall{}
	if err := remoteClient.Get(ctx, client.ObjectKey{Name: pkgInstallName, Namespace: pkgInstallNamespace}, pkgInstall); err != nil {
		if apierrors.IsNotFound(err) {
			return false
		}
		log.Error(err, "could not verify status of "+pkgInstallNamespace+"/"+pkgInstallName)
		return true
	}
	for _, condition := range pkgInstall.Status.Conditions {
		if condition.Type == kappctrlv1alpha1.DeleteFailed {
			log.Info("ignoring " + pkgInstallNamespace + "/" + pkgInstallName + " packageInstall because it is in  " + string(kappctrlv1alpha1.DeleteFailed) + " state. ")
			return false
		}
	}
	return true
}

func (r *ClusterBootstrapReconciler) removeAdditionalPackageInstalls(remoteClient client.Client, cluster *clusterapiv1beta1.Cluster, clusterBootstrap *runtanzuv1alpha3.ClusterBootstrap, log logr.Logger) error {
	// Removes all additional package install CRs from the cluster
	log.Info("queueing additional packageInstalls for deletion")
	err := r.Client.Get(r.context, client.ObjectKeyFromObject(cluster), clusterBootstrap)
	if err != nil {
		return err
	}
	if clusterBootstrap == nil {
		return nil
	}
	for _, additionalPkg := range clusterBootstrap.Spec.AdditionalPackages {
		additionalPkgInstall := &kapppkgiv1alpha1.PackageInstall{
			ObjectMeta: metav1.ObjectMeta{
				Name:      util.GeneratePackageInstallName(cluster.Name, additionalPkg.RefName),
				Namespace: r.Config.SystemNamespace,
			},
		}
		err = remoteClient.Delete(r.context, additionalPkgInstall)
		if err != nil && !apierrors.IsNotFound(err) {
			log.Error(err, fmt.Sprintf("unable to delete package install for %s/%s",
				additionalPkgInstall.Namespace, additionalPkg.RefName))
			return err
		}
	}
	return nil
}

func (r *ClusterBootstrapReconciler) generateSecretForPackagesWithEmptyValuesFrom(cbPkg *runtanzuv1alpha3.ClusterBootstrapPackage, cluster *clusterapiv1beta1.Cluster, packageRefName string) (string, error) {
	clusterBootstrap := &runtanzuv1alpha3.ClusterBootstrap{}
	if err := r.Client.Get(r.context, client.ObjectKeyFromObject(cluster), clusterBootstrap); err != nil {
		return "", err
	}

	packageSecret := &corev1.Secret{}
	packageSecret.Name = util.GeneratePackageSecretName(cluster.Name, packageRefName)
	packageSecret.Namespace = cluster.Namespace

	if _, err := controllerutil.CreateOrPatch(r.context, r.Client, packageSecret, func() error {
		packageSecret.StringData = make(map[string]string)

		packageSecret.OwnerReferences = []metav1.OwnerReference{
			{
				APIVersion: clusterapiv1beta1.GroupVersion.String(),
				Kind:       cluster.Kind,
				Name:       cluster.Name,
				UID:        cluster.UID,
			},
			{
				APIVersion:         runtanzuv1alpha3.GroupVersion.String(),
				Kind:               "ClusterBootstrap",
				Name:               clusterBootstrap.Name,
				UID:                clusterBootstrap.UID,
				Controller:         pointer.BoolPtr(true),
				BlockOwnerDeletion: pointer.BoolPtr(true),
			},
		}

		// Set secret.Type to ClusterBootstrapManagedSecret to enable clusterbootstrap_controller to Watch these secrets
		packageSecret.Type = constants.ClusterBootstrapManagedSecret

		return nil
	}); err != nil {
		r.Log.Error(err, "error creating or patching addon package secret")
		return "", err
	}

	r.Log.Info(fmt.Sprintf("created secret %v for ClusterBootstrapPackage.RefName: %s", packageSecret.Name, cbPkg.RefName))
	return packageSecret.Name, nil
}
