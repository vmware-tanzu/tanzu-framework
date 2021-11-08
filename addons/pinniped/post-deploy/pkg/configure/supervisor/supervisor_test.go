// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package supervisor

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	configv1alpha1 "go.pinniped.dev/generated/1.19/apis/supervisor/config/v1alpha1"
	idpv1alpha1 "go.pinniped.dev/generated/1.19/apis/supervisor/idp/v1alpha1"
	pinnipedsupervisorfake "go.pinniped.dev/generated/1.19/client/supervisor/clientset/versioned/fake"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubefake "k8s.io/client-go/kubernetes/fake"
	kubetesting "k8s.io/client-go/testing"
)

// nolint:funlen
func TestCreateOrUpdateFederationDomain(t *testing.T) {
	federationDomainGVR := configv1alpha1.SchemeGroupVersion.WithResource("federationdomains")

	federationDomain := &configv1alpha1.FederationDomain{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "some-namespace",
			Name:      "some-name",
		},
		Spec: configv1alpha1.FederationDomainSpec{
			Issuer: "some-issuer",
		},
	}

	tests := []struct {
		name         string
		newClientset func() *pinnipedsupervisorfake.Clientset
		wantError    string
		wantActions  []kubetesting.Action
	}{
		{
			name: "getting jwt authenticator fails",
			newClientset: func() *pinnipedsupervisorfake.Clientset {
				c := pinnipedsupervisorfake.NewSimpleClientset()
				c.PrependReactor("get", "federationdomains", func(a kubetesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("some get error")
				})
				return c
			},
			wantError: fmt.Sprintf("could not get federationdomain %s/%s: some get error", federationDomain.Namespace, federationDomain.Name),
			wantActions: []kubetesting.Action{
				kubetesting.NewGetAction(federationDomainGVR, federationDomain.Namespace, federationDomain.Name),
			},
		},
		{
			name: "jwt authenticator does not exist",
			newClientset: func() *pinnipedsupervisorfake.Clientset {
				return pinnipedsupervisorfake.NewSimpleClientset()
			},
			wantActions: []kubetesting.Action{
				kubetesting.NewGetAction(federationDomainGVR, federationDomain.Namespace, federationDomain.Name),
				kubetesting.NewCreateAction(federationDomainGVR, federationDomain.Namespace, federationDomain),
			},
		},
		{
			name: "jwt authenticator does not exist and creating federationdomain fails",
			newClientset: func() *pinnipedsupervisorfake.Clientset {
				c := pinnipedsupervisorfake.NewSimpleClientset()
				c.PrependReactor("create", "federationdomains", func(a kubetesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("some create error")
				})
				return c
			},
			wantError: fmt.Sprintf("could not create federationdomain %s/%s: some create error", federationDomain.Namespace, federationDomain.Name),
			wantActions: []kubetesting.Action{
				kubetesting.NewGetAction(federationDomainGVR, federationDomain.Namespace, federationDomain.Name),
				kubetesting.NewCreateAction(federationDomainGVR, federationDomain.Namespace, federationDomain),
			},
		},
		{
			name: "jwt authenticator exists and is up to date",
			newClientset: func() *pinnipedsupervisorfake.Clientset {
				return pinnipedsupervisorfake.NewSimpleClientset(federationDomain.DeepCopy())
			},
			wantActions: []kubetesting.Action{
				kubetesting.NewGetAction(federationDomainGVR, federationDomain.Namespace, federationDomain.Name),
				kubetesting.NewUpdateAction(federationDomainGVR, federationDomain.Namespace, federationDomain),
			},
		},
		{
			name: "jwt authenticator exists and is not up to date",
			newClientset: func() *pinnipedsupervisorfake.Clientset {
				existingFederationDomain := federationDomain.DeepCopy()
				existingFederationDomain.Spec.Issuer = "some-other-issuer"
				return pinnipedsupervisorfake.NewSimpleClientset(federationDomain.DeepCopy())
			},
			wantActions: []kubetesting.Action{
				kubetesting.NewGetAction(federationDomainGVR, federationDomain.Namespace, federationDomain.Name),
				kubetesting.NewUpdateAction(federationDomainGVR, federationDomain.Namespace, federationDomain),
			},
		},
		{
			name: "updating federationdomain fails",
			newClientset: func() *pinnipedsupervisorfake.Clientset {
				c := pinnipedsupervisorfake.NewSimpleClientset(federationDomain.DeepCopy())
				c.PrependReactor("update", "federationdomains", func(a kubetesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("some update error")
				})
				return c
			},
			wantError: fmt.Sprintf("could not update federationdomain %s/%s: some update error", federationDomain.Namespace, federationDomain.Name),
			wantActions: []kubetesting.Action{
				kubetesting.NewGetAction(federationDomainGVR, federationDomain.Namespace, federationDomain.Name),
				kubetesting.NewUpdateAction(federationDomainGVR, federationDomain.Namespace, federationDomain),
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			clientset := test.newClientset()
			err := Configurator{
				Clientset: clientset,
			}.CreateOrUpdateFederationDomain(
				context.Background(),
				federationDomain.Namespace,
				federationDomain.Name,
				federationDomain.Spec.Issuer,
			)
			if test.wantError != "" {
				require.EqualError(t, err, test.wantError)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, test.wantActions, clientset.Actions())
		})
	}
}

