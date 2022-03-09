// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package controllers contains the pinniped-config-controller-manager controller code.
package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

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
		Log:    ctrl.Log.WithName("Pinniped Config Controller"),
	}
}

func (c *PinnipedController) SetupWithManager(manager ctrl.Manager) error {
	// CM gets deleted: do nothing for now...should it get logged?
	// CM generic func: do nothing
	// Addons secret deleted: recreate it User only manages addons secret on mgmt cluster

	err := ctrl.
		NewControllerManagedBy(manager).
		For(&clusterapiv1beta1.Cluster{}).
		Watches(
			&source.Kind{Type: &corev1.ConfigMap{}},
			handler.EnqueueRequestsFromMapFunc(c.configMapToCluster),
			withNamespacedName(types.NamespacedName{Namespace: "kube-public", Name: "pinniped-info"}),
		).
		Watches(
			&source.Kind{Type: &corev1.Secret{}},
			handler.EnqueueRequestsFromMapFunc(c.addonSecretToCluster),
			builder.WithPredicates(
				c.withAddonLabel("pinniped"),
			),
		).
		// TODO: uncomment this to filter based on TKr version (only check v1alpha3 clusters/secrets)
		// WithEventFilter(utils.ClusterHasLabel(constants.TKRLabelClassyClusters, c.Log)).
		Complete(c)
	if err != nil {
		c.Log.Error(err, "Error creating pinniped config controller")
		return err
	}
	return nil
}

