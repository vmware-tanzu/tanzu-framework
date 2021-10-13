// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package pinnipedclientset

import (
	"context"

	authenticationv1alpha1 "go.pinniped.dev/generated/1.19/apis/concierge/authentication/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// NewConcierge returns a new client for Concierge APIs that are being served with the provided
// apiGroupSuffix.
func NewConcierge(client dynamic.Interface, apiGroupSuffix string) Concierge {
	return &concierge{client: client, apiGroupSuffix: apiGroupSuffix}
}

type concierge struct {
	client         dynamic.Interface
	apiGroupSuffix string
}

func (c *concierge) AuthenticationV1alpha1() ConciergeAuthenticationV1alpha1 {
	return &conciergeAuthenticationV1alpha1{client: c.client, apiGroupSuffix: c.apiGroupSuffix}
}

type conciergeAuthenticationV1alpha1 struct {
	client         dynamic.Interface
	apiGroupSuffix string
}

func (c *conciergeAuthenticationV1alpha1) JWTAuthenticators(namespace string) ConciergeJWTAuthenticators {
	return &conciergeJWTAuthenticators{
		client: c.client.Resource(
			schema.GroupVersionResource{
				Group:    translateAPIGroup(authenticationv1alpha1.GroupName, c.apiGroupSuffix),
				Version:  authenticationv1alpha1.SchemeGroupVersion.Version,
				Resource: "jwtauthenticators",
			},
		).Namespace(namespace),
	}
}

type conciergeJWTAuthenticators struct {
	client dynamic.ResourceInterface
}

func (c *conciergeJWTAuthenticators) Create(ctx context.Context, obj *authenticationv1alpha1.JWTAuthenticator, opts metav1.CreateOptions) (*authenticationv1alpha1.JWTAuthenticator, error) {
	newObj := &authenticationv1alpha1.JWTAuthenticator{}
	err := create(ctx, c.client, obj, opts, newObj, "JWTAuthenticator")
	return newObj, err
}

func (c *conciergeJWTAuthenticators) Update(ctx context.Context, obj *authenticationv1alpha1.JWTAuthenticator, opts metav1.UpdateOptions) (*authenticationv1alpha1.JWTAuthenticator, error) {
	newObj := &authenticationv1alpha1.JWTAuthenticator{}
	err := update(ctx, c.client, obj, opts, newObj, "JWTAuthenticator")
	return newObj, err
}

func (c *conciergeJWTAuthenticators) Get(ctx context.Context, name string, opts metav1.GetOptions) (*authenticationv1alpha1.JWTAuthenticator, error) {
	newObj := &authenticationv1alpha1.JWTAuthenticator{}
	err := get(ctx, c.client, name, opts, newObj, "JWTAuthenticator")
	return newObj, err
}
