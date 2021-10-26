// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package concierge implements concierge functionality.
package concierge

import (
	"context"
	"fmt"

	authv1alpha1 "go.pinniped.dev/generated/1.19/apis/concierge/authentication/v1alpha1"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/post-deploy/pkg/pinnipedclientset"
)

// Configurator contains concierge client information.
type Configurator struct {
	Clientset pinnipedclientset.Concierge
}

// CreateOrUpdateJWTAuthenticator creates a new JWT or updates an existing one.
func (c Configurator) CreateOrUpdateJWTAuthenticator(ctx context.Context, namespace, name, issuer, audience, caData string) error {
	var err error
	var jwtAuthenticator *authv1alpha1.JWTAuthenticator
	if jwtAuthenticator, err = c.Clientset.AuthenticationV1alpha1().JWTAuthenticators(namespace).Get(ctx, name, metav1.GetOptions{}); err != nil {
		if errors.IsNotFound(err) {
			// create if not found
			zap.S().Infof("Creating the JWTAuthenticator %s/%s", namespace, name)
			newJWTAuthenticator := &authv1alpha1.JWTAuthenticator{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Spec: authv1alpha1.JWTAuthenticatorSpec{
					Issuer:   issuer,
					Audience: audience,
					TLS: &authv1alpha1.TLSSpec{
						CertificateAuthorityData: caData,
					},
				},
			}
			if _, err = c.Clientset.AuthenticationV1alpha1().JWTAuthenticators(namespace).Create(ctx, newJWTAuthenticator, metav1.CreateOptions{}); err != nil {
				err = fmt.Errorf("could not create jwtauthenticator %s: %w", name, err)
				zap.S().Error(err)
				return err
			}

			zap.S().Infof("Created the JWTAuthenticator %s/%s", namespace, name)
			return nil
		}
		err = fmt.Errorf("could not get jwtauthenticator %s: %w", name, err)
		zap.S().Error(err)
		return err
	}

	// update existing JWTAuthenticator
	zap.S().Infof("Updating existing JWTAuthenticator %s/%s", namespace, name)
	copiedJwtAuthenticator := jwtAuthenticator.DeepCopy()
	copiedJwtAuthenticator.Spec.Issuer = issuer
	copiedJwtAuthenticator.Spec.Audience = audience
	copiedJwtAuthenticator.Spec.TLS = &authv1alpha1.TLSSpec{
		CertificateAuthorityData: caData,
	}
	if _, err = c.Clientset.AuthenticationV1alpha1().JWTAuthenticators(namespace).Update(ctx, copiedJwtAuthenticator, metav1.UpdateOptions{}); err != nil {
		err = fmt.Errorf("could not update jwtauthenticator %s: %w", name, err)
		zap.S().Error(err)
		return err
	}

	zap.S().Infof("Updated the JWTAuthenticator %s/%s", namespace, name)
	return nil
}
