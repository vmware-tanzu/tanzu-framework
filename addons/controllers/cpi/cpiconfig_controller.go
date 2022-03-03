// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package controllers implements k8s controller functionality for vsphere-cpi.
package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
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

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	cpiv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cpi/v1alpha1"
)

// CPIConfigReconciler reconciles a CPIConfig object
type CPIConfigReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=cpi.tanzu.vmware.com,resources=cpiconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cpi.tanzu.vmware.com,resources=cpiconfigs/status,verbs=get;update;patch

// Reconcile the CPIConfig CRD
func (r *CPIConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log = r.Log.WithValues("CPIConfig", req.NamespacedName)

	r.Log.Info("Start reconciliation for CPIConfig")

	// fetch CPIConfig resource
	cpiConfig := &cpiv1alpha1.CPIConfig{}
	if err := r.Client.Get(ctx, req.NamespacedName, cpiConfig); err != nil {
		if apierrors.IsNotFound(err) {
			r.Log.Info("CPIConfig resource not found")
			return ctrl.Result{}, nil
		}
		r.Log.Error(err, "Unable to fetch CPIConfig resource")
		return ctrl.Result{}, err
	}

	cpiConfig = cpiConfig.DeepCopy()
	cluster, err := r.getOwnerCluster(ctx, cpiConfig)
	if cluster == nil {
		return ctrl.Result{}, err // no need to requeue if cluster is not found
	}
	if res, err := r.reconcileCPIConfig(ctx, cpiConfig, cluster); err != nil {
		r.Log.Error(err, "Failed to reconcile CPIConfig")
		return res, err
	}
	return ctrl.Result{}, nil
}

// reconcileCPIConfig reconciles CPIConfig with its owner cluster
func (r *CPIConfigReconciler) reconcileCPIConfig(ctx context.Context, cpiConfig *cpiv1alpha1.CPIConfig, cluster *clusterapiv1beta1.Cluster) (_ ctrl.Result, retErr error) {
	patchHelper, err := clusterapipatchutil.NewHelper(cpiConfig, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}
	// patch CPIConfig before returning the function
	defer func() {
		r.Log.Info("Patching CPIConfig")
		if err := patchHelper.Patch(ctx, cpiConfig); err != nil {
			r.Log.Error(err, "Error patching CPIConfig")
			retErr = err
		}
		r.Log.Info("Successfully patched CPIConfig")
	}()

	if !cpiConfig.GetDeletionTimestamp().IsZero() {
		return ctrl.Result{}, nil
	}
	if err = r.reconcileCPIConfigNormal(ctx, cpiConfig, cluster); err != nil {
		r.Log.Error(err, "Error reconciling CPIConfig to create/patch data values secret")
		return ctrl.Result{}, err
	}
	r.Log.Info("Successfully reconciled CPIConfig")
	return ctrl.Result{}, nil
}

// reconcileCPIConfigNormal triggers when a CPIConfig is not being deleted
// it ensures the owner reference of the CPIConfig and generates the data values secret for CPI
func (r *CPIConfigReconciler) reconcileCPIConfigNormal(ctx context.Context,
	cpiConfig *cpiv1alpha1.CPIConfig, cluster *clusterapiv1beta1.Cluster) (retErr error) {
	// add owner reference to CPIConfig if not already added by TanzuClusterBootstrap Controller
	ownerReference := metav1.OwnerReference{
		APIVersion: clusterapiv1beta1.GroupVersion.String(),
		Kind:       cluster.Kind,
		Name:       cluster.Name,
		UID:        cluster.UID,
	}
	r.Log.Info("Ensure CPIConfig has the cluster as owner reference")
	if !clusterapiutil.HasOwnerRef(cpiConfig.OwnerReferences, ownerReference) {
		r.Log.Info("Adding owner reference to CPIConfig")
		cpiConfig.OwnerReferences = clusterapiutil.EnsureOwnerRef(cpiConfig.OwnerReferences, ownerReference)
	}
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GenerateDataValueSecretName(cluster.Name, constants.CPIAddonName),
			Namespace: cpiConfig.Namespace,
		},
		Type: v1.SecretTypeOpaque,
	}
	secret.SetOwnerReferences([]metav1.OwnerReference{ownerReference})

	mutateFn := func() error {
		secret.Data = make(map[string][]byte)
		cpiConfigSpec, err := r.mapCPIConfigToDataValues(ctx, cpiConfig, cluster)
		if err != nil {
			r.Log.Error(err, "Error while mapping CPIConfig to data values")
			return err
		}
		yamlBytes, err := yaml.Marshal(cpiConfigSpec)
		if err != nil {
			r.Log.Error(err, "Error marshaling CPIConfig to Yaml")
			return err
		}
		secret.Data[constants.TKGDataValueFileName] = yamlBytes
		r.Log.Info("Mutated CPIConfig data values")
		return nil
	}
	result, err := controllerutil.CreateOrPatch(ctx, r.Client, secret, mutateFn)
	if err != nil {
		r.Log.Error(err, "Error creating or patching CPIConfig data values secret")
		return err
	}
	r.Log.Info(fmt.Sprintf("Resource '%s' data values secret '%s'", constants.CPIAddonName, result))
	// update the secret reference in CPIConfig status
	cpiConfig.Status.SecretRef = secret.Name
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CPIConfigReconciler) SetupWithManager(_ context.Context, mgr ctrl.Manager, options controller.Options) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cpiv1alpha1.CPIConfig{}).
		WithOptions(options).
		Complete(r)
}
