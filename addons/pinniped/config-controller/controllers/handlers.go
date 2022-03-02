package controllers

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/config-controller/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/config-controller/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)


func (c *pinnipedController) configMapToCluster(o client.Object) []ctrl.Request {
	// return empty object, if pinniped-info CM changes, update all the secrets
	c.Log.Info("Configmap created/updated/deleted, sending back empty request to reconcile all clusters/secrets")
	return []ctrl.Request{}
}

func withNamespacedName(namespacedName types.NamespacedName) builder.Predicates {
	isNamespacedName := func(o client.Object) bool {
		return o.GetNamespace() == namespacedName.Namespace && o.GetName() == namespacedName.Name
	}
	return builder.WithPredicates(
		predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool { return isNamespacedName(e.Object) },
			UpdateFunc: func(e event.UpdateEvent) bool {
				return isNamespacedName(e.ObjectOld) || isNamespacedName(e.ObjectNew)
			},
			DeleteFunc: func(e event.DeleteEvent) bool { return false },
			GenericFunc: func(e event.GenericEvent) bool { return isNamespacedName(e.Object) },
		},
	)
}

func (c *pinnipedController) addonSecretToCluster(o client.Object) []ctrl.Request {
	log := c.Log.WithValues(constants.NamespaceLogKey, o.GetName(), constants.NameLogKey, o.GetNamespace())

	log.Info("Mapping Addon Secret to cluster")
	clusterName, labelExists := o.GetLabels()[constants.TKGClusterNameLabel]

	if !labelExists || clusterName == "" {
		log.Error(errors.New("cluster name label not found on resource"),
			"Expected to find cluster name label")
		return nil
	}

	log.Info("Adding cluster for reconciliation")

	return []ctrl.Request{{
		NamespacedName: client.ObjectKey{Namespace: o.GetNamespace(), Name: clusterName},
	}}
}

func (c *pinnipedController) withAddonLabel(addonLabel string) predicate.Funcs {
	// Predicate func will get called for all events (create, update, delete, generic)
	return predicate.NewPredicateFuncs(func(o client.Object) bool {
		var secret *corev1.Secret
		switch obj := o.(type) {
		case *corev1.Secret:
			secret = obj
		default:
			c.Log.Info("Expected object type of secret. Got object type", "actualType", fmt.Sprintf("%T", o))
			return false
		}
		// TODO: do we care if secret is paused?
		if utils.IsAddonType(secret) && utils.HasAddonLabel(secret, addonLabel) {
			return true
		}

		c.Log.Info("Secret is not an addon or does not have the given label",
			constants.NamespaceLogKey, secret.Namespace, constants.NameLogKey, secret.Name, "label", addonLabel)

		return false
	})
}