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
	corev1 "k8s.io/api/core/v1"
	kubeerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	kubedynamicfake "k8s.io/client-go/dynamic/fake"
	kubefake "k8s.io/client-go/kubernetes/fake"
	kubetesting "k8s.io/client-go/testing"

	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/post-deploy/pkg/pinnipedclientset"
)

// nolint:funlen
func TestCreateOrUpdateFederationDomain(t *testing.T) {
	const apiGroupSuffix = "some.tuna.api.group.suffix.com"

	configv1alpha1GV := configv1alpha1.SchemeGroupVersion
	configv1alpha1GV.Group = "config.supervisor." + apiGroupSuffix

	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(configv1alpha1GV, &configv1alpha1.FederationDomain{}, &configv1alpha1.FederationDomainList{})

	federationDomainGVR := configv1alpha1GV.WithResource("federationdomains")

	federationDomain := &configv1alpha1.FederationDomain{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "some-namespace",
			Name:      "some-name",
		},
		Spec: configv1alpha1.FederationDomainSpec{
			Issuer: "some-issuer",
		},
	}
	federationDomain.APIVersion, federationDomain.Kind = configv1alpha1GV.WithKind("FederationDomain").ToAPIVersionAndKind()

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
			newKubeDynamicClient: func() *kubedynamicfake.FakeDynamicClient {
				return kubedynamicfake.NewSimpleDynamicClient(scheme)
			},
			wantActions: []kubetesting.Action{
				kubetesting.NewGetAction(federationDomainGVR, federationDomain.Namespace, federationDomain.Name),
				kubetesting.NewCreateAction(federationDomainGVR, federationDomain.Namespace, toUnstructured(federationDomain, true)),
			},
		},
		{
			name: "jwt authenticator does not exist and creating federationdomain fails",
			newKubeDynamicClient: func() *kubedynamicfake.FakeDynamicClient {
				c := kubedynamicfake.NewSimpleDynamicClient(scheme)
				c.PrependReactor("create", "federationdomains", func(a kubetesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("some create error")
				})
				return c
			},
			wantError: fmt.Sprintf("could not create federationdomain %s/%s: some create error", federationDomain.Namespace, federationDomain.Name),
			wantActions: []kubetesting.Action{
				kubetesting.NewGetAction(federationDomainGVR, federationDomain.Namespace, federationDomain.Name),
				kubetesting.NewCreateAction(federationDomainGVR, federationDomain.Namespace, toUnstructured(federationDomain, true)),
			},
		},
		{
			name: "jwt authenticator exists and is up to date",
			newKubeDynamicClient: func() *kubedynamicfake.FakeDynamicClient {
				return kubedynamicfake.NewSimpleDynamicClient(scheme, federationDomain.DeepCopy())
			},
			wantActions: []kubetesting.Action{
				kubetesting.NewGetAction(federationDomainGVR, federationDomain.Namespace, federationDomain.Name),
				kubetesting.NewUpdateAction(federationDomainGVR, federationDomain.Namespace, toUnstructured(federationDomain, false)),
			},
		},
		{
			name: "jwt authenticator exists and is not up to date",
			newKubeDynamicClient: func() *kubedynamicfake.FakeDynamicClient {
				existingFederationDomain := federationDomain.DeepCopy()
				existingFederationDomain.Spec.Issuer = "some-other-issuer"
				return kubedynamicfake.NewSimpleDynamicClient(scheme, federationDomain.DeepCopy())
			},
			wantActions: []kubetesting.Action{
				kubetesting.NewGetAction(federationDomainGVR, federationDomain.Namespace, federationDomain.Name),
				kubetesting.NewUpdateAction(federationDomainGVR, federationDomain.Namespace, toUnstructured(federationDomain, false)),
			},
		},
		{
			name: "updating federationdomain fails",
			newKubeDynamicClient: func() *kubedynamicfake.FakeDynamicClient {
				c := kubedynamicfake.NewSimpleDynamicClient(scheme, federationDomain.DeepCopy())
				c.PrependReactor("update", "federationdomains", func(a kubetesting.Action) (bool, runtime.Object, error) {
					return true, nil, errors.New("some update error")
				})
				return c
			},
			wantError: fmt.Sprintf("could not update federationdomain %s/%s: some update error", federationDomain.Namespace, federationDomain.Name),
			wantActions: []kubetesting.Action{
				kubetesting.NewGetAction(federationDomainGVR, federationDomain.Namespace, federationDomain.Name),
				kubetesting.NewUpdateAction(federationDomainGVR, federationDomain.Namespace, toUnstructured(federationDomain, false)),
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			kubeDynamicClient := test.newKubeDynamicClient()
			err := Configurator{
				Clientset: pinnipedclientset.NewSupervisor(kubeDynamicClient, apiGroupSuffix),
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
			require.Equal(t, test.wantActions, kubeDynamicClient.Actions())
		})
	}
}

