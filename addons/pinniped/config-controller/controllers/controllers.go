// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package controllers contains the pinniped-config-controller-manager controller code.
package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"gopkg.in/yaml.v3"
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

	tkgconstants "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"

	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/config-controller/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/config-controller/utils"
)

type PinnipedController struct {
	client client.Client
	Log    logr.Logger
}

func NewController(c client.Client) *PinnipedController {
	return &PinnipedController{
		client: c,
		Log:    ctrl.Log.WithName("pinniped config controller"),
	}
}

func (c *PinnipedController) SetupWithManager(manager ctrl.Manager) error {
	err := ctrl.
		NewControllerManagedBy(manager).
		For(
			&corev1.Secret{},
			c.withPackageName(constants.PinnipedPackageLabel),
		).
		Watches(
			&source.Kind{Type: &corev1.ConfigMap{}},
			handler.EnqueueRequestsFromMapFunc(c.configMapToSecret),
			withNamespacedName(types.NamespacedName{Namespace: "kube-public", Name: "pinniped-info"}),
		).
		Complete(c)
	if err != nil {
		c.Log.Error(err, "error creating pinniped config controller")
		return err
	}
	return nil
}

// +kubebuilder:rbac:groups="",resources=secrets,verbs=list;watch;get;patch;update
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=list;watch;get
// +kubebuilder:rbac:groups="cluster.x-k8s.io",resources=clusters,verbs=get

func (c *PinnipedController) Reconcile(ctx context.Context, req ctrl.Request) (reconcile.Result, error) {
	log := c.Log.WithName("reconcile").WithValues("request object", req)
	log.Info("starting reconciliation")
	pinnipedInfoCM, err := utils.GetPinnipedInfoConfigMap(ctx, c.client, log)
	if err != nil {
		log.Error(err, "error getting pinniped-info configmap")
		return reconcile.Result{}, err
	}

	if (req == ctrl.Request{}) {
		log.V(1).Info("empty request provided, checking all secrets")

		secrets, err := utils.ListSecretsContainingPackageName(ctx, c.client, constants.PinnipedPackageLabel)
		if err != nil {
			log.Error(err, "error retrieving secrets", "package name", constants.PinnipedPackageLabel)
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

func (c *PinnipedController) reconcileSecret(ctx context.Context, secret *corev1.Secret, pinnipedInfoCM *corev1.ConfigMap, log logr.Logger) error {
	// check if secret is scheduled for deletion, if so, skip reconcile
	if !secret.GetDeletionTimestamp().IsZero() {
		log.V(1).Info("secret is getting deleted, skipping reconcile")
		return nil
	}

	cluster, err := utils.GetClusterFromSecret(ctx, c.client, secret)
	if err != nil {
		if k8serror.IsNotFound(err) {
			// when cluster is deleted, secret will get deleted since it has an owner ref
			log.V(1).Info("cluster is getting deleted, skipping secret reconcile")
			return nil
		}
		log.Error(err, "error getting cluster")
	}

	log = log.WithValues(constants.ClusterNamespaceLogKey, cluster.Namespace, constants.ClusterNameLogKey, cluster.Name)

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

func (c *PinnipedController) reconcileDataValues(ctx context.Context, secret *corev1.Secret, cluster *clusterapiv1beta1.Cluster, pinnipedInfoCM *corev1.ConfigMap, log logr.Logger) error {
	pinnipedSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: secret.Namespace,
			Name:      secret.Name,
		},
	}

	log.V(1).Info("creating or patching secret")
	// TODO: Create or Patch here vs. just patch since it should already be there?
	result, err := controllerutil.CreateOrPatch(ctx, c.client, pinnipedSecret, getMutateFn(pinnipedSecret, cluster, pinnipedInfoCM, log))
	if err != nil {
		log.Error(err, "error creating or patching data values")
		return err
	}

	log.Info("finished creating/patching", "result", result)

	return nil
}

func getMutateFn(secret *corev1.Secret, cluster *clusterapiv1beta1.Cluster, pinnipedInfoCM *corev1.ConfigMap, log logr.Logger) func() error {
	return func() error {
		supervisorAddress := ""
		supervisorCABundle := ""
		identityManagementType := constants.None

		if pinnipedInfoCM.Data != nil {
			var (
				issuerExists bool
				bundleExists bool
			)
			supervisorAddress, issuerExists = pinnipedInfoCM.Data[constants.IssuerKey]
			supervisorCABundle, bundleExists = pinnipedInfoCM.Data[constants.IssuerCABundleKey]
			if issuerExists && bundleExists {
				identityManagementType = constants.OIDC
				log.V(1).Info("retrieved data from pinniped-info configmap",
					"supervisorAddress", supervisorAddress,
					"supervisorCABundle", supervisorCABundle)
			} else {
				log.Error(fmt.Errorf("supervisor address and/or CA bundle not found in pinniped info configmap"),
					"setting identity_management_type to none",
					"address exists", issuerExists,
					"CA bundle exists", bundleExists)
			}
		}

		if secret.Data == nil {
			secret.Data = map[string][]byte{}
		}

		pinnipedDataValues := &pinnipedDataValues{}
		existingDataValues, labelExists := secret.Data[constants.TKGDataValueFieldName]
		if labelExists {
			if err := yaml.Unmarshal(existingDataValues, pinnipedDataValues); err != nil {
				log.Error(err, "unable to unmarshal existing data values from secret")
			}
		}

		pinnipedDataValues.IdentityManagementType = identityManagementType
		infraProvider, err := utils.GetInfraProvider(cluster)
		if err != nil {
			if pinnipedDataValues.Infrastructure != "" {
				infraProvider = pinnipedDataValues.Infrastructure
			} else {
				log.Error(err, "unable to get infrastructure_provider, setting to vSphere")
				infraProvider = tkgconstants.InfrastructureProviderVSphere
			}
		}
		pinnipedDataValues.Infrastructure = infraProvider
		pinnipedDataValues.ClusterRole = "workload"
		pinnipedDataValues.Pinniped.SupervisorEndpoint = supervisorAddress
		pinnipedDataValues.Pinniped.SupervisorCABundle = supervisorCABundle
		// TODO: Do we want to include concierge audience if idmgmttype is none?
		pinnipedDataValues.Pinniped.Concierge.Audience = fmt.Sprintf("%s-%s", cluster.Name, string(cluster.UID))
		dataValueYamlBytes, err := yaml.Marshal(pinnipedDataValues)
		if err != nil {
			log.Error(err, "error marshaling Pinniped Secret values to yaml")
			return err
		}
		secret.Data[constants.TKGDataValueFieldName] = dataValueYamlBytes

		return nil
	}
}

type pinnipedDataValues struct {
	IdentityManagementType string   `yaml:"identity_management_type,omitempty"`
	Infrastructure         string   `yaml:"infrastructure_provider,omitempty"`
	ClusterRole            string   `yaml:"tkg_cluster_role,omitempty"`
	Pinniped               pinniped `yaml:"pinniped,omitempty"`
}

type pinniped struct {
	SupervisorEndpoint string    `yaml:"supervisor_svc_endpoint,omitempty"`
	SupervisorCABundle string    `yaml:"supervisor_ca_bundle_data,omitempty"`
	Concierge          concierge `yaml:"concierge"`
}

type concierge struct {
	Audience string `yaml:"audience,omitempty"`
}
