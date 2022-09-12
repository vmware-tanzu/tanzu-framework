// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package controllers implements k8s controller functionality for vsphere-cpi.
package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	capvvmwarev1beta1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/vmware/v1beta1"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	clusterapiutil "sigs.k8s.io/cluster-api/util"
	clusterapipatchutil "sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	addonconfig "github.com/vmware-tanzu/tanzu-framework/addons/pkg/config"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	"github.com/vmware-tanzu/tanzu-framework/addons/predicates"
	cpiv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/addonconfigs/cpi/v1alpha1"
)

// VSphereCPIConfigReconciler reconciles a VSphereCPIConfig object
type VSphereCPIConfigReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Config addonconfig.VSphereCPIConfigControllerConfig
}

var providerServiceAccountRBACRules = []rbacv1.PolicyRule{
	{
		Verbs:     []string{"get", "create", "update", "patch", "delete"},
		APIGroups: []string{"vmoperator.vmware.com"},
		Resources: []string{"virtualmachineservices", "virtualmachineservices/status"},
	},
	{
		Verbs:     []string{"get", "list"},
		APIGroups: []string{"vmoperator.vmware.com"},
		Resources: []string{"virtualmachines", "virtualmachines/status"},
	},
	{
		Verbs:     []string{"get", "create", "update", "list", "patch", "delete", "watch"},
		APIGroups: []string{"nsx.vmware.com"},
		Resources: []string{"ippools", "ippools/status"},
	},
	{
		Verbs:     []string{"get", "create", "update", "list", "patch", "delete"},
		APIGroups: []string{"nsx.vmware.com"},
		Resources: []string{"routesets", "routesets/status"},
	},
}

// VsphereCPIProviderServiceAccountAggregatedClusterRole is the cluster role to assign permissions to capv provider
var vsphereCPIProviderServiceAccountAggregatedClusterRole = &rbacv1.ClusterRole{
	ObjectMeta: metav1.ObjectMeta{
		Name: constants.VsphereCPIProviderServiceAccountAggregatedClusterRole,
		Labels: map[string]string{
			constants.CAPVClusterRoleAggregationRuleLabelSelectorKey: constants.CAPVClusterRoleAggregationRuleLabelSelectorValue,
		},
	},
}

//+kubebuilder:rbac:groups=cpi.tanzu.vmware.com,resources=vspherecpiconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cpi.tanzu.vmware.com,resources=vspherecpiconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=vmware.infrastructure.cluster.x-k8s.io,resources=providerserviceaccounts,verbs=get;create;list;watch;update;patch

// Reconcile the VSphereCPIConfig CRD
func (r *VSphereCPIConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("VSphereCPIConfig", req.NamespacedName)

	log.Info("Start reconciliation for VSphereCPIConfig")

	// fetch VSphereCPIConfig resource
	cpiConfig := &cpiv1alpha1.VSphereCPIConfig{}
	if err := r.Client.Get(ctx, req.NamespacedName, cpiConfig); err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("VSphereCPIConfig resource not found")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Unable to fetch VSphereCPIConfig resource")
		return ctrl.Result{}, err
	}

	// deep copy VSphereCPIConfig to avoid issues if in the future other controllers where interacting with the same copy
	cpiConfig = cpiConfig.DeepCopy()

	cluster, err := r.getOwnerCluster(ctx, cpiConfig)
	if cluster == nil {
		return ctrl.Result{}, err // no need to requeue if cluster is not found
	}
	if cpiConfig.Spec.VSphereCPI.Mode == nil {
		log.Info("VSphere CPI mode is not provided.")
		return ctrl.Result{}, nil // no need to requeue if CPI mode is not provided
	}
	if res, err := r.reconcileVSphereCPIConfig(ctx, cpiConfig, cluster); err != nil {
		log.Error(err, "Failed to reconcile VSphereCPIConfig")
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
		secret.StringData = make(map[string]string)
		cpiConfigSpec, err := r.mapCPIConfigToDataValues(ctx, cpiConfig, cluster)
		if err != nil {
			r.Log.Error(err, "Error while mapping VSphereCPIConfig to data values")
			return err
		}
		yamlBytes, err := cpiConfigSpec.Serialize()
		if err != nil {
			r.Log.Error(err, "Error marshaling VSphereCPIConfig to Yaml")
			return err
		}
		secret.StringData[constants.TKGDataValueFileName] = string(yamlBytes)
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
		// create an aggregated cluster role RBAC that will be inherited by CAPV (https://kubernetes.io/docs/reference/access-authn-authz/rbac/#aggregated-clusterroles)
		// CAPV needs to hold these rules before it can grant it to serviceAccount for CPI

		_, err := controllerutil.CreateOrPatch(ctx, r.Client, vsphereCPIProviderServiceAccountAggregatedClusterRole, func() error {
			vsphereCPIProviderServiceAccountAggregatedClusterRole.Rules = providerServiceAccountRBACRules
			return nil
		})
		if err != nil {
			r.Log.Error(err, "Error creating or patching cluster role", "name", vsphereCPIProviderServiceAccountAggregatedClusterRole)
			return err
		}

		vsphereClusters := &capvvmwarev1beta1.VSphereClusterList{}
		labelMatch, err := labels.NewRequirement(clusterapiv1beta1.ClusterLabelName, selection.Equals, []string{cluster.Name})
		if err != nil {
			r.Log.Error(err, "Error creating label")
			return err
		}
		labelSelector := labels.NewSelector()
		labelSelector = labelSelector.Add(*labelMatch)
		if err := r.Client.List(ctx, vsphereClusters, &client.ListOptions{LabelSelector: labelSelector, Namespace: cluster.Namespace}); err != nil {
			r.Log.Error(err, "error retrieving clusters")
			return err
		}
		if len(vsphereClusters.Items) != 1 {
			return fmt.Errorf("expected to find 1 VSphereCluster object for label key %s and value %s but found %d",
				clusterapiv1beta1.ClusterLabelName, cluster.Name, len(vsphereClusters.Items))
		}
		vsphereCluster := vsphereClusters.Items[0]
		serviceAccount := &capvvmwarev1beta1.ProviderServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      getCCMName(&vsphereCluster),
				Namespace: vsphereCluster.Namespace,
			},
		}
		_, err = controllerutil.CreateOrUpdate(ctx, r.Client, serviceAccount, func() error {
			serviceAccount.Spec = r.mapCPIConfigToProviderServiceAccountSpec(&vsphereCluster)
			return controllerutil.SetControllerReference(&vsphereCluster, serviceAccount, r.Scheme)
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
		WithEventFilter(predicates.ConfigOfKindWithoutAnnotation(constants.TKGAnnotationTemplateConfig, constants.VSphereCPIConfigKind, r.Config.SystemNamespace, r.Log)).
		Watches(
			&source.Kind{Type: &capvvmwarev1beta1.VSphereCluster{}},
			handler.EnqueueRequestsFromMapFunc(r.VSphereClusterToVSphereCPIConfig),
		).
		Complete(r)
}
