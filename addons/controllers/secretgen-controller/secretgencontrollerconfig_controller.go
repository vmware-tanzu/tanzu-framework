// Copyright YEAR VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"

	"k8s.io/utils/pointer"

	clusterapipatchutil "sigs.k8s.io/cluster-api/util/patch"

	clusterapiutil "sigs.k8s.io/cluster-api/util"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"

	"gopkg.in/yaml.v3"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	addonsv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/addons/v1alpha1"
)

// SecretGenControllerConfigReconciler reconciles a SecretGenControllerConfig object
type SecretGenControllerConfigReconciler struct {
	Client client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//SecretGenControllerConfigSpecYaml is used for yaml marshalling for secretGen controller config
type SecretGenControllerConfigSpecYaml struct {
	Namespace       string `yaml:"namespace,omitempty"`
	CreateNamespace bool   `yaml:"createNamespace,omitempty"`
}

//+kubebuilder:rbac:groups=addons.tanzu.vmware.com,resources=secretgencontrollerconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=addons.tanzu.vmware.com,resources=secretgencontrollerconfigs/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *SecretGenControllerConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("secretgencontrollerconfig", req.NamespacedName)

	// get secret-gen config object
	secretGenConfig := &addonsv1alpha1.SecretGenControllerConfig{}
	if err := r.Client.Get(ctx, req.NamespacedName, secretGenConfig); err != nil {
		log.Error(err, "unable to fetch SecretGenControllerConfig")
		return ctrl.Result{}, err
	}

	// verify that the cluster related to config is present
	cluster := &clusterapiv1beta1.Cluster{}
	if err := r.Client.Get(ctx, req.NamespacedName, cluster); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Cluster not found")
			return ctrl.Result{}, nil
		}

		log.Error(err, "unable to fetch cluster")
		return ctrl.Result{}, err
	}

	// Get client for the cluster that we deploy data value secret
	remoteClient, err := util.GetClusterClient(ctx, r.Client, r.Scheme, clusterapiutil.ObjectKey(cluster))
	if err != nil {
		log.Error(err, "Error getting remote cluster client")
		return ctrl.Result{}, err
	}

	if retResult, err := r.ReconcileSecretGenControllerConfig(ctx, log, cluster, remoteClient, secretGenConfig); err != nil {
		log.Error(err, "unable to reconcile SecretGenControllerConfig")
		return retResult, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SecretGenControllerConfigReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&addonsv1alpha1.SecretGenControllerConfig{}).
		Complete(r)
}

