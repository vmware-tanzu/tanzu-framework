// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package predicates

import (
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	addonconstants "github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
)

// processBomConfigMap returns true if configmap should be processed.
// ConfigMap can be processed if it is has tkr label
func processBomConfigMap(o client.Object, log logr.Logger) bool {
	var configmap *corev1.ConfigMap
	switch obj := o.(type) {
	case *corev1.ConfigMap:
		configmap = obj
	default:
		// Defaults to true so we don't filter out other objects as the
		// filters are global
		log.Info("Expected object type of configmap. Got object type", "actualType", fmt.Sprintf("%T", o))
		return true
	}

	if configmap.Namespace == addonconstants.TKGBomNamespace && isABom(configmap) {
		return true
	}

	log.V(7).Info("Configmap is not a BOM", "configmap-namespace", configmap.Namespace, "configmap-name", configmap.Name)

	return false
}

// isABom returns true if configmap holds a BOM
func isABom(configMap *corev1.ConfigMap) bool {
	tkrName := util.GetTKRNameFromBOMConfigMap(configMap)
	return tkrName != ""
}

// BomConfigMap returns a predicate.Predicate that filters configmap
// that holds bom
func BomConfigMap(log logr.Logger) predicate.Funcs {
	return predicate.Funcs{
		CreateFunc:  func(e event.CreateEvent) bool { return processBomConfigMap(e.Object, log) },
		UpdateFunc:  func(e event.UpdateEvent) bool { return processBomConfigMap(e.ObjectNew, log) },
		DeleteFunc:  func(e event.DeleteEvent) bool { return false },
		GenericFunc: func(e event.GenericEvent) bool { return processBomConfigMap(e.Object, log) },
	}
}
