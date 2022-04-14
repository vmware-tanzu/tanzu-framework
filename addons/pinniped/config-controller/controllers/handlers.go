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
)

func configMapHandler(o client.Object) []ctrl.Request {
	// return empty object, if pinniped-info CM changes, update all secrets/clusters
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

// withLabel determines if the input object contains the given label
func withLabel(label string) builder.Predicates {
	hasLabel := func(o client.Object) bool {
		_, labelExists := o.GetLabels()[label]

		return labelExists
	}

	return builder.WithPredicates(
		predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool { return hasLabel(e.Object) },
			UpdateFunc: func(e event.UpdateEvent) bool {
				// TODO: do we want to process if either old or new cluster has/had tkrLabel??
				return hasLabel(e.ObjectOld) || hasLabel(e.ObjectNew)
			},
			DeleteFunc:  func(e event.DeleteEvent) bool { return hasLabel(e.Object) },
			GenericFunc: func(e event.GenericEvent) bool { return hasLabel(e.Object) },
		},
	)
}

func (c *PinnipedV3Controller) withPackageName(packageName string) builder.Predicates {
	var log logr.Logger
	containsPackageName := func(o client.Object, packageName string) bool {
		var secret *corev1.Secret
		log = c.Log.WithValues(secretNamespaceLogKey, o.GetName(), secretNameLogKey, o.GetNamespace())
		switch obj := o.(type) {
		case *corev1.Secret:
			secret = obj
		default:
			c.Log.V(1).Info("expected secret, got", "type", fmt.Sprintf("%T", o))
			return false
		}
		// TODO: do we care if secret is paused?
		if isClusterBootstrapType(secret) && containsPackageName(secret, packageName) {
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
					secretNamespaceLogKey, e.Object.GetNamespace(),
					secretNameLogKey, e.Object.GetName(),
				)
				return false
			},
			GenericFunc: func(e event.GenericEvent) bool { return containsPackageName(e.Object, packageName) },
		},
	)
}
