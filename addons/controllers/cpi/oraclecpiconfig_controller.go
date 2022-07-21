// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package controllers implements k8s controller functionality for oracle-cpi.
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
	clusterapipatchutil "sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	addonconfig "github.com/vmware-tanzu/tanzu-framework/addons/pkg/config"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	"github.com/vmware-tanzu/tanzu-framework/addons/predicates"
	cpiv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cpi/v1alpha1"
)

// OracleCPIConfigReconciler reconciles a OracleCPIConfig object
type OracleCPIConfigReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Config addonconfig.OracleCPIConfigControllerConfig
}

//+kubebuilder:rbac:groups=cpi.tanzu.vmware.com,resources=oraclecpiconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cpi.tanzu.vmware.com,resources=oraclecpiconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=vmware.infrastructure.cluster.x-k8s.io,resources=providerserviceaccounts,verbs=get;create;list;watch;update;patch

// Reconcile the OracleCPIConfig CRD
func (r *OracleCPIConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	r.Log = r.Log.WithValues("OracleCPIConfig", req.NamespacedName)

	// fetch OracleCPIConfig resource
	cpiConfig := &cpiv1alpha1.OracleCPIConfig{}
	if err := r.Client.Get(ctx, req.NamespacedName, cpiConfig); err != nil {
		if apierrors.IsNotFound(err) {
			r.Log.Info("OracleCPIConfig resource not found")
			return ctrl.Result{}, nil
		}
		r.Log.Error(err, "Unable to fetch OracleCPIConfig resource")
		return ctrl.Result{}, err
	}

	// deep copy OracleCPIConfig to avoid issues if in the future other controllers where interacting with the same copy
	cpiConfig = cpiConfig.DeepCopy()

	cluster, err := getOwnerCluster(ctx, r.Client, r.Log, cpiConfig)
	if cluster == nil {
		return ctrl.Result{}, err // no need to requeue if cluster is not found
	}

	if res, err := r.reconcileOracleCPIConfig(ctx, cpiConfig, cluster); err != nil {
		r.Log.Error(err, "Failed to reconcile VSphereCPIConfig")
		return res, err
	}

	return ctrl.Result{}, nil
}

// reconcileOracleCPIConfig reconciles OracleCPIConfig with its owner cluster
func (r *OracleCPIConfigReconciler) reconcileOracleCPIConfig(ctx context.Context, cpiConfig *cpiv1alpha1.OracleCPIConfig, cluster *clusterapiv1beta1.Cluster) (_ ctrl.Result, retErr error) {
	// patch the CPIConfig CR in the end
	patchHelper, err := clusterapipatchutil.NewHelper(cpiConfig, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}
	// patch OracleCPIConfig before returning the function
	defer func() {
		r.Log.Info("Patching OracleCPIConfig")
		if err := patchHelper.Patch(ctx, cpiConfig); err != nil {
			r.Log.Error(err, "Error patching OracleCPIConfig")
			retErr = err
		}
		r.Log.Info("Successfully patched OracleCPIConfig")
	}()

	// convert the CPIConfig CR to data values
	d := &OracleCPIDataValues{
		Compartment: cpiConfig.Spec.Compartment,
		VCN:         cpiConfig.Spec.VCN,
		LoadBalancer: struct {
			Subnet1 string `yaml:"subnet1"`
			Subnet2 string `yaml:"subnet2"`
		}{
			Subnet1: cpiConfig.Spec.LoadBalancer.Subnet1,
			Subnet2: cpiConfig.Spec.LoadBalancer.Subnet2,
		},
	}

	// generate data value secrets for the Oracle CPI package
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GenerateDataValueSecretName(cluster.Name, constants.OracleCPIAddonName),
			Namespace: cpiConfig.Namespace,
		},
		Type: v1.SecretTypeOpaque,
	}

	// add owner reference to OracleCPIConfig if not already added by TanzuClusterBootstrap Controller
	r.Log.Info("Ensure OracleCPIConfig has the cluster as owner reference")
	ownerReference := metav1.OwnerReference{
		APIVersion: clusterapiv1beta1.GroupVersion.String(),
		Kind:       cluster.Kind,
		Name:       cluster.Name,
		UID:        cluster.UID,
	}
	secret.SetOwnerReferences([]metav1.OwnerReference{ownerReference})

	// create or patch the data value secret
	result, err := controllerutil.CreateOrUpdate(ctx, r.Client, secret, func() error {
		secret.StringData = make(map[string]string)
		yamlBytes, err := d.Serialize()
		if err != nil {
			r.Log.Error(err, "Error marshaling OracleCPIConfig to Yaml")
			return err
		}
		secret.StringData[constants.TKGDataValueFileName] = string(yamlBytes)
		r.Log.Info("Mutated OracleCPIConfig data values")
		return nil
	})
	if err != nil {
		r.Log.Error(err, "Error creating or patching OracleCPIConfig data values secret")
		return ctrl.Result{}, err
	}
	r.Log.Info(fmt.Sprintf("Resource '%s' data values secret '%s'", constants.CPIAddonName, result))

	// update the secret reference in CPIConfig CR status
	cpiConfig.Status.SecretRef = secret.Name

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OracleCPIConfigReconciler) SetupWithManager(_ context.Context, mgr ctrl.Manager, options controller.Options) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cpiv1alpha1.OracleCPIConfig{}).
		WithOptions(options).
		WithEventFilter(predicates.ConfigOfKindWithoutAnnotation(constants.TKGAnnotationTemplateConfig, constants.OracleCPIConfigKind, r.Config.SystemNamespace, r.Log)).
		Complete(r)
}
