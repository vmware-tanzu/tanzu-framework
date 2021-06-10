// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/go-logr/logr"
	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterapiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/vmware-tanzu-private/core/addons/constants"
	addonconstants "github.com/vmware-tanzu-private/core/addons/pkg/constants"
	addontypes "github.com/vmware-tanzu-private/core/addons/pkg/types"
	"github.com/vmware-tanzu-private/core/addons/pkg/util"
	bomtypes "github.com/vmware-tanzu-private/core/pkg/v1/tkr/pkg/types"
)

func (r *AddonReconciler) reconcileAddonNamespace(
	ctx context.Context,
	log logr.Logger,
	clusterClient client.Client) error {

	addonNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: addonconstants.TKGAddonsAppNamespace,
		},
	}

	result, err := controllerutil.CreateOrPatch(ctx, clusterClient, addonNamespace, nil)
	if err != nil {
		log.Error(err, "Error creating or patching addon namespace")
		return err
	}

	r.logOperationResult(log, "addon namespace", result)

	return nil
}

func (r *AddonReconciler) reconcileAddonServiceAccount(
	ctx context.Context,
	log logr.Logger,
	clusterClient client.Client) error {

	addonServiceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      addonconstants.TKGAddonsAppServiceAccount,
			Namespace: addonconstants.TKGAddonsAppNamespace,
		},
	}

	result, err := controllerutil.CreateOrPatch(ctx, clusterClient, addonServiceAccount, nil)
	if err != nil {
		log.Error(err, "Error creating or patching addon service account")
		return err
	}

	r.logOperationResult(log, "addon service account", result)

	return nil
}

func (r *AddonReconciler) reconcileAddonRole(
	ctx context.Context,
	log logr.Logger,
	clusterClient client.Client) error {

	addonRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: addonconstants.TKGAddonsAppClusterRole,
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

	r.logOperationResult(log, "addon role", roleResult)

	addonRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: addonconstants.TKGAddonsAppClusterRoleBinding,
		},
	}

	addonRoleBindingMutateFn := func() error {
		addonRoleBinding.Subjects = []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      addonconstants.TKGAddonsAppServiceAccount,
				Namespace: addonconstants.TKGAddonsAppNamespace,
			},
		}

		addonRoleBinding.RoleRef = rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     addonconstants.TKGAddonsAppClusterRole,
		}

		return nil
	}

	roleBindingResult, err := controllerutil.CreateOrPatch(ctx, clusterClient, addonRoleBinding, addonRoleBindingMutateFn)
	if err != nil {
		log.Error(err, "Error creating or patching addon role binding")
		return err
	}

	r.logOperationResult(log, "addon role binding", roleBindingResult)

	return nil
}

func (r *AddonReconciler) logOperationResult(log logr.Logger, resourceName string, result controllerutil.OperationResult) {
	switch result {
	case controllerutil.OperationResultCreated,
		controllerutil.OperationResultUpdated,
		controllerutil.OperationResultUpdatedStatus,
		controllerutil.OperationResultUpdatedStatusOnly:
		log.Info(fmt.Sprintf("Resource %s %s", resourceName, result))
	default:
	}
}

// nolint:dupl
func (r *AddonReconciler) reconcileAddonDataValuesSecretDelete(
	ctx context.Context,
	log logr.Logger,
	clusterClient client.Client,
	addonSecret *corev1.Secret) error {

	addonDataValuesSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GenerateAppSecretNameFromAddonSecret(addonSecret),
			Namespace: util.GenerateAppNamespaceFromAddonSecret(addonSecret),
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

func (r *AddonReconciler) reconcileAddonDataValuesSecretNormal(
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
			Namespace: util.GenerateAppNamespaceFromAddonSecret(addonSecret),
		},
	}

	imageInfoBytes, err := util.GetImageInfo(addonConfig, imageRepository, bom)
	if err != nil {
		log.Error(err, "Error retrieving addon image info")
		return err
	}

	addonDataValuesSecretMutateFn := func() error {
		addonDataValuesSecret.Type = corev1.SecretTypeOpaque
		addonDataValuesSecret.Data = addonSecret.Data
		addonDataValuesSecret.Data["imageInfo.yaml"] = imageInfoBytes
		return nil
	}

	result, err := controllerutil.CreateOrPatch(ctx, clusterClient, addonDataValuesSecret, addonDataValuesSecretMutateFn)
	if err != nil {
		log.Error(err, "Error creating or patching addon data values secret")
		return err
	}

	r.logOperationResult(log, "addon data values secret", result)

	return nil
}

// nolint:dupl
func (r *AddonReconciler) reconcileAddonAppDelete(
	ctx context.Context,
	log logr.Logger,
	clusterClient client.Client,
	addonSecret *corev1.Secret) error {

	app := &kappctrl.App{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GenerateAppNameFromAddonSecret(addonSecret),
			Namespace: util.GenerateAppNamespaceFromAddonSecret(addonSecret),
		},
	}

	if err := clusterClient.Delete(ctx, app); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Addon app not found")
			return nil
		}
		log.Error(err, "Error deleting addon app")
		return err
	}

	log.Info("Deleted app")

	return nil
}

