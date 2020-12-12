// Copyright (c) 2020 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/vmware-tanzu-private/core/addons/constants"
	addonpredicates "github.com/vmware-tanzu-private/core/addons/predicates"
	"github.com/vmware-tanzu-private/core/addons/util"
	addonsv1alpha1 "github.com/vmware-tanzu-private/core/apis/addons/v1alpha1"
	runtanzuv1alpha1 "github.com/vmware-tanzu-private/core/apis/run/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	clusterapiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	capiremote "sigs.k8s.io/cluster-api/controllers/remote"
	controlplanev1alpha3 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1alpha3"
	clusterapiutil "sigs.k8s.io/cluster-api/util"
	clusterApiPredicates "sigs.k8s.io/cluster-api/util/predicates"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type AddonReconciler struct {
	Client  client.Client
	Log     logr.Logger
	Scheme  *runtime.Scheme
	Tracker *capiremote.ClusterCacheTracker
}

func (r *AddonReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	return ctrl.NewControllerManagedBy(mgr).
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
		Complete(r)
}

func (r *AddonReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	log := r.Log.WithValues("cluster", req.NamespacedName)

	log.Info("Reconciling cluster")

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

	if !cluster.GetDeletionTimestamp().IsZero() {
		if err := r.reconcileDelete(ctx, log, cluster); err != nil {
			log.Error(err, "failed to reconcile cluster")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

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
	// Deletion of secret is handled by ownerreference. So, we just remove finalizer from secret here.
	for _, addonSecret := range addonSecrets.Items {
		if !addonSecret.GetDeletionTimestamp().IsZero() {
			addonName := util.GetAddonNameFromAddonSecret(&addonSecret)

			if err := r.removeFinalizerFromAddonSecret(ctx, log, cluster, &addonSecret); err != nil {
				log.Error(err, "Error removing metadata from addon secret", constants.ADDON_NAME_LOG_KEY, addonName)
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

	//TODO: All the below code can be refactored into a library
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

	// Get remote cluster client
	remoteClient, err := r.Tracker.GetClient(ctx, clusterapiutil.ObjectKey(cluster))
	if err != nil {
		log.Error(err, "Error getting remote cluster client")
		return err
	}

	// reconcile each addon secret
	var errors []error
	for _, addonSecret := range addonSecrets.Items {
		log := log.WithValues(constants.ADDON_SECRET_NAMESPACE_LOG_KEY, addonSecret.Namespace, constants.ADDON_SECRET_NAME_LOG_KEY, addonSecret.Name)

		log.Info("Reconciling addon secret")

		addonName := util.GetAddonNameFromAddonSecret(&addonSecret)
		if addonName == "" {
			log.Info("Addon name not found from addon secret")
			continue
		}

		log.Info("Reconciling addon", constants.ADDON_NAME_LOG_KEY, addonName)

		addonConfig := util.GetAddonConfigFromBom(addonName, bomConfig)
		if addonConfig == nil {
			log.Info("Addon config not found from BOM for addon", constants.ADDON_NAME_LOG_KEY, addonName)
			continue
		}

		// If addon secret is marked for deletion then delete the addon, else create/patch it
		if !addonSecret.GetDeletionTimestamp().IsZero() {
			if err := r.reconcileAddonDelete(ctx, log, remoteClient, &addonSecret); err != nil {
				log.Error(err, "Error reconciling addon delete", constants.ADDON_NAME_LOG_KEY, addonName)
				errors = append(errors, err)
				continue
			}

			// Remove finalizer from addon secret
			// TODO: Figure out how to wait until app is deleted in the remote cluster before we remove finalizer
			if err := r.removeFinalizerFromAddonSecret(ctx, log, cluster, &addonSecret); err != nil {
				log.Error(err, "Error removing metadata from addon secret", constants.ADDON_NAME_LOG_KEY, addonName)
				errors = append(errors, err)
				continue
			}
		} else {
			// Add finalizer and owner reference to addon secret
			if err := r.addMetadataToAddonSecret(ctx, log, cluster, &addonSecret); err != nil {
				log.Error(err, "Error adding metadata to addon secret", constants.ADDON_NAME_LOG_KEY, addonName)
				errors = append(errors, err)
				continue
			}

			log.Info(fmt.Sprintf("%#v", addonSecret.Finalizers))
			log.Info(fmt.Sprintf("%#v", addonSecret.OwnerReferences))

			if err := r.reconcileAddonNormal(ctx, log, remoteClient, &addonSecret, addonConfig); err != nil {
				log.Error(err, "Error reconciling addon", constants.ADDON_NAME_LOG_KEY, addonName)
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

// removeFinalizerFromAddonSecret removes finalizer from addon secret
func (r *AddonReconciler) removeFinalizerFromAddonSecret(
	ctx context.Context,
	log logr.Logger,
	cluster *clusterapiv1alpha3.Cluster,
	addonSecret *corev1.Secret) error {

	addonName := util.GetAddonNameFromAddonSecret(addonSecret)

	addonSecretPatchObj := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      addonSecret.Name,
			Namespace: addonSecret.Namespace,
		},
	}

	mutateFn := func() error {
		controllerutil.RemoveFinalizer(addonSecretPatchObj, addonsv1alpha1.AddonFinalizer)
		return nil
	}

	if _, err := controllerutil.CreateOrPatch(ctx, r.Client, addonSecretPatchObj, mutateFn); err != nil {
		log.Error(err, "Error patching addon secret with finalizer and owner reference", constants.ADDON_NAME_LOG_KEY, addonName)
		return err
	}

	return nil
}

// addMetadataToAddonSecret adds finalizer and owner reference to the addon secret
func (r *AddonReconciler) addMetadataToAddonSecret(
	ctx context.Context,
	log logr.Logger,
	cluster *clusterapiv1alpha3.Cluster,
	addonSecret *corev1.Secret) error {

	addonName := util.GetAddonNameFromAddonSecret(addonSecret)

	addonSecretPatchObj := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      addonSecret.Name,
			Namespace: addonSecret.Namespace,
		},
	}

	mutateFn := func() error {

		// Add finalizer first if not exist to avoid the race condition between init and delete
		if !controllerutil.ContainsFinalizer(addonSecretPatchObj, addonsv1alpha1.AddonFinalizer) {
			log.Info("Adding finalizer to addon secret", constants.ADDON_NAME_LOG_KEY, addonName)
			controllerutil.AddFinalizer(addonSecretPatchObj, addonsv1alpha1.AddonFinalizer)
		}

		addonSecretPatchObj.OwnerReferences = clusterapiutil.EnsureOwnerRef(addonSecretPatchObj.OwnerReferences, metav1.OwnerReference{
			APIVersion: clusterapiv1alpha3.GroupVersion.String(),
			Kind:       "Cluster",
			Name:       cluster.Name,
			UID:        cluster.UID,
		})

		return nil
	}

	if _, err := controllerutil.CreateOrPatch(ctx, r.Client, addonSecretPatchObj, mutateFn); err != nil {
		log.Error(err, "Error patching addon secret with finalizer and owner reference", constants.ADDON_NAME_LOG_KEY, addonName)
		return err
	}

	return nil
}
