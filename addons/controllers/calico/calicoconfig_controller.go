// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package controllers implements k8s controller functionality for calico.
package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	yaml "gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	clusterapiutil "sigs.k8s.io/cluster-api/util"
	clusterapipatchutil "sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	addonconfig "github.com/vmware-tanzu/tanzu-framework/addons/pkg/config"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	"github.com/vmware-tanzu/tanzu-framework/addons/predicates"
	cniv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cni/v1alpha1"
)

// CalicoConfigReconciler reconciles CalicoConfig resource
type CalicoConfigReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Ctx    context.Context
	Config addonconfig.CalicoConfigControllerConfig
}

//+kubebuilder:rbac:groups=cni.tanzu.vmware.com,resources=calicoconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cni.tanzu.vmware.com,resources=calicoconfigs/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *CalicoConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log = r.Log.WithValues("CalicoConfig", req.NamespacedName)

	r.Log.Info("Start reconciliation")

	// fetch CalicoConfig resource, ignore not-found errors
	calicoConfig := &cniv1alpha1.CalicoConfig{}
	if err := r.Client.Get(ctx, req.NamespacedName, calicoConfig); err != nil {
		if apierrors.IsNotFound(err) {
			r.Log.Info("CalicoConfig resource not found")
			return ctrl.Result{}, nil
		}
		r.Log.Error(err, "Unable to fetch CalicoConfig resource")
		return ctrl.Result{}, err
	}

	// deep copy CalicoConfig to avoid issues if in the future other controllers where interacting with the same copy
	calicoConfig = calicoConfig.DeepCopy()

	// config resources are expected to have the same name as the cluster. However, we ideally try to read the cluster name from the owner reference of the addon config object
	clusterNamespacedName := req.NamespacedName
	cluster := &clusterapiv1beta1.Cluster{}
	for _, ownerRef := range calicoConfig.GetOwnerReferences() {
		if ownerRef.Kind == constants.ClusterKind {
			clusterNamespacedName.Name = ownerRef.Name
			break
		}
	}

	// verify that the cluster related to config is present
	if err := r.Client.Get(ctx, clusterNamespacedName, cluster); err != nil {
		if apierrors.IsNotFound(err) {
			r.Log.Info(fmt.Sprintf("Cluster resource '%s/%s' not found", clusterNamespacedName.Namespace, clusterNamespacedName.Name))
			return ctrl.Result{}, nil
		}
		r.Log.Error(err, fmt.Sprintf("Unable to fetch cluster '%s/%s'", clusterNamespacedName.Namespace, clusterNamespacedName.Name))
		return ctrl.Result{}, err
	}

	// reconcile CalicoConfig resource
	if retResult, err := r.ReconcileCalicoConfig(calicoConfig, cluster); err != nil {
		r.Log.Error(err, "Unable to reconcile CalicoConfig")
		return retResult, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CalicoConfigReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	r.Ctx = ctx
	return ctrl.NewControllerManagedBy(mgr).
		For(&cniv1alpha1.CalicoConfig{}).
		WithOptions(options).
		WithEventFilter(predicates.ConfigOfKindWithoutAnnotation(constants.TKGAnnotationTemplateConfig, constants.CalicoConfigKind, r.Config.SystemNamespace, r.Log)).
		Complete(r)
}

