// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"
	"time"

	yaml "gopkg.in/yaml.v3"
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
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	csiv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/csi/v1alpha1"
)

// VSphereCSIConfigReconciler reconciles a VSphereCSIConfig object
type VSphereCSIConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=csi.tanzu.vmware.com,resources=vspherecsiconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=csi.tanzu.vmware.com,resources=vspherecsiconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=csi.tanzu.vmware.com,resources=vspherecsiconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *VSphereCSIConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("VSphereCSIConfig", req.NamespacedName)

	vcsiConfig := &csiv1alpha1.VSphereCSIConfig{}
	if err := r.Get(ctx, req.NamespacedName, vcsiConfig); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("VSphereCSIConfig resource not found")
			return ctrl.Result{}, nil
		}

		logger.Error(err, "Unable to fetch VSphereCSIConfig resource")
		return ctrl.Result{}, err
	}

	cluster, err := r.getOwnerCluster(ctx, vcsiConfig)
	if cluster == nil {
		return ctrl.Result{RequeueAfter: 20 * time.Second}, err // retry until corresponding cluster is found
	}

	return r.reconcileVSphereCSIConfig(ctx, vcsiConfig, cluster)
}

// SetupWithManager sets up the controller with the Manager.
func (r *VSphereCSIConfigReconciler) SetupWithManager(_ context.Context, mgr ctrl.Manager,
	options controller.Options) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&csiv1alpha1.VSphereCSIConfig{}).
		WithOptions(options).
		Complete(r)
}

func (r *VSphereCSIConfigReconciler) reconcileVSphereCSIConfig(ctx context.Context,
	csiCfg *csiv1alpha1.VSphereCSIConfig,
	cluster *clusterapiv1beta1.Cluster) (result ctrl.Result, retErr error) {

	logger := log.FromContext(ctx)

	patchHelper, err := clusterapipatchutil.NewHelper(csiCfg, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}

	defer func() {
		if retErr != nil {
			// don't modify VSphereCSIConfig if there is an error
			return
		}

		if err := patchHelper.Patch(ctx, csiCfg); err != nil {
			logger.Error(err, "Failed to patch VSphereCSIConfig")
			retErr = err
		}
	}()

	if !csiCfg.GetDeletionTimestamp().IsZero() {
		return ctrl.Result{}, nil // deleted
	}

	if result, err = r.reconcileVSphereCSIConfigNormal(ctx, csiCfg, cluster); err != nil {
		logger.Error(err, "Error reconciling VSphereCSIConfig")
		return result, err
	}

	return result, nil
}

func (r *VSphereCSIConfigReconciler) reconcileVSphereCSIConfigNormal(ctx context.Context,
	csiCfg *csiv1alpha1.VSphereCSIConfig,
	cluster *clusterapiv1beta1.Cluster) (ctrl.Result, error) {

	logger := log.FromContext(ctx)

	ownerRef := metav1.OwnerReference{
		APIVersion: clusterapiv1beta1.GroupVersion.String(),
		Kind:       cluster.Kind,
		Name:       cluster.Name,
		UID:        cluster.UID,
	}

	if !clusterapiutil.HasOwnerRef(csiCfg.OwnerReferences, ownerRef) {
		csiCfg.OwnerReferences = clusterapiutil.EnsureOwnerRef(csiCfg.OwnerReferences, ownerRef)
	}

	addonName := ""
	switch csiCfg.Spec.VSphereCSI.Mode {
	case VSphereCSINonParavirtualMode:
		addonName = constants.CSIAddonName
	case VSphereCSIParavirtualMode:
		addonName = constants.PVCSIAddonName
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GenerateDataValueSecretName(cluster.Name, addonName),
			Namespace: csiCfg.Namespace},
		Type: v1.SecretTypeOpaque,
	}

	secret.SetOwnerReferences([]metav1.OwnerReference{ownerRef})

	mutateFn := func() error {
		secret.Data = make(map[string][]byte)
		dvs, err := r.mapVSphereCSIConfigToDataValues(ctx, csiCfg, cluster)
		if err != nil {
			logger.Error(err, "Error while mapping VSphereCSIConfig to data values")
			return err
		}
		yamlBytes, err := yaml.Marshal(dvs)
		if err != nil {
			logger.Error(err, "Error marshaling CSI config data values to yaml")
			return err
		}
		secret.Data[constants.TKGDataValueFileName] = yamlBytes
		return nil
	}

	opResult, err := controllerutil.CreateOrPatch(ctx, r.Client, secret, mutateFn)

	if err != nil {
		logger.Error(err, "Error creating or patching VSphereCSIConfig data values secret")
		return ctrl.Result{}, err
	}

	logger.Info(fmt.Sprintf("'%s' the secret '%s'", opResult,
		fmt.Sprintf("%s/%s", secret.Namespace, secret.Name)))

	csiCfg.Status.SecretRef = &secret.Name

	return ctrl.Result{}, nil
}
