// Copyright 2020 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/utils/pointer"
	clusterapiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	controlplanev1alpha3 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1alpha3"
	clusterapiutil "sigs.k8s.io/cluster-api/util"
	clusterapipatchutil "sigs.k8s.io/cluster-api/util/patch"
	clusterApiPredicates "sigs.k8s.io/cluster-api/util/predicates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	addonconfig "github.com/vmware-tanzu-private/core/addons/pkg/config"
	"github.com/vmware-tanzu-private/core/addons/pkg/constants"
	addontypes "github.com/vmware-tanzu-private/core/addons/pkg/types"
	"github.com/vmware-tanzu-private/core/addons/pkg/util"
	addonpredicates "github.com/vmware-tanzu-private/core/addons/predicates"
	runtanzuv1alpha1 "github.com/vmware-tanzu-private/core/apis/run/v1alpha1"
	bomtypes "github.com/vmware-tanzu-private/core/pkg/v1/tkr/pkg/types"
)

const (
	deleteRequeueAfter = 10 * time.Second
)

type AddonKappResourceReconciler interface {
	ReconcileAddonKappResourceNormal(
		remoteApp bool,
		remoteCluster *clusterapiv1alpha3.Cluster,
		addonSecret *corev1.Secret,
		addonConfig *bomtypes.Addon,
		imageRepository string,
		bom *bomtypes.Bom) error

	ReconcileAddonKappResourceDelete(
		addonSecret *corev1.Secret) error
}

// AddonReconciler contains the reconciler information for addon controllers.
type AddonReconciler struct {
	Client     client.Client
	Log        logr.Logger
	Scheme     *runtime.Scheme
	controller controller.Controller
	Config     addonconfig.Config
}

// SetupWithManager performs the setup actions for an add on controller, using the passed in mgr.
func (r *AddonReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	addonController, err := ctrl.NewControllerManagedBy(mgr).
		For(&clusterapiv1alpha3.Cluster{}).
		Watches(
			&source.Kind{Type: &corev1.Secret{}},
			handler.EnqueueRequestsFromMapFunc(r.AddonSecretToClusters),
			builder.WithPredicates(
				addonpredicates.AddonSecret(r.Log),
			),
		).
		Watches(
			&source.Kind{Type: &runtanzuv1alpha1.TanzuKubernetesRelease{}},
			handler.EnqueueRequestsFromMapFunc(r.TKRToClusters),
			builder.WithPredicates(
				addonpredicates.TKR(r.Log),
			),
		).
		Watches(
			&source.Kind{Type: &corev1.ConfigMap{}},
			handler.EnqueueRequestsFromMapFunc(r.BOMConfigMapToClusters),
			builder.WithPredicates(
				addonpredicates.BomConfigMap(r.Log),
			),
		).
		Watches(
			&source.Kind{Type: &controlplanev1alpha3.KubeadmControlPlane{}},
			handler.EnqueueRequestsFromMapFunc(r.KubeadmControlPlaneToClusters),
			builder.WithPredicates(
				addonpredicates.KubeadmControlPlane(r.Log),
			),
		).
		WithOptions(options).
		WithEventFilter(clusterApiPredicates.ResourceNotPaused(r.Log)).
		Build(r)
	if err != nil {
		r.Log.Error(err, "Error creating an addon controller")
		return err
	}

	r.controller = addonController

	return nil
}

// Reconcile performs the reconciliation action for the controller.
func (r *AddonReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	log := r.Log.WithValues(constants.ClusterNamespaceLogKey, req.Namespace, constants.ClusterNameLogKey, req.Name)

	log.Info("Reconciling cluster")

	// get cluster object
	cluster := &clusterapiv1alpha3.Cluster{}
	if err := r.Client.Get(ctx, req.NamespacedName, cluster); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Cluster not found")
			return ctrl.Result{}, nil
		}

		log.Error(err, "unable to fetch cluster")
		return ctrl.Result{}, err
	}

	// if deletion timestamp is set, handle cluster deletion
	if !cluster.GetDeletionTimestamp().IsZero() {
		result, err := r.reconcileDelete(ctx, log, cluster)
		if err != nil {
			log.Error(err, "failed to reconcile cluster")
			return ctrl.Result{}, err
		}
		return result, nil
	}

	// reconcile addons in cluster
	result, err := r.reconcileNormal(ctx, log, cluster)
	if err != nil {
		log.Error(err, "failed to reconcile cluster")
		return ctrl.Result{}, err
	}

	return result, nil
}

