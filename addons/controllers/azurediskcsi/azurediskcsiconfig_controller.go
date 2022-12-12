// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"

	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	clusterapiutil "sigs.k8s.io/cluster-api/util"
	clusterapipatchutil "sigs.k8s.io/cluster-api/util/patch"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	addonconfig "github.com/vmware-tanzu/tanzu-framework/addons/pkg/config"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	"github.com/vmware-tanzu/tanzu-framework/addons/predicates"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	csiv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/addonconfigs/csi/v1alpha1"
)

// AzureDiskCSIConfigReconciler reconciles a AzureDiskCSIConfig object
type AzureDiskCSIConfigReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Config addonconfig.AzureDiskCSIConfigControllerConfig
}

//+kubebuilder:rbac:groups=csi.tanzu.vmware.com,resources=azurediskcsiconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=csi.tanzu.vmware.com,resources=azurediskcsiconfigs/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *AzureDiskCSIConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	logger := r.Log.WithValues("azurediskcsiconfig", req.NamespacedName)
	logger.Info("azureDiskConfig start Reconcile ")
	azureDiskCSIConfig := &csiv1alpha1.AzureDiskCSIConfig{}
	if err := r.Get(ctx, req.NamespacedName, azureDiskCSIConfig); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("azureDiskCSIConfig resource not found")
			return ctrl.Result{}, nil
		}

		logger.Error(err, "Unable to fetch azureDiskCSIConfig resource")
		return ctrl.Result{}, err
	}

	// deep copy azureDiskCSIConfig to avoid issues if in the future other controllers where interacting with the same copy
	azureDiskCSIConfig = azureDiskCSIConfig.DeepCopy()
	cluster, err := r.getOwnerCluster(ctx, azureDiskCSIConfig)
	if cluster == nil {
		return ctrl.Result{}, err
	}

	return r.reconcileAzureDiskCSIConfig(ctx, azureDiskCSIConfig, cluster)
}

func (r *AzureDiskCSIConfigReconciler) reconcileAzureDiskCSIConfig(ctx context.Context,
	csiCfg *csiv1alpha1.AzureDiskCSIConfig,
	cluster *clusterapiv1beta1.Cluster) (result ctrl.Result, retErr error) {

	logger := log.FromContext(ctx)

	patchHelper, err := clusterapipatchutil.NewHelper(csiCfg, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}

	defer func() {
		if retErr != nil {
			// don't modify azureDiskCSIConfig if there is an error
			return
		}

		if err := patchHelper.Patch(ctx, csiCfg); err != nil {
			logger.Error(err, "Error patching azureDiskCSIConfig")
			retErr = err
		}
	}()

	if !csiCfg.GetDeletionTimestamp().IsZero() {
		return ctrl.Result{}, nil // deleted
	}

	if result, err = r.reconcileAzureDiskCSIConfigNormal(ctx, csiCfg, cluster); err != nil {
		logger.Error(err, "Error reconciling AzureDiskConfigCSI")
		return result, err
	}

	return result, nil
}

func (r *AzureDiskCSIConfigReconciler) reconcileAzureDiskCSIConfigNormal(ctx context.Context,
	csiCfg *csiv1alpha1.AzureDiskCSIConfig,
	cluster *clusterapiv1beta1.Cluster) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	ownerRef := metav1.OwnerReference{
		APIVersion: clusterapiv1beta1.GroupVersion.String(),
		Kind:       cluster.Kind,
		Name:       cluster.Name,
		UID:        cluster.UID,
	}

	if !clusterapiutil.HasOwnerRef(csiCfg.OwnerReferences, ownerRef) {
		// csiCfg object is patched in defer func in 'azureDiskCSIConfig'
		csiCfg.OwnerReferences = clusterapiutil.EnsureOwnerRef(csiCfg.OwnerReferences, ownerRef)
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GenerateDataValueSecretName(cluster.Name, constants.AzureDiskCSIAddonName),
			Namespace: csiCfg.Namespace},
		Type: v1.SecretTypeOpaque,
	}

	mutateFn := func() error {
		secret.StringData = make(map[string]string)
		dvs, err := r.mapAzureDiskCSIConfigToDataValues(ctx, csiCfg, cluster)
		if err != nil {
			logger.Error(err, "Error while mapping azureDiskCSIConfig to data values")
			return err
		}
		yamlBytes, err := yaml.Marshal(dvs)
		if err != nil {
			logger.Error(err, "Error marshaling CSI config data values to yaml")
			return err
		}
		secret.StringData[constants.TKGDataValueFileName] = string(yamlBytes)
		return nil
	}

	secret.SetOwnerReferences([]metav1.OwnerReference{ownerRef})
	_, err := controllerutil.CreateOrPatch(ctx, r.Client, secret, mutateFn)
	if err != nil {
		logger.Error(err, "Error creating or patching azureDiskCsiConfiga data values secret")
		return ctrl.Result{}, err
	}

	csiCfg.Status.SecretRef = &secret.Name

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AzureDiskCSIConfigReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager,
	options controller.Options) error {
	logger := log.FromContext(ctx)
	logger.Info("SetupWithManager azureDiskcsicontroller start")
	return ctrl.NewControllerManagedBy(mgr).
		For(&csiv1alpha1.AzureDiskCSIConfig{}).
		WithOptions(options).
		WithEventFilter(predicates.ConfigOfKindWithoutAnnotation(constants.TKGAnnotationTemplateConfig, constants.AzureDiskCSIConfigKind, r.Config.SystemNamespace, r.Log)).
		Complete(r)
}
