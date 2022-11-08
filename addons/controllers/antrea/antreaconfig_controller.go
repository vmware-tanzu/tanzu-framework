// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package controllers implements k8s controller functionality for antrea.
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
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	clusterapiutil "sigs.k8s.io/cluster-api/util"
	clusterapipatchutil "sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	cutil "github.com/vmware-tanzu/tanzu-framework/addons/controllers/utils"
	addonconfig "github.com/vmware-tanzu/tanzu-framework/addons/pkg/config"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	"github.com/vmware-tanzu/tanzu-framework/addons/predicates"
	cniv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/addonconfigs/cni/v1alpha1"
)

// AntreaConfigReconciler reconciles a AntreaConfig object
type AntreaConfigReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Config addonconfig.AntreaConfigControllerConfig
}

// +kubebuilder:rbac:groups=addons.tanzu.vmware.com,resources=antreaconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=addons.tanzu.vmware.com,resources=antreaconfigs/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// In Reconcile function, we compare the state specified by
// the AntreaConfig object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.

func (r *AntreaConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("antreaconfig", req.NamespacedName)

	r.Log.Info("Start reconciliation")

	// fetch AntreaConfig resource, ignore not-found errors
	antreaConfig := &cniv1alpha1.AntreaConfig{}
	if err := r.Client.Get(ctx, req.NamespacedName, antreaConfig); err != nil {
		if apierrors.IsNotFound(err) {
			r.Log.Info(fmt.Sprintf("AntreaConfig resource '%v' not found", req.NamespacedName))
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	annotations := antreaConfig.GetAnnotations()
	if _, ok := annotations[constants.TKGAnnotationTemplateConfig]; ok {
		log.Info(fmt.Sprintf("resource '%v' is a config template. Skipping reconciling", req.NamespacedName))
		return ctrl.Result{}, nil
	}

	// deep copy AntreaConfig to avoid issues if in the future other controllers where interacting with the same copy
	antreaConfig = antreaConfig.DeepCopy()

	cluster, err := cutil.GetOwnerCluster(ctx, r.Client, antreaConfig, req.Namespace, constants.AntreaDefaultRefName)

	if err != nil {
		if apierrors.IsNotFound(err) && cluster != nil {
			log.Info(fmt.Sprintf("'%s/%s' is listed as owner reference but could not be found",
				cluster.Namespace, cluster.Name))
			return ctrl.Result{}, nil
		}
		log.Error(err, "could not determine owner cluster")
		return ctrl.Result{}, err
	}

	if retResult, err := r.ReconcileAntreaConfig(ctx, antreaConfig, cluster, log); err != nil {
		log.Error(err, "unable to reconcile AntreaConfig")
		return retResult, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AntreaConfigReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cniv1alpha1.AntreaConfig{}).
		WithOptions(options).
		Watches(
			&source.Kind{Type: &clusterapiv1beta1.Cluster{}},
			handler.EnqueueRequestsFromMapFunc(r.ClusterToAntreaConfig),
		).
		WithEventFilter(predicates.ConfigOfKindWithoutAnnotation(constants.TKGAnnotationTemplateConfig, constants.AntreaConfigKind, r.Config.SystemNamespace, r.Log)).
		Complete(r)
}

// ReconcileAntreaConfig reconciles AntreaConfig CR
func (r *AntreaConfigReconciler) ReconcileAntreaConfig(
	ctx context.Context,
	antreaConfig *cniv1alpha1.AntreaConfig,
	cluster *clusterapiv1beta1.Cluster,
	log logr.Logger) (_ ctrl.Result, retErr error) {

	patchHelper, err := clusterapipatchutil.NewHelper(antreaConfig, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}
	// Patch AntreaConfig before returning the function
	defer func() {
		log.Info("Patching AntreaConfig")
		if err := patchHelper.Patch(ctx, antreaConfig); err != nil {
			log.Error(err, "Error patching AntreaConfig")
			retErr = err
		}
		log.Info("Successfully patched AntreaConfig")
	}()

	// If AntreaConfig is marked for deletion, then no reconciliation is needed
	if !antreaConfig.GetDeletionTimestamp().IsZero() {
		return ctrl.Result{}, nil
	}

	if err := r.ReconcileAntreaConfigNormal(ctx, antreaConfig, cluster, log); err != nil {
		log.Error(err, "Error reconciling AntreaConfig to create data value secret")
		return ctrl.Result{}, err
	}

	log.Info("Successfully reconciled AntreaConfig")
	return ctrl.Result{}, nil
}

// ReconcileAntreaConfigNormal reconciles AntreaConfig by creating/patching data values secret
func (r *AntreaConfigReconciler) ReconcileAntreaConfigNormal(
	ctx context.Context,
	antreaConfig *cniv1alpha1.AntreaConfig,
	cluster *clusterapiv1beta1.Cluster,
	log logr.Logger) (retErr error) {

	// add owner reference to antreaConfig
	ownerReference := metav1.OwnerReference{
		APIVersion: clusterapiv1beta1.GroupVersion.String(),
		Kind:       cluster.Kind,
		Name:       cluster.Name,
		UID:        cluster.UID,
	}

	if !clusterapiutil.HasOwnerRef(antreaConfig.OwnerReferences, ownerReference) {
		log.Info("Adding owner reference to AntreaConfig")
		antreaConfig.OwnerReferences = clusterapiutil.EnsureOwnerRef(antreaConfig.OwnerReferences, ownerReference)
	}

	if err := r.ReconcileAntreaConfigDataValue(ctx, antreaConfig, cluster, log); err != nil {
		log.Error(err, "Error creating antreaConfig data value secret")
		return err
	}

	// update status.secretRef
	dataValueSecretName := util.GenerateDataValueSecretName(cluster.Name, constants.AntreaAddonName)
	antreaConfig.Status.SecretRef = dataValueSecretName

	return nil
}

// ReconcileAntreaConfigDataValue reconciles AntreaConfig data values secret
func (r *AntreaConfigReconciler) ReconcileAntreaConfigDataValue(
	ctx context.Context,
	antreaConfig *cniv1alpha1.AntreaConfig,
	cluster *clusterapiv1beta1.Cluster,
	log logr.Logger) (retErr error) {

	// prepare data values secret for AntreaConfig
	antreaDataValuesSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GenerateDataValueSecretName(cluster.Name, constants.AntreaAddonName),
			Namespace: antreaConfig.Namespace,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: clusterapiv1beta1.GroupVersion.String(),
				Kind:       cluster.Kind,
				Name:       cluster.Name,
				UID:        cluster.UID,
			}},
		},
	}

	antreaDataValuesSecretMutateFn := func() error {
		antreaDataValuesSecret.Type = corev1.SecretTypeOpaque
		antreaDataValuesSecret.StringData = make(map[string]string)

		// marshall the yaml contents
		antreaConfigYaml, err := mapAntreaConfigSpec(cluster, antreaConfig)
		if err != nil {
			return err
		}

		dataValueYamlBytes, err := yaml.Marshal(antreaConfigYaml)
		if err != nil {
			log.Error(err, "Error marshaling AntreaConfig to Yaml")
			return err
		}

		antreaDataValuesSecret.StringData[constants.TKGDataValueFileName] = string(dataValueYamlBytes)

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
