// Copyright (c) 2020 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/vmware-tanzu-private/core/addons/constants"
	addonpredicates "github.com/vmware-tanzu-private/core/addons/predicates"
	"github.com/vmware-tanzu-private/core/addons/util"
	addonsv1alpha1 "github.com/vmware-tanzu-private/core/apis/addons/v1alpha1"
	bomv1alpha1 "github.com/vmware-tanzu-private/core/apis/bom/v1alpha1"
	runtanzuv1alpha1 "github.com/vmware-tanzu-private/core/apis/run/v1alpha1"
	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/utils/pointer"
	clusterapiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	capiremote "sigs.k8s.io/cluster-api/controllers/remote"
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
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type AddonReconciler struct {
	Client     client.Client
	Log        logr.Logger
	Scheme     *runtime.Scheme
	Tracker    *capiremote.ClusterCacheTracker
	controller controller.Controller
}

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

func (r *AddonReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	log := r.Log.WithValues("cluster", req.NamespacedName)

	log.Info("Reconciling cluster")

	// get cluster object
	cluster := &clusterapiv1alpha3.Cluster{}
	if err := r.Client.Get(ctx, req.NamespacedName, cluster); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Cluster not found")
			return ctrl.Result{}, nil
		} else {
			log.Error(err, "unable to fetch cluster")
			return ctrl.Result{}, err
		}
	}

	// if deletion timestamp is set, handle cluster deletion
	if !cluster.GetDeletionTimestamp().IsZero() {
		if err := r.reconcileDelete(ctx, log, cluster); err != nil {
			log.Error(err, "failed to reconcile cluster")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// reconcile addons in cluster
	if err := r.reconcileNormal(ctx, log, cluster); err != nil {
		log.Error(err, "failed to reconcile cluster")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, reterr
}

// reconcileDelete deletes the addon secrets that belong to the cluster
func (r *AddonReconciler) reconcileDelete(
	ctx context.Context,
	log logr.Logger,
	cluster *clusterapiv1alpha3.Cluster) error {

	// Get addon secrets for the cluster
	addonSecrets, err := util.GetAddonSecretsForCluster(ctx, r.Client, cluster)
	if err != nil {
		log.Error(err, "Error getting addon secrets for cluster")
		return err
	}

	var errors []error

	// When cluster is deleted, we need to delete all the secrets.
	// Deletion of secret is handled by owner reference. So, we just remove finalizer from secret here.
	for _, addonSecret := range addonSecrets.Items {
		if !addonSecret.GetDeletionTimestamp().IsZero() {
			addonName := util.GetAddonNameFromAddonSecret(&addonSecret)

			if _, err := r.removeFinalizerFromAddonSecret(ctx, log, false, nil, &addonSecret); err != nil {
				log.Error(err, "Error removing metadata from addon secret", constants.AddonNameLogKey, addonName)
				errors = append(errors, err)
				continue
			}
		}
	}

	if len(errors) > 0 {
		return kerrors.NewAggregate(errors)
	}

	return nil
}

// reconcileNormal reconciles the addons belonging to the cluster
func (r *AddonReconciler) reconcileNormal(
	ctx context.Context,
	log logr.Logger,
	cluster *clusterapiv1alpha3.Cluster) error {

	// Get addon secrets for the cluster
	addonSecrets, err := util.GetAddonSecretsForCluster(ctx, r.Client, cluster)
	if err != nil {
		log.Error(err, "Error getting addon secrets for cluster")
		return err
	}

	// Get bom for cluster
	bomConfig, err := util.GetBOMForCluster(ctx, r.Client, cluster)
	if err != nil || bomConfig == nil {
		log.Error(err, "Error getting BOM")
		return err
	}

	if bomConfig.Addons == nil {
		log.Error(err, "Error getting BOM addons")
		return err
	}

	// Get remote cluster live client.
	// Do not use cached client since it creates watches implicitly for all objects that we GET/LIST.
	remoteClient, err := r.Tracker.GetLiveClient(ctx, clusterapiutil.ObjectKey(cluster))
	if err != nil {
		log.Error(err, "Error getting remote cluster client")
		return err
	}

	// reconcile each addon secret
	var errors []error
	for _, addonSecret := range addonSecrets.Items {
		log := log.WithValues(constants.AddonSecretNamespaceLogKey, addonSecret.Namespace, constants.AddonSecretNameLogKey, addonSecret.Name)

		if err := r.reconcileAddonSecret(ctx, log, cluster, remoteClient, &addonSecret, bomConfig); err != nil {
			log.Error(err, "Error reconciling addon secret")
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return kerrors.NewAggregate(errors)
	}

	return nil
}

// reconcileNormal reconciles the addons belonging to the cluster
func (r *AddonReconciler) reconcileAddonSecret(
	ctx context.Context,
	log logr.Logger,
	cluster *clusterapiv1alpha3.Cluster,
	clusterClient client.Client,
	addonSecret *corev1.Secret,
	bomConfig *bomv1alpha1.BomConfig) error {

	var (
		retErr           error
		patchAddonSecret bool
	)

	log.Info("Reconciling addon secret")

	addonName := util.GetAddonNameFromAddonSecret(addonSecret)
	if addonName == "" {
		log.Info("Addon name not found from addon secret")
		return nil
	}

	log.Info("Reconciling addon", constants.AddonNameLogKey, addonName)

	// get addon config from BOM
	addonConfig := util.GetAddonConfigFromBom(addonName, bomConfig)
	if addonConfig == nil {
		log.Info("Addon config not found from BOM for addon", constants.AddonNameLogKey, addonName)
		return nil
	}

	// Create a patch helper for addon secret
	patchHelper, err := clusterapipatchutil.NewHelper(addonSecret, r.Client)
	if err != nil {
		return err
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
		if err := r.reconcileAddonSecretDelete(ctx, log, addonName, clusterClient, addonSecret, &patchAddonSecret); err != nil {
			log.Error(err, "Error reconciling addon secret delete", constants.AddonNameLogKey, addonName)
			return err
		}
	} else {
		if err := r.reconcileAddonSecretNormal(ctx, log, addonName, cluster, clusterClient, addonSecret, &patchAddonSecret, bomConfig); err != nil {
			log.Error(err, "Error reconciling addon secret", constants.AddonNameLogKey, addonName)
			return err
		}
	}

	return retErr
}

// reconcileAddonSecretDelete reconciles a deletion of addon secret
func (r *AddonReconciler) reconcileAddonSecretDelete(
	ctx context.Context,
	log logr.Logger,
	addonName string,
	clusterClient client.Client,
	addonSecret *corev1.Secret,
	patchAddonSecret *bool) error {

	// delete remote app and data values secret
	if err := r.reconcileAddonDelete(ctx, log, clusterClient, addonSecret); err != nil {
		log.Error(err, "Error reconciling addon delete", constants.AddonNameLogKey, addonName)
		return err
	}

	// Remove finalizer from addon secret
	finalizerRemoved, err := r.removeFinalizerFromAddonSecret(ctx, log, true, clusterClient, addonSecret)
	if err != nil {
		log.Error(err, "Error removing metadata from addon secret", constants.AddonNameLogKey, addonName)
		return err
	}

	*patchAddonSecret = finalizerRemoved

	return nil
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
	bomConfig *bomv1alpha1.BomConfig) error {

	// get addon config from BOM
	addonConfig := util.GetAddonConfigFromBom(addonName, bomConfig)
	if addonConfig == nil {
		log.Info("Addon config not found from BOM for addon", constants.AddonNameLogKey, addonName)
		return nil
	}

	// Add finalizer and owner reference to addon secret
	metadataAdded, err := r.addMetadataToAddonSecret(ctx, log, cluster, addonSecret)
	if err != nil {
		log.Error(err, "Error adding metadata to addon secret", constants.AddonNameLogKey, addonName)
		return err
	}

	*patchAddonSecret = metadataAdded

	// create/patch remote app and data values secret
	if err := r.reconcileAddonNormal(ctx, log, cluster, clusterClient, addonSecret, addonConfig); err != nil {
		log.Error(err, "Error reconciling addon", constants.AddonNameLogKey, addonName)
		return err
	}

	// setup app watches
	if err := r.setupAppWatches(ctx, log, cluster, addonSecret); err != nil {
		log.Error(err, "Error setting up app watcher", constants.AddonNameLogKey, addonName)
		return err
	}

	return nil
}

// getAddonsToBeDeleted returns the addons to be deleted in the cluster
func (r *AddonReconciler) getAddonsToBeDeleted(
	ctx context.Context,
	log logr.Logger,
	clusterClient client.Client,
	addonSecrets *corev1.SecretList) ([]string, error) {

	var addonsToBeDeleted []string

	clusterAddons, err := util.GetAddonsInCluster(ctx, clusterClient)
	if err != nil {
		log.Error(err, "Error getting addons in cluster")
		return nil, err
	}

	// Make a map of addon secrets present for easier lookup
	addonSecretMap := make(map[string]bool)
	for _, addonSecret := range addonSecrets.Items {
		addonSecretMap[addonSecret.Name] = true
	}

	// For each cluster addon check if a corresponding secret is present. If not add for deletion.
	for _, clusterAddon := range clusterAddons {
		if _, ok := addonSecretMap[clusterAddon]; !ok {
			addonsToBeDeleted = append(addonsToBeDeleted, clusterAddon)
		}
	}

	return addonsToBeDeleted, nil
}

// removeFinalizerFromAddonSecret removes finalizer from addon secret if it is present and returns true if it is removed
func (r *AddonReconciler) removeFinalizerFromAddonSecret(
	ctx context.Context,
	log logr.Logger,
	checkAppBeforeRemoval bool,
	clusterClient client.Client,
	addonSecret *corev1.Secret) (bool, error) {

	addonName := util.GetAddonNameFromAddonSecret(addonSecret)

	if checkAppBeforeRemoval {
		appPresent, err := util.IsAppPresent(ctx, r.Client, clusterClient, addonSecret)
		if err != nil {
			log.Error(err, "Error checking if app is present", constants.AddonNameLogKey, addonName)
			return false, err
		}
		// If app is present, return without removing finalizer
		if appPresent {
			log.V(4).Info("App still present. Not removing finalizer", constants.AddonNameLogKey, addonName)
			return false, nil
		}
	}

	// remove finalizer from addon secret
	if controllerutil.ContainsFinalizer(addonSecret, addonsv1alpha1.AddonFinalizer) {
		log.Info("Removing finalizer to addon secret", constants.AddonNameLogKey, addonName)
		controllerutil.RemoveFinalizer(addonSecret, addonsv1alpha1.AddonFinalizer)
		return true, nil
	}

	return false, nil
}

// addMetadataToAddonSecret adds finalizer and owner reference to the addon secret if not present and
// returns true if finalizer or owner reference is added
func (r *AddonReconciler) addMetadataToAddonSecret(
	ctx context.Context,
	log logr.Logger,
	cluster *clusterapiv1alpha3.Cluster,
	addonSecret *corev1.Secret) (bool, error) {

	var patchAddonSecret bool

	addonName := util.GetAddonNameFromAddonSecret(addonSecret)

	// add finalizer to addon secret
	if !controllerutil.ContainsFinalizer(addonSecret, addonsv1alpha1.AddonFinalizer) {
		log.Info("Adding finalizer to addon secret", constants.AddonNameLogKey, addonName)
		controllerutil.AddFinalizer(addonSecret, addonsv1alpha1.AddonFinalizer)
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

	return patchAddonSecret, nil
}

// setupAppWatches watches the App objects on a cluster
func (r *AddonReconciler) setupAppWatches(
	ctx context.Context,
	log logr.Logger,
	cluster *clusterapiv1alpha3.Cluster,
	addonSecret *corev1.Secret) error {

	addonName := util.GetAddonNameFromAddonSecret(addonSecret)

	log.Info("Setting up watch for app")

	remoteApp := util.IsRemoteApp(addonSecret)
	if remoteApp {
		// if remote app i.e.App lives on management cluster, then setup a local watch
		if err := r.controller.Watch(&source.Kind{Type: &kappctrl.App{}},
			handler.EnqueueRequestsFromMapFunc(r.AppToClusters),
			addonpredicates.App(log)); err != nil {
			r.Log.Error(err, "Error setting up app watch on local cluster")
			return err
		}
	} else {
		if err := r.Tracker.Watch(ctx, capiremote.WatchInput{
			Name:         addonName,
			Cluster:      clusterapiutil.ObjectKey(cluster),
			Watcher:      r.controller,
			Kind:         &kappctrl.App{},
			EventHandler: handler.EnqueueRequestsFromMapFunc(r.AppToClusters),
			Predicates:   []predicate.Predicate{addonpredicates.App(log)},
		}); err != nil {
			r.Log.Error(err, "Error setting up app watch on remote cluster")
			return err
		}
	}

	return nil
}
