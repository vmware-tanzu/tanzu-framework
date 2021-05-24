// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterapiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	bomtypes "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/types"
)

func (r *AddonReconciler) reconcileAddonNamespace(
	ctx context.Context,
	log logr.Logger,
	clusterClient client.Client) error {

	addonNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: r.Config.AddonNamespace,
		},
	}

	result, err := controllerutil.CreateOrPatch(ctx, clusterClient, addonNamespace, nil)
	if err != nil {
		log.Error(err, "Error creating or patching addon namespace")
		return err
	}

	logOperationResult(log, "addon namespace", result)

	return nil
}

func (r *AddonReconciler) reconcileAddonServiceAccount(
	ctx context.Context,
	log logr.Logger,
	clusterClient client.Client) error {

	addonServiceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Config.AddonServiceAccount,
			Namespace: r.Config.AddonNamespace,
		},
	}

	result, err := controllerutil.CreateOrPatch(ctx, clusterClient, addonServiceAccount, nil)
	if err != nil {
		log.Error(err, "Error creating or patching addon service account")
		return err
	}

	logOperationResult(log, "addon service account", result)

	return nil
}

func (r *AddonReconciler) reconcileAddonRole(
	ctx context.Context,
	log logr.Logger,
	clusterClient client.Client) error {

	addonRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: r.Config.AddonClusterRole,
		},
	}

	addonRoleMutateFn := func() error {
		addonRole.Rules = []rbacv1.PolicyRule{
			{
				APIGroups: []string{"*"},
				Verbs:     []string{"*"},
				Resources: []string{"*"},
			},
		}

		return nil
	}

	roleResult, err := controllerutil.CreateOrPatch(ctx, clusterClient, addonRole, addonRoleMutateFn)
	if err != nil {
		log.Error(err, "Error creating or patching addon role")
		return err
	}

	logOperationResult(log, "addon role", roleResult)

	addonRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: r.Config.AddonClusterRoleBinding,
		},
	}

	addonRoleBindingMutateFn := func() error {
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
	}

	roleBindingResult, err := controllerutil.CreateOrPatch(ctx, clusterClient, addonRoleBinding, addonRoleBindingMutateFn)
	if err != nil {
		log.Error(err, "Error creating or patching addon role binding")
		return err
	}

	logOperationResult(log, "addon role binding", roleBindingResult)

	return nil
}

func (r *AddonReconciler) reconcileAddonDataValuesSecretDelete(
	ctx context.Context,
	log logr.Logger,
	clusterClient client.Client,
	addonSecret *corev1.Secret) error {

	addonDataValuesSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GenerateAppSecretNameFromAddonSecret(addonSecret),
			Namespace: util.GenerateAppNamespaceFromAddonSecret(addonSecret, r.Config.AddonNamespace),
		},
	}

	if err := clusterClient.Delete(ctx, addonDataValuesSecret); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Addon data values secret not found")
			return nil
		}
		log.Error(err, "Error deleting addon data values secret")
		return err
	}

	log.Info("Deleted app data value secret")

	return nil
}

// ReconcileAddonDataValuesSecretNormal reconciles addons data values secrets
func (r *AddonReconciler) ReconcileAddonDataValuesSecretNormal(
	ctx context.Context,
	log logr.Logger,
	clusterClient client.Client,
	addonSecret *corev1.Secret,
	addonConfig *bomtypes.Addon,
	imageRepository string,
	bom *bomtypes.Bom) error {

	addonDataValuesSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GenerateAppSecretNameFromAddonSecret(addonSecret),
			Namespace: util.GenerateAppNamespaceFromAddonSecret(addonSecret, r.Config.AddonNamespace),
		},
	}

	addonDataValuesSecretMutateFn := func() error {
		addonDataValuesSecret.Type = corev1.SecretTypeOpaque
		if addonDataValuesSecret.Data == nil {
			addonDataValuesSecret.Data = map[string][]byte{}
		}
		for k, v := range addonSecret.Data {
			addonDataValuesSecret.Data[k] = v
		}
		// Add or updates the imageInfo if container image reference exists
		if len(addonConfig.AddonContainerImages) > 0 {
			imageInfoBytes, err := util.GetImageInfo(addonConfig, imageRepository, r.Config.AddonImagePullPolicy, bom)
			if err != nil {
				log.Error(err, "Error retrieving addon image info")
				return err
			}
			addonDataValuesSecret.Data["imageInfo.yaml"] = imageInfoBytes
		}

		return nil
	}

	result, err := controllerutil.CreateOrPatch(ctx, clusterClient, addonDataValuesSecret, addonDataValuesSecretMutateFn)
	if err != nil {
		log.Error(err, "Error creating or patching addon data values secret")
		return err
	}

	logOperationResult(log, "addon app data values secret", result)

	return nil
}

