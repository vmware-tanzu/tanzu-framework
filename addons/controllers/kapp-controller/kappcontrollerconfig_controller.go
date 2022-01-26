// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"

	"gopkg.in/yaml.v2"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	corev1 "k8s.io/api/core/v1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"

	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	clusterapiutil "sigs.k8s.io/cluster-api/util"
	clusterapipatchutil "sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/go-logr/logr"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	runv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

const KappControllerAddonName = "kapp-controller"

// KappControllerConfigReconciler reconciles a KappControllerConfig object
type KappControllerConfigReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=run.tanzu.vmware.com,resources=kappcontrollerconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=run.tanzu.vmware.com,resources=kappcontrollerconfigs/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *KappControllerConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("kappControllerConfig", req.NamespacedName)

	// get kapp-controller config object
	kappControllerConfig := &runv1alpha3.KappControllerConfig{}
	if err := r.Client.Get(ctx, req.NamespacedName, kappControllerConfig); err != nil {
		log.Error(err, "unable to fetch kappControllerConfig")
		return ctrl.Result{}, err
	}

	// get the parent cluster name from owner reference
	// if the owner reference doesn't exist, use the same name as config CRD
	clusterNamespacedName := req.NamespacedName
	cluster := &clusterapiv1beta1.Cluster{}
	for _, owner := range kappControllerConfig.OwnerReferences {
		if owner.Kind == cluster.Kind {
			clusterNamespacedName.Name = owner.Name
			break
		}
	}

	// verify that the cluster related to config is present
	if err := r.Client.Get(ctx, clusterNamespacedName, cluster); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("Cluster not found")
			return ctrl.Result{}, nil
		}

		log.Error(err, "unable to fetch cluster")
		return ctrl.Result{}, err
	}

	if retResult, err := r.ReconcileKappControllerConfig(ctx, log, cluster, kappControllerConfig); err != nil {
		log.Error(err, "unable to reconcile kappControllerConfig")
		return retResult, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KappControllerConfigReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&runv1alpha3.KappControllerConfig{}).
		WithOptions(options).
		Complete(r)
}

