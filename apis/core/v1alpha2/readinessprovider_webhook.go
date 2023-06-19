// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var (
	// log is for logging in this package.
	readinessproviderlog = logf.Log.WithName("readinessprovider-resource").WithValues("apigroup", "core")
	// cli is the client used for making calls to the k8s API server.
	cli client.Client
)

// SetupWebhookWithManager adds the webhook to the manager.
func (r *ReadinessProvider) SetupWebhookWithManager(mgr ctrl.Manager) error {
	cli = mgr.GetClient()
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
func (r *ReadinessProvider) ValidateUpdate(_ runtime.Object) error {
	readinessproviderlog.Info("validate update", "name", r.Name)
	return r.validateObject()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *ReadinessProvider) ValidateDelete() error {
	readinessproviderlog.Info("validate delete", "name", r.Name)
	return nil
}

func (r *ReadinessProvider) validateObject() error {
	var allErrors field.ErrorList
	specPath := field.NewPath("spec")

	saSource := r.Spec.ServiceAccount
	if saSource != nil {
		if saSource.Name == "" {
			allErrors = append(allErrors, field.Required(specPath.Child("serviceAccount").Child("Name"), "missing required field"))
		}

		if saSource.Namespace == "" {
			allErrors = append(allErrors, field.Required(specPath.Child("serviceAccount").Child("Namespace"), "missing required field"))
		}
	}

	// Checking if provided serviceaccount is valid
	if saSource != nil && len(allErrors) == 0 {
		sa := &corev1.ServiceAccount{}
		readinessproviderlog.Info("checking if service account is present", "source", saSource)
		err := cli.Get(context.Background(), client.ObjectKey{
			Namespace: saSource.Namespace,
			Name:      saSource.Name,
		}, sa)

		if err != nil {
			allErrors = append(allErrors, field.Invalid(specPath.Child("serviceAccount"), saSource, err.Error()))
		}
	}

	for _, condition := range r.Spec.Conditions {
		if condition.ResourceExistenceCondition == nil {
			allErrors = append(
				allErrors,
				field.Invalid(
					specPath.Child("conditions"),
					r.Spec.Conditions, fmt.Sprintf("Expected condition %s to have exactly one type defined", condition.Name)))
		}
	}

	if len(allErrors) == 0 {
		return nil
	}

	return apierrors.NewInvalid(GroupVersion.WithKind("ReadinessProvider").GroupKind(), r.Name, allErrors)
}
