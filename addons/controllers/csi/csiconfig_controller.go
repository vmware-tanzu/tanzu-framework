/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"

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

// CSIConfigReconciler reconciles a CSIConfig object
type CSIConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=csi.tanzu.vmware.com,resources=csiconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=csi.tanzu.vmware.com,resources=csiconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=csi.tanzu.vmware.com,resources=csiconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CSIConfig object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *CSIConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("CSIConfig", req.NamespacedName)

	csiConfig := &csiv1alpha1.CSIConfig{}
	if err := r.Get(ctx, req.NamespacedName, csiConfig); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("CSIConfig resource not found", "namespacedName", req.NamespacedName)
			return ctrl.Result{}, nil
		}

		logger.Error(err, "Unable to fetch CSIConfig resource", "namespacedName", req.NamespacedName)
		return ctrl.Result{}, err
	}

	if cluster, err := r.getOwnerCluster(ctx, csiConfig); cluster == nil {
		return ctrl.Result{}, err // no need to requeue if cluster is not found
	}

	return r.reconcileCSIConfig(ctx, csiConfig)
}

// SetupWithManager sets up the controller with the Manager.
func (r *CSIConfigReconciler) SetupWithManager(_ context.Context, mgr ctrl.Manager,
	options controller.Options) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&csiv1alpha1.CSIConfig{}).
		WithOptions(options).
		Complete(r)
}

func (r *CSIConfigReconciler) reconcileCSIConfig(ctx context.Context,
	csiCfg *csiv1alpha1.CSIConfig) (result ctrl.Result, retErr error) {

	logger := log.FromContext(ctx)

	patchHelper, err := clusterapipatchutil.NewHelper(csiCfg, r.Client)
	if err != nil {
		return ctrl.Result{}, err
	}

	defer func() {
		if retErr != nil {
			// don't modify CSIConfig if there is an error
			return
		}

		if err := patchHelper.Patch(ctx, csiCfg); err != nil {
			logger.Error(err, "Failed to patch CSIConfig")
			return
		}
	}()

	if !csiCfg.GetDeletionTimestamp().IsZero() {
		return ctrl.Result{}, nil // deleted
	}

	if result, err = r.reconcileCSIConfigNormal(ctx, csiCfg); err != nil {
		logger.Error(err, "Error reconciling CSIConfig")
		return result, err
	}

	return result, nil
}

func (r *CSIConfigReconciler) reconcileCSIConfigNormal(ctx context.Context,
	csiCfg *csiv1alpha1.CSIConfig) (result ctrl.Result, retErr error) {

	logger := log.FromContext(ctx)

	// add owner reference to CSIConfig if not already added
	cluster, err := r.getOwnerCluster(ctx, csiCfg)
	if err != nil {
		return ctrl.Result{}, err
	}

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
	case CSINonParavirtualMode:
		addonName = "vsphere-csi"
	case CSIParavirtualMode:
		addonName = "vsphere-pv-csi"
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
		dvs, err := r.mapCSIConfigToDataValues(ctx, csiCfg)
		if err != nil {
			logger.Error(err, "Error while mapping CSIConfig to data values")
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
		logger.Error(err, "Error creating or patching CSIConfig data values secret")
		return result, err
	}

	logger.Info(fmt.Sprintf("'%s' the secret '%s'", opResult,
		fmt.Sprintf("%s/%s", secret.Namespace, secret.Name)))

	csiCfg.Status.SecretRef = &secret.Name

	return ctrl.Result{}, nil
}