// nolint:funlen
func (r *AddonReconciler) reconcileAddonAppNormal(
	ctx context.Context,
	log logr.Logger,
	remoteApp bool,
	remoteCluster *clusterapiv1alpha3.Cluster,
	clusterClient client.Client,
	addonSecret *corev1.Secret,
	addonConfig *bomtypes.Addon,
	imageRepository string,
	bom *bomtypes.Bom) error {

	addonName := util.GetAddonNameFromAddonSecret(addonSecret)

	app := &kappctrl.App{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GenerateAppNameFromAddonSecret(addonSecret),
			Namespace: util.GenerateAppNamespaceFromAddonSecret(addonSecret),
		},
	}

	appMutateFn := func() error {
		if app.ObjectMeta.Annotations == nil {
			app.ObjectMeta.Annotations = make(map[string]string)
		}

		app.ObjectMeta.Annotations[addontypes.AddonTypeAnnotation] = fmt.Sprintf("%s/%s", addonConfig.Category, addonName)
		app.ObjectMeta.Annotations[addontypes.AddonNameAnnotation] = addonSecret.Name
		app.ObjectMeta.Annotations[addontypes.AddonNamespaceAnnotation] = addonSecret.Namespace

		/*
		 * remoteApp means App CR on the management cluster that kapp-controller uses to remotely manages set of objects deployed in a workload cluster.
		 * workload clusters kubeconfig details need to be added for remote App so that kapp-controller on management
		 * cluster can reconcile and push the addon/app to the workload cluster
		 */
		if remoteApp {
			clusterKubeconfigDetails := util.GetClusterKubeconfigSecretDetails(remoteCluster)

			app.Spec.Cluster = &kappctrl.AppCluster{
				KubeconfigSecretRef: &kappctrl.AppClusterKubeconfigSecretRef{
					Name: clusterKubeconfigDetails.Name,
					Key:  clusterKubeconfigDetails.Key,
				},
			}
		} else {
			app.Spec.ServiceAccountName = addonconstants.TKGAddonsAppServiceAccount
		}

		app.Spec.SyncPeriod = &metav1.Duration{Duration: r.Config.AppSyncPeriod}

		/*example addon section in BOM:

		  app-controller:
		    category: addons-management
		    clusterTypes:
		    - management
		    - workload
		    templatesImagePath: tanzu_core/addons/kapp-controller-templates (legacy)
		    templatesImageTag: v1.3.0 (legacy)
		    addonTemplatesImage:
		    - componentRef: tanzu_core_addons
		      imageRefs:
		      - kappControllerTemplatesImage
		    addonContainerImages:
		    - componentRef: kapp-controller
		      imageRefs:
		      - kappControllerImage
		*/
		var templateImagePath, templateImageTag string
		if len(addonConfig.AddonTemplatesImage) < 1 || len(addonConfig.AddonTemplatesImage[0].ImageRefs) < 1 {
			// if AddonTemplatesImage and AddonTemplatesImage are not present, use the older BOM format
			templateImagePath = addonConfig.TemplatesImagePath
			templateImageTag = addonConfig.TemplatesImageTag
		} else {
			templateImageComponentName := addonConfig.AddonTemplatesImage[0].ComponentRef
			templateImageName := addonConfig.AddonTemplatesImage[0].ImageRefs[0]

			templateImage, err := bom.GetImageInfo(templateImageComponentName, "", templateImageName)
			if err != nil {
				log.Error(err, "Error getting template image info from BOM", constants.ComponentNameLogKey, templateImageComponentName, constants.ImageNameLogKey, templateImageName)
				return err
			}
			templateImagePath = templateImage.ImagePath
			templateImageTag = templateImage.Tag
		}

		if templateImagePath == "" || templateImageTag == "" {
			err := errors.New(fmt.Sprintf("unable to get template image for %s", addonName))
			log.Error(err, "Error getting valid template image info from BOM")
			return err
		}

		templateImageURL := fmt.Sprintf("%s/%s:%s", imageRepository, templateImagePath, templateImageTag)
		log.Info("Addon template image found", constants.ImageURLLogKey, templateImageURL)

		app.Spec.Fetch = []kappctrl.AppFetch{
			{
				Image: &kappctrl.AppFetchImage{
					URL: templateImageURL,
				},
			},
		}

		app.Spec.Template = []kappctrl.AppTemplate{
			{
				Ytt: &kappctrl.AppTemplateYtt{
					IgnoreUnknownComments: true,
					Strict:                false,
					Inline: &kappctrl.AppFetchInline{
						PathsFrom: []kappctrl.AppFetchInlineSource{
							{
								SecretRef: &kappctrl.AppFetchInlineSourceRef{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: util.GenerateAppSecretNameFromAddonSecret(addonSecret),
									},
								},
							},
						},
					},
				},
			},
		}

		app.Spec.Deploy = []kappctrl.AppDeploy{
			{
				Kapp: &kappctrl.AppDeployKapp{
					// --wait-timeout flag specifies the maximum time to wait for App deployment. In some corner cases,
					// current App could have the dependency on the deployment of another App, so current App could get
					// stuck in wait phase.
					RawOptions: []string{fmt.Sprintf("--wait-timeout=%s", r.Config.AppWaitTimeout)},
				},
			},
		}

		// If its a remoteApp set delete to no-op since the app doesnt have to be deleted when cluster is deleted.
		if remoteApp {
			app.Spec.NoopDelete = true
		}

		return nil
	}

	result, err := controllerutil.CreateOrPatch(ctx, clusterClient, app, appMutateFn)
	if err != nil {
		log.Error(err, "Error creating or patching addon")
		return err
	}

	r.logOperationResult(log, "app", result)

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

	if err := r.reconcileAddonAppDelete(ctx, logWithContext, clusterClient, addonSecret); err != nil {
		log.Error(err, "Error reconciling addon app delete")
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

	if err := r.reconcileAddonDataValuesSecretNormal(ctx, logWithContext, clusterClient, addonSecret, addonConfig, imageRepository, bom); err != nil {
		log.Error(err, "Error reconciling addon data values secret")
		return err
	}

	if err := r.reconcileAddonAppNormal(ctx, logWithContext, remoteApp, remoteCluster, clusterClient, addonSecret, addonConfig, imageRepository, bom); err != nil {
		log.Error(err, "Error reconciling addon app")
		return err
	}

	return nil
}