// nolint:funlen
func TestRecreateIDPForDex(t *testing.T) {
	const (
		dexServiceIP   = "1.2.3.4"
		dexServicePort = 12345
	)

	dexCAData := []byte("some-dex-ca-data")

	dexService := &corev1.Service{
		Spec: corev1.ServiceSpec{
			Type:  corev1.ServiceTypeLoadBalancer,
			Ports: []corev1.ServicePort{corev1.ServicePort{Port: dexServicePort}},
		},
		Status: corev1.ServiceStatus{
			LoadBalancer: corev1.LoadBalancerStatus{
				Ingress: []corev1.LoadBalancerIngress{corev1.LoadBalancerIngress{IP: dexServiceIP}},
			},
		},
	}
	dexTLSSecret := &corev1.Secret{
		Data: map[string][]byte{
			"ca.crt": dexCAData,
		},
	}

	oidcIdentityProviderGVR := idpv1alpha1.SchemeGroupVersion.WithResource("oidcidentityproviders")

	oidcIdentityProvider := &idpv1alpha1.OIDCIdentityProvider{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "pinniped-supervisor",
			Name:      "upstream-oidc-identity-provider",
		},
		Spec: idpv1alpha1.OIDCIdentityProviderSpec{
			Issuer: fmt.Sprintf("https://%s:%d", dexServiceIP, dexServicePort),
			TLS: &idpv1alpha1.TLSSpec{
				CertificateAuthorityData: base64.StdEncoding.EncodeToString(dexCAData),
			},
		},
	}

	tests := []struct {
		name                     string
		kubeClientError          string
		newClientset             func() *pinnipedsupervisorfake.Clientset
		wantError                string
		wantOIDCIdentityProvider *idpv1alpha1.OIDCIdentityProvider
		wantActions              []kubetesting.Action
	}{
		{
			name:            "inspecting dex service endpoint fails",
			kubeClientError: "some get error",
			newClientset: func() *pinnipedsupervisorfake.Clientset {
				return pinnipedsupervisorfake.NewSimpleClientset()
			},
			wantError:   "could not get dex service endpoint: some get error",
			wantActions: []kubetesting.Action{},
		},
		{
			name: "oidcidentityprovider does not exist",
			newClientset: func() *pinnipedsupervisorfake.Clientset {
				return pinnipedsupervisorfake.NewSimpleClientset()
			},
			wantError: fmt.Sprintf(
				`could not get oidcidentityprovider %s/%s: oidcidentityproviders.idp.supervisor.pinniped.dev "upstream-oidc-identity-provider" not found`,
				oidcIdentityProvider.Namespace,
				oidcIdentityProvider.Name,
			),
			wantActions: []kubetesting.Action{
				// We retry 3 times to get the oidcidentityprovider but we fail every time because it
				// doesn't exist.
				kubetesting.NewGetAction(oidcIdentityProviderGVR, oidcIdentityProvider.Namespace, oidcIdentityProvider.Name),
				kubetesting.NewGetAction(oidcIdentityProviderGVR, oidcIdentityProvider.Namespace, oidcIdentityProvider.Name),
				kubetesting.NewGetAction(oidcIdentityProviderGVR, oidcIdentityProvider.Namespace, oidcIdentityProvider.Name),
			},
		},
		{
			name: "oidcidentityprovider exists and is up to date",
			newClientset: func() *pinnipedsupervisorfake.Clientset {
				return pinnipedsupervisorfake.NewSimpleClientset(oidcIdentityProvider.DeepCopy())
			},
			wantOIDCIdentityProvider: oidcIdentityProvider,
			wantActions: []kubetesting.Action{
				kubetesting.NewGetAction(oidcIdentityProviderGVR, oidcIdentityProvider.Namespace, oidcIdentityProvider.Name),
				kubetesting.NewDeleteAction(oidcIdentityProviderGVR, oidcIdentityProvider.Namespace, oidcIdentityProvider.Name),
				kubetesting.NewCreateAction(oidcIdentityProviderGVR, oidcIdentityProvider.Namespace, oidcIdentityProvider),
			},
		},
		{
			name: "oidcidentityprovider exists and needs update",
			newClientset: func() *pinnipedsupervisorfake.Clientset {
				existingOIDCIdentityProvider := oidcIdentityProvider.DeepCopy()
				existingOIDCIdentityProvider.Spec.Issuer = "some-incorrect-issuer"
				existingOIDCIdentityProvider.Spec.TLS = nil
				return pinnipedsupervisorfake.NewSimpleClientset(existingOIDCIdentityProvider)
			},
			wantOIDCIdentityProvider: oidcIdentityProvider,
			wantActions: []kubetesting.Action{
				kubetesting.NewGetAction(oidcIdentityProviderGVR, oidcIdentityProvider.Namespace, oidcIdentityProvider.Name),
				kubetesting.NewDeleteAction(oidcIdentityProviderGVR, oidcIdentityProvider.Namespace, oidcIdentityProvider.Name),
				kubetesting.NewCreateAction(oidcIdentityProviderGVR, oidcIdentityProvider.Namespace, oidcIdentityProvider),
			},
		},
		{
			name: "oidcidentityprovider exists after the first get",
			newClientset: func() *pinnipedsupervisorfake.Clientset {
				c := pinnipedsupervisorfake.NewSimpleClientset(oidcIdentityProvider.DeepCopy())
				once := &sync.Once{}
				c.PrependReactor("get", "oidcidentityproviders", func(a kubetesting.Action) (bool, runtime.Object, error) {
					var err error
					once.Do(func() {
						// When the first get is called, we should return this NotFound error. The second time
						// update is called (i.e., on retry), this Do() func will not run and we will succeed.
						err = kubeerrors.NewNotFound(oidcIdentityProviderGVR.GroupResource(), oidcIdentityProvider.Name)
					})
					return true, oidcIdentityProvider, err
				})
				return c
			},
			wantOIDCIdentityProvider: oidcIdentityProvider,
			wantActions: []kubetesting.Action{
				kubetesting.NewGetAction(oidcIdentityProviderGVR, oidcIdentityProvider.Namespace, oidcIdentityProvider.Name),
				kubetesting.NewGetAction(oidcIdentityProviderGVR, oidcIdentityProvider.Namespace, oidcIdentityProvider.Name),
				kubetesting.NewDeleteAction(oidcIdentityProviderGVR, oidcIdentityProvider.Namespace, oidcIdentityProvider.Name),
				kubetesting.NewCreateAction(oidcIdentityProviderGVR, oidcIdentityProvider.Namespace, oidcIdentityProvider),
			},
		},
		{
			name: "deleting oidcidentityprovider fails",
			newClientset: func() *pinnipedsupervisorfake.Clientset {
				c := pinnipedsupervisorfake.NewSimpleClientset(oidcIdentityProvider.DeepCopy())
				c.PrependReactor("delete", "oidcidentityproviders", func(a kubetesting.Action) (bool, runtime.Object, error) {
					return true, oidcIdentityProvider, errors.New("some delete error")
				})
				return c
			},
			wantError: fmt.Sprintf(
				`could not create oidcidentityprovider %s/%s: oidcidentityproviders.idp.supervisor.pinniped.dev "upstream-oidc-identity-provider" already exists`,
				oidcIdentityProvider.Namespace,
				oidcIdentityProvider.Name,
			),
			wantActions: []kubetesting.Action{
				kubetesting.NewGetAction(oidcIdentityProviderGVR, oidcIdentityProvider.Namespace, oidcIdentityProvider.Name),
				kubetesting.NewDeleteAction(oidcIdentityProviderGVR, oidcIdentityProvider.Namespace, oidcIdentityProvider.Name),
				// We retry 3 times to create the oidcidentityprovider, but fail each time since we failed to delete it.
				kubetesting.NewCreateAction(oidcIdentityProviderGVR, oidcIdentityProvider.Namespace, oidcIdentityProvider),
				kubetesting.NewCreateAction(oidcIdentityProviderGVR, oidcIdentityProvider.Namespace, oidcIdentityProvider),
				kubetesting.NewCreateAction(oidcIdentityProviderGVR, oidcIdentityProvider.Namespace, oidcIdentityProvider),
			},
		},
		{
			name: "deleting oidcidentityprovider takes time to complete",
			newClientset: func() *pinnipedsupervisorfake.Clientset {
				c := pinnipedsupervisorfake.NewSimpleClientset(oidcIdentityProvider.DeepCopy())
				once := &sync.Once{}
				c.PrependReactor("create", "oidcidentityproviders", func(a kubetesting.Action) (bool, runtime.Object, error) {
					var err error
					once.Do(func() {
						// When the first get is called, we should return this Conflict error. The second time
						// update is called (i.e., on retry), this Do() func will not run and we will succeed.
						err = kubeerrors.NewAlreadyExists(oidcIdentityProviderGVR.GroupResource(), oidcIdentityProvider.Name)
					})
					return true, oidcIdentityProvider, err
				})
				return c
			},
			wantOIDCIdentityProvider: oidcIdentityProvider,
			wantActions: []kubetesting.Action{
				kubetesting.NewGetAction(oidcIdentityProviderGVR, oidcIdentityProvider.Namespace, oidcIdentityProvider.Name),
				kubetesting.NewDeleteAction(oidcIdentityProviderGVR, oidcIdentityProvider.Namespace, oidcIdentityProvider.Name),
				kubetesting.NewCreateAction(oidcIdentityProviderGVR, oidcIdentityProvider.Namespace, oidcIdentityProvider),
				kubetesting.NewCreateAction(oidcIdentityProviderGVR, oidcIdentityProvider.Namespace, oidcIdentityProvider),
			},
		},
		{
			name: "creating oidcidentityprovider fails",
			newClientset: func() *pinnipedsupervisorfake.Clientset {
				c := pinnipedsupervisorfake.NewSimpleClientset(oidcIdentityProvider.DeepCopy())
				c.PrependReactor("create", "oidcidentityproviders", func(a kubetesting.Action) (bool, runtime.Object, error) {
					return true, oidcIdentityProvider, errors.New("some create error")
				})
				return c
			},
			wantError: fmt.Sprintf(
				"could not create oidcidentityprovider %s/%s: some create error",
				oidcIdentityProvider.Namespace,
				oidcIdentityProvider.Name,
			),
			wantActions: []kubetesting.Action{
				kubetesting.NewGetAction(oidcIdentityProviderGVR, oidcIdentityProvider.Namespace, oidcIdentityProvider.Name),
				kubetesting.NewDeleteAction(oidcIdentityProviderGVR, oidcIdentityProvider.Namespace, oidcIdentityProvider.Name),
				kubetesting.NewCreateAction(oidcIdentityProviderGVR, oidcIdentityProvider.Namespace, oidcIdentityProvider),
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			kubeClient := kubefake.NewSimpleClientset(dexService)
			if test.kubeClientError != "" {
				kubeClient.PrependReactor("get", "services", func(a kubetesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New(test.kubeClientError)
				})
			}

			clientset := test.newClientset()

			returnedOIDCIdentityProvider, err := Configurator{
				K8SClientset: kubeClient,
				Clientset:    clientset,
			}.RecreateIDPForDex(
				context.Background(),
				dexService.Namespace,
				dexService.Name,
				dexTLSSecret,
			)
			if test.wantError != "" {
				require.EqualError(t, err, test.wantError)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, test.wantOIDCIdentityProvider, returnedOIDCIdentityProvider)
			require.Equal(t, test.wantActions, clientset.Actions())
		})
	}
}