// ReconcileCalicoConfig reconciles CalicoConfig CR
func (r *CalicoConfigReconciler) ReconcileCalicoConfig(
	calicoConfig *cniv1alpha1.CalicoConfig,
	cluster *clusterapiv1beta1.Cluster) (_ ctrl.Result, retErr error) {

	patchHelper, err := clusterapipatchutil.NewHelper(calicoConfig, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}
	// patch CalicoConfig before returning the function
	defer func() {
		r.Log.Info("Patching CalicoConfig")
		if err := patchHelper.Patch(r.Ctx, calicoConfig); err != nil {
			r.Log.Error(err, "Error patching CalicoConfig")
			retErr = err
		}
		r.Log.Info("Successfully patched CalicoConfig")
	}()

	// if CalicoConfig is marked for deletion, then no reconciliation is needed
	if !calicoConfig.GetDeletionTimestamp().IsZero() {
		return ctrl.Result{}, nil
	}

	// reconcile CalicoConfig by creating the data values secret for CalicoConfig
	if err := r.ReconcileCalicoConfigNormal(calicoConfig, cluster); err != nil {
		r.Log.Error(err, "Error reconciling CalicoConfig to create/patch data values secret")
		return ctrl.Result{}, err
	}

	r.Log.Info("Successfully reconciled CalicoConfig")
	return ctrl.Result{}, nil
}

// ReconcileCalicoConfigNormal reconciles CalicoConfig by creating/patching data values secret
func (r *CalicoConfigReconciler) ReconcileCalicoConfigNormal(
	calicoConfig *cniv1alpha1.CalicoConfig,
	cluster *clusterapiv1beta1.Cluster) (retErr error) {

	// add owner reference to CalicoConfig if not already added by TanzuClusterBootstrap Controller
	ownerReference := metav1.OwnerReference{
		APIVersion: clusterapiv1beta1.GroupVersion.String(),
		Kind:       cluster.Kind,
		Name:       cluster.Name,
		UID:        cluster.UID,
	}
	if !clusterapiutil.HasOwnerRef(calicoConfig.OwnerReferences, ownerReference) {
		r.Log.Info("Adding owner reference to CalicoConfig")
		calicoConfig.OwnerReferences = clusterapiutil.EnsureOwnerRef(calicoConfig.OwnerReferences, ownerReference)
	}

	// create/patch data values secret for CalicoConfig
	secretNamespacedName := types.NamespacedName{
		Name:      util.GenerateDataValueSecretName(cluster.Name, constants.CalicoAddonName),
		Namespace: calicoConfig.Namespace,
	}
	if err := r.ReconcileCalicoDataValuesSecret(calicoConfig, cluster, secretNamespacedName); err != nil {
		return err
	}

	// add the name of the data values secret in the CalicoConfig Status field
	calicoConfig.Status.SecretRef = secretNamespacedName.Name

	return nil
}

func (r *CalicoConfigReconciler) ReconcileCalicoDataValuesSecret(
	calicoConfig *cniv1alpha1.CalicoConfig,
	cluster *clusterapiv1beta1.Cluster,
	secretNamespacedName types.NamespacedName) error {

	// prepare data values secret for CalicoConfig
	calicoDataValuesSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretNamespacedName.Name,
			Namespace: secretNamespacedName.Namespace,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: clusterapiv1beta1.GroupVersion.String(),
				Kind:       cluster.Kind,
				Name:       cluster.Name,
				UID:        cluster.UID,
			}},
		},
		Type: corev1.SecretTypeOpaque,
	}

	calicoDataValuesSecretMutateFn := func() error {
		calicoDataValuesSecret.Data = make(map[string][]byte)

		calicoConfigYaml, err := mapCalicoConfigSpec(cluster, calicoConfig)
		if err != nil {
			return err
		}

		dataValueYamlBytes, err := yaml.Marshal(calicoConfigYaml)
		if err != nil {
			r.Log.Error(err, "Error marshaling CalicoConfig to Yaml")
			return err
		}
		calicoDataValuesSecret.Data[constants.TKGDataValueFileName] = dataValueYamlBytes
		return nil
	}

	// create/patch the data values secret for CalicoConfig
	result, err := controllerutil.CreateOrPatch(r.Ctx, r.Client, calicoDataValuesSecret, calicoDataValuesSecretMutateFn)
	if err != nil {
		r.Log.Error(err, "Error creating or patching CalicoConfig data values secret")
		return err
	}
	r.Log.Info(fmt.Sprintf("Resource '%s' data values secret '%s'", constants.CalicoAddonName, result))

	return nil
}
