// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"context"
	"errors"
	"testing"

	certmanagerv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubefake "k8s.io/client-go/kubernetes/fake"
	kubetesting "k8s.io/client-go/testing"
)

func TestRemoveDefaultTLSPort(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantURL   string
		wantError string
	}{
		{
			name:      "invalid url",
			url:       "not a valid url \x01",
			wantError: `cannot parse url: parse "not a valid url \x01": net/url: invalid control character in URL`,
		},
		{
			name:    "contains no port",
			url:     "https://example.com",
			wantURL: "https://example.com",
		},
		{
			name:    "contains non-443 port",
			url:     "https://example.com:12345",
			wantURL: "https://example.com:12345",
		},
		{
			name:    "contains 443 port",
			url:     "https://example.com:443",
			wantURL: "https://example.com",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			gotURL, err := RemoveDefaultTLSPort(test.url)
			if test.wantError != "" {
				require.EqualError(t, err, test.wantError)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, test.wantURL, gotURL)
		})
	}
}

func TestGetSecretFromCert(t *testing.T) {
	secretGVR := corev1.SchemeGroupVersion.WithResource("secrets")
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: "some-namespace", Name: "some-secret"},
	}

	cert := &certmanagerv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{Namespace: secret.Namespace},
		Spec:       certmanagerv1.CertificateSpec{SecretName: secret.Name},
	}

	tests := []struct {
		name          string
		newKubeClient func() *kubefake.Clientset
		wantSecret    *corev1.Secret
		wantError     string
		wantActions   []kubetesting.Action
	}{
		{
			name:          "happy path",
			newKubeClient: func() *kubefake.Clientset { return kubefake.NewSimpleClientset(secret) },
			wantSecret:    secret,
			wantActions: []kubetesting.Action{
				kubetesting.NewGetAction(secretGVR, secret.Namespace, secret.Name),
			},
		},
		{
			name: "secret is not there until 3rd retry",
			newKubeClient: func() *kubefake.Clientset {
				c := kubefake.NewSimpleClientset(secret)
				retries := 2
				c.PrependReactor("get", "secrets", func(_ kubetesting.Action) (bool, runtime.Object, error) {
					if retries == 0 {
						return true, secret, nil
					}
					retries--
					return true, nil, kubeerrors.NewNotFound(secretGVR.GroupResource(), secret.Name)
				})
				return c
			},
			wantSecret: secret,
			wantActions: []kubetesting.Action{
				kubetesting.NewGetAction(secretGVR, secret.Namespace, secret.Name),
				kubetesting.NewGetAction(secretGVR, secret.Namespace, secret.Name),
				kubetesting.NewGetAction(secretGVR, secret.Namespace, secret.Name),
			},
		},
		{
			name: "secret get fails",
			newKubeClient: func() *kubefake.Clientset {
				c := kubefake.NewSimpleClientset(secret)
				c.PrependReactor("get", "secrets", func(_ kubetesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("some get error")
				})
				return c
			},
			wantError: "some get error",
			wantActions: []kubetesting.Action{
				kubetesting.NewGetAction(secretGVR, secret.Namespace, secret.Name),
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			fakeKubeClient := test.newKubeClient()
			gotSecret, err := GetSecretFromCert(context.Background(), fakeKubeClient, cert)
			if test.wantError != "" {
				require.EqualError(t, err, test.wantError)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, test.wantSecret, gotSecret)
			require.Equal(t, test.wantActions, fakeKubeClient.Actions())
		})
	}
}
