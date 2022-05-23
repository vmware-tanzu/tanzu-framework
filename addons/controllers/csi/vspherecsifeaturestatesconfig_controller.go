// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util/kubeconfig"
)

// VSphereCSIFeatureStatesConfigReconciler reconciles ConfigMap used to specify csi feature states
type VSphereCSIFeatureStatesConfigReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// SingleClusterReconciler reconciles ConfigMap used to specify csi feature states for 1 cluster
type SingleClusterReconciler struct {
	client.Client
	Log           logr.Logger
	Scheme        *runtime.Scheme
	vfsreconciler *VSphereCSIFeatureStatesConfigReconciler
}

// SetupWithManager sets up the controller with the Manager.
func (r *VSphereCSIFeatureStatesConfigReconciler) SetupWithManager(_ context.Context, mgr ctrl.Manager,
	options controller.Options) error {

	cmpredicates := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return isFeatureStatesConfigMap(e.ObjectNew)
		},
		CreateFunc: func(e event.CreateEvent) bool {
			return isFeatureStatesConfigMap(e.Object)
		},
		// Delete is not expected to occur
	}

	if err := ctrl.NewControllerManagedBy(mgr).
		For(&v1.ConfigMap{}).
		WithEventFilter(cmpredicates).
		WithOptions(options).
		Complete(r); err != nil {
		return err
	}

	scr := &SingleClusterReconciler{Client: r.Client,
		Log:           r.Log,
		Scheme:        r.Scheme,
		vfsreconciler: r}
	clusterpredicates := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return e.ObjectOld.GetResourceVersion() != e.ObjectNew.GetResourceVersion()
		},
		CreateFunc: func(e event.CreateEvent) bool {
			return true
		},
		DeleteFunc:  func(event.DeleteEvent) bool { return false },
		GenericFunc: func(event.GenericEvent) bool { return false },
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&clusterapiv1beta1.Cluster{}).
		WithEventFilter(clusterpredicates).
		WithOptions(options).
		Complete(scr)
}

//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *VSphereCSIFeatureStatesConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log = r.Log.WithValues("VSphereCSIFeatureStatesConfig", req.NamespacedName)
	ctx = logr.NewContext(ctx, r.Log)
	logger := log.FromContext(ctx)

	cm := &v1.ConfigMap{}
	if err := r.Get(ctx, req.NamespacedName, cm); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("ConfigMap resource not found")
			return ctrl.Result{}, nil
		}

		logger.Error(err, "Unable to fetch ConfigMap resource")
		return ctrl.Result{}, err
	}

	clusterKeys, err := r.getClusters(ctx)
	if err != nil {
		logger.Error(err, "Unable to fetch workload clusters")
		return ctrl.Result{}, err
	}

	if len(clusterKeys) == 0 {
		logger.Info("No CAPI clusters available to update")
		return ctrl.Result{}, nil
	}

	err = r.syncFeatureStatesToWorkloadClusters(ctx, clusterKeys, cm.Data)
	if err != nil {
		logger.Error(err, "Unable to update workload clusters with csi feature states")
		return ctrl.Result{RequeueAfter: 60 * time.Second}, err
	}

	return ctrl.Result{}, nil
}

func (r *VSphereCSIFeatureStatesConfigReconciler) syncFeatureStatesToWorkloadClusters(ctx context.Context,
	clusterKeys []client.ObjectKey,
	featureStates map[string]string) error {

	configMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      VSphereCSIFeatureStateConfigMapName,
			Namespace: VSphereCSIFeatureStateNamespace,
		},
	}

	var errs error
	for _, ck := range clusterKeys {
		clt, err := r.getClusterClient(ctx, ck)
		if err != nil {
			err = fmt.Errorf("%w; Failed to get k8s client for workload cluster '%s/%s'. Retry later",
				err, ck.Namespace, ck.Name)
			if errs == nil {
				errs = err
			} else {
				errs = fmt.Errorf("%w; %v", errs, err)
			}
			continue
		}

		_, err = controllerutil.CreateOrUpdate(ctx, clt, configMap, func() error {
			configMap.Data = featureStates
			return nil
		})

		if err != nil {
			if errs == nil {
				errs = err
			} else {
				errs = fmt.Errorf("%w; %v", errs, err)
			}
			continue
		}
	}

	return errs
}

func (r *VSphereCSIFeatureStatesConfigReconciler) getClusterClient(ctx context.Context,
	clusterKey client.ObjectKey) (client.Client, error) {

	restConfig, err := r.getRESTConfig(ctx, r.Client, clusterKey.Namespace, clusterKey.Name)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create client configuration for Cluster %s/%s",
			clusterKey.Namespace, clusterKey.Name)
	}
	restConfig.Timeout = 5 * time.Second

	return client.New(restConfig, client.Options{})
}

// getRESTConfig returns the *rest.Config for the provided cluster.
func (r *VSphereCSIFeatureStatesConfigReconciler) getRESTConfig(ctx context.Context, c client.Client,
	namespace, clusterName string) (*rest.Config, error) {

	secret, err := kubeconfig.GetSecret(ctx, c, namespace, clusterName)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to retrieve kubeconfig secret for Cluster %s/%s", namespace, clusterName)
	}

	data, err := kubeconfig.FromSecret(secret)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get kubeconfig from secret for Cluster %s/%s", namespace, clusterName)
	}

	return clientcmd.RESTConfigFromKubeConfig(data)
}

func (r *VSphereCSIFeatureStatesConfigReconciler) getClusters(ctx context.Context) ([]client.ObjectKey, error) {
	clusterList := &clusterapiv1beta1.ClusterList{}
	if err := r.List(ctx, clusterList, &client.ListOptions{}); err != nil {
		return nil, err
	}

	clusterKeys := []client.ObjectKey{}
	for i := 0; i < len(clusterList.Items); i++ {
		if clusterList.Items[i].GetDeletionTimestamp().IsZero() {
			key := client.ObjectKey{Namespace: clusterList.Items[i].Namespace, Name: clusterList.Items[i].Name}
			clusterKeys = append(clusterKeys, key)
		}
	}

	return clusterKeys, nil
}

func isFeatureStatesConfigMap(o metav1.Object) bool {
	return o.GetNamespace() == VSphereCSIFeatureStateNamespace &&
		o.GetName() == VSphereCSIFeatureStateConfigMapName
}

//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *SingleClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	cmnsname := types.NamespacedName{Namespace: VSphereCSIFeatureStateNamespace,
		Name: VSphereCSIFeatureStateConfigMapName}
	r.Log = r.Log.WithValues("VSphereCSIFeatureStatesConfig", cmnsname).WithValues("Cluster", req.NamespacedName)
	ctx = logr.NewContext(ctx, r.Log)
	logger := log.FromContext(ctx)

	cm := &v1.ConfigMap{}
	if err := r.Get(ctx, cmnsname, cm); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("ConfigMap resource not found") // no feature states config
			return ctrl.Result{}, nil
		}

		logger.Error(err, "Unable to fetch ConfigMap resource")
		return ctrl.Result{}, err
	}

	clusterKeys := []client.ObjectKey{req.NamespacedName}

	if err := r.vfsreconciler.syncFeatureStatesToWorkloadClusters(ctx, clusterKeys, cm.Data); err != nil {
		logger.Error(err, "Unable to update workload cluster with csi feature states")
		return ctrl.Result{RequeueAfter: 60 * time.Second}, err
	}

	return ctrl.Result{}, nil
}
