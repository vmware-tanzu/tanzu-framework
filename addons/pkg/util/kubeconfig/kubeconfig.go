// Copyright (c) 2019 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package kubeconfig contains utility functions to retrieve kubeconfig data from cluster-api secrets
package kubeconfig

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	kubeconfigSecretKey = "value"
)

var (
	// ErrSecretNotFound is returned when a Kubeconfig secret is not found for
	// a cluster.
	ErrSecretNotFound = errors.New("secret not found")

	// ErrSecretMissingValue is returned when a Kubeconfig secret is missing
	// the kubeconfig data.
	ErrSecretMissingValue = errors.New("missing value in secret")
)

// GetSecret retrieves the KubeConfig Secret (if any) from the given cluster
// name and namespace.
func GetSecret(ctx context.Context, c client.Client, namespace, clusterName string) (*corev1.Secret, error) {
	secret := &corev1.Secret{}
	secretKey := client.ObjectKey{
		Namespace: namespace,
		Name:      GetSecretName(clusterName),
	}

	if err := c.Get(ctx, secretKey, secret); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, ErrSecretNotFound
		}

		return nil, err
	}

	return secret, nil
}

// FromSecret uses the Secret to retrieve the KubeConfig.
func FromSecret(secret *corev1.Secret) ([]byte, error) {
	data, ok := secret.Data[kubeconfigSecretKey]
	if !ok {
		return nil, ErrSecretMissingValue
	}

	return data, nil
}

// GetSecretName returns the name of the KubeConfig secret for the provided
// cluster.
func GetSecretName(clusterName string) string {
	return fmt.Sprintf("%s-kubeconfig", clusterName)
}
