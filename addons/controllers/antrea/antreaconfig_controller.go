// Copyright YEAR VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"

	clusterapipatchutil "sigs.k8s.io/cluster-api/util/patch"

	clusterapiutil "sigs.k8s.io/cluster-api/util"

	yaml "gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cniv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cni/v1alpha1"
)

// AntreaConfigReconciler reconciles a AntreaConfig object
type AntreaConfigReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=addons.tanzu.vmware.com,resources=antreaconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=addons.tanzu.vmware.com,resources=antreaconfigs/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// In Reconcile function, we compare the state specified by
// the AntreaConfig object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.

func (r *AntreaConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("antreaconfig", req.NamespacedName)

	// get antrea config object
	antreaConfig := &cniv1alpha1.AntreaConfig{}
	if err := r.Client.Get(ctx, req.NamespacedName, antreaConfig); err != nil {
		log.Error(err, "unable to fetch AntreaConfig")
		return ctrl.Result{}, err
	}

	// get the parent cluster name from owner reference
	// if the owner reference doesn't exist, use the same name as config CRD
	clusterNamespacedName := req.NamespacedName
	cluster := &clusterapiv1beta1.Cluster{}
	for _, owner := range antreaConfig.OwnerReferences {
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

	if retResult, err := r.ReconcileAntreaConfig(ctx, log, cluster, antreaConfig); err != nil {
		log.Error(err, "unable to reconcile AntreaConfig")
		return retResult, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AntreaConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cniv1alpha1.AntreaConfig{}).
		Complete(r)
}

// ReconcileAntreaConfigNormal reconciles AntreaConfig data values secrets
func (r *AntreaConfigReconciler) ReconcileAntreaConfig(
	ctx context.Context,
	log logr.Logger,
	cluster *clusterapiv1beta1.Cluster,
	antreaConfig *cniv1alpha1.AntreaConfig) (_ ctrl.Result, retErr error) {
	var patchCRD bool

	patchHelper, err := clusterapipatchutil.NewHelper(antreaConfig, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}
	// Patch AntreaConfig before returning the function
	defer func() {
		// patchCRD will be true if finalizer or owner reference is added or deleted
		if patchCRD {
			log.Info("Patching AntreaConfig")

			if err := patchHelper.Patch(ctx, antreaConfig.DeepCopy()); err != nil {
				log.Error(err, "Error patching AntreaConfig")
				retErr = err
			}
		}
	}()

	// If AntreaConfig is marked for deletion then delete the data value secret
	if !antreaConfig.GetDeletionTimestamp().IsZero() {
		log.Info("Deleting antreaConfig")
		err := r.ReconcileAntreaConfigDelete(ctx, log, cluster, antreaConfig, &patchCRD)
		if err != nil {
			log.Error(err, "Error reconciling AntreaConfig delete")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	if err := r.ReconcileAntreaConfigNormal(ctx, log, cluster, antreaConfig, &patchCRD); err != nil {
		log.Error(err, "Error reconciling AntreaConfig to create data value secret")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// ReconcileAntreaConfigNormal reconciles AntreaConfig data values secret
func (r *AntreaConfigReconciler) ReconcileAntreaConfigNormal(
	ctx context.Context,
	log logr.Logger,
	cluster *clusterapiv1beta1.Cluster,
	antreaConfig *cniv1alpha1.AntreaConfig,
	patchCRD *bool) (retErr error) {

	// Add finalizer to addon secret
	*patchCRD = util.AddFinalizerToCRD(log, constants.AntreaAddonName, antreaConfig)

	// add owner reference to antreaConfig
	ownerReference := metav1.OwnerReference{
		APIVersion:         clusterapiv1beta1.GroupVersion.String(),
		Kind:               cluster.Kind,
		Name:               cluster.Name,
		UID:                cluster.UID,
		Controller:         pointer.BoolPtr(true),
		BlockOwnerDeletion: pointer.BoolPtr(true),
	}

	if !clusterapiutil.HasOwnerRef(antreaConfig.OwnerReferences, ownerReference) {
		log.Info("Adding owner reference to AntreaConfig")
		antreaConfig.OwnerReferences = clusterapiutil.EnsureOwnerRef(antreaConfig.OwnerReferences, ownerReference)
		*patchCRD = true
	}

	if err := r.ReconcileAntreaConfigDataValue(ctx, log, cluster, antreaConfig); err != nil {
		log.Error(err, "Error creating antreaConfig data value secret")
		return err
	}

	// update status.secretRef
	antreaConfig.Status.SecretRef.Name = util.GenerateDataValueSecretNameFromAddonAndClusterNames(cluster.Name, constants.AntreaAddonName)
	*patchCRD = true

	return nil
}

// ReconcileAntreaConfigDataValue reconciles AntreaConfig data values secret
func (r *AntreaConfigReconciler) ReconcileAntreaConfigDataValue(
	ctx context.Context,
	log logr.Logger,
	cluster *clusterapiv1beta1.Cluster,
	antreaConfig *cniv1alpha1.AntreaConfig) (retErr error) {

	antreaDataValuesSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GenerateDataValueSecretNameFromAddonAndClusterNames(antreaConfig.Name, constants.AntreaAddonName),
			Namespace: antreaConfig.Namespace,
		},
	}

	antreaDataValuesSecretMutateFn := func() error {
		antreaDataValuesSecret.Type = corev1.SecretTypeOpaque
		antreaDataValuesSecret.Data = map[string][]byte{}

		// marshall the yaml contents
		antreaConfigYaml, err := mapAntreaConfigSpec(cluster, antreaConfig)
		if err != nil {
			return err
		}

		yamlBytes, err := yaml.Marshal(antreaConfigYaml)
		if err != nil {
			return err
		}

		dataValueBytes := append([]byte(constants.TKGDataValueFormatString), yamlBytes...)
		antreaDataValuesSecret.Data[constants.TKGDataValueFileName] = dataValueBytes

		return nil
	}

	result, err := controllerutil.CreateOrPatch(ctx, r.Client, antreaDataValuesSecret, antreaDataValuesSecretMutateFn)
	if err != nil {
		log.Error(err, "Error creating or patching antrea data values secret")
		return err
	}

	log.Info(fmt.Sprintf("Resource %s data values secret %s", constants.AntreaAddonName, result))

	return nil
}

// ReconcileAntreaConfigDelete reconciles AntreaConfig deletion
func (r *AntreaConfigReconciler) ReconcileAntreaConfigDelete(
	ctx context.Context,
	log logr.Logger,
	cluster *clusterapiv1beta1.Cluster,
	antreaConfig *cniv1alpha1.AntreaConfig,
	patchCRD *bool) (retErr error) {

	// delete data value secret
	addonDataValuesSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GenerateDataValueSecretNameFromAddonAndClusterNames(antreaConfig.Name, constants.AntreaAddonName),
			Namespace: antreaConfig.Namespace,
		},
	}
	if err := r.Client.Delete(ctx, addonDataValuesSecret); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("AntreaConfig data values secret not found")
			return nil
		}
		log.Error(err, "Error deleting AntreaConfig data values secret")
		return err
	}

	// Remove finalizer from addon secret
	*patchCRD = util.RemoveFinalizerFromCRD(log, constants.AntreaAddonName, antreaConfig)

	return nil
}