// reconcileDelete deletes the addon secrets that belong to the cluster
func (r *AddonReconciler) reconcileDelete(
	ctx context.Context,
	log logr.Logger,
	cluster *clusterapiv1alpha3.Cluster) (ctrl.Result, error) {

	log.Info("Reconciling cluster deletion")

	// Get addon secrets for the cluster
	addonSecrets, err := util.GetAddonSecretsForCluster(ctx, r.Client, cluster)
	if err != nil {
		log.Error(err, "Error getting addon secrets for cluster")
		return ctrl.Result{}, err
	}

	var errors []error

	// When cluster is deleted, we need to delete all the secrets.
	for i := range addonSecrets.Items {
		addonSecret := addonSecrets.Items[i]
		addonName := util.GetAddonNameFromAddonSecret(&addonSecret)

		logWithContext := log.WithValues(constants.AddonSecretNamespaceLogKey, addonSecret.Namespace,
			constants.AddonSecretNameLogKey, addonSecret.Name, constants.AddonNameLogKey, addonName)

		// Create a patch helper for addon secret
		patchHelper, err := clusterapipatchutil.NewHelper(&addonSecret, r.Client)
		if err != nil {
			errors = append(errors, err)
			continue
		}

		// If App is remote i.e. App resides in the management cluster, delete the app and its secret
		// from management cluster when the workload cluster is deleted.
		// For Apps residing in workload cluster, it is not necessary to delete the app and its secret since the
		// cluster itself is deleted.
		if util.IsRemoteApp(&addonSecret) {
			if err := r.reconcileAddonDelete(ctx, logWithContext, nil, &addonSecret); err != nil {
				logWithContext.Error(err, "Error deleting remote app for addon")
				errors = append(errors, err)
				continue
			}
		}

		// Remove finalizer from addon secret
		finalizerRemoved, _, err := r.removeFinalizerFromAddonSecret(ctx, logWithContext, false, nil, &addonSecret)
		if err != nil {
			logWithContext.Error(err, "Error removing metadata from addon secret")
			errors = append(errors, err)
			continue
		}

		// Patch addon secret
		if finalizerRemoved {
			// Patch addon secret before returning the function
			logWithContext.Info("Patching addon secret to remove finalizer")
			if err := patchHelper.Patch(ctx, addonSecret.DeepCopy()); err != nil {
				logWithContext.Error(err, "Error patching addon secret to remove finalizer")
				errors = append(errors, err)
				continue
			}
		}

		// Delete addon secret
		if err := r.Client.Delete(ctx, &addonSecret); err != nil {
			if apierrors.IsNotFound(err) {
				logWithContext.Info("Addon secret not found")
				continue
			}
			logWithContext.Error(err, "Error deleting addon secret")
			errors = append(errors, err)
			continue
		}
	}

	if len(errors) > 0 {
		return ctrl.Result{}, kerrors.NewAggregate(errors)
	}

	return ctrl.Result{}, nil
}

// reconcileNormal reconciles the addons belonging to the cluster
func (r *AddonReconciler) reconcileNormal(
	ctx context.Context,
	log logr.Logger,
	cluster *clusterapiv1alpha3.Cluster) (ctrl.Result, error) {

	// Get addon secrets for the cluster
	addonSecrets, err := util.GetAddonSecretsForCluster(ctx, r.Client, cluster)
	if err != nil {
		log.Error(err, "Error getting addon secrets for cluster")
		return ctrl.Result{}, err
	}

	// Get bom for cluster
	bom, err := util.GetBOMForCluster(ctx, r.Client, cluster)
	if err != nil {
		log.Error(err, "Error getting BOM")
		return ctrl.Result{}, err
	}

	if bom == nil {
		log.Info("Bom not found")
		return ctrl.Result{}, nil
	}

	// Get remote cluster live client.
	remoteClient, err := util.GetClusterClient(ctx, r.Client, r.Scheme, clusterapiutil.ObjectKey(cluster))
	if err != nil {
		log.Error(err, "Error getting remote cluster client")
		return ctrl.Result{}, err
	}

	// Get the repository used for all images
	imageRepository, err := util.GetAddonImageRepository(ctx, r.Client, bom)
	if err != nil || imageRepository == "" {
		log.Info("Error getting image repository")
		return ctrl.Result{}, err
	}

	var (
		errors []error
		result ctrl.Result
	)
	// Reconcile core package repository in the cluster
	err = PackageReconciler{ctx: ctx, log: log, clusterClient: remoteClient, Config: r.Config}.reconcileCorePackageRepository(imageRepository, bom)
	if err != nil {
		log.Error(err, "Error reconciling core package repository")
		errors = append(errors, err)
	}

	for i := range addonSecrets.Items {
		addonSecret := addonSecrets.Items[i]
		logWithContext := log.WithValues(constants.AddonSecretNamespaceLogKey, addonSecret.Namespace, constants.AddonSecretNameLogKey, addonSecret.Name)

		result, err = r.reconcileAddonSecret(ctx, logWithContext, cluster, remoteClient, &addonSecret, imageRepository, bom)
		if err != nil {
			logWithContext.Error(err, "Error reconciling addon secret")
			errors = append(errors, err)
			continue
		}
	}

	if len(errors) > 0 {
		return ctrl.Result{}, kerrors.NewAggregate(errors)
	}

	return result, nil
}