// nolint:funlen
func TestRecreateIDPForDex(t *testing.T) {
	const (
		apiGroupSuffix = "some.marlin.api.group.suffix.com"

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

	idpv1alpha1GV := idpv1alpha1.SchemeGroupVersion
	idpv1alpha1GV.Group = "idp.supervisor." + apiGroupSuffix

	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(idpv1alpha1GV, &idpv1alpha1.OIDCIdentityProvider{}, &idpv1alpha1.OIDCIdentityProviderList{})

	oidcIdentityProviderGVR := idpv1alpha1GV.WithResource("oidcidentityproviders")

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
	oidcIdentityProvider.APIVersion, oidcIdentityProvider.Kind = idpv1alpha1GV.WithKind("OIDCIdentityProvider").ToAPIVersionAndKind()

	tests := []struct {
		name                     string
		kubeClientError          string
		newKubeDynamicClient     func() *kubedynamicfake.FakeDynamicClient
		wantError                string
		wantOIDCIdentityProvider *idpv1alpha1.OIDCIdentityProvider
		wantActions              []kubetesting.Action
	}{
		{
			name:            "inspecting dex service endpoint fails",
			kubeClientError: "some get error",
			newKubeDynamicClient: func() *kubedynamicfake.FakeDynamicClient {
				return kubedynamicfake.NewSimpleDynamicClient(scheme)
			},
			wantError:   "could not get dex service endpoint: some get error",
			wantActions: []kubetesting.Action{},
		},
		{
			name: "oidcidentityprovider does not exist",
			newKubeDynamicClient: func() *kubedynamicfake.FakeDynamicClient {
				return kubedynamicfake.NewSimpleDynamicClient(scheme)
			},
			wantError: fmt.Sprintf(
				`could not get oidcidentityprovider %s/%s: oidcidentityproviders.idp.supervisor.some.marlin.api.group.suffix.com "upstream-oidc-identity-provider" not found`,
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
			newKubeDynamicClient: func() *kubedynamicfake.FakeDynamicClient {
				return kubedynamicfake.NewSimpleDynamicClient(scheme, oidcIdentityProvider.DeepCopy())
			},
			wantOIDCIdentityProvider: oidcIdentityProvider,
			wantActions: []kubetesting.Action{
				kubetesting.NewGetAction(oidcIdentityProviderGVR, oidcIdentityProvider.Namespace, oidcIdentityProvider.Name),
				kubetesting.NewDeleteAction(oidcIdentityProviderGVR, oidcIdentityProvider.Namespace, oidcIdentityProvider.Name),
				kubetesting.NewCreateAction(oidcIdentityProviderGVR, oidcIdentityProvider.Namespace, toUnstructured(oidcIdentityProvider, false)),
			},
		},
		{
			name: "oidcidentityprovider exists and needs update",
			newKubeDynamicClient: func() *kubedynamicfake.FakeDynamicClient {
				existingOIDCIdentityProvider := oidcIdentityProvider.DeepCopy()
				existingOIDCIdentityProvider.Spec.Issuer = "some-incorrect-issuer"
				existingOIDCIdentityProvider.Spec.TLS = nil
				return kubedynamicfake.NewSimpleDynamicClient(scheme, existingOIDCIdentityProvider)
			},
			wantOIDCIdentityProvider: oidcIdentityProvider,
			wantActions: []kubetesting.Action{
				kubetesting.NewGetAction(oidcIdentityProviderGVR, oidcIdentityProvider.Namespace, oidcIdentityProvider.Name),
				kubetesting.NewDeleteAction(oidcIdentityProviderGVR, oidcIdentityProvider.Namespace, oidcIdentityProvider.Name),
				kubetesting.NewCreateAction(oidcIdentityProviderGVR, oidcIdentityProvider.Namespace, toUnstructured(oidcIdentityProvider, false)),
			},
		},
		{
			name: "oidcidentityprovider exists after the first get",
			newKubeDynamicClient: func() *kubedynamicfake.FakeDynamicClient {
				c := kubedynamicfake.NewSimpleDynamicClient(scheme, oidcIdentityProvider.DeepCopy())
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
				kubetesting.NewCreateAction(oidcIdentityProviderGVR, oidcIdentityProvider.Namespace, toUnstructured(oidcIdentityProvider, false)),
			},
		},
		{
			name: "deleting oidcidentityprovider fails",
			newKubeDynamicClient: func() *kubedynamicfake.FakeDynamicClient {
				c := kubedynamicfake.NewSimpleDynamicClient(scheme, oidcIdentityProvider.DeepCopy())
				c.PrependReactor("delete", "oidcidentityproviders", func(a kubetesting.Action) (bool, runtime.Object, error) {
					return true, oidcIdentityProvider, errors.New("some delete error")
				})
				return c
			},
			wantError: fmt.Sprintf(
				`could not create oidcidentityprovider %s/%s: oidcidentityproviders.idp.supervisor.some.marlin.api.group.suffix.com "upstream-oidc-identity-provider" already exists`,
				oidcIdentityProvider.Namespace,
				oidcIdentityProvider.Name,
			),
			wantActions: []kubetesting.Action{
				kubetesting.NewGetAction(oidcIdentityProviderGVR, oidcIdentityProvider.Namespace, oidcIdentityProvider.Name),
				kubetesting.NewDeleteAction(oidcIdentityProviderGVR, oidcIdentityProvider.Namespace, oidcIdentityProvider.Name),
				// We retry 3 times to create the oidcidentityprovider, but fail each time since we failed to delete it.
				kubetesting.NewCreateAction(oidcIdentityProviderGVR, oidcIdentityProvider.Namespace, toUnstructured(oidcIdentityProvider, false)),
				kubetesting.NewCreateAction(oidcIdentityProviderGVR, oidcIdentityProvider.Namespace, toUnstructured(oidcIdentityProvider, false)),
				kubetesting.NewCreateAction(oidcIdentityProviderGVR, oidcIdentityProvider.Namespace, toUnstructured(oidcIdentityProvider, false)),
			},
		},
		{
			name: "deleting oidcidentityprovider takes time to complete",
			newKubeDynamicClient: func() *kubedynamicfake.FakeDynamicClient {
				c := kubedynamicfake.NewSimpleDynamicClient(scheme, oidcIdentityProvider.DeepCopy())
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
				kubetesting.NewCreateAction(oidcIdentityProviderGVR, oidcIdentityProvider.Namespace, toUnstructured(oidcIdentityProvider, false)),
				kubetesting.NewCreateAction(oidcIdentityProviderGVR, oidcIdentityProvider.Namespace, toUnstructured(oidcIdentityProvider, false)),
			},
		},
		{
			name: "creating oidcidentityprovider fails",
			newKubeDynamicClient: func() *kubedynamicfake.FakeDynamicClient {
				c := kubedynamicfake.NewSimpleDynamicClient(scheme, oidcIdentityProvider.DeepCopy())
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
				kubetesting.NewCreateAction(oidcIdentityProviderGVR, oidcIdentityProvider.Namespace, toUnstructured(oidcIdentityProvider, false)),
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

			kubeDynamicClient := test.newKubeDynamicClient()

			returnedOIDCIdentityProvider, err := Configurator{
				K8SClientset: kubeClient,
				Clientset:    pinnipedclientset.NewSupervisor(kubeDynamicClient, apiGroupSuffix),
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
