// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package predicates

import (
	"reflect"
	"strings"

	"github.com/go-logr/logr"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// TKR returns a predicate.Predicate that filters tkr
func TKR(log logr.Logger) predicate.Funcs {
	return predicate.Funcs{
		CreateFunc:  func(e event.CreateEvent) bool { return true },
		UpdateFunc:  func(e event.UpdateEvent) bool { return true },
		DeleteFunc:  func(e event.DeleteEvent) bool { return false },
		GenericFunc: func(e event.GenericEvent) bool { return true },
	}
}

// ClusterHasLabel checks if the cluster has the given label
func ClusterHasLabel(label string, logger logr.Logger) predicate.Funcs {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return processIfClusterHasLabel(label, e.ObjectNew, logger.WithValues("predicate", "updateEvent"))
		},
		CreateFunc: func(e event.CreateEvent) bool {
			return processIfClusterHasLabel(label, e.Object, logger.WithValues("predicate", "createEvent"))
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return processIfClusterHasLabel(label, e.Object, logger.WithValues("predicate", "deleteEvent"))
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return processIfClusterHasLabel(label, e.Object, logger.WithValues("predicate", "genericEvent"))
		},
	}
}

// processIfClusterHasLabel determines if the input object is a cluster with a non-empty
// value for the specified label. For other input object types, it returns true
func processIfClusterHasLabel(label string, obj client.Object, logger logr.Logger) bool {
	kind := obj.GetObjectKind().GroupVersionKind().Kind
	clusterKind := reflect.TypeOf(clusterapiv1beta1.Cluster{}).Name()

	if kind != clusterKind {
		return true
	}

	labels := obj.GetLabels()
	if labels != nil {
		if l, ok := labels[label]; ok && l != "" {
			return true
		}
	}

	log := logger.WithValues("namespace", obj.GetNamespace(), strings.ToLower(kind), obj.GetName())
	log.V(6).Info("Cluster resource does not have label", "label", label)
	return false
}
