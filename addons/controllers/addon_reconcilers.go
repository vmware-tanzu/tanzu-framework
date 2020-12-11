package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/vmware-tanzu-private/core/addons/constants"
	"github.com/vmware-tanzu-private/core/addons/util"
	addonsv1alpha1 "github.com/vmware-tanzu-private/core/apis/addons/v1alpha1"
	bomv1alpha1 "github.com/vmware-tanzu-private/core/apis/bom/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *AddonReconciler) reconcileAddonNamespace(
	ctx context.Context,
	log logr.Logger,
	clusterClient client.Client) error {

	addonNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.TKG_ADDONS_APP_NAMESPACE,
		},
	}

	if _, err := controllerutil.CreateOrPatch(ctx, clusterClient, addonNamespace, nil); err != nil {
		log.Error(err, "Error creating or patching addon namespace")
		return err
	}

	return nil
}

func (r *AddonReconciler) reconcileAddonServiceAccount(
	ctx context.Context,
	log logr.Logger,
	clusterClient client.Client) error {

	addonServiceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.TKG_ADDONS_APP_SERVICE_ACCOUNT,
			Namespace: constants.TKG_ADDONS_APP_NAMESPACE,
		},
	}

	if _, err := controllerutil.CreateOrPatch(ctx, clusterClient, addonServiceAccount, nil); err != nil {
		log.Error(err, "Error creating or patching addon service account")
		return err
	}

	return nil
}

func (r *AddonReconciler) reconcileAddonRole(
	ctx context.Context,
	log logr.Logger,
	clusterClient client.Client) error {

	addonRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.TKG_ADDONS_APP_CLUSTER_ROLE,
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

	if _, err := controllerutil.CreateOrPatch(ctx, clusterClient, addonRole, addonRoleMutateFn); err != nil {
		log.Error(err, "Error creating or patching addon role")
		return err
	}

	addonRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.TKG_ADDONS_APP_CLUSTER_ROLE_BINDING,
		},
	}

	addonRoleBindingMutateFn := func() error {
		addonRoleBinding.Subjects = []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      constants.TKG_ADDONS_APP_SERVICE_ACCOUNT,
				Namespace: constants.TKG_ADDONS_APP_NAMESPACE,
			},
		}

		addonRoleBinding.RoleRef = rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     constants.TKG_ADDONS_APP_CLUSTER_ROLE,
		}

		return nil
	}

	if _, err := controllerutil.CreateOrPatch(ctx, clusterClient, addonRoleBinding, addonRoleBindingMutateFn); err != nil {
		log.Error(err, "Error creating or patching addon role binding")
		return err
	}

	return nil
}

func (r *AddonReconciler) reconcileAddonDataValuesSecretDelete(
	ctx context.Context,
	log logr.Logger,
	clusterClient client.Client,
	addonName string) error {

	addonDataValuesSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      addonName,
			Namespace: constants.TKG_ADDONS_APP_NAMESPACE,
		},
	}

	if err := clusterClient.Delete(ctx, addonDataValuesSecret); err != nil {
		log.Error(err, "Error deleting addon app")
		return err
	}

	return nil
}

func (r *AddonReconciler) reconcileAddonDataValuesSecretNormal(
	ctx context.Context,
	log logr.Logger,
	clusterClient client.Client,
	addonSecret *corev1.Secret) error {

	addonName := util.GetAddonNameFromAddonSecret(addonSecret)

	addonDataValuesSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      addonName,
			Namespace: constants.TKG_ADDONS_APP_NAMESPACE,
		},
	}

	addonDataValuesSecretMutateFn := func() error {
		addonDataValuesSecret.Type = corev1.SecretTypeOpaque
		addonDataValuesSecret.Data = addonSecret.Data
		return nil
	}

	if _, err := controllerutil.CreateOrPatch(ctx, clusterClient, addonDataValuesSecret, addonDataValuesSecretMutateFn); err != nil {
		log.Error(err, "Error creating or patching addon data values secret")
		return err
	}

	return nil
}

