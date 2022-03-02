package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/config-controller/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/config-controller/utils"
	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type pinnipedController struct {
	client client.Client
	Log    logr.Logger
}

func NewController(client client.Client) *pinnipedController {
	return &pinnipedController{
		client: client,
		Log:    ctrl.Log.WithName("Pinniped Config Controller"),
	}
}

func (c *pinnipedController) SetupWithManager(manager ctrl.Manager) error {
	// CM gets deleted: do nothing for now...should it get logged?
	// CM generic func: do nothing
	// Addons secret deleted: recreate it User only manages addons secret on mgmt cluster
	// 		- do we want to delete all WLC secrets when MC secret gets deleted?

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
		Complete(c)
	if err != nil {
		c.Log.Error(err, "Error creating pinniped config controller")
		return err
	}
	return nil
}

func (c *pinnipedController) Reconcile(ctx context.Context, req ctrl.Request) (reconcile.Result, error) {
	log := c.Log.WithName("Pinniped Config Controller Reconcile Function")
	// if req is empty, CM changed, let's loop through all clusters and create/update/delete secrets
	if (req == ctrl.Request{}) {
		clusters := &clusterapiv1beta1.ClusterList{}
		if err := c.client.List(ctx, clusters); err != nil {
			log.Error(err, "Error listing clusters")
			return reconcile.Result{}, err
		}

		for _, cluster := range clusters.Items {
			if utils.IsManagementCluster(cluster) {
				continue
			}
			if err := c.reconcileAddonSecret(ctx, cluster); err != nil {
				log.Error(err, "Error reconciling addon secret")
				return reconcile.Result{}, err
			}
		}
		return reconcile.Result{}, nil
	}

	log = log.WithValues( constants.NamespaceLogKey, req.Namespace, constants.NameLogKey, req.Name)
	// Get cluster from rec
	cluster := clusterapiv1beta1.Cluster{}
	if err := c.client.Get(ctx, req.NamespacedName, &cluster); err != nil {
		if k8serror.IsNotFound(err) {
			if err := c.client.DeleteAllOf(ctx, &corev1.Secret{}, client.InNamespace(req.Namespace),
				client.MatchingLabels{
					constants.TKGClusterNameLabel: req.Name,
					constants.TKGAddonLabel: constants.PinnipedAddonLabel});
				err != nil {
				if k8serror.IsNotFound(err) {
					return reconcile.Result{}, nil
				}
				log.Error(err,"Error deleting addons secrets")
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, nil
		}
		log.Error(err, "Error getting cluster")
		return reconcile.Result{}, err
	}

	if utils.IsManagementCluster(cluster) {
		log.Info("Cluster is management cluster")
		return reconcile.Result{}, nil
	}

	if err := c.reconcileAddonSecret(ctx, cluster); err != nil {
		log.Error(err, "Error reconciling addon secret")
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (c *pinnipedController) reconcileAddonSecret(ctx context.Context, cluster clusterapiv1beta1.Cluster) error {
	log := c.Log.WithValues(constants.NamespaceLogKey, cluster.Namespace, constants.NameLogKey, cluster.Name)
	// check if cluster is scheduled for deletion, if so, delete addon secret on mgmt cluster
	if !cluster.GetDeletionTimestamp().IsZero() {
		c.Log.Info("Cluster is getting deleted, deleting addon secret")
		if err := c.client.DeleteAllOf(ctx, &corev1.Secret{}, client.InNamespace(cluster.Namespace),
			client.MatchingLabels{
				constants.TKGAddonLabel: constants.PinnipedAddonLabel,
				constants.TKGClusterNameLabel: cluster.Name});
			err != nil {
				if k8serror.IsNotFound(err) {
					return nil
				}
			return err
		}
		return nil
	}

	pinnipedAddonSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cluster.Namespace,
			Name:      fmt.Sprintf("%s-pinniped-addon", cluster.Name),
			Labels: map[string]string{
				constants.TKGAddonLabel: constants.PinnipedAddonLabel,
				constants.TKGClusterNameLabel: cluster.Name,
			},
			Annotations: map[string]string{
				constants.TKGAddonTypeAnnotation: constants.PinnipedAddonTypeAnnotation,
			},
		},
	}
	pinnipedAddonSecret.Type = "tkg.tanzu.vmware.com/addon"
	pinnipedAddonSecret.Data = make(map[string][]byte)

	pinnipedAddonSecretMutateFn := func() error {
		// TODO: add the following fields:
		//    infrastructure_provider:
		//    tkg_cluster_role:
		//    pinniped:
		//      cert_duration:
		//      cert_renew_before:
		//      supervisor_svc_endpoint:
		//      supervisor_ca_bundle_data
		pinnipedAddonSecret.Data[constants.TKGDataValueFieldName] = []byte(`#@data/values
#@overlay/match-child-defaults missing_ok=True
---
identity_management_type: none
`)

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
