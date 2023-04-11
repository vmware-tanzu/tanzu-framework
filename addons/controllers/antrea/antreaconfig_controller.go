// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package controllers implements k8s controller functionality for antrea.
package controllers

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	vsphere "sigs.k8s.io/cluster-api-provider-vsphere/apis/vmware/v1beta1"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	clusterapiutil "sigs.k8s.io/cluster-api/util"
	clusterapipatchutil "sigs.k8s.io/cluster-api/util/patch"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	nsxoperatorapi "github.com/vmware-tanzu/nsx-operator/pkg/apis/v1alpha1"
	addontypes "github.com/vmware-tanzu/tanzu-framework/addons/pkg/types"

	cutil "github.com/vmware-tanzu/tanzu-framework/addons/controllers/utils"
	addonconfig "github.com/vmware-tanzu/tanzu-framework/addons/pkg/config"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	"github.com/vmware-tanzu/tanzu-framework/addons/predicates"
	cniv1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/addonconfigs/cni/v1alpha2"
)

const (
	antreaTargetNameSpace     = "vmware-system-antrea"
	antreaSecretName          = "supervisor-cred"
	nsxServiceAccountAPIGroup = "nsx.vmware.com"
	nsxServiceAccountKind     = "nsxserviceaccounts"
	clusterNameLabel          = "tkg.tanzu.vmware.com/cluster-name"
)

// vsphereAntreaConfigProviderServiceAccountAggregatedClusterRole is the cluster role to assign permissions to capv provider
var vsphereAntreaConfigProviderServiceAccountAggregatedClusterRole = &rbacv1.ClusterRole{
	ObjectMeta: metav1.ObjectMeta{
		Name: constants.VsphereAntreaConfigProviderServiceAccountAggregatedClusterRole,
		Labels: map[string]string{
			constants.CAPVClusterRoleAggregationRuleLabelSelectorKey: constants.CAPVClusterRoleAggregationRuleLabelSelectorValue,
		},
	},
}

// AntreaConfigReconciler reconciles a AntreaConfig object
type AntreaConfigReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Config addonconfig.AntreaConfigControllerConfig
}

// +kubebuilder:rbac:groups=addons.tanzu.vmware.com,resources=antreaconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=addons.tanzu.vmware.com,resources=antreaconfigs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=vmware.infrastructure.cluster.x-k8s.io,resources=providerserviceaccounts,verbs=get;create;list;watch;update;patch;delete
// +kubebuilder:rbac:groups=nsx.vmware.com,resources=nsxserviceaccounts,verbs=get;create;list;watch;update;patch;delete

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
	antreaConfig := &cniv1alpha2.AntreaConfig{}
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

	labels := antreaConfig.GetLabels()
	if _, ok := labels[addontypes.PackageNameLabel]; !ok {
		r.Log.Info(fmt.Sprintf("AntreaConfig resource '%v' does not contains package name label", req.NamespacedName))
		return ctrl.Result{}, errors.New("AntreaConfig does not contains package name label")
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
		For(&cniv1alpha2.AntreaConfig{}).
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
	antreaConfig *cniv1alpha2.AntreaConfig,
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
		r.deregisterAntreaNSX(ctx, antreaConfig, cluster)
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
	antreaConfig *cniv1alpha2.AntreaConfig,
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

	if antreaConfig.Spec.AntreaNsx.BootstrapFrom.ProviderRef != nil && antreaConfig.Spec.AntreaNsx.BootstrapFrom.Inline != nil {
		err := fmt.Errorf("providerRef and inline should not be both set in AntreaConfig.spec.antreaNsx.bootstrapFrom")
		antreaConfig.Status.Message = err.Error()
	} else {
		// clear the message here.
		antreaConfig.Status.Message = ""
	}
	// update status.secretRef
	dataValueSecretName := util.GenerateDataValueSecretName(cluster.Name, constants.AntreaAddonName)
	antreaConfig.Status.SecretRef = dataValueSecretName

	return r.registerAntreaNSX(ctx, antreaConfig, cluster)
}

func getClusterName(antreaConfig *cniv1alpha2.AntreaConfig) (name string, exists bool) {
	name, exists = antreaConfig.Labels[clusterNameLabel]
	if !exists {
		index := strings.Index(antreaConfig.Name, "-antrea-package")
		if index > 0 {
			name = antreaConfig.Name[:index]
			exists = true
		}
	}
	return
}