// ReconcileSecretGenControllerConfigNormal reconciles SecretGenControllerConfig data values secrets
func (r *SecretGenControllerConfigReconciler) ReconcileSecretGenControllerConfig(
	ctx context.Context,
	log logr.Logger,
	cluster *clusterapiv1beta1.Cluster,
	clusterClient client.Client,
	secretGenConfig *addonsv1alpha1.SecretGenControllerConfig) (_ ctrl.Result, retErr error) {

	var (
		patchCRD bool
	)

	patchHelper, err := clusterapipatchutil.NewHelper(secretGenConfig, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}
	// Patch SecretGenControllerConfig before returning the function
	defer func() {
		// patchCRD will be true if finalizer or ownerrefence is added
		if patchCRD {
			log.Info("Patching SecretGenControllerConfig")

			if err := patchHelper.Patch(ctx, secretGenConfig.DeepCopy()); err != nil {
				log.Error(err, "Error patching SecretGenControllerConfig")
				retErr = err
			}
		}
	}()

	// If SecretGenControllerConfig is marked for deletion then delete the data value secret
	if !secretGenConfig.GetDeletionTimestamp().IsZero() {
		err := r.ReconcileSecretGenControllerConfigDelete(ctx, log, clusterClient, secretGenConfig, &patchCRD)
		if err != nil {
			log.Error(err, "Error reconciling SecretGenControllerConfig delete")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if err := r.ReconcileSecretGenControllerConfigNormal(ctx, log, cluster, clusterClient, secretGenConfig, &patchCRD); err != nil {
		log.Error(err, "Error reconciling SecretGenControllerConfig to create data value secret")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// ReconcileSecretGenControllerConfigNormal reconciles SecretGenControllerConfig data values secret
func (r *SecretGenControllerConfigReconciler) ReconcileSecretGenControllerConfigNormal(
	ctx context.Context,
	log logr.Logger,
	cluster *clusterapiv1beta1.Cluster,
	clusterClient client.Client,
	secretGenConfig *addonsv1alpha1.SecretGenControllerConfig,
	patchCRD *bool) (retErr error) {

	// Add finalizer to addon secret
	*patchCRD = util.AddFinalizerToCRD(log, constants.SecretGenControllerAddonName, secretGenConfig)

	// add owner reference to addon secret
	ownerReference := metav1.OwnerReference{
		APIVersion:         clusterapiv1beta1.GroupVersion.String(),
		Kind:               "Cluster",
		Name:               cluster.Name,
		UID:                cluster.UID,
		Controller:         pointer.BoolPtr(true),
		BlockOwnerDeletion: pointer.BoolPtr(true),
	}

	if !clusterapiutil.HasOwnerRef(secretGenConfig.OwnerReferences, ownerReference) {
		log.Info("Adding owner reference to SecretGenControllerConfig")
		secretGenConfig.OwnerReferences = clusterapiutil.EnsureOwnerRef(secretGenConfig.OwnerReferences, ownerReference)
		*patchCRD = true
	}

	addonDataValuesSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GenerateDataValueSecretNameFromAddonNames(secretGenConfig.Name, constants.SecretGenControllerAddonName),
			Namespace: secretGenConfig.Namespace,
		},
	}

	addonDataValuesSecretMutateFn := func() error {
		addonDataValuesSecret.Type = corev1.SecretTypeOpaque
		addonDataValuesSecret.Data = map[string][]byte{}

		// marshall the yaml contents
		secretGenConfigYaml := &SecretGenControllerConfigSpecYaml{Namespace: secretGenConfig.Spec.Namespace, CreateNamespace: secretGenConfig.Spec.CreateNamespace}

		yamlBytes, err := yaml.Marshal(secretGenConfigYaml)
		if err != nil {
			return err
		}

		dataValueBytes := append([]byte(constants.TKGDataValueFormatString), yamlBytes...)
		addonDataValuesSecret.Data[constants.DataValueFileName] = dataValueBytes

		return nil
	}

	result, err := controllerutil.CreateOrPatch(ctx, clusterClient, addonDataValuesSecret, addonDataValuesSecretMutateFn)
	if err != nil {
		log.Error(err, "Error creating or patching addon data values secret")
		return err
	}

	log.Info(fmt.Sprintf("Resource %s data values secret %s", constants.SecretGenControllerAddonName, result))

	return nil
}

// ReconcileSecretGenControllerConfigDelete reconciles SecretGenControllerConfig deletion
func (r *SecretGenControllerConfigReconciler) ReconcileSecretGenControllerConfigDelete(
	ctx context.Context,
	log logr.Logger,
	clusterClient client.Client,
	secretGenConfig *addonsv1alpha1.SecretGenControllerConfig,
	patchCRD *bool) (retErr error) {

	// delete data value secret
	addonDataValuesSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GenerateDataValueSecretNameFromAddonNames(secretGenConfig.Name, constants.SecretGenControllerAddonName),
			Namespace: secretGenConfig.Namespace,
		},
	}
	if err := clusterClient.Delete(ctx, addonDataValuesSecret); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("SecretGenControllerConfig data values secret not found")
			return nil
		}
		log.Error(err, "Error deleting SecretGenControllerConfig data values secret")
		return err
	}

	// Remove finalizer from addon secret
	*patchCRD = util.RemoveFinalizerFromCRD(log, constants.SecretGenControllerAddonName, secretGenConfig)

	return nil
}
