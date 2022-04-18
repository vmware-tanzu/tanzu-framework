// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package controllers contains the pinniped-config-controller code.
package controllers

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/builder"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type PinnipedV1Controller struct {
	client client.Client
	Log    logr.Logger
}

func NewV1Controller(c client.Client) *PinnipedV1Controller {
	return &PinnipedV1Controller{
		client: c,
		Log:    ctrl.Log.WithName("pinniped cascade v1 controller"),
	}
}

func (c *PinnipedV1Controller) SetupWithManager(manager ctrl.Manager) error {
	err := ctrl.
		NewControllerManagedBy(manager).
		For(&clusterapiv1beta1.Cluster{},
			withLabel(tkrLabel)).
		Watches(
			&source.Kind{Type: &corev1.ConfigMap{}},
			handler.EnqueueRequestsFromMapFunc(configMapHandler),
			withNamespacedName(types.NamespacedName{Namespace: "kube-public", Name: "pinniped-info"}),
		).
		Watches(
			&source.Kind{Type: &corev1.Secret{}},
			handler.EnqueueRequestsFromMapFunc(c.addonSecretToCluster),
			builder.WithPredicates(
				c.withAddonLabel("pinniped"),
			),
		).
		Complete(c)
	if err != nil {
		c.Log.Error(err, "error creating pinniped config controller")
		return err
	}
	return nil
}

// +kubebuilder:rbac:groups="",resources=secrets,verbs=list;watch;get;patch;create;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=list;watch;get
// +kubebuilder:rbac:groups="cluster.x-k8s.io",resources=clusters,verbs=list;watch;get

func (c *PinnipedV1Controller) Reconcile(ctx context.Context, req ctrl.Request) (reconcile.Result, error) {
	log := c.Log.WithName("reconcile").WithValues("request object", req)
	log.Info("starting reconciliation")
	pinnipedInfoCM, err := getPinnipedInfoConfigMap(ctx, c.client, log)
	if err != nil {
		log.Error(err, "error getting pinniped-info configmap")
		return reconcile.Result{}, err
	}

	if (req == ctrl.Request{}) {
		log.V(1).Info("configmap changed, checking all clusters")
		clusters, err := listClustersContainingLabel(ctx, c.client, tkrLabel)
		if err != nil {
			log.Error(err, "error retrieving clusters", "cluster label", tkrLabel)
			return reconcile.Result{}, err
		}

		for i := range clusters.Items {
			cluster := &clusters.Items[i]
			log = log.WithValues(clusterNamespaceLogKey, cluster.Namespace, clusterNameLogKey, cluster.Name)
			if isManagementCluster(cluster) {
				log.V(1).Info("skipping reconciliation of management cluster")
				continue
			}

			if err := c.reconcileAddonSecret(ctx, cluster, pinnipedInfoCM, log); err != nil {
				log.Error(err, "error reconciling addon secret")
				return reconcile.Result{}, err
			}
		}
		return reconcile.Result{}, nil
	}

	// Get cluster from rec
	cluster := clusterapiv1beta1.Cluster{}
	if err := c.client.Get(ctx, req.NamespacedName, &cluster); err != nil {
		if k8serror.IsNotFound(err) {
			secretNamespacedName := secretNameFromClusterName(req.NamespacedName)
			log = log.WithValues(
				secretNamespaceLogKey, secretNamespacedName.Namespace,
				secretNameLogKey, secretNamespacedName.Name)
			if err := c.reconcileDelete(ctx, secretNamespacedName, log); err != nil {
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, nil
		}
		log.Error(err, "error getting cluster")
		return reconcile.Result{}, err
	}

	log = log.WithValues(clusterNamespaceLogKey, cluster.Namespace, clusterNameLogKey, cluster.Name)

	if isManagementCluster(&cluster) {
		log.V(1).Info("skipping reconciliation of management cluster")
		return reconcile.Result{}, nil
	}

	if err := c.reconcileAddonSecret(ctx, &cluster, pinnipedInfoCM, log); err != nil {
		log.Error(err, "error reconciling addon secret")
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (c *PinnipedV1Controller) reconcileAddonSecret(ctx context.Context, cluster *clusterapiv1beta1.Cluster, pinnipedInfoCM *corev1.ConfigMap, log logr.Logger) error {
	secretNamespacedName := secretNameFromClusterName(client.ObjectKeyFromObject(cluster))
	log = log.WithValues(secretNamespaceLogKey, secretNamespacedName.Namespace, secretNameLogKey, secretNamespacedName.Name)
	// For v1alpha1 we will delete secret if CM is not found
	// also check if cluster is scheduled for deletion, if so, delete addon secret on mgmt cluster
	if pinnipedInfoCM.Data == nil || !cluster.GetDeletionTimestamp().IsZero() {
		log.V(1).Info("deleting secret")
		if err := c.reconcileDelete(ctx, client.ObjectKeyFromObject(cluster), log); err != nil {
			return err
		}
		return nil
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: secretNamespacedName.Namespace,
			Name:      secretNamespacedName.Name,
			Labels: map[string]string{
				tkgAddonLabel:       pinnipedAddonLabel,
				tkgClusterNameLabel: cluster.Name,
			},
			Annotations: map[string]string{
				tkgAddonTypeAnnotation: pinnipedAddonTypeAnnotation,
			},
		},
	}
	secret.Type = tkgAddonType

	log.V(1).Info("creating or patching addon secret")
	result, err := controllerutil.CreateOrPatch(ctx, c.client, secret, getMutateFn(secret, pinnipedInfoCM, cluster, log, true))
	if err != nil && !k8serror.IsAlreadyExists(err) {
		log.Error(err, "error creating or patching data values")
		return err
	}

	log.Info("finished reconciling secret", "result", result)

	return nil
}

func (c *PinnipedV1Controller) reconcileDelete(ctx context.Context, secretNamespacedName types.NamespacedName, log logr.Logger) error {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: secretNamespacedName.Namespace,
			Name:      secretNamespacedName.Name,
			Labels: map[string]string{
				tkgAddonLabel: pinnipedAddonLabel,
			},
			Annotations: map[string]string{
				tkgAddonTypeAnnotation: pinnipedAddonTypeAnnotation,
			},
		},
	}
	secret.Type = tkgAddonType

	if err := c.client.Delete(ctx, secret); err != nil {
		if k8serror.IsNotFound(err) {
			return nil
		}
		log.Error(err, "error deleting addon secret")
		return err
	}

	// made this V1 since the logs were pretty excessive... and I couldn't seem to avoid it even w/the notFound clause above
	log.V(1).Info("finished reconciling secret", "result", "deleted")
	return nil
}
