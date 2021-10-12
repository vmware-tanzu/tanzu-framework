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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	kubedynamicfake "k8s.io/client-go/dynamic/fake"
	kubetesting "k8s.io/client-go/testing"

	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/post-deploy/pkg/pinnipedclientset"
)

// nolint:funlen
func TestCreateOrUpdateJWTAuthenticator(t *testing.T) {
	const apiGroupSuffix = "some.fish.api.group.suffix.com"

	authv1alpha1GV := authv1alpha1.SchemeGroupVersion
	authv1alpha1GV.Group = "authentication.concierge." + apiGroupSuffix

	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(authv1alpha1GV, &authv1alpha1.JWTAuthenticator{}, &authv1alpha1.JWTAuthenticatorList{})

	jwtAuthenticatorGVR := authv1alpha1GV.WithResource("jwtauthenticators")

	jwtAuthenticator := &authv1alpha1.JWTAuthenticator{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "some-namespace",
			Name:      "some-name",
		},
		Spec: authv1alpha1.JWTAuthenticatorSpec{
			Issuer:   "some-issuer",
			Audience: "some-audience",
			TLS: &authv1alpha1.TLSSpec{
				CertificateAuthorityData: "some-ca-data",
			},
		},
	}
	jwtAuthenticator.APIVersion, jwtAuthenticator.Kind = authv1alpha1GV.WithKind("JWTAuthenticator").ToAPIVersionAndKind()

	tests := []struct {
		name                 string
		newKubeDynamicClient func() *kubedynamicfake.FakeDynamicClient
		wantError            string
		wantActions          []kubetesting.Action
	}{
		{
			name: "getting jwt authenticator fails",
			newKubeDynamicClient: func() *kubedynamicfake.FakeDynamicClient {
				c := kubedynamicfake.NewSimpleDynamicClient(scheme)
				c.PrependReactor("get", "jwtauthenticators", func(a kubetesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("some get error")
				})
				return c
			},
			wantError: fmt.Sprintf("could not get jwtauthenticator %s: some get error", jwtAuthenticator.Name),
			wantActions: []kubetesting.Action{
				kubetesting.NewGetAction(jwtAuthenticatorGVR, jwtAuthenticator.Namespace, jwtAuthenticator.Name),
			},
		},
		{
			name: "jwt authenticator does not exist",
			newKubeDynamicClient: func() *kubedynamicfake.FakeDynamicClient {
				return kubedynamicfake.NewSimpleDynamicClient(scheme)
			},
			wantActions: []kubetesting.Action{
				kubetesting.NewGetAction(jwtAuthenticatorGVR, jwtAuthenticator.Namespace, jwtAuthenticator.Name),
				kubetesting.NewCreateAction(jwtAuthenticatorGVR, jwtAuthenticator.Namespace, toUnstructured(jwtAuthenticator, true)),
			},
		},
		{
			name: "jwt authenticator does not exist and creating jwtauthenticator fails",
			newKubeDynamicClient: func() *kubedynamicfake.FakeDynamicClient {
				c := kubedynamicfake.NewSimpleDynamicClient(scheme)
				c.PrependReactor("create", "jwtauthenticators", func(a kubetesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("some create error")
				})
				return c
			},
			wantError: fmt.Sprintf("could not create jwtauthenticator %s: some create error", jwtAuthenticator.Name),
			wantActions: []kubetesting.Action{
				kubetesting.NewGetAction(jwtAuthenticatorGVR, jwtAuthenticator.Namespace, jwtAuthenticator.Name),
				kubetesting.NewCreateAction(jwtAuthenticatorGVR, jwtAuthenticator.Namespace, toUnstructured(jwtAuthenticator, true)),
			},
		},
		{
			name: "jwt authenticator exists and is up to date",
			newKubeDynamicClient: func() *kubedynamicfake.FakeDynamicClient {
				return kubedynamicfake.NewSimpleDynamicClient(scheme, jwtAuthenticator.DeepCopy())
			},
			wantActions: []kubetesting.Action{
				kubetesting.NewGetAction(jwtAuthenticatorGVR, jwtAuthenticator.Namespace, jwtAuthenticator.Name),
				kubetesting.NewUpdateAction(jwtAuthenticatorGVR, jwtAuthenticator.Namespace, toUnstructured(jwtAuthenticator, false)),
			},
		},
		{
			name: "jwt authenticator exists and is not up to date",
			newKubeDynamicClient: func() *kubedynamicfake.FakeDynamicClient {
				existingJWTAuthenticator := jwtAuthenticator.DeepCopy()
				existingJWTAuthenticator.Spec.Issuer = "some-other-issuer"
				existingJWTAuthenticator.Spec.Audience = "some-other-audience"
				existingJWTAuthenticator.Spec.TLS = nil
				return kubedynamicfake.NewSimpleDynamicClient(scheme, jwtAuthenticator.DeepCopy())
			},
			wantActions: []kubetesting.Action{
				kubetesting.NewGetAction(jwtAuthenticatorGVR, jwtAuthenticator.Namespace, jwtAuthenticator.Name),
				kubetesting.NewUpdateAction(jwtAuthenticatorGVR, jwtAuthenticator.Namespace, toUnstructured(jwtAuthenticator, false)),
			},
		},
		{
			name: "updating jwtauthenticator fails",
			newKubeDynamicClient: func() *kubedynamicfake.FakeDynamicClient {
				c := kubedynamicfake.NewSimpleDynamicClient(scheme, jwtAuthenticator.DeepCopy())
				c.PrependReactor("update", "jwtauthenticators", func(a kubetesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("some update error")
				})
				return c
			},
			wantError: fmt.Sprintf("could not update jwtauthenticator %s: some update error", jwtAuthenticator.Name),
			wantActions: []kubetesting.Action{
				kubetesting.NewGetAction(jwtAuthenticatorGVR, jwtAuthenticator.Namespace, jwtAuthenticator.Name),
				kubetesting.NewUpdateAction(jwtAuthenticatorGVR, jwtAuthenticator.Namespace, toUnstructured(jwtAuthenticator, false)),
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			kubeDynamicClient := test.newKubeDynamicClient()
			err := Configurator{
				Clientset: pinnipedclientset.NewConcierge(kubeDynamicClient, apiGroupSuffix),
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
			require.Equal(t, test.wantActions, kubeDynamicClient.Actions())
		})
	}
}

func toUnstructured(obj runtime.Object, removeTypeMeta bool) runtime.Object {
	unstructuredObjData, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		panic(err)
	}

	if removeTypeMeta {
		delete(unstructuredObjData, "apiVersion")
		delete(unstructuredObjData, "kind")
	}

	return &unstructured.Unstructured{Object: unstructuredObjData}
}
