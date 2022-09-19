// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
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
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	cutil "github.com/vmware-tanzu/tanzu-framework/addons/controllers/utils"
	addonconfig "github.com/vmware-tanzu/tanzu-framework/addons/pkg/config"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	"github.com/vmware-tanzu/tanzu-framework/addons/predicates"
	csiv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/addonconfigs/csi/v1alpha1"
)

// VSphereCSIConfigReconciler reconciles a VSphereCSIConfig object
type VSphereCSIConfigReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Config addonconfig.VSphereCSIConfigControllerConfig
}

var providerServiceAccountRBACRules = []rbacv1.PolicyRule{
	{
		APIGroups: []string{"vmoperator.vmware.com"},
		Resources: []string{"virtualmachines"},
		Verbs:     []string{"get", "list", "watch", "update", "patch"},
	},
	{
		APIGroups: []string{"cns.vmware.com"},
		Resources: []string{"cnsvolumemetadatas", "cnsfileaccessconfigs"},
		Verbs:     []string{"get", "list", "watch", "update", "create", "delete"},
	},
	{
		APIGroups: []string{"cns.vmware.com"},
		Resources: []string{"cnscsisvfeaturestates"},
		Verbs:     []string{"get", "list", "watch"},
	},
	{
		APIGroups: []string{""},
		Resources: []string{"persistentvolumeclaims"},
		Verbs:     []string{"get", "list", "watch", "update", "create", "delete"},
	},
	{
		APIGroups: []string{""},
		Resources: []string{"persistentvolumeclaims/status"},
		Verbs:     []string{"get", "update", "patch"},
	},
	{
		APIGroups: []string{""},
		Resources: []string{"events"},
		Verbs:     []string{"list"},
	},
}

// VsphereCSIProviderServiceAccountAggregatedClusterRole is the cluster role to assign permissions to capv provider
var vsphereCSIProviderServiceAccountAggregatedClusterRole = &rbacv1.ClusterRole{
	ObjectMeta: metav1.ObjectMeta{
		Name: constants.VsphereCSIProviderServiceAccountAggregatedClusterRole,
		Labels: map[string]string{
			constants.CAPVClusterRoleAggregationRuleLabelSelectorKey: constants.CAPVClusterRoleAggregationRuleLabelSelectorValue,
		},
	},
}

// SetupWithManager sets up the controller with the Manager.
func (r *VSphereCSIConfigReconciler) SetupWithManager(_ context.Context, mgr ctrl.Manager,
	options controller.Options) error {

	c, err := ctrl.NewControllerManagedBy(mgr).
		For(&csiv1alpha1.VSphereCSIConfig{}).
		WithOptions(options).
		WithEventFilter(predicates.ConfigOfKindWithoutAnnotation(constants.TKGAnnotationTemplateConfig, constants.VSphereCSIConfigKind, r.Config.SystemNamespace, r.Log)).
		Build(r)
	if err != nil {
		return errors.Wrap(err, "failed to setup vspherecsiconfig controller")
	}

	fsPredicates := predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return isFeatureStatesConfigMap(e.Object)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return isFeatureStatesConfigMap(e.ObjectNew) &&
				e.ObjectOld.GetResourceVersion() != e.ObjectNew.GetResourceVersion()
		},
		// Delete is not expected to occur
	}

	if err = c.Watch(&source.Kind{Type: &v1.ConfigMap{}},
		handler.EnqueueRequestsFromMapFunc(r.ConfigMapToVSphereCSIConfig),
		fsPredicates); err != nil {
		return errors.Wrapf(err,
			"Failed to watch for ConfigMap '%s/%s' while setting vspherecsiconfig controller",
			VSphereCSIFeatureStateNamespace,
			VSphereCSIFeatureStateConfigMapName)
	}

	// (deliberate decision): There is no watch on AvailabilityZone in vspherecsiconfig_controller so any change to it will not trigger reconcile
	// of resources. Based on discussions with TKGS team, availability zone is created at supervisor cluster init time
	// and does not really change after that.
	return nil
}

