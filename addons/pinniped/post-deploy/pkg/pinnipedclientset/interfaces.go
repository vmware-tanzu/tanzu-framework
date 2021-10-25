// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package pinnipedclientset

import (
	"context"

	authenticationv1alpha1 "go.pinniped.dev/generated/1.19/apis/concierge/authentication/v1alpha1"
	configv1alpha1 "go.pinniped.dev/generated/1.19/apis/supervisor/config/v1alpha1"
	idpv1alpha1 "go.pinniped.dev/generated/1.19/apis/supervisor/idp/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Supervisor returns clients for a particular Supervisor API group.
type Supervisor interface {
	IDPV1alpha1() SupervisorIDPV1alpha1
	ConfigV1alpha1() SupervisorConfigV1alpha1
}

// SupervisorIDPV1alpha1 returns clients for a particular Supervisor IDP API kind.
type SupervisorIDPV1alpha1 interface {
	OIDCIdentityProviders(namespace string) SupervisorOIDCIdentityProviders
}

// SupervisorOIDCIdentityProviders can perform OIDCIdentityProvider CRUD operations.
type SupervisorOIDCIdentityProviders interface {
	Create(ctx context.Context, obj *idpv1alpha1.OIDCIdentityProvider, opts metav1.CreateOptions) (*idpv1alpha1.OIDCIdentityProvider, error)
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*idpv1alpha1.OIDCIdentityProvider, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
}

// SupervisorConfigV1alpha1 returns clients for a particular Supervisor Config API kind.
type SupervisorConfigV1alpha1 interface {
	FederationDomains(namespace string) SupervisorFederationDomains
}

// SupervisorFederationDomains can perform FederationDomain CRUD operations.
type SupervisorFederationDomains interface {
	Create(ctx context.Context, obj *configv1alpha1.FederationDomain, opts metav1.CreateOptions) (*configv1alpha1.FederationDomain, error)
	Update(ctx context.Context, obj *configv1alpha1.FederationDomain, opts metav1.UpdateOptions) (*configv1alpha1.FederationDomain, error)
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*configv1alpha1.FederationDomain, error)
}

// Concierge returns clients for a particular Concierge API group.
type Concierge interface {
	AuthenticationV1alpha1() ConciergeAuthenticationV1alpha1
}

// ConciergeAuthenticationV1alpha1 returns clients for a particular Concierge Authentication API
// kind.
type ConciergeAuthenticationV1alpha1 interface {
	JWTAuthenticators(namespace string) ConciergeJWTAuthenticators
}

// ConciergeJWTAuthenticators can perform JWTAuthenticator CRUD operations.
type ConciergeJWTAuthenticators interface {
	Create(ctx context.Context, obj *authenticationv1alpha1.JWTAuthenticator, opts metav1.CreateOptions) (*authenticationv1alpha1.JWTAuthenticator, error)
	Update(ctx context.Context, obj *authenticationv1alpha1.JWTAuthenticator, opts metav1.UpdateOptions) (*authenticationv1alpha1.JWTAuthenticator, error)
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*authenticationv1alpha1.JWTAuthenticator, error)
}
