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
func ConfigOfKindWithoutAnnotation(annotation string, configKind string, logger logr.Logger) predicate.Funcs {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return processIfConfigOfKindWithoutAnnotation(annotation, configKind, e.ObjectNew, logger.WithValues("predicate", "updateEvent"))
		},
		CreateFunc: func(e event.CreateEvent) bool {
			return processIfConfigOfKindWithoutAnnotation(annotation, configKind, e.Object, logger.WithValues("predicate", "createEvent"))
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return processIfConfigOfKindWithoutAnnotation(annotation, configKind, e.Object, logger.WithValues("predicate", "deleteEvent"))
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return processIfConfigOfKindWithoutAnnotation(annotation, configKind, e.Object, logger.WithValues("predicate", "genericEvent"))
		},
	}
}

// processIfConfigOfKindWithoutAnnotation determines if the input object is of the specified Kind without the
// given annotation. For input objects do not match with the specified Kind, it returns true.
func processIfConfigOfKindWithoutAnnotation(annotation string, configKind string, obj client.Object, logger logr.Logger) bool {
	if kind := obj.GetObjectKind().GroupVersionKind().Kind; kind != configKind {
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