// ReconcileKappControllerConfig reconciles KappControllerConfig CRD
func (r *KappControllerConfigReconciler) ReconcileKappControllerConfig(
	ctx context.Context,
	log logr.Logger,
	cluster *clusterapiv1beta1.Cluster,
	kappControllerConfig *runv1alpha3.KappControllerConfig) (_ ctrl.Result, retErr error) {

	var (
		patchConfig bool
	)

	patchHelper, err := clusterapipatchutil.NewHelper(kappControllerConfig, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}
	// Patch KappControllerConfig before returning the function
	defer func() {
		// patchConfig will be true if finalizer, ownerrefence or secretRef is added or deleted
		if patchConfig {
			log.Info("Patching kappControllerConfig")

			if err := patchHelper.Patch(ctx, kappControllerConfig.DeepCopy()); err != nil {
				log.Error(err, "Error patching kappControllerConfig")
				retErr = err
			}
		}
	}()

	// If KappControllerConfig is marked for deletion then delete the data value secret
	if !kappControllerConfig.GetDeletionTimestamp().IsZero() {
		log.Info("Deleting kappControllerConfig")
		err := r.ReconcileKappControllerConfigDelete(ctx, log, cluster, kappControllerConfig, &patchConfig)
		if err != nil {
			log.Error(err, "Error reconciling kappControllerConfig delete")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if err := r.ReconcileKappControllerConfigNormal(ctx, log, cluster, kappControllerConfig, &patchConfig); err != nil {
		log.Error(err, "Error reconciling kappControllerConfig")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// ReconcileKappControllerConfigNormal reconciles KappControllerConfig CRD
func (r *KappControllerConfigReconciler) ReconcileKappControllerConfigNormal(
	ctx context.Context,
	log logr.Logger,
	cluster *clusterapiv1beta1.Cluster,
	kappControllerConfig *runv1alpha3.KappControllerConfig,
	patchConfig *bool) (retErr error) {

	// Add finalizer to kappControllerConfig
	*patchConfig = util.AddFinalizerToCRD(log, KappControllerAddonName, kappControllerConfig)

	// add owner reference to kappControllerConfig
	ownerReference := metav1.OwnerReference{
		APIVersion:         clusterapiv1beta1.GroupVersion.String(),
		Kind:               cluster.Kind,
		Name:               cluster.Name,
		UID:                cluster.UID,
		BlockOwnerDeletion: pointer.BoolPtr(true),
	}

	if !clusterapiutil.HasOwnerRef(kappControllerConfig.OwnerReferences, ownerReference) {
		log.Info("Adding owner reference to kappControllerConfig")
		kappControllerConfig.OwnerReferences = clusterapiutil.EnsureOwnerRef(kappControllerConfig.OwnerReferences, ownerReference)
		*patchConfig = true
	}

	if err := r.ReconcileKappControllerConfigDataValue(ctx, log, cluster, kappControllerConfig); err != nil {
		log.Error(err, "Error creating kappControllerConfig data value secret")
		return err
	}

	// update status.secretRef
	dataValueSecretName := util.GenerateDataValueSecretNameFromAddonNames(cluster.Name, KappControllerAddonName)
	if kappControllerConfig.Status.SecretRef.Name != dataValueSecretName {
		kappControllerConfig.Status.SecretRef.Name = dataValueSecretName
		*patchConfig = true
	}

	return nil
}

// ReconcileKappControllerConfigDataValue reconciles KappControllerConfig data values secret
func (r *KappControllerConfigReconciler) ReconcileKappControllerConfigDataValue(
	ctx context.Context,
	log logr.Logger,
	cluster *clusterapiv1beta1.Cluster,
	kappControllerConfig *runv1alpha3.KappControllerConfig) (retErr error) {

	dataValuesSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GenerateDataValueSecretNameFromAddonNames(cluster.Name, KappControllerAddonName),
			Namespace: cluster.Namespace,
		},
	}

	dataValuesSecretMutateFn := func() error {
		dataValuesSecret.Type = corev1.SecretTypeOpaque
		dataValuesSecret.Data = map[string][]byte{}

		// marshall the yaml contents
		kappConfig, err := mapKappControllerConfigSpec(cluster, kappControllerConfig)
		if err != nil {
			return err
		}

		yamlBytes, err := yaml.Marshal(kappConfig)
		if err != nil {
			return err
		}

		dataValueBytes := append([]byte(constants.TKGDataValueFormatString), yamlBytes...)
		dataValuesSecret.Data[constants.TKGDataValueFileName] = dataValueBytes

		return nil
	}

	result, err := controllerutil.CreateOrPatch(ctx, r.Client, dataValuesSecret, dataValuesSecretMutateFn)
	if err != nil {
		log.Error(err, "Error creating or patching kappControllerConfig data values secret")
		return err
	}

	log.Info(fmt.Sprintf("Resource %s data values secret %s", KappControllerAddonName, result))

	return nil
}

// ReconcileKappControllerConfigDelete reconciles kappControllerConfig deletion
func (r *KappControllerConfigReconciler) ReconcileKappControllerConfigDelete(
	ctx context.Context,
	log logr.Logger,
	cluster *clusterapiv1beta1.Cluster,
	kappControllerConfig *runv1alpha3.KappControllerConfig,
	patchConfig *bool) (retErr error) {

	// delete data value secret
	addonDataValuesSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GenerateDataValueSecretNameFromAddonNames(cluster.Name, KappControllerAddonName),
			Namespace: kappControllerConfig.Namespace,
		},
	}
	if err := r.Client.Delete(ctx, addonDataValuesSecret); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("kappControllerConfig data values secret not found")
			return nil
		}
		log.Error(err, "Error deleting kappControllerConfig data values secret")
		return err
	}

	// Remove finalizer from addon secret
	*patchConfig = util.RemoveFinalizerFromCRD(log, KappControllerAddonName, kappControllerConfig)

	return nil
}
