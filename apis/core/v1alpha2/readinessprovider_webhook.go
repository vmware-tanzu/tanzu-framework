// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

import (
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var readinessproviderlog = logf.Log.WithName("readinessprovider-resource").WithValues("apigroup", "core")

// SetupWebhookWithManager adds the webhook to the manager.
func (r *ReadinessProvider) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

var _ webhook.Validator = &ReadinessProvider{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *ReadinessProvider) ValidateCreate() error {
	readinessproviderlog.Info("validate create", "name", r.Name)
	return r.validateObject()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *ReadinessProvider) ValidateUpdate(old runtime.Object) error {
	readinessproviderlog.Info("validate update", "name", r.Name)
	return r.validateObject()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *ReadinessProvider) ValidateDelete() error {
	readinessproviderlog.Info("validate delete", "name", r.Name)
	return nil
}

func (r *ReadinessProvider) validateObject() error {
	for _, condition := range r.Spec.Conditions {
		if condition.ResourceExistenceCondition == nil {
			return apierrors.NewBadRequest(fmt.Sprintf("Expected condition %s to have exactly one type defined", condition.Name))
		}
	}

	return nil
}
