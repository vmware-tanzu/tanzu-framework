// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha3

import (
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// log is for logging in this package.
var tanzukubernetesreleaselog = logf.Log.WithName("tanzukubernetesrelease-resource")

func (r *TanzuKubernetesRelease) SetupWebhookWithManager(mgr ctrl.Manager) error {
	// TODO: Remove logger call below, as it is only used to prevent deadcode lint errors.
	tanzukubernetesreleaselog.Info("TODO: Remove this - this is only to pass lint")

	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