func (r *AddonReconciler) reconcileAddonDelete(
	ctx context.Context,
	log logr.Logger,
	remoteClusterClient client.Client,
	addonSecret *corev1.Secret) error {

	addonName := util.GetAddonNameFromAddonSecret(addonSecret)

	logWithContext := r.Log.WithValues(constants.AddonNameLogKey, addonName)
	logWithContext.Info("Reconciling addon delete")

	clusterClient := util.GetClientFromAddonSecret(addonSecret, r.Client, remoteClusterClient)

	var reconcilerKey string
	// When deleting, check if the corresponding packageInstall is created.
	// If so, delete packageInstall CR. Otherwise, delete App CR.
	pkgiPresent, err := util.IsPackageInstallPresent(ctx, clusterClient, addonSecret, r.Config.AddonNamespace)
	if err != nil {
		log.Error(err, "Error checking if PackageInstall is present", constants.AddonNameLogKey, addonName)
		return err
	}
	if pkgiPresent {
		log.Info("Deleting PackageInstall")
		reconcilerKey = constants.TKGPackageReconcilerKey
	} else {
		log.Info("Deleting App")
		reconcilerKey = constants.TKGAppReconcilerKey
	}
	kappResourceReconciler, err := r.GetAddonKappResourceReconciler(ctx, logWithContext, clusterClient, reconcilerKey)
	if err != nil {
		log.Error(err, "Error finding kapp resource reconciler")
		return err
	}

	if err := kappResourceReconciler.ReconcileAddonKappResourceDelete(addonSecret); err != nil {
		log.Error(err, "Error reconciling addon kapp resource delete")
		return err
	}

	if err := r.reconcileAddonDataValuesSecretDelete(ctx, logWithContext, clusterClient, addonSecret); err != nil {
		log.Error(err, "Error reconciling addon data values secret delete")
		return err
	}

	return nil
}

func (r *AddonReconciler) reconcileAddonNormal(
	ctx context.Context,
	log logr.Logger,
	remoteCluster *clusterapiv1alpha3.Cluster,
	remoteClusterClient client.Client,
	addonSecret *corev1.Secret,
	addonConfig *bomtypes.Addon,
	imageRepository string,
	bom *bomtypes.Bom) error {

	addonName := util.GetAddonNameFromAddonSecret(addonSecret)

	logWithContext := r.Log.WithValues(constants.AddonNameLogKey, addonName)
	logWithContext.Info("Reconciling addon")

	remoteApp := util.IsRemoteApp(addonSecret)
	clusterClient := util.GetClientFromAddonSecret(addonSecret, r.Client, remoteClusterClient)

	/* remoteApp means App that lives in management cluster. but deployed in workload cluster.
	 * Since App doesn't deploy on workload cluster, namespace, serviceaccount, roles and rolebindings dont need to be created
	 * on management cluster.
	 */
	if !remoteApp {
		if err := r.reconcileAddonNamespace(ctx, logWithContext, clusterClient); err != nil {
			log.Error(err, "Error reconciling addon namespace")
			return err
		}

		if err := r.reconcileAddonServiceAccount(ctx, logWithContext, clusterClient); err != nil {
			log.Error(err, "Error reconciling addon service account")
			return err
		}

		if err := r.reconcileAddonRole(ctx, logWithContext, clusterClient); err != nil {
			log.Error(err, "Error reconciling addon roles and role bindings")
			return err
		}
	}

	if err := r.ReconcileAddonDataValuesSecretNormal(ctx, logWithContext, clusterClient, addonSecret, addonConfig, imageRepository, bom); err != nil {
		log.Error(err, "Error reconciling addon data values secret")
		return err
	}

	var reconcilerKey string
	if addonConfig.PackageName != "" {
		log.Info("Reconciling PackageInstall")
		reconcilerKey = constants.TKGPackageReconcilerKey
	} else {
		log.Info("Reconciling App")
		reconcilerKey = constants.TKGAppReconcilerKey
	}
	kappResourceReconciler, err := r.GetAddonKappResourceReconciler(ctx, logWithContext, clusterClient, reconcilerKey)
	if err != nil {
		log.Error(err, "Error finding kapp resource reconciler")
		return err
	}

	if err := kappResourceReconciler.ReconcileAddonKappResourceNormal(remoteApp, remoteCluster, addonSecret, addonConfig, imageRepository, bom); err != nil {
		log.Error(err, "Error reconciling addon kapp resource")
		return err
	}

	return nil
}
