// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package controllers contains the pinniped-config-controller code.
package controllers

import (
	"context"
	"fmt"

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
	// Addons secret deleted: recreate it User only manages addons secret on mgmt cluster
	err := ctrl.
		NewControllerManagedBy(manager).
		For(&clusterapiv1beta1.Cluster{}).
		Watches(
			&source.Kind{Type: &corev1.ConfigMap{}},
			handler.EnqueueRequestsFromMapFunc(configMapHandler),
			withNamespacedName(types.NamespacedName{Namespace: "kube-public", Name: "pinniped-info"}),
		).
		// only watch v1alpha1 clusters
		WithEventFilter(clusterHasLabel(tkrLabel, c.Log)).
		Complete(c)
	if err != nil {
		c.Log.Error(err, "error creating pinniped config controller")
		return err
	}
	return nil
}

// +kubebuilder:rbac:groups="",resources=secrets,verbs=list;watch;get;patch;create
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
			if isManagementCluster(&clusters.Items[i]) {
				log.V(1).Info("skipping reconciliation of management cluster")
				continue
			}

			// For v1alpha1 we will delete secret if CM is not found
			// if pinnipedInfoCM.Data == nil {
			//	// TODO: Reconcile Delete
			// }

			if err := c.reconcileAddonSecret(ctx, &clusters.Items[i], pinnipedInfoCM, log); err != nil {
				log.Error(err, "error reconciling addon secret")
				return reconcile.Result{}, err
			}
		}
		return reconcile.Result{}, nil
	}

	// TODO: handle other scenarios in future stories
	return reconcile.Result{}, nil
}

func (c *PinnipedV1Controller) reconcileAddonSecret(ctx context.Context, cluster *clusterapiv1beta1.Cluster, pinnipedInfoCM *corev1.ConfigMap, log logr.Logger) error {
	log = log.WithValues(clusterNamespaceLogKey, cluster.Namespace, clusterNameLogKey, cluster.Name)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cluster.Namespace,
			Name:      fmt.Sprintf("%s-pinniped-addon", cluster.Name),
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

	// check if cluster is scheduled for deletion, if so, delete addon secret on mgmt cluster
	if !cluster.GetDeletionTimestamp().IsZero() {
		log.V(1).Info("cluster is getting deleted, deleting secret")
		// TODO: Delete secret
		return nil
	}

	log.V(1).Info("creating or patching addon secret")
	result, err := controllerutil.CreateOrPatch(ctx, c.client, secret, getMutateFn(secret, pinnipedInfoCM, cluster, log, true))
	if err != nil && !k8serror.IsAlreadyExists(err) {
		log.Error(err, "error creating or patching data values")
		return err
	}

	log.Info("finished creating/patching", "result", result)

	return nil
}