func (r *AddonReconciler) reconcileAddonAppDelete(
	ctx context.Context,
	log logr.Logger,
	clusterClient client.Client,
	addonName string) error {

	app := &kappctrl.App{
		ObjectMeta: metav1.ObjectMeta{
			Name:      addonName,
			Namespace: constants.TKG_ADDONS_APP_NAMESPACE,
		},
	}

	if err := clusterClient.Delete(ctx, app); err != nil {
		log.Error(err, "Error deleting app")
		return err
	}

	return nil
}

func (r *AddonReconciler) reconcileAddonAppNormal(
	ctx context.Context,
	log logr.Logger,
	clusterClient client.Client,
	addonSecret *corev1.Secret,
	addonConfig *bomv1alpha1.BomAddon) error {

	addonName := util.GetAddonNameFromAddonSecret(addonSecret)

	app := &kappctrl.App{
		ObjectMeta: metav1.ObjectMeta{
			Name:      addonName,
			Namespace: constants.TKG_ADDONS_APP_NAMESPACE,
		},
	}

	appMutateFn := func() error {
		if app.ObjectMeta.Annotations == nil {
			app.ObjectMeta.Annotations = make(map[string]string)
		}

		app.ObjectMeta.Annotations[addonsv1alpha1.AddonTypeAnnotation] = fmt.Sprintf("%s/%s", addonConfig.Category, addonName)

		app.Spec.ServiceAccountName = constants.TKG_ADDONS_APP_SERVICE_ACCOUNT

		app.Spec.Fetch = []kappctrl.AppFetch{
			{
				Image: &kappctrl.AppFetchImage{
					URL: addonConfig.Image,
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
										Name: addonName,
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
				Kapp: &kappctrl.AppDeployKapp{},
			},
		}

		return nil
	}

	if _, err := controllerutil.CreateOrPatch(ctx, clusterClient, app, appMutateFn); err != nil {
		log.Error(err, "Error creating or patching addon")
		return err
	}

	return nil
}

func (r *AddonReconciler) reconcileAddonDelete(
	ctx context.Context,
	log logr.Logger,
	clusterClient client.Client,
	addonSecret *corev1.Secret) error {

	addonName := util.GetAddonNameFromAddonSecret(addonSecret)

	log = r.Log.WithValues(constants.ADDON_NAME_LOG_KEY, addonName)

	log.Info("Reconciling addon delete")

	if err := r.reconcileAddonAppDelete(ctx, log, clusterClient, addonName); err != nil {
		log.Error(err, "Error reconcling addon app delete")
		return err
	}

	if err := r.reconcileAddonDataValuesSecretDelete(ctx, log, clusterClient, addonName); err != nil {
		log.Error(err, "Error reconciling addon data values secret delete")
		return err
	}

	return nil
}

func (r *AddonReconciler) reconcileAddonNormal(
	ctx context.Context,
	log logr.Logger,
	clusterClient client.Client,
	addonSecret *corev1.Secret,
	addonConfig *bomv1alpha1.BomAddon) error {

	addonName := util.GetAddonNameFromAddonSecret(addonSecret)

	log = r.Log.WithValues(constants.ADDON_NAME_LOG_KEY, addonName)

	log.Info("Reconciling addon")

	if err := r.reconcileAddonNamespace(ctx, log, clusterClient); err != nil {
		log.Error(err, "Error reconciling addon namespace")
		return err
	}

	if err := r.reconcileAddonServiceAccount(ctx, log, clusterClient); err != nil {
		log.Error(err, "Error reconciling addon service account")
		return err
	}

	if err := r.reconcileAddonRole(ctx, log, clusterClient); err != nil {
		log.Error(err, "Error reconciling addon roles and role bindings")
		return err
	}

	if err := r.reconcileAddonDataValuesSecretNormal(ctx, log, clusterClient, addonSecret); err != nil {
		log.Error(err, "Error reconciling addon data values secret")
		return err
	}

	if err := r.reconcileAddonAppNormal(ctx, log, clusterClient, addonSecret, addonConfig); err != nil {
		log.Error(err, "Error reconciling addon app")
		return err
	}

	return nil
}
