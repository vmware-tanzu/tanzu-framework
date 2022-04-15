// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"

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

type PinnipedV3Controller struct {
	client client.Client
	Log    logr.Logger
}

func NewV3Controller(c client.Client) *PinnipedV3Controller {
	return &PinnipedV3Controller{
		client: c,
		Log:    ctrl.Log.WithName("pinniped cascade v3 controller"),
	}
}

func (c *PinnipedV3Controller) SetupWithManager(manager ctrl.Manager) error {
	err := ctrl.
		NewControllerManagedBy(manager).
		For(
			&corev1.Secret{},
			c.withPackageName(pinnipedPackageLabel),
		).
		Watches(
			&source.Kind{Type: &corev1.ConfigMap{}},
			handler.EnqueueRequestsFromMapFunc(configMapHandler),
			withNamespacedName(types.NamespacedName{Namespace: "kube-public", Name: "pinniped-info"}),
		).
		Complete(c)
	if err != nil {
		c.Log.Error(err, "error creating pinniped config controller")
		return err
	}
	return nil
}

// +kubebuilder:rbac:groups="",resources=secrets,verbs=list;watch;get;patch
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=list;watch;get
// +kubebuilder:rbac:groups="cluster.x-k8s.io",resources=clusters,verbs=get

func (c *PinnipedV3Controller) Reconcile(ctx context.Context, req ctrl.Request) (reconcile.Result, error) {
	log := c.Log.WithName("reconcile").WithValues("request object", req)
	log.Info("starting reconciliation")
	pinnipedInfoCM, err := getPinnipedInfoConfigMap(ctx, c.client, log)
	if err != nil {
		log.Error(err, "error getting pinniped-info configmap")
		return reconcile.Result{}, err
	}

	if (req == ctrl.Request{}) {
		log.V(1).Info("empty request provided, checking all secrets")

		secrets, err := listSecretsContainingPackageName(ctx, c.client, pinnipedPackageLabel)
		if err != nil {
			log.Error(err, "error retrieving secrets", "package name", pinnipedPackageLabel)
			return reconcile.Result{}, err
		}

		for i := range secrets.Items {
			if err := c.reconcileSecret(ctx, &secrets.Items[i], pinnipedInfoCM, log); err != nil {
				log.Error(err, "error reconciling secret")
				return reconcile.Result{}, err
			}
		}

		return reconcile.Result{}, nil
	}

	secret := corev1.Secret{}
	if err := c.client.Get(ctx, req.NamespacedName, &secret); err != nil {
		if k8serror.IsNotFound(err) {
			// If secret not found, assume cluster was deleted, secret will be deleted via OwnerRef
			log.V(1).Info("could not find secret, assuming it has been deleted")
			return reconcile.Result{}, nil
		}
		log.Error(err, "error getting secret")
		return reconcile.Result{}, err
	}

	if err := c.reconcileSecret(ctx, &secret, pinnipedInfoCM, log); err != nil {
		log.Error(err, "Error reconciling secret")
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (c *PinnipedV3Controller) reconcileSecret(ctx context.Context, secret *corev1.Secret, pinnipedInfoCM *corev1.ConfigMap, log logr.Logger) error {
	// check if secret is scheduled for deletion, if so, skip reconcile
	log = log.WithValues(secretNamespaceLogKey, secret.Namespace, secretNameLogKey, secret.Name)
	if !secret.GetDeletionTimestamp().IsZero() {
		log.V(1).Info("secret is getting deleted, skipping reconcile")
		return nil
	}

	cluster, err := getClusterFromSecret(ctx, c.client, secret)
	if err != nil {
		if k8serror.IsNotFound(err) {
			// when cluster is deleted, secret will get deleted since it has an owner ref
			log.V(1).Info("cluster is getting deleted, skipping secret reconcile")
			return nil
		}
		log.Error(err, "error getting cluster, skipping reconciliation")
		return nil
	}

	log = log.WithValues(clusterNamespaceLogKey, cluster.Namespace, clusterNameLogKey, cluster.Name)

	// check if cluster is scheduled for deletion, if so, skip reconciling secret
	if !cluster.GetDeletionTimestamp().IsZero() {
		log.V(1).Info("cluster is getting deleted, skipping secret reconcile")
		return nil
	}

	if err := c.reconcileDataValues(ctx, secret, cluster, pinnipedInfoCM, log); err != nil {
		return err
	}

	return nil
}

func (c *PinnipedV3Controller) reconcileDataValues(ctx context.Context, secret *corev1.Secret, cluster *clusterapiv1beta1.Cluster, pinnipedInfoCM *corev1.ConfigMap, log logr.Logger) error {
	pinnipedSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: secret.Namespace,
			Name:      secret.Name,
		},
	}

	log.V(1).Info("creating or patching secret")
	// TODO: Create or Patch here vs. just patch since it should already be there?
	result, err := controllerutil.CreateOrPatch(ctx, c.client, pinnipedSecret, getMutateFn(pinnipedSecret, pinnipedInfoCM, cluster, log, false))
	if err != nil {
		log.Error(err, "error creating or patching data values")
		return err
	}

	log.Info("finished reconciling secret", "result", result)

	return nil
}
