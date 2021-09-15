// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package predicates

import (
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	addontypes "github.com/vmware-tanzu/tanzu-framework/addons/pkg/types"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
)

// processAddonSecret returns true if secret should be processed.
// Secret can be processed if it is of type AddonSecretType and
// has addon related labels.
func processAddonSecret(o client.Object, log logr.Logger) bool {
	var secret *corev1.Secret
	switch obj := o.(type) {
	case *corev1.Secret:
		secret = obj
	default:
		// Defaults to true so we don't filter out other objects as the
		// filters are global
		log.Info("Expected object type of secret. Got object type", "actualType", fmt.Sprintf("%T", o))
		return true
	}

	if isAddonType(secret) && hasAddonLabels(secret) && !util.IsAddonPaused(secret) {
		return true
	}

	log.V(7).Info("Secret is not a addon", constants.NamespaceLogKey, secret.Namespace, constants.NameLogKey, secret.Name)

	return false
}

// isAddonType returns true if secret is of addon type
func isAddonType(secret *corev1.Secret) bool {
	return secret.Type == addontypes.AddonSecretType
}

// hasAddonLabels returns true if secret has addon-name and cluster-name labels with non-empty values
func hasAddonLabels(secret *corev1.Secret) bool {
	addonName := util.GetAddonNameFromAddonSecret(secret)
	clusterName := util.GetClusterNameFromAddonSecret(secret)

	if addonName != "" && clusterName != "" {
		return true
	}

	return false
}

// AddonSecret returns predicate funcs for addon secret
func AddonSecret(log logr.Logger) predicate.Funcs {
	return predicate.NewPredicateFuncs(func(object client.Object) bool {
		return processAddonSecret(object, log)
	})
}
