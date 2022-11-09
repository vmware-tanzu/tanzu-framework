// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package controllers implements k8s controller functionality for kube-vip-cloud-provider config.
package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
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
	kvcpiv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/addonconfigs/cpi/v1alpha1"
)

// KubevipCPIConfigReconciler reconciles a KubevipCPIConfig object
type KubevipCPIConfigReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Config addonconfig.KubevipCPIConfigControllerConfig
}

//+kubebuilder:rbac:groups=cpi.tanzu.vmware.com,resources=KubevipCPIConfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cpi.tanzu.vmware.com,resources=KubevipCPIConfigs/status,verbs=get;update;patch

// Reconcile the KubevipCPIConfig CRD
func (r *KubevipCPIConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.Log.WithValues("KubevipCPIConfig", req.NamespacedName)

	logger.Info("Start reconciliation for KubevipCPIConfig")

	// fetch KubevipCPIConfig resource
	kvcpConfig := &kvcpiv1alpha1.KubevipCPIConfig{}
	if err := r.Client.Get(ctx, req.NamespacedName, kvcpConfig); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("KubevipCPIConfig resource not found")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Unable to fetch KubevipCPIConfig resource")
		return ctrl.Result{}, err
	}

	// deep copy KubevipCPIConfig to avoid issues if in the future other controllers where interacting with the same copy
	kvcpConfig = kvcpConfig.DeepCopy()

	cluster, err := cutil.GetOwnerCluster(ctx, r.Client, kvcpConfig, req.Namespace, constants.CPIDefaultRefName)
	if err != nil {
		if apierrors.IsNotFound(err) && cluster != nil {
			logger.Info(fmt.Sprintf("'%s/%s' is listed as owner reference but could not be found",
				cluster.Namespace, cluster.Name))
			return ctrl.Result{}, nil
		}
		logger.Error(err, "could not determine owner cluster")
		return ctrl.Result{}, err
	}

	if res, err := r.reconcileKubevipCPIConfig(ctx, kvcpConfig, cluster); err != nil {
		logger.Error(err, "Failed to reconcile KubevipCPIConfig")
		return res, err
	}
	return ctrl.Result{}, nil
}

// reconcileKubevipCPIConfig reconciles KubevipCPIConfig with its owner cluster
func (r *KubevipCPIConfigReconciler) reconcileKubevipCPIConfig(ctx context.Context, kvcpConfig *kvcpiv1alpha1.KubevipCPIConfig, cluster *clusterapiv1beta1.Cluster) (_ ctrl.Result, retErr error) {
	patchHelper, err := clusterapipatchutil.NewHelper(kvcpConfig, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}
	// patch KubevipCPIConfig before returning the function
	defer func() {
		r.Log.Info("Patching KubevipCPIConfig")
		if err := patchHelper.Patch(ctx, kvcpConfig); err != nil {
			r.Log.Error(err, "Error patching KubevipCPIConfig")
			retErr = err
		}
		r.Log.Info("Successfully patched KubevipCPIConfig")
	}()

	if !kvcpConfig.GetDeletionTimestamp().IsZero() {
		return ctrl.Result{}, nil
	}
	if err = r.reconcileKubevipCPIConfigNormal(ctx, kvcpConfig, cluster); err != nil {
		r.Log.Error(err, "Error reconciling KubevipCPIConfig to create/patch data values secret")
		return ctrl.Result{}, err
	}
	r.Log.Info("Successfully reconciled KubevipCPIConfig")
	return ctrl.Result{}, nil
}

// reconcileKubevipCPIConfigNormal triggers when a KubevipCPIConfig is not being deleted
// it ensures the owner reference of the KubevipCPIConfig and generates the data values secret for Kubevip CloudProvider
func (r *KubevipCPIConfigReconciler) reconcileKubevipCPIConfigNormal(ctx context.Context,
	kvcpConfig *kvcpiv1alpha1.KubevipCPIConfig, cluster *clusterapiv1beta1.Cluster) (retErr error) {
	// add owner reference to KubevipCPIConfig if not already added by TanzuClusterBootstrap Controller
	ownerReference := metav1.OwnerReference{
		APIVersion: clusterapiv1beta1.GroupVersion.String(),
		Kind:       cluster.Kind,
		Name:       cluster.Name,
		UID:        cluster.UID,
	}
	r.Log.Info("Ensure KubevipCPIConfig has the cluster as owner reference")
	if !clusterapiutil.HasOwnerRef(kvcpConfig.OwnerReferences, ownerReference) {
		r.Log.Info("Adding owner reference to KubevipCPIConfig")
		kvcpConfig.OwnerReferences = clusterapiutil.EnsureOwnerRef(kvcpConfig.OwnerReferences, ownerReference)
	}
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GenerateDataValueSecretName(cluster.Name, constants.KubevipCloudProviderAddonName),
			Namespace: kvcpConfig.Namespace,
		},
		Type: v1.SecretTypeOpaque,
	}
	secret.SetOwnerReferences([]metav1.OwnerReference{ownerReference})

	mutateFn := func() error {
		secret.StringData = make(map[string]string)
		kvcpConfigSpec, err := r.mapKubevipCPIConfigToDataValues(ctx, kvcpConfig, cluster)
		if err != nil {
			r.Log.Error(err, "Error while mapping KubevipCPIConfig to data values")
			return err
		}
		yamlBytes, err := kvcpConfigSpec.Serialize()
		if err != nil {
			r.Log.Error(err, "Error marshaling KubevipCPIConfig to Yaml")
			return err
		}
		secret.StringData[constants.TKGDataValueFileName] = string(yamlBytes)
		r.Log.Info("Mutated KubevipCPIConfig data values")
		return nil
	}
	result, err := controllerutil.CreateOrPatch(ctx, r.Client, secret, mutateFn)
	if err != nil {
		r.Log.Error(err, "Error creating or patching KubevipCPIConfig data values secret")
		return err
	}

	r.Log.Info(fmt.Sprintf("Resource '%s' data values secret '%s'", constants.KubevipCloudProviderAddonName, result))
	// update the secret reference in KubevipCPIConfig status
	kvcpConfig.Status.SecretRef = &secret.Name
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KubevipCPIConfigReconciler) SetupWithManager(_ context.Context, mgr ctrl.Manager, options controller.Options) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kvcpiv1alpha1.KubevipCPIConfig{}).
		WithOptions(options).
		WithEventFilter(predicates.ConfigOfKindWithoutAnnotation(constants.TKGAnnotationTemplateConfig, constants.KubevipCPIConfigKind, r.Config.SystemNamespace, r.Log)).
		Watches(
			&source.Kind{Type: &clusterapiv1beta1.Cluster{}},
			handler.EnqueueRequestsFromMapFunc(r.ClusterToKubevipCPIConfig),
		).
		Complete(r)
}