//+kubebuilder:rbac:groups=csi.tanzu.vmware.com,resources=vspherecsiconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=csi.tanzu.vmware.com,resources=vspherecsiconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=csi.tanzu.vmware.com,resources=vspherecsiconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters,verbs=get;list;watch
//+kubebuilder:rbac:groups=topology.tanzu.vmware.com,resources=availabilityzones,verbs=get;list
//+kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=kubeadmcontrolplanes,verbs=get
//+kubebuilder:rbac:groups=vmware.infrastructure.cluster.x-k8s.io,resources=providerserviceaccounts,verbs=get;create;list;watch;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *VSphereCSIConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log = r.Log.WithValues("VSphereCSIConfig", req.NamespacedName)
	ctx = logr.NewContext(ctx, r.Log)
	logger := log.FromContext(ctx)

	vcsiConfig := &csiv1alpha1.VSphereCSIConfig{}
	if err := r.Get(ctx, req.NamespacedName, vcsiConfig); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("VSphereCSIConfig resource not found")
			return ctrl.Result{}, nil
		}

		logger.Error(err, "Unable to fetch VSphereCSIConfig resource")
		return ctrl.Result{}, err
	}

	// deep copy VSphereCSIConfig to avoid issues if in the future other controllers where interacting with the same copy
	vcsiConfig = vcsiConfig.DeepCopy()

	cluster, err := r.getOwnerCluster(ctx, vcsiConfig)
	if cluster == nil {
		return ctrl.Result{RequeueAfter: 20 * time.Second}, err // retry until corresponding cluster is found
	}

	return r.reconcileVSphereCSIConfig(ctx, vcsiConfig, cluster)
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
			logger.Error(err, "Error patching VSphereCSIConfig")
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
		// csiCfg object is patched in defer func in 'reconcileVSphereCSIConfig'
		csiCfg.OwnerReferences = clusterapiutil.EnsureOwnerRef(csiCfg.OwnerReferences, ownerRef)
	}

	addonName := ""
	switch csiCfg.Spec.VSphereCSI.Mode {
	case VSphereCSINonParavirtualMode:
		addonName = constants.CSIAddonName
	case VSphereCSIParavirtualMode:
		addonName = constants.PVCSIAddonName
	default:
		err := errors.Errorf("Invalid CSI mode '%s', must either be '%s' or '%s'",
			csiCfg.Spec.VSphereCSI.Mode, VSphereCSIParavirtualMode, VSphereCSINonParavirtualMode)
		logger.Error(err, "")
		return ctrl.Result{}, err
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GenerateDataValueSecretName(cluster.Name, addonName),
			Namespace: csiCfg.Namespace},
		Type: v1.SecretTypeOpaque,
	}

	secret.SetOwnerReferences([]metav1.OwnerReference{ownerRef})

	mutateFn := func() error {
		secret.StringData = make(map[string]string)
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
		secret.StringData[constants.TKGDataValueFileName] = string(yamlBytes)
		return nil
	}

	_, err := controllerutil.CreateOrPatch(ctx, r.Client, secret, mutateFn)

	if err != nil {
		logger.Error(err, "Error creating or patching VSphereCSIConfig data values secret")
		return ctrl.Result{}, err
	}

	// deploy the provider service account for paravirtual mode
	if csiCfg.Spec.VSphereCSI.Mode == VSphereCSIParavirtualMode {
		// create an aggregated cluster role RBAC that will be inherited by CAPV (https://kubernetes.io/docs/reference/access-authn-authz/rbac/#aggregated-clusterroles)
		// CAPV needs to hold these rules before it can grant it to serviceAccount for CSI
		_, err := controllerutil.CreateOrPatch(ctx, r.Client, vsphereCSIProviderServiceAccountAggregatedClusterRole, func() error {
			vsphereCSIProviderServiceAccountAggregatedClusterRole.Rules = providerServiceAccountRBACRules
			return nil
		})

		if err != nil {
			r.Log.Error(err, "Error creating or patching cluster role", "name", vsphereCSIProviderServiceAccountAggregatedClusterRole)
			return ctrl.Result{}, err
		}

		vsphereCluster, err := cutil.VSphereClusterParavirtualForCAPICluster(ctx, r.Client, cluster)
		if err != nil {
			return ctrl.Result{}, err
		}
		serviceAccount := r.mapCSIConfigToProviderServiceAccount(vsphereCluster)
		_, err = controllerutil.CreateOrUpdate(ctx, r.Client, serviceAccount, func() error {
			return controllerutil.SetControllerReference(vsphereCluster, serviceAccount, r.Scheme)
		})
		if err != nil {
			logger.Error(err, "Error creating or updating ProviderServiceAccount for VSphere CSI")
			return ctrl.Result{}, err
		}
	}

	csiCfg.Status.SecretRef = &secret.Name

	return ctrl.Result{}, nil
}

func (r *VSphereCSIConfigReconciler) ConfigMapToVSphereCSIConfig(o client.Object) []ctrl.Request {
	configs := &csiv1alpha1.VSphereCSIConfigList{}
	_ = r.List(context.Background(), configs)
	requests := []ctrl.Request{}
	for i := 0; i < len(configs.Items); i++ {
		// avoid enqueuing reconcile requests for template vSphereCSIConfig CRs in event handler of ConfigMap CR
		if _, ok := configs.Items[i].Annotations[constants.TKGAnnotationTemplateConfig]; ok && configs.Items[i].Namespace == r.Config.SystemNamespace {
			continue
		}
		if configs.Items[i].Spec.VSphereCSI.Mode == VSphereCSIParavirtualMode {
			requests = append(requests,
				ctrl.Request{NamespacedName: client.ObjectKey{Namespace: configs.Items[i].Namespace,
					Name: configs.Items[i].Name}})
		}
	}
	return requests
}

func isFeatureStatesConfigMap(o metav1.Object) bool {
	return o.GetNamespace() == VSphereCSIFeatureStateNamespace &&
		o.GetName() == VSphereCSIFeatureStateConfigMapName
}
