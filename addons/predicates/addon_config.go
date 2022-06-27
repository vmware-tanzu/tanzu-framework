// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package predicates

import (
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// ConfigOfKindWithoutAnnotation checks if the config is of the given Kind and does not have the given annotation
func ConfigOfKindWithoutAnnotation(annotation, configKind, namespace string, logger logr.Logger) predicate.Funcs {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return processIfConfigOfKindWithoutAnnotation(annotation, configKind, namespace, e.ObjectNew, logger.WithValues("predicate", "updateEvent"))
		},
		CreateFunc: func(e event.CreateEvent) bool {
			return processIfConfigOfKindWithoutAnnotation(annotation, configKind, namespace, e.Object, logger.WithValues("predicate", "createEvent"))
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return processIfConfigOfKindWithoutAnnotation(annotation, configKind, namespace, e.Object, logger.WithValues("predicate", "deleteEvent"))
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return processIfConfigOfKindWithoutAnnotation(annotation, configKind, namespace, e.Object, logger.WithValues("predicate", "genericEvent"))
		},
	}
}

// processIfConfigOfKindWithoutAnnotation determines if the input object is of the specified Kind in the given namespace
// without the given annotation. For input objects do not match with the specified Kind or not in the given namespace, it returns true.
func processIfConfigOfKindWithoutAnnotation(annotation, configKind, namespace string, obj client.Object, logger logr.Logger) bool {
	if kind := obj.GetObjectKind().GroupVersionKind().Kind; kind != configKind {
		return true
	}
	if obj.GetNamespace() != namespace {
		return true
	}

	annotations := obj.GetAnnotations()
	if annotations != nil {
		if _, ok := annotations[annotation]; ok {
			log := logger.WithValues("kind", configKind, "namespace", obj.GetNamespace(), "name", obj.GetName())
			log.V(6).Info("resource has annotation", "annotation", annotation)
			return false
		}
	}
	return true
}