func (c *PinnipedController) Reconcile(ctx context.Context, req ctrl.Request) (reconcile.Result, error) {
	log := c.Log.WithName("Pinniped Config Controller Reconcile Function")
	pinnipedInfoCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: constants.KubePublicNamespace,
			Name:      constants.PinnipedInfoConfigMapName,
		},
	}

	if err := c.client.Get(ctx, client.ObjectKeyFromObject(pinnipedInfoCM), pinnipedInfoCM); err != nil {
		if !k8serror.IsNotFound(err) {
			log.Error(err, "Error getting pinniped-info configmap")
			return reconcile.Result{}, err
		}

		log.Info("pinniped-info configmap not found, setting value to nil")
		pinnipedInfoCM.Data = nil
	}
	// if req is empty, CM changed, let's loop through all clusters and create/update/delete secrets
	if (req == ctrl.Request{}) {
		clusters := &clusterapiv1beta1.ClusterList{}
		if err := c.client.List(ctx, clusters); err != nil {
			log.Error(err, "Error listing clusters")
			return reconcile.Result{}, err
		}

		for i := range clusters.Items {
			if utils.IsManagementCluster(&clusters.Items[i]) {
				continue
			}

			if err := c.reconcileAddonSecret(ctx, &clusters.Items[i], pinnipedInfoCM); err != nil {
				log.Error(err, "Error reconciling addon secret")
				return reconcile.Result{}, err
			}
		}
		return reconcile.Result{}, nil
	}

	log = log.WithValues(constants.NamespaceLogKey, req.Namespace, constants.NameLogKey, req.Name)
	// Get cluster from rec
	cluster := clusterapiv1beta1.Cluster{}
	if err := c.client.Get(ctx, req.NamespacedName, &cluster); err != nil {
		if k8serror.IsNotFound(err) {
			if err := c.reconcileClusterDelete(ctx, req.NamespacedName); err != nil {
				return reconcile.Result{}, fmt.Errorf("failed to reconcile cluster delete: %w", err)
			}
			return reconcile.Result{}, nil
		}
		log.Error(err, "Error getting cluster")
		return reconcile.Result{}, err
	}

	if utils.IsManagementCluster(&cluster) {
		log.Info("Cluster is management cluster")
		return reconcile.Result{}, nil
	}

	if err := c.reconcileAddonSecret(ctx, &cluster, pinnipedInfoCM); err != nil {
		log.Error(err, "Error reconciling addon secret")
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
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

// nolint:funlen,nolintlint // Eh, we can live with a function of this length
func (c *PinnipedController) reconcileAddonSecret(ctx context.Context, cluster *clusterapiv1beta1.Cluster, pinnipedInfoCM *corev1.ConfigMap) error {
	log := c.Log.WithValues(constants.NamespaceLogKey, cluster.Namespace, constants.NameLogKey, cluster.Name)
	// check if cluster is scheduled for deletion, if so, delete addon secret on mgmt cluster
	if !cluster.GetDeletionTimestamp().IsZero() {
		c.Log.Info("Cluster is getting deleted, deleting addon secret")
		if err := c.reconcileClusterDelete(ctx, client.ObjectKeyFromObject(cluster)); err != nil {
			return fmt.Errorf("failed to reconcile cluster delete: %w", err)
		}
		return nil
	}
	var (
		supervisorAddress  string
		supervisorCABundle string
	)

	identityManagementType := "none"
	if pinnipedInfoCM.Data != nil {
		var labelExists bool
		identityManagementType = "oidc"                                // nolint:goconst
		supervisorAddress, labelExists = pinnipedInfoCM.Data["issuer"] // TODO: get rid of raw strings...
		if !labelExists {
			err := errors.New("could not find issuer")
			log.Error(err, "Error retrieving issuer from pinniped-info configmap")
			return err
		}
		supervisorCABundle, labelExists = pinnipedInfoCM.Data["issuer_ca_bundle_data"] // TODO: get rid of raw strings...
		if !labelExists {
			err := errors.New("could not find ca bundle")
			log.Error(err, "Error retrieving ca bundle from pinniped-info configmap")
			return err
		}
	} else {
		supervisorAddress = ""
		supervisorCABundle = ""
	}

	log.Info(fmt.Sprintf("supervisorAddress: %q, supervisorCABundle: %q", supervisorAddress, supervisorCABundle))

	pinnipedAddonSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cluster.Namespace,
			Name:      fmt.Sprintf("%s-pinniped-addon", cluster.Name),
			Labels: map[string]string{
				constants.TKGAddonLabel:       constants.PinnipedAddonLabel,
				constants.TKGClusterNameLabel: cluster.Name,
			},
			Annotations: map[string]string{
				constants.TKGAddonTypeAnnotation: constants.PinnipedAddonTypeAnnotation,
			},
		},
	}
	// TODO: remove this
	pinnipedAddonSecret.Type = constants.TKGAddonType
	pinnipedAddonSecret.Data = make(map[string][]byte)

	pinnipedAddonSecretMutateFn := func() error {
		// TODO: do we need to add these fields?!?!?!?!?!?:
		//    pinniped:
		//      cert_duration:
		//      cert_renew_before:
		pinnipedDataValues := &pinnipedDataValues{}
		pinnipedDataValues.IdentityManagementType = identityManagementType
		// TODO: should we fail here if we can't find infra provider?
		infraProvider, err := utils.GetInfraProvider(cluster)
		if err != nil {
			log.Error(err, "Unable to get infrastructure_provider for ", "cluster", cluster.Name)
			return err
		}
		pinnipedDataValues.Infrastructure = infraProvider
		pinnipedDataValues.ClusterRole = "workload"
		pinnipedDataValues.Pinniped.SupervisorEndpoint = supervisorAddress
		pinnipedDataValues.Pinniped.SupervisorCABundle = supervisorCABundle
		pinnipedDataValues.Pinniped.Concierge.Audience = string(cluster.UID)
		dataValueYamlBytes, err := yaml.Marshal(pinnipedDataValues)
		if err != nil {
			log.Error(err, "Error marshaling Pinniped Addon Secret values to Yaml")
			return err
		}
		pinnipedAddonSecret.Data[constants.TKGDataValueFieldName] = dataValueYamlBytes

		return nil
	}

	log.Info("Creating or patching addon secret")
	result, err := controllerutil.CreateOrPatch(ctx, c.client, pinnipedAddonSecret, pinnipedAddonSecretMutateFn)
	if err != nil {
		log.Error(err, "Error creating or patching Pinniped addon secret data values")
		return err
	}

	log.Info(fmt.Sprintf("Result of create/patch: '%s'", result))

	return nil
}

func (c *PinnipedController) reconcileClusterDelete(ctx context.Context, clusterName types.NamespacedName) error {
	secretName := secretNameFromClusterName(clusterName)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: secretName.Namespace,
			Name:      secretName.Name,
		},
	}
	if err := c.client.Delete(ctx, secret); err != nil && !k8serror.IsNotFound(err) {
		return fmt.Errorf("could not delete secret %s for cluster %s: %w", secretName, clusterName, err)
	}
	return nil
}

func secretNameFromClusterName(clusterName types.NamespacedName) types.NamespacedName {
	return types.NamespacedName{
		Namespace: clusterName.Namespace,
		Name:      fmt.Sprintf("%s-pinniped-addon", clusterName.Name),
	}
}
