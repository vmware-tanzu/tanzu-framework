// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package pinnipedclientset

import (
	"context"

	configv1alpha1 "go.pinniped.dev/generated/1.19/apis/supervisor/config/v1alpha1"
	idpv1alpha1 "go.pinniped.dev/generated/1.19/apis/supervisor/idp/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// NewSupervisor returns a new client for Supervisor APIs that are being served with the provided
// apiGroupSuffix.
func NewSupervisor(client dynamic.Interface, apiGroupSuffix string) Supervisor {
	return &supervisor{client: client, apiGroupSuffix: apiGroupSuffix}
}

type supervisor struct {
	client         dynamic.Interface
	apiGroupSuffix string
}

func (s *supervisor) IDPV1alpha1() SupervisorIDPV1alpha1 {
	return &supervisorIDPV1alpha1{client: s.client, apiGroupSuffix: s.apiGroupSuffix}
}

type supervisorIDPV1alpha1 struct {
	client         dynamic.Interface
	apiGroupSuffix string
}

func (s *supervisor) ConfigV1alpha1() SupervisorConfigV1alpha1 {
	return &supervisorConfigV1alpha1{client: s.client, apiGroupSuffix: s.apiGroupSuffix}
}

type supervisorConfigV1alpha1 struct {
	client         dynamic.Interface
	apiGroupSuffix string
}

func (s *supervisorIDPV1alpha1) OIDCIdentityProviders(namespace string) SupervisorOIDCIdentityProviders {
	return &supervisorOIDCIdentityProviders{
		client: s.client.Resource(
			schema.GroupVersionResource{
				Group:    translateAPIGroup(idpv1alpha1.GroupName, s.apiGroupSuffix),
				Version:  idpv1alpha1.SchemeGroupVersion.Version,
				Resource: "oidcidentityproviders",
			},
		).Namespace(namespace),
	}
}

func (s *supervisorConfigV1alpha1) FederationDomains(namespace string) SupervisorFederationDomains {
	return &supervisorFederationDomains{
		client: s.client.Resource(
			schema.GroupVersionResource{
				Group:    translateAPIGroup(configv1alpha1.GroupName, s.apiGroupSuffix),
				Version:  configv1alpha1.SchemeGroupVersion.Version,
				Resource: "federationdomains",
			},
		).Namespace(namespace),
	}
}

type supervisorOIDCIdentityProviders struct {
	client dynamic.ResourceInterface
}

func (s *supervisorOIDCIdentityProviders) Create(ctx context.Context, obj *idpv1alpha1.OIDCIdentityProvider, opts metav1.CreateOptions) (*idpv1alpha1.OIDCIdentityProvider, error) {
	newObj := &idpv1alpha1.OIDCIdentityProvider{}
	err := create(ctx, s.client, obj, opts, newObj, "OIDCIdentityProvider")
	return newObj, err
}

func (s *supervisorOIDCIdentityProviders) Get(ctx context.Context, name string, opts metav1.GetOptions) (*idpv1alpha1.OIDCIdentityProvider, error) {
	newObj := &idpv1alpha1.OIDCIdentityProvider{}
	err := get(ctx, s.client, name, opts, newObj, "OIDCIdentityProvider")
	return newObj, err
}

// nolint:gocritic // DeleteOptions is usually passed by value, so keep the same convention here.
func (s *supervisorOIDCIdentityProviders) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return deleete(ctx, s.client, name, opts, "OIDCIdentityProvider")
}

type supervisorFederationDomains struct {
	client dynamic.ResourceInterface
}

func (s *supervisorFederationDomains) Create(ctx context.Context, obj *configv1alpha1.FederationDomain, opts metav1.CreateOptions) (*configv1alpha1.FederationDomain, error) {
	newObj := &configv1alpha1.FederationDomain{}
	err := create(ctx, s.client, obj, opts, newObj, "FederationDomain")
	return newObj, err
}

func (s *supervisorFederationDomains) Update(ctx context.Context, obj *configv1alpha1.FederationDomain, opts metav1.UpdateOptions) (*configv1alpha1.FederationDomain, error) {
	newObj := &configv1alpha1.FederationDomain{}
	err := update(ctx, s.client, obj, opts, newObj, "FederationDomain")
	return newObj, err
}

func (s *supervisorFederationDomains) Get(ctx context.Context, name string, opts metav1.GetOptions) (*configv1alpha1.FederationDomain, error) {
	newObj := &configv1alpha1.FederationDomain{}
	err := get(ctx, s.client, name, opts, newObj, "FederationDomain")
	return newObj, err
}
