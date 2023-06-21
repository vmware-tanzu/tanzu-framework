// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

import (
	"context"
	"fmt"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var (
	// log is for logging in this package.
	readinessproviderlog = logf.Log.WithName("readinessprovider-resource").WithValues("apigroup", "core")
	// kubeClient is the client used for making calls to the k8s API server.
	kubeClient client.Client
)

// SetupWebhookWithManager adds the webhook to the manager.
func (r *ReadinessProvider) SetupWebhookWithManager(mgr ctrl.Manager) error {
	s, err := getScheme()
	if err != nil {
		return err
	}

	kubeClient, err = client.New(mgr.GetConfig(), client.Options{Scheme: s})
	if err != nil {
		return err
	}

	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

var _ webhook.Validator = &ReadinessProvider{}

// Get a cached client.
func (r *ReadinessProvider) getClient() (client.Client, error) {
	if kubeClient != nil && !reflect.ValueOf(kubeClient).IsNil() {
		return kubeClient, nil
	}

	s, err := getScheme()
	if err != nil {
		return nil, err
	}

	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}
	return client.New(cfg, client.Options{Scheme: s})
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *ReadinessProvider) ValidateCreate() error {
	readinessproviderlog.Info("validate create", "name", r.Name)
	ctx := context.Background()

	c, err := r.getClient()
	if err != nil {
		return apierrors.NewInternalError(err)
	}

	return r.validateObject(ctx, c)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *ReadinessProvider) ValidateUpdate(_ runtime.Object) error {
	readinessproviderlog.Info("validate update", "name", r.Name)

	ctx := context.Background()

	c, err := r.getClient()
	if err != nil {
		return apierrors.NewInternalError(err)
	}

	return r.validateObject(ctx, c)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *ReadinessProvider) ValidateDelete() error {
	readinessproviderlog.Info("validate delete", "name", r.Name)
	return nil
}

func (r *ReadinessProvider) validateObject(ctx context.Context, k8sClient client.Client) error {
	var allErrors field.ErrorList
	specPath := field.NewPath("spec")

	saRef := r.Spec.ServiceAccountRef
	if saRef != nil {
		if saRef.Name == "" {
			allErrors = append(allErrors, field.Required(specPath.Child("serviceAccount").Child("Name"), "missing required field"))
		}

		if saRef.Namespace == "" {
			allErrors = append(allErrors, field.Required(specPath.Child("serviceAccount").Child("Namespace"), "missing required field"))
		}
	}

	// Checking if provided serviceaccount is valid
	if saRef != nil && len(allErrors) == 0 {
		sa := &corev1.ServiceAccount{}
		readinessproviderlog.Info("checking if service account is present", "source", saRef)
		err := k8sClient.Get(ctx, client.ObjectKey{
			Namespace: saRef.Namespace,
			Name:      saRef.Name,
		}, sa)

		if err != nil {
			allErrors = append(allErrors, field.Invalid(specPath.Child("serviceAccount"), saRef, err.Error()))
		}
	}

	// Validate conditions
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
