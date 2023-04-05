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
		Log:    ctrl.Log.WithName(CascadeControllerV1alpha3Name),
	}
}

func (c *PinnipedV3Controller) SetupWithManager(manager ctrl.Manager) error {
	// Note that this controller is not directly watching Clusters, unlike the v1 controller.
	// It watches Secrets with a certain type and label, and also watches the pinniped-info configmap.
	// This controller is not responsible for creating or deleting the pinniped clusterbootstrap-secrets,
	// which are created elsewhere during cluster creation for classy clusters.
	// Instead, it is only responsible for updating the content of the pinniped clusterbootstrap-secrets based on the
	// content of the pinniped-info configmap. When the pinniped-info configmap does not exist, the pinniped
	// clusterbootstrap-secrets will be updated by this controller to contain some defaults.
	err := ctrl.
		NewControllerManagedBy(manager).
		For(
			&corev1.Secret{},
			c.withPackageName(pinnipedPackageLabel), // type="clusterbootstrap-secret" and "tkg.tanzu.vmware.com/package-name" label has value containing "pinniped"
		).
		Watches(
			&source.Kind{Type: &corev1.ConfigMap{}},
			handler.EnqueueRequestsFromMapFunc(configMapHandler), // enqueues an empty ctrl.Request to indicate that the configmap changed
			withNamespacedName(types.NamespacedName{Namespace: "kube-public", Name: "pinniped-info"}),
		).
		Complete(c)
	if err != nil {
		c.Log.Error(err, "error creating controller")
		return err
	}
	return nil
}

// +kubebuilder:rbac:groups="",resources=secrets,verbs=list;watch;get;patch;delete
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
		// The pinniped-info configmap has changed. Find all the secrets and update their contents.
		log.V(1).Info("empty request provided, checking all secrets")

		// List secrets with same type and label as those secrets being watched by this controller.
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

	// A particular secret has changed. Update its contents.
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

	// Get the Cluster name from the Secret's ownerRefs or from its "tkg.tanzu.vmware.com/cluster-name" label,
	// and use it to get the Cluster resource. While updating the Secret's contents below, the Cluster will be
	// used to determine the infrastructure type.
	cluster, err := getClusterFromSecret(ctx, c.client, secret)
	if err != nil {
		if k8serror.IsNotFound(err) {
			// When cluster is deleted, secret will get deleted since it has an owner ref.
			// Or it could be the case that the Secret was created just moments before its
			// corresponding Cluster was created.
			log.V(1).Info("cluster for secret was not found, skipping secret reconcile")
			return nil
		}
		log.Error(err, "error getting cluster for secret, skipping reconciliation")
		return nil
	}

	// Check the Cluster's labels to determine if it is a management cluster. Do not update the pinniped
	// clusterbootstrap-secret for the management cluster to avoid overwriting its contents. This controller
	// is not responsible for configuring pinniped on management clusters. The user will provide the management
	// cluster's pinniped configuration either during `tanzu management-cluster create`, or by following the
	// documentation to update the pinniped addon secret for the management cluster on an existing management cluster.
	// Either way, this controller should not interfere with the user's configuration for the management cluster.
	if isManagementCluster(cluster) {
		log.V(1).Info("skipping reconciliation of secret for management cluster")
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
		// TODO: Return err if not found here or nah?  (maybe depends on contract w/CB controller)
		log.Error(err, "error creating or patching data values")
		return err
	}

	log.Info("finished reconciling secret", "result", result)

	return nil
}