func (r *AntreaConfigReconciler) getProviderServiceAccountName(clusterName string) string {
	return fmt.Sprintf("%s-antrea", clusterName)
}

func (r *AntreaConfigReconciler) getNSXServiceAccountName(clusterName string) string {
	return fmt.Sprintf("%s-antrea", clusterName)
}

func (r *AntreaConfigReconciler) ensureNsxServiceAccount(ctx context.Context, antreaConfig *cniv1alpha2.AntreaConfig, cluster *clusterapiv1beta1.Cluster) error {
	account := &nsxoperatorapi.NSXServiceAccount{}

	account.Name = r.getNSXServiceAccountName(cluster.Name)
	account.Namespace = antreaConfig.Namespace
	account.OwnerReferences = []metav1.OwnerReference{
		{
			APIVersion: cluster.APIVersion,
			Kind:       cluster.Kind,
			Name:       cluster.Name,
			UID:        cluster.UID,
		},
	}

	err := r.Client.Get(ctx, types.NamespacedName{
		Namespace: account.Namespace,
		Name:      account.Name,
	}, account)
	if err == nil {
		r.Log.Info("NSXServiceAccount %s/%s already exists", account.Namespace, account.Name)
		return nil
	}
	if err != nil && !apierrors.IsNotFound(err) {
		r.Log.Info("Found no existing NSXServiceAccount %s/%s", account.Namespace, account.Name)
		return err
	}

	result, err := controllerutil.CreateOrPatch(ctx, r.Client, account, nil)
	if err != nil {
		r.Log.Error(err, "Error creating or patching NSXServiceAccount", account.Namespace, account.Name)
	} else {
		r.Log.Info(fmt.Sprintf("NSXServiceAccount %s/%s created %s", account.Namespace, account.Name, result))
	}
	return err
}

func (r *AntreaConfigReconciler) ensureProviderServiceAccount(ctx context.Context, antreaConfig *cniv1alpha2.AntreaConfig, cluster *clusterapiv1beta1.Cluster) error {
	provider := &vsphere.ProviderServiceAccount{}
	vsphereCluster, err := cutil.VSphereClusterParavirtualForCAPICluster(ctx, r.Client, cluster)
	if err != nil {
		return err
	}
	clusterName, _ := getClusterName(antreaConfig)
	nsxSAName := clusterName + "-antrea"
	nsxSecretName := clusterName + "-antrea-nsx-cert"
	clusterName = vsphereCluster.Name
	providerServiceAccountRBACRules := []rbacv1.PolicyRule{
		{
			APIGroups:     []string{nsxServiceAccountAPIGroup},
			Resources:     []string{nsxServiceAccountKind},
			ResourceNames: []string{nsxSAName},
			Verbs:         []string{"get", "list", "watch"},
		},
		{
			APIGroups:     []string{""},
			Resources:     []string{"secrets"},
			ResourceNames: []string{fmt.Sprintf(nsxSecretName)},
			Verbs:         []string{"get", "list", "watch"},
		},
	}
	_, err = controllerutil.CreateOrPatch(ctx, r.Client, vsphereAntreaConfigProviderServiceAccountAggregatedClusterRole, func() error {
		vsphereAntreaConfigProviderServiceAccountAggregatedClusterRole.Rules = providerServiceAccountRBACRules
		return nil
	})
	if err != nil {
		r.Log.Error(err, "Error creating or patching cluster role", "name", vsphereAntreaConfigProviderServiceAccountAggregatedClusterRole)
		return err
	}
	provider.Name = r.getProviderServiceAccountName(clusterName)
	provider.Namespace = antreaConfig.Namespace
	provider.Spec = vsphere.ProviderServiceAccountSpec{
		Ref: &corev1.ObjectReference{
			APIVersion: cluster.APIVersion,
			Kind:       cluster.Kind,
			Name:       clusterName,
			UID:        cluster.UID,
		},
		TargetNamespace:  antreaTargetNameSpace,
		TargetSecretName: antreaSecretName,
		Rules:            providerServiceAccountRBACRules,
	}
	result, err := controllerutil.CreateOrPatch(ctx, r.Client, provider, func() error {
		return controllerutil.SetControllerReference(vsphereCluster, provider, r.Scheme)
	})
	if err != nil {
		r.Log.Error(err, "Error creating or patching ProviderServiceAccount", provider.Namespace, provider.Name)
	} else {
		r.Log.Info(fmt.Sprintf("ProviderServiceAccount %s/%s created %s： %+v", provider.Namespace, provider.Name, result, provider))
	}
	return err
}

