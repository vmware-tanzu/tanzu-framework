// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetConfigForServiceAccount returns a *rest.Config which uses the service account for talking to a Kubernetes API server.
func GetConfigForServiceAccount(ctx context.Context, coreClient client.Client, nsName, saName, host string) (*rest.Config, error) {
	serviceAccount := &corev1.ServiceAccount{}
	if err := coreClient.Get(ctx, client.ObjectKey{
		Namespace: nsName,
		Name:      saName,
	}, serviceAccount); err != nil {
		return nil, fmt.Errorf("couldn't get service account: %w", err)
	}

	for _, secretRef := range serviceAccount.Secrets {
		secret := &corev1.Secret{}
		if err := coreClient.Get(ctx, client.ObjectKey{
			Namespace: nsName,
			Name:      secretRef.Name,
		}, secret); err != nil {
			return nil, fmt.Errorf("couldn't get service account secret: %w", err)
		}

		if secret.Type != corev1.SecretTypeServiceAccountToken {
			continue
		}

		return buildConfig(secret, host)
	}

	return nil, fmt.Errorf("expected to find one service account token secret, but found none")
}

// buildConfig builds a *rest.Config from the service account secret
func buildConfig(secret *corev1.Secret, host string) (*rest.Config, error) {
	caBytes, found := secret.Data[corev1.ServiceAccountRootCAKey]
	if !found {
		return nil, fmt.Errorf("couldn't find service account token ca")
	}

	tokenBytes, found := secret.Data[corev1.ServiceAccountTokenKey]
	if !found {
		return nil, fmt.Errorf("couldn't find service account token value")
	}

	tlsClientConfig := rest.TLSClientConfig{
		CAData: caBytes,
	}

	return &rest.Config{
		Host:            host,
		TLSClientConfig: tlsClientConfig,
		BearerToken:     string(tokenBytes),
	}, nil
}
