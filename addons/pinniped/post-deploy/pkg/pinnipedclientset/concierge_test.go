// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package pinnipedclientset

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	authv1alpha1 "go.pinniped.dev/generated/1.19/apis/concierge/authentication/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	kubedynamicfake "k8s.io/client-go/dynamic/fake"
	kubetesting "k8s.io/client-go/testing"
)

func TestConciergeIsClusterScoped(t *testing.T) {
	const (
		apiGroupSuffix = "some.mackerel.api.group.suffix.com"
	)

	authv1alpha1GV := authv1alpha1.SchemeGroupVersion
	authv1alpha1GV.Group = "authentication.concierge." + apiGroupSuffix

	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(authv1alpha1GV, &authv1alpha1.JWTAuthenticator{}, &authv1alpha1.JWTAuthenticatorList{})

	jwtAuthenticatorGVR := authv1alpha1GV.WithResource("jwtauthenticators")

	namespaceScopedJWTAuthenticator := &authv1alpha1.JWTAuthenticator{
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
	clusterScopedJWTAuthenticator := namespaceScopedJWTAuthenticator.DeepCopy()
	clusterScopedJWTAuthenticator.Namespace = ""

	tests := []struct {
		name                     string
		conciergeIsClusterScoped bool
		jwtAuthenticator         *authv1alpha1.JWTAuthenticator
		wantActions              []kubetesting.Action
	}{
		{
			name:                     "namespace scoped",
			conciergeIsClusterScoped: false,
			jwtAuthenticator:         namespaceScopedJWTAuthenticator,
			wantActions: []kubetesting.Action{
				kubetesting.NewCreateAction(
					jwtAuthenticatorGVR,
					namespaceScopedJWTAuthenticator.Namespace,
					toUnstructured(namespaceScopedJWTAuthenticator, false),
				),
				kubetesting.NewGetAction(
					jwtAuthenticatorGVR,
					namespaceScopedJWTAuthenticator.Namespace,
					namespaceScopedJWTAuthenticator.Name,
				),
			},
		},
		{
			name:                     "cluster scoped",
			conciergeIsClusterScoped: true,
			jwtAuthenticator:         clusterScopedJWTAuthenticator,
			wantActions: []kubetesting.Action{
				kubetesting.NewRootCreateAction(jwtAuthenticatorGVR, toUnstructured(clusterScopedJWTAuthenticator, false)),
				kubetesting.NewRootGetAction(jwtAuthenticatorGVR, clusterScopedJWTAuthenticator.Name),
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			kubeDynamicClient := kubedynamicfake.NewSimpleDynamicClient(scheme)

			jwtAuthenticators := NewConcierge(kubeDynamicClient, apiGroupSuffix, test.conciergeIsClusterScoped).
				AuthenticationV1alpha1().
				JWTAuthenticators(test.jwtAuthenticator.Namespace)

			createdJWTAuthenticator, err := jwtAuthenticators.
				Create(context.Background(), test.jwtAuthenticator, metav1.CreateOptions{})
			require.NoError(t, err)
			require.Equal(t, createdJWTAuthenticator, test.jwtAuthenticator)

			gotJWTAuthenticator, err := jwtAuthenticators.
				Get(context.Background(), test.jwtAuthenticator.Name, metav1.GetOptions{})
			require.NoError(t, err)
			require.Equal(t, gotJWTAuthenticator, createdJWTAuthenticator)

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