// reconcileNormal reconciles the addons belonging to the cluster
func (r *AddonReconciler) reconcileAddonSecret(
	ctx context.Context,
	log logr.Logger,
	cluster *clusterapiv1alpha3.Cluster,
	clusterClient client.Client,
	addonSecret *corev1.Secret,
	imageRepository string,
	bom *bomtypes.Bom) (_ ctrl.Result, retErr error) {

	var (
		patchAddonSecret bool
	)

	log.Info("Reconciling addon secret")

	addonName := util.GetAddonNameFromAddonSecret(addonSecret)
	if addonName == "" {
		log.Info("Addon name not found from addon secret")
		return ctrl.Result{}, nil
	}

	log.Info("Reconciling addon", constants.AddonNameLogKey, addonName)

	// Create a patch helper for addon secret
	patchHelper, err := clusterapipatchutil.NewHelper(addonSecret, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Patch addon secret before returning the function
	defer func() {
		// patchAddonSecret will be true if finalizer or ownerrefence is added
		if patchAddonSecret {
			log.Info("Patching addon secret", constants.AddonNameLogKey, addonName)

			if err := patchHelper.Patch(ctx, addonSecret.DeepCopy()); err != nil {
				log.Error(err, "Error patching addon secret", constants.AddonNameLogKey, addonName)
				retErr = err
			}
		}
	}()

	// If addon secret is marked for deletion then delete the addon, else create/patch it
	if !addonSecret.GetDeletionTimestamp().IsZero() {
		result, err := r.reconcileAddonSecretDelete(ctx, log, addonName, clusterClient, addonSecret, &patchAddonSecret)
		if err != nil {
			log.Error(err, "Error reconciling addon secret delete", constants.AddonNameLogKey, addonName)
			return ctrl.Result{}, err
		}
		return result, nil
	}

	result, err := r.reconcileAddonSecretNormal(ctx, log, addonName, cluster, clusterClient, addonSecret, &patchAddonSecret, imageRepository, bom)
	if err != nil {
		log.Error(err, "Error reconciling addon secret", constants.AddonNameLogKey, addonName)
		return ctrl.Result{}, err
	}
	return result, nil
}

// reconcileAddonSecretDelete reconciles a deletion of addon secret
func (r *AddonReconciler) reconcileAddonSecretDelete(
	ctx context.Context,
	log logr.Logger,
	addonName string,
	clusterClient client.Client,
	addonSecret *corev1.Secret,
	patchAddonSecret *bool) (ctrl.Result, error) {

	if r.shouldNotReconcile(log, addonSecret) {
		return ctrl.Result{}, nil
	}

	// delete remote app and data values secret
	if err := r.reconcileAddonDelete(ctx, log, clusterClient, addonSecret); err != nil {
		log.Error(err, "Error reconciling addon delete", constants.AddonNameLogKey, addonName)
		return ctrl.Result{}, err
	}

	// Remove finalizer from addon secret
	finalizerRemoved, requeue, err := r.removeFinalizerFromAddonSecret(ctx, log, true, clusterClient, addonSecret)
	if err != nil {
		log.Error(err, "Error removing metadata from addon secret", constants.AddonNameLogKey, addonName)
		return ctrl.Result{}, err
	}

	*patchAddonSecret = finalizerRemoved

	if requeue {
		return ctrl.Result{RequeueAfter: deleteRequeueAfter}, nil
	}

	return ctrl.Result{}, nil
}

// reconcileAddonSecretNormal reconciles a addon secret
func (r *AddonReconciler) reconcileAddonSecretNormal(
	ctx context.Context,
	log logr.Logger,
	addonName string,
	cluster *clusterapiv1alpha3.Cluster,
	clusterClient client.Client,
	addonSecret *corev1.Secret,
	patchAddonSecret *bool,
	imageRepository string,
	bom *bomtypes.Bom) (ctrl.Result, error) {

	// get addon config from BOM
	addonConfig, err := bom.GetAddon(addonName)
	if err != nil {
		log.Info("Addon config not found from BOM for addon", constants.AddonNameLogKey, addonName)
		return ctrl.Result{}, err
	}

	// Add finalizer and owner reference to addon secret
	metadataAdded := r.addMetadataToAddonSecret(log, cluster, addonSecret)

	*patchAddonSecret = metadataAdded

	if r.shouldNotReconcile(log, addonSecret) {
		return ctrl.Result{}, nil
	}

	// create/patch remote app and data values secret
	if err := r.reconcileAddonNormal(ctx, log, cluster, clusterClient, addonSecret, &addonConfig, imageRepository, bom); err != nil {
		log.Error(err, "Error reconciling addon", constants.AddonNameLogKey, addonName)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// removeFinalizerFromAddonSecret removes finalizer from addon secret if it is present and returns true if it is removed
func (r *AddonReconciler) removeFinalizerFromAddonSecret(
	ctx context.Context,
	log logr.Logger,
	checkAppBeforeRemoval bool,
	clusterClient client.Client,
	addonSecret *corev1.Secret) (finalizerRemoved, requeue bool, err error) {

	addonName := util.GetAddonNameFromAddonSecret(addonSecret)

	if checkAppBeforeRemoval {
		appPresent, err := util.IsAppPresent(ctx, r.Client, clusterClient, addonSecret, r.Config.AddonNamespace)
		if err != nil {
			log.Error(err, "Error checking if app is present", constants.AddonNameLogKey, addonName)
			return false, false, err
		}
		// If app is present, return without removing finalizer
		if appPresent {
			log.V(4).Info("App still present. Not removing finalizer", constants.AddonNameLogKey, addonName)
			return false, true, nil
		}
	}

	// remove finalizer from addon secret
	if controllerutil.ContainsFinalizer(addonSecret, addontypes.AddonFinalizer) {
		log.Info("Removing finalizer to addon secret", constants.AddonNameLogKey, addonName)
		controllerutil.RemoveFinalizer(addonSecret, addontypes.AddonFinalizer)
		return true, false, nil
	}

	return false, false, nil
}

// addMetadataToAddonSecret adds finalizer and owner reference to the addon secret if not present and
// returns true if finalizer or owner reference is added
func (r *AddonReconciler) addMetadataToAddonSecret(
	log logr.Logger,
	cluster *clusterapiv1alpha3.Cluster,
	addonSecret *corev1.Secret) bool {

	var patchAddonSecret bool

	addonName := util.GetAddonNameFromAddonSecret(addonSecret)

	// add finalizer to addon secret
	if !controllerutil.ContainsFinalizer(addonSecret, addontypes.AddonFinalizer) {
		log.Info("Adding finalizer to addon secret", constants.AddonNameLogKey, addonName)
		controllerutil.AddFinalizer(addonSecret, addontypes.AddonFinalizer)
		patchAddonSecret = true
	}

	// add owner reference to addon secret
	ownerReference := metav1.OwnerReference{
		APIVersion:         clusterapiv1alpha3.GroupVersion.String(),
		Kind:               "Cluster",
		Name:               cluster.Name,
		UID:                cluster.UID,
		Controller:         pointer.BoolPtr(true),
		BlockOwnerDeletion: pointer.BoolPtr(true),
	}

	if !clusterapiutil.HasOwnerRef(addonSecret.OwnerReferences, ownerReference) {
		log.Info("Adding owner reference to addon secret", constants.AddonNameLogKey, addonName)
		addonSecret.OwnerReferences = clusterapiutil.EnsureOwnerRef(addonSecret.OwnerReferences, ownerReference)
		patchAddonSecret = true
	}

	return patchAddonSecret
}

func (r *AddonReconciler) shouldNotReconcile(
	log logr.Logger,
	addonSecret *corev1.Secret) bool {

	if util.IsAddonPaused(addonSecret) {
		log.Info("Addon paused")
		return true
	}

	return false
}

// logOperationResult logs the reconcile operation results
func logOperationResult(log logr.Logger, resourceName string, result controllerutil.OperationResult) {
	switch result {
	case controllerutil.OperationResultCreated,
		controllerutil.OperationResultUpdated,
		controllerutil.OperationResultUpdatedStatus,
		controllerutil.OperationResultUpdatedStatusOnly:
		log.Info(fmt.Sprintf("Resource %s %s", resourceName, result))
	default:
	}
}

func (r *AddonReconciler) GetAddonKappResourceReconciler(
	ctx context.Context,
	log logr.Logger,
	clusterClient client.Client,
	reconcilerType string) (error, AddonKappResourceReconciler) {
	switch reconcilerType {
	case constants.TKGAppReconcilerKey:
		return nil, AppReconciler{ctx: ctx, log: log, clusterClient: clusterClient, Config: r.Config}
	case constants.TKGPackageReconcilerKey:
		return nil, PackageReconciler{ctx: ctx, log: log, clusterClient: clusterClient, Config: r.Config}
	}
	return fmt.Errorf("invalid reconciler type: %s", reconcilerType), nil

}
