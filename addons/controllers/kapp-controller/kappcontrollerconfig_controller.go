// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package controllers implements k8s controller functionality for kapp-controller config CRD.
package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"gopkg.in/yaml.v2"
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
	runv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
)

// KappControllerConfigReconciler reconciles a KappControllerConfig object
type KappControllerConfigReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Config addonconfig.KappControllerConfigControllerConfig
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
		if apierrors.IsNotFound(err) {
			r.Log.Info("kappControllerConfig resource not found")
			return ctrl.Result{}, nil
		}
		r.Log.Error(err, "Unable to fetch kappControllerConfig resource")
		return ctrl.Result{}, err
	}

	annotations := kappControllerConfig.GetAnnotations()
	if _, ok := annotations[constants.TKGAnnotationTemplateConfig]; ok {
		r.Log.Info(fmt.Sprintf("resource '%v' is a config template. Skipping reconciling", req.NamespacedName))
		return ctrl.Result{}, nil
	}

	// Deepcopy to prevent client-go cache conflict
	kappControllerConfig = kappControllerConfig.DeepCopy()

	cluster, err := cutil.GetOwnerCluster(ctx, r.Client, kappControllerConfig, req.Namespace, constants.KappControllerDefaultRefName)
	if err != nil {
		if apierrors.IsNotFound(err) && cluster != nil {
			r.Log.Info(fmt.Sprintf("'%s/%s' is listed as owner reference but could not be found",
				cluster.Namespace, cluster.Name))
			return ctrl.Result{}, nil
		}
		r.Log.Info("could not determine owner cluster")
		return ctrl.Result{}, err
	}

	if retResult, err := r.ReconcileKappControllerConfig(ctx, kappControllerConfig, cluster, log); err != nil {
		log.Error(err, "unable to reconcile kappControllerConfig")
		return retResult, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KappControllerConfigReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager, options controller.Options) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&runv1alpha3.KappControllerConfig{}).
		Watches(
			&source.Kind{Type: &clusterapiv1beta1.Cluster{}},
			handler.EnqueueRequestsFromMapFunc(r.ClusterToKappControllerConfig),
		).
		WithOptions(options).
		WithEventFilter(predicates.ConfigOfKindWithoutAnnotation(constants.TKGAnnotationTemplateConfig, constants.KappControllerConfigKind, r.Config.SystemNamespace, r.Log)).
		Complete(r)
}

// ReconcileKappControllerConfig reconciles KappControllerConfig CR
func (r *KappControllerConfigReconciler) ReconcileKappControllerConfig(
	ctx context.Context,
	kappControllerConfig *runv1alpha3.KappControllerConfig,
	cluster *clusterapiv1beta1.Cluster,
	log logr.Logger) (_ ctrl.Result, retErr error) {

	patchHelper, err := clusterapipatchutil.NewHelper(kappControllerConfig, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}
	// Patch KappControllerConfig before returning the function
	defer func() {
		log.Info("Patching kappControllerConfig")

		if err := patchHelper.Patch(ctx, kappControllerConfig); err != nil {
			log.Error(err, "Error patching kappControllerConfig")
			retErr = err
		}
	}()

	if err := r.ReconcileKappControllerConfigNormal(ctx, kappControllerConfig, cluster, log); err != nil {
		log.Error(err, "Error reconciling kappControllerConfig")
		return ctrl.Result{}, err
	}

	log.Info("Successfully reconciled kappControllerConfig")

	return ctrl.Result{}, nil
}

// ReconcileKappControllerConfigNormal reconciles KappControllerConfig CR
func (r *KappControllerConfigReconciler) ReconcileKappControllerConfigNormal(
	ctx context.Context,
	kappControllerConfig *runv1alpha3.KappControllerConfig,
	cluster *clusterapiv1beta1.Cluster,
	log logr.Logger) (retErr error) {

	// add owner reference to kappControllerConfig
	ownerReference := metav1.OwnerReference{
		APIVersion: clusterapiv1beta1.GroupVersion.String(),
		Kind:       cluster.Kind,
		Name:       cluster.Name,
		UID:        cluster.UID,
	}

	if !clusterapiutil.HasOwnerRef(kappControllerConfig.OwnerReferences, ownerReference) {
		log.Info("Adding owner reference to kappControllerConfig")
		kappControllerConfig.OwnerReferences = clusterapiutil.EnsureOwnerRef(kappControllerConfig.OwnerReferences, ownerReference)
	}

	if err := r.ReconcileKappControllerConfigDataValue(ctx, kappControllerConfig, cluster, log); err != nil {
		log.Error(err, "Error creating kappControllerConfig data value secret")
		return err
	}

	// update status.secretRef
	dataValueSecretName := util.GenerateDataValueSecretName(cluster.Name, constants.KappControllerAddonName)
	kappControllerConfig.Status.SecretRef = dataValueSecretName

	return nil
}

// ReconcileKappControllerConfigDataValue reconciles KappControllerConfig data values secret
func (r *KappControllerConfigReconciler) ReconcileKappControllerConfigDataValue(
	ctx context.Context,
	kappControllerConfig *runv1alpha3.KappControllerConfig,
	cluster *clusterapiv1beta1.Cluster,
	log logr.Logger) (retErr error) {

	dataValuesSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GenerateDataValueSecretName(cluster.Name, constants.KappControllerAddonName),
			Namespace: cluster.Namespace,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: clusterapiv1beta1.GroupVersion.String(),
				Kind:       cluster.Kind,
				Name:       cluster.Name,
				UID:        cluster.UID,
			}},
		},
	}

	dataValuesSecretMutateFn := func() error {
		dataValuesSecret.Type = corev1.SecretTypeOpaque
		dataValuesSecret.StringData = make(map[string]string)

		// marshall the yaml contents
		kappConfig, err := mapKappControllerConfigSpec(cluster, kappControllerConfig)
		if err != nil {
			return err
		}

		dataValueYamlBytes, err := yaml.Marshal(kappConfig)
		if err != nil {
			return err
		}

		dataValuesSecret.StringData[constants.TKGDataValueFileName] = string(dataValueYamlBytes)

		return nil
	}

	result, err := controllerutil.CreateOrPatch(ctx, r.Client, dataValuesSecret, dataValuesSecretMutateFn)
	if err != nil {
		log.Error(err, "Error creating or patching kappControllerConfig data values secret")
		return err
	}

	log.Info(fmt.Sprintf("Resource %s data values secret %s", constants.KappControllerAddonName, result))

	return nil
}
