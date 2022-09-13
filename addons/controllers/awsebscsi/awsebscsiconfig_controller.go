// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"gopkg.in/yaml.v2"
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

	addonconfig "github.com/vmware-tanzu/tanzu-framework/addons/pkg/config"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	"github.com/vmware-tanzu/tanzu-framework/addons/predicates"
	csiv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/addonconfigs/csi/v1alpha1"
)

// AwsEbsCSIConfigReconciler reconciles a AwsEbsCSIConfig object
type AwsEbsCSIConfigReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Config addonconfig.AwsEbsCSIConfigControllerConfig
}

//+kubebuilder:rbac:groups=csi.tanzu.vmware.com,resources=awsebscsiconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=csi.tanzu.vmware.com,resources=awsebscsiconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=csi.tanzu.vmware.com,resources=vspherecsiconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters,verbs=get;list;watch
//+kubebuilder:rbac:groups=topology.tanzu.vmware.com,resources=availabilityzones,verbs=get;list
//+kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=kubeadmcontrolplanes,verbs=get
//+kubebuilder:rbac:groups=vmware.infrastructure.cluster.x-k8s.io,resources=providerserviceaccounts,verbs=get;create;list;watch;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *AwsEbsCSIConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log = r.Log.WithValues("AwsEbsCSIConfig", req.NamespacedName)
	ctx = logr.NewContext(ctx, r.Log)
	logger := log.FromContext(ctx)
	_ = context.Background()
	_ = r.Log.WithValues("AwsEbsCSIConfig", req.NamespacedName)

	awsebsCSIConfig := &csiv1alpha1.AwsEbsCSIConfig{}
	if err := r.Get(ctx, req.NamespacedName, awsebsCSIConfig); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("AwsEbsCSIConfig resource not found")
			return ctrl.Result{}, nil
		}

		logger.Error(err, "Unable to fetch AwsEbsCSIConfig resource")
		return ctrl.Result{}, err
	}

	// deep copy awsebsCSIConfig to avoid issues if in the future other controllers where interacting with the same copy
	awsebsCSIConfig = awsebsCSIConfig.DeepCopy()
	cluster, err := r.getOwnerCluster(ctx, awsebsCSIConfig)
	if cluster == nil {
		return ctrl.Result{RequeueAfter: 20 * time.Second}, err // retry until corresponding cluster is found
	}

	return r.reconcileAwsEbsCSIConfig(ctx, awsebsCSIConfig, cluster)
}

func (r *AwsEbsCSIConfigReconciler) reconcileAwsEbsCSIConfig(ctx context.Context,
	csiCfg *csiv1alpha1.AwsEbsCSIConfig,
	cluster *clusterapiv1beta1.Cluster) (result ctrl.Result, retErr error) {

	logger := log.FromContext(ctx)

	patchHelper, err := clusterapipatchutil.NewHelper(csiCfg, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}

	defer func() {
		if retErr != nil {
			// don't modify AwsEbsCSIConfig if there is an error
			return
		}

		if err := patchHelper.Patch(ctx, csiCfg); err != nil {
			logger.Error(err, "Error patching AwsEbsCSIConfig")
			retErr = err
		}
	}()

	if !csiCfg.GetDeletionTimestamp().IsZero() {
		return ctrl.Result{}, nil // deleted
	}

	if result, err = r.reconcileAwsEbsCSIConfigNormal(ctx, csiCfg, cluster); err != nil {
		logger.Error(err, "Error reconciling AwsEbsCSIConfig")
		return result, err
	}

	return result, nil
}

func (r *AwsEbsCSIConfigReconciler) reconcileAwsEbsCSIConfigNormal(ctx context.Context,
	csiCfg *csiv1alpha1.AwsEbsCSIConfig,
	cluster *clusterapiv1beta1.Cluster) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	ownerRef := metav1.OwnerReference{
		APIVersion: clusterapiv1beta1.GroupVersion.String(),
		Kind:       cluster.Kind,
		Name:       cluster.Name,
		UID:        cluster.UID,
	}

	if !clusterapiutil.HasOwnerRef(csiCfg.OwnerReferences, ownerRef) {
		// csiCfg object is patched in defer func in 'reconcileAwsEbsCSIConfig'
		csiCfg.OwnerReferences = clusterapiutil.EnsureOwnerRef(csiCfg.OwnerReferences, ownerRef)
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GenerateDataValueSecretName(cluster.Name, constants.AwsEbsCSIAddonName),
			Namespace: csiCfg.Namespace},
		Type: v1.SecretTypeOpaque,
	}

	mutateFn := func() error {
		secret.StringData = make(map[string]string)
		dvs, err := r.mapAwsEbsCSIConfigToDataValues(ctx, csiCfg, cluster)
		if err != nil {
			logger.Error(err, "Error while mapping AwsEbsCSIConfig to data values")
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
		logger.Error(err, "Error creating or patching AwsEbsCSIConfig data values secret")
		return ctrl.Result{}, err
	}

	csiCfg.Status.SecretRef = &secret.Name

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AwsEbsCSIConfigReconciler) SetupWithManager(_ context.Context, mgr ctrl.Manager,
	options controller.Options) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&csiv1alpha1.AwsEbsCSIConfig{}).
		WithOptions(options).
		WithEventFilter(predicates.ConfigOfKindWithoutAnnotation(constants.TKGAnnotationTemplateConfig, constants.AwsEbsCSIConfigKind, r.Config.SystemNamespace, r.Log)).
		Complete(r)
}