func (r *AntreaConfigReconciler) registerAntreaNSX(ctx context.Context, antreaConfig *cniv1alpha2.AntreaConfig, cluster *clusterapiv1beta1.Cluster) error {
	if !r.Config.AntreaNsxEnabledFSS || !antreaConfig.Spec.AntreaNsx.Enable || antreaConfig.Spec.AntreaNsx.BootstrapFrom.Inline != nil {
		r.Log.Info("antreaNsx is not enabled or inline is set, there is no ProviderServiceAccount or NsxServiceAccount to be created")
		r.deregisterAntreaNSX(ctx, antreaConfig, cluster)
		return nil
	}
	if antreaConfig.Spec.AntreaNsx.BootstrapFrom.ProviderRef != nil {
		if strings.ToLower(antreaConfig.Spec.AntreaNsx.BootstrapFrom.ProviderRef.Kind) != nsxServiceAccountKind ||
			strings.ToLower(antreaConfig.Spec.AntreaNsx.BootstrapFrom.ProviderRef.ApiGroup) != nsxServiceAccountAPIGroup {
			err := fmt.Errorf("either ProviderRef.Kind(%s) or ProviderRef.ApiGroup(%s) is invalid, expcted:ProviderRef.Kind(%s) ProviderRef.ApiGroup(%s)",
				antreaConfig.Spec.AntreaNsx.BootstrapFrom.ProviderRef.Kind, antreaConfig.Spec.AntreaNsx.BootstrapFrom.ProviderRef.ApiGroup,
				nsxServiceAccountKind, nsxServiceAccountAPIGroup)
			antreaConfig.Status.Message = err.Error()
			return err
		}
	}
	antreaConfig.Status.Message = ""
	err := r.ensureProviderServiceAccount(ctx, antreaConfig, cluster)
	if err != nil {
		return err
	}
	err = r.ensureNsxServiceAccount(ctx, antreaConfig, cluster)
	return err
}

func (r *AntreaConfigReconciler) deregisterAntreaNSX(ctx context.Context, antreaConfig *cniv1alpha2.AntreaConfig, cluster *clusterapiv1beta1.Cluster) error {
	if !r.Config.AntreaNsxEnabledFSS || !antreaConfig.Spec.AntreaNsx.Enable {
		r.Log.Info("antreaNsx is not enabled, there is no ProviderServiceAccount or NsxServiceAccount to be deleted")
		return nil
	}
	vsphereCluster, err := cutil.VSphereClusterParavirtualForCAPICluster(ctx, r.Client, cluster)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	clusterName, exists := getClusterName(antreaConfig)
	if !exists {
		return fmt.Errorf("invalid antreaConfig Name")
	}
	account := &nsxoperatorapi.NSXServiceAccount{}

	account.Name = r.getNSXServiceAccountName(clusterName)
	account.Namespace = antreaConfig.Namespace
	err = r.Client.Delete(ctx, account)
	if err != nil && !apierrors.IsNotFound(err) {
		r.Log.Error(err, "failed to delete NSXServiceAccount", account.Namespace, account.Name)
		return err
	}

	provider := &vsphere.ProviderServiceAccount{}
	provider.Name = r.getProviderServiceAccountName(vsphereCluster.Name)
	provider.Namespace = vsphereCluster.Namespace
	err = r.Client.Delete(ctx, provider)
	if err != nil && !apierrors.IsNotFound(err) {
		r.Log.Error(err, "failed to delete ProviderServiceAccount", provider.Namespace, provider.Name)
		return err
	}
	return nil
}

// ReconcileAntreaConfigDataValue reconciles AntreaConfig data values secret
func (r *AntreaConfigReconciler) ReconcileAntreaConfigDataValue(
	ctx context.Context,
	antreaConfig *cniv1alpha2.AntreaConfig,
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
		antreaConfigYaml, err := mapAntreaConfigSpec(cluster, antreaConfig, r.Client)
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
