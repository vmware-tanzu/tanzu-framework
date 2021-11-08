// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package concierge

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	authv1alpha1 "go.pinniped.dev/generated/1.19/apis/concierge/authentication/v1alpha1"
	pinnipedconciergefake "go.pinniped.dev/generated/1.19/client/concierge/clientset/versioned/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubetesting "k8s.io/client-go/testing"
)

// nolint:funlen
func TestCreateOrUpdateJWTAuthenticator(t *testing.T) {
	jwtAuthenticatorGVR := authv1alpha1.SchemeGroupVersion.WithResource("jwtauthenticators")

	jwtAuthenticator := &authv1alpha1.JWTAuthenticator{
		ObjectMeta: metav1.ObjectMeta{
			Name: "some-name",
		},
		Spec: authv1alpha1.JWTAuthenticatorSpec{
			Issuer:   "some-issuer",
			Audience: "some-audience",
			TLS: &authv1alpha1.TLSSpec{
				CertificateAuthorityData: "some-ca-data",
			},
		},
	}

	tests := []struct {
		name         string
		newClientset func() *pinnipedconciergefake.Clientset
		wantError    string
		wantActions  []kubetesting.Action
	}{
		{
			name: "getting jwt authenticator fails",
			newClientset: func() *pinnipedconciergefake.Clientset {
				c := pinnipedconciergefake.NewSimpleClientset()
				c.PrependReactor("get", "jwtauthenticators", func(a kubetesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("some get error")
				})
				return c
			},
			wantError: fmt.Sprintf("could not get jwtauthenticator %s: some get error", jwtAuthenticator.Name),
			wantActions: []kubetesting.Action{
				kubetesting.NewRootGetAction(jwtAuthenticatorGVR, jwtAuthenticator.Name),
			},
		},
		{
			name: "jwt authenticator does not exist",
			newClientset: func() *pinnipedconciergefake.Clientset {
				return pinnipedconciergefake.NewSimpleClientset()
			},
			wantActions: []kubetesting.Action{
				kubetesting.NewRootGetAction(jwtAuthenticatorGVR, jwtAuthenticator.Name),
				kubetesting.NewRootCreateAction(jwtAuthenticatorGVR, jwtAuthenticator),
			},
		},
		{
			name: "jwt authenticator does not exist and creating jwtauthenticator fails",
			newClientset: func() *pinnipedconciergefake.Clientset {
				c := pinnipedconciergefake.NewSimpleClientset()
				c.PrependReactor("create", "jwtauthenticators", func(a kubetesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("some create error")
				})
				return c
			},
			wantError: fmt.Sprintf("could not create jwtauthenticator %s: some create error", jwtAuthenticator.Name),
			wantActions: []kubetesting.Action{
				kubetesting.NewRootGetAction(jwtAuthenticatorGVR, jwtAuthenticator.Name),
				kubetesting.NewRootCreateAction(jwtAuthenticatorGVR, jwtAuthenticator),
			},
		},
		{
			name: "jwt authenticator exists and is up to date",
			newClientset: func() *pinnipedconciergefake.Clientset {
				return pinnipedconciergefake.NewSimpleClientset(jwtAuthenticator.DeepCopy())
			},
			wantActions: []kubetesting.Action{
				kubetesting.NewRootGetAction(jwtAuthenticatorGVR, jwtAuthenticator.Name),
				kubetesting.NewRootUpdateAction(jwtAuthenticatorGVR, jwtAuthenticator),
			},
		},
		{
			name: "jwt authenticator exists and is not up to date",
			newClientset: func() *pinnipedconciergefake.Clientset {
				existingJWTAuthenticator := jwtAuthenticator.DeepCopy()
				existingJWTAuthenticator.Spec.Issuer = "some-other-issuer"
				existingJWTAuthenticator.Spec.Audience = "some-other-audience"
				existingJWTAuthenticator.Spec.TLS = nil
				return pinnipedconciergefake.NewSimpleClientset(jwtAuthenticator.DeepCopy())
			},
			wantActions: []kubetesting.Action{
				kubetesting.NewRootGetAction(jwtAuthenticatorGVR, jwtAuthenticator.Name),
				kubetesting.NewRootUpdateAction(jwtAuthenticatorGVR, jwtAuthenticator),
			},
		},
		{
			name: "updating jwtauthenticator fails",
			newClientset: func() *pinnipedconciergefake.Clientset {
				c := pinnipedconciergefake.NewSimpleClientset(jwtAuthenticator.DeepCopy())
				c.PrependReactor("update", "jwtauthenticators", func(a kubetesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("some update error")
				})
				return c
			},
			wantError: fmt.Sprintf("could not update jwtauthenticator %s: some update error", jwtAuthenticator.Name),
			wantActions: []kubetesting.Action{
				kubetesting.NewRootGetAction(jwtAuthenticatorGVR, jwtAuthenticator.Name),
				kubetesting.NewRootUpdateAction(jwtAuthenticatorGVR, jwtAuthenticator),
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			clientset := test.newClientset()
			err := Configurator{
				Clientset: clientset,
			}.CreateOrUpdateJWTAuthenticator(
				context.Background(),
				jwtAuthenticator.Namespace,
				jwtAuthenticator.Name,
				jwtAuthenticator.Spec.Issuer,
				jwtAuthenticator.Spec.Audience,
				jwtAuthenticator.Spec.TLS.CertificateAuthorityData,
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
