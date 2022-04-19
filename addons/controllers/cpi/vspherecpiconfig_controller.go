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
	capvvmwarev1beta1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/vmware/v1beta1"
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

// VSphereCPIConfigReconciler reconciles a VSphereCPIConfig object
type VSphereCPIConfigReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=cpi.tanzu.vmware.com,resources=vspherecpiconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cpi.tanzu.vmware.com,resources=vspherecpiconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=vmware.infrastructure.cluster.x-k8s.io,resources=providerserviceaccounts,verbs=get;create;list;watch;update;patch

// Reconcile the VSphereCPIConfig CRD
func (r *VSphereCPIConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log = r.Log.WithValues("VSphereCPIConfig", req.NamespacedName)

	r.Log.Info("Start reconciliation for VSphereCPIConfig")

	// fetch VSphereCPIConfig resource
	cpiConfig := &cpiv1alpha1.VSphereCPIConfig{}
	if err := r.Client.Get(ctx, req.NamespacedName, cpiConfig); err != nil {
		if apierrors.IsNotFound(err) {
			r.Log.Info("VSphereCPIConfig resource not found")
			return ctrl.Result{}, nil
		}
		r.Log.Error(err, "Unable to fetch VSphereCPIConfig resource")
		return ctrl.Result{}, err
	}

	cpiConfig = cpiConfig.DeepCopy()
	cluster, err := r.getOwnerCluster(ctx, cpiConfig)
	if cluster == nil {
		return ctrl.Result{}, err // no need to requeue if cluster is not found
	}
	if cpiConfig.Spec.VSphereCPI.Mode == nil {
		r.Log.Info("VSphere CPI mode is not provided.")
		return ctrl.Result{}, nil // no need to requeue if CPI mode is not provided
	}
	if res, err := r.reconcileVSphereCPIConfig(ctx, cpiConfig, cluster); err != nil {
		r.Log.Error(err, "Failed to reconcile VSphereCPIConfig")
		return res, err
	}
	return ctrl.Result{}, nil
}

// reconcileVSphereCPIConfig reconciles VSphereCPIConfig with its owner cluster
func (r *VSphereCPIConfigReconciler) reconcileVSphereCPIConfig(ctx context.Context, cpiConfig *cpiv1alpha1.VSphereCPIConfig, cluster *clusterapiv1beta1.Cluster) (_ ctrl.Result, retErr error) {
	patchHelper, err := clusterapipatchutil.NewHelper(cpiConfig, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}
	// patch VSphereCPIConfig before returning the function
	defer func() {
		r.Log.Info("Patching VSphereCPIConfig")
		if err := patchHelper.Patch(ctx, cpiConfig); err != nil {
			r.Log.Error(err, "Error patching VSphereCPIConfig")
			retErr = err
		}
		r.Log.Info("Successfully patched VSphereCPIConfig")
	}()

	if !cpiConfig.GetDeletionTimestamp().IsZero() {
		return ctrl.Result{}, nil
	}
	if err = r.reconcileVSphereCPIConfigNormal(ctx, cpiConfig, cluster); err != nil {
		r.Log.Error(err, "Error reconciling VSphereCPIConfig to create/patch data values secret")
		return ctrl.Result{}, err
	}
	r.Log.Info("Successfully reconciled VSphereCPIConfig")
	return ctrl.Result{}, nil
}

// reconcileVSphereCPIConfigNormal triggers when a VSphereCPIConfig is not being deleted
// it ensures the owner reference of the VSphereCPIConfig and generates the data values secret for CPI
func (r *VSphereCPIConfigReconciler) reconcileVSphereCPIConfigNormal(ctx context.Context,
	cpiConfig *cpiv1alpha1.VSphereCPIConfig, cluster *clusterapiv1beta1.Cluster) (retErr error) {
	// add owner reference to VSphereCPIConfig if not already added by TanzuClusterBootstrap Controller
	ownerReference := metav1.OwnerReference{
		APIVersion: clusterapiv1beta1.GroupVersion.String(),
		Kind:       cluster.Kind,
		Name:       cluster.Name,
		UID:        cluster.UID,
	}
	r.Log.Info("Ensure VSphereCPIConfig has the cluster as owner reference")
	if !clusterapiutil.HasOwnerRef(cpiConfig.OwnerReferences, ownerReference) {
		r.Log.Info("Adding owner reference to VSphereCPIConfig")
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
			r.Log.Error(err, "Error while mapping VSphereCPIConfig to data values")
			return err
		}
		yamlBytes, err := yaml.Marshal(cpiConfigSpec)
		if err != nil {
			r.Log.Error(err, "Error marshaling VSphereCPIConfig to Yaml")
			return err
		}
		secret.Data[constants.TKGDataValueFileName] = yamlBytes
		r.Log.Info("Mutated VSphereCPIConfig data values")
		return nil
	}
	result, err := controllerutil.CreateOrPatch(ctx, r.Client, secret, mutateFn)
	if err != nil {
		r.Log.Error(err, "Error creating or patching VSphereCPIConfig data values secret")
		return err
	}

	// deploy the provider service account for paravirtual mode
	if *cpiConfig.Spec.VSphereCPI.Mode == VSphereCPIParavirtualMode {
		r.Log.Info("Create or update provider serviceAccount for VSphere CPI")
		serviceAccount := &capvvmwarev1beta1.ProviderServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      getCCMName(cluster),
				Namespace: cluster.Namespace,
			},
		}
		_, err := controllerutil.CreateOrUpdate(ctx, r.Client, serviceAccount, func() error {
			serviceAccount.Spec = r.mapCPIConfigToProviderServiceAccountSpec(cluster)
			return controllerutil.SetControllerReference(cluster, serviceAccount, r.Scheme)
		})
		if err != nil {
			r.Log.Error(err, "Error creating or updating ProviderServiceAccount for VSphere CPI")
		}
	}

	r.Log.Info(fmt.Sprintf("Resource '%s' data values secret '%s'", constants.CPIAddonName, result))
	// update the secret reference in VSphereCPIConfig status
	cpiConfig.Status.SecretRef = secret.Name
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *VSphereCPIConfigReconciler) SetupWithManager(_ context.Context, mgr ctrl.Manager, options controller.Options) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cpiv1alpha1.VSphereCPIConfig{}).
		WithOptions(options).
		Complete(r)
}
