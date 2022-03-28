// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/config-controller/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/config-controller/utils"
)

func (c *PinnipedController) configMapToSecret(o client.Object) []ctrl.Request {
	// return empty object, if pinniped-info CM changes, update all the secrets
	return []ctrl.Request{{}}
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
			DeleteFunc:  func(e event.DeleteEvent) bool { return isNamespacedName(e.Object) },
			GenericFunc: func(e event.GenericEvent) bool { return isNamespacedName(e.Object) },
		},
	)
}

func (c *PinnipedController) withPackageName(packageName string) builder.Predicates {
	var log logr.Logger
	containsPackageName := func(o client.Object, packageName string) bool {
		var secret *corev1.Secret
		log = c.Log.WithValues(constants.SecretNamespaceLogKey, o.GetName(), constants.SecretNameLogKey, o.GetNamespace())
		switch obj := o.(type) {
		case *corev1.Secret:
			secret = obj
		default:
			c.Log.V(1).Info("expected secret, got", "type", fmt.Sprintf("%T", o))
			return false
		}
		// TODO: do we care if secret is paused?
		if utils.IsClusterBootstrapType(secret) && utils.ContainsPackageName(secret, packageName) {
			log.V(1).Info("adding secret for reconciliation")
			return true
		}

		log.V(1).Info(
			"secret is not a cluster bootstrap type or does not have the given name",
			"name", packageName)
		return false
	}
	// Predicate func will get called for all events (createObject, update, deleteObject, generic)
	return builder.WithPredicates(
		predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool { return containsPackageName(e.Object, packageName) },
			UpdateFunc: func(e event.UpdateEvent) bool {
				return containsPackageName(e.ObjectOld, packageName) || containsPackageName(e.ObjectNew, packageName)
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				log.V(1).Info(
					"secret is being deleted, skipping reconcile",
					constants.SecretNamespaceLogKey, e.Object.GetNamespace(),
					constants.SecretNameLogKey, e.Object.GetName(),
				)
				return false
			},
			GenericFunc: func(e event.GenericEvent) bool { return containsPackageName(e.Object, packageName) },
		},
	)
}
