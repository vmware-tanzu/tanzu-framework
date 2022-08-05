// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package configure

import (
	"context"
	"encoding/base64"
	"fmt"
	"testing"

	certmanagerv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	certmanagerfake "github.com/jetstack/cert-manager/pkg/client/clientset/versioned/fake"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubefake "k8s.io/client-go/kubernetes/fake"
	kubetesting "k8s.io/client-go/testing"
	"k8s.io/utils/pointer"

	authv1alpha1 "go.pinniped.dev/generated/1.20/apis/concierge/authentication/v1alpha1"
	configv1alpha1 "go.pinniped.dev/generated/1.20/apis/supervisor/config/v1alpha1"
	idpv1alpha1 "go.pinniped.dev/generated/1.20/apis/supervisor/idp/v1alpha1"
	supervisoridpv1alpha1 "go.pinniped.dev/generated/1.20/apis/supervisor/idp/v1alpha1"
	pinnipedconciergefake "go.pinniped.dev/generated/1.20/client/concierge/clientset/versioned/fake"
	pinnipedsupervisorfake "go.pinniped.dev/generated/1.20/client/supervisor/clientset/versioned/fake"

	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/post-deploy/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/post-deploy/pkg/inspect"
	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/post-deploy/pkg/vars"
)

func TestTKGAuthentication(t *testing.T) {

	const (
		supervisorNamespace = "pinniped-supervisor" // vars.SupervisorNamespace default
	)

	enableLogging() // Comment me out for less verbose test logs

	supervisorCertificateSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: supervisorNamespace,
			Name:      "some-supervisor-certificate-secret-name",
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			"ca.crt": []byte("some-ca-bundle-data"),
		},
	}

	supervisorService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: supervisorNamespace,
			// This is set via flag on post-deploy job, the default value is empty string. This simplifies the test.
			Name: "",
		},
		Spec: corev1.ServiceSpec{
			Type:  corev1.ServiceTypeLoadBalancer,
			Ports: []corev1.ServicePort{{Port: 12345}},
		},
		Status: corev1.ServiceStatus{
			LoadBalancer: corev1.LoadBalancerStatus{
				Ingress: []corev1.LoadBalancerIngress{{IP: "1.2.3.4"}},
			},
		},
	}

	supervisorCertificate := &certmanagerv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: supervisorNamespace,
			// This is set via flag on post-deploy job, the default value is empty string. This simplifies the test.
			Name: "",
		},
		Spec: certmanagerv1.CertificateSpec{
			SecretName: supervisorCertificateSecret.Name,
			IPAddresses: []string{
				supervisorService.Status.LoadBalancer.Ingress[0].IP,
			},
		},
	}

	federationDomainGVR := configv1alpha1.SchemeGroupVersion.WithResource("federationdomains")
	federationDomain := &configv1alpha1.FederationDomain{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: supervisorNamespace,
			Name:      "my-federation-domain",
		},
		Spec: configv1alpha1.FederationDomainSpec{
			Issuer: serviceHTTPSEndpoint(supervisorService),
		},
	}

	jwtAuthenticatorGVR := authv1alpha1.SchemeGroupVersion.WithResource("jwtauthenticators")
	jwtAuthenticator := func(audience string) *authv1alpha1.JWTAuthenticator {
		return &authv1alpha1.JWTAuthenticator{
			ObjectMeta: metav1.ObjectMeta{
				Name: "my-cool-authenticator",
			},
			Spec: authv1alpha1.JWTAuthenticatorSpec{
				Issuer:   federationDomain.Spec.Issuer,
				Audience: audience,
				TLS: &authv1alpha1.TLSSpec{
					CertificateAuthorityData: base64.StdEncoding.EncodeToString(supervisorCertificateSecret.Data["ca.crt"]),
				},
			},
		}
	}

	tkgMetadataConfigmapManagementCluster := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: constants.TKGSystemPublicNamespace,
			Name:      constants.TKGMetaConfigMapName,
		},
		Data: map[string]string{
			"metadata.yaml": `cluster:
  name: tkg-mgmt-vc
  type: management
  plan: dev
  kubernetesProvider: VMware Tanzu Kubernetes Grid
  tkgVersion: v1.6.0-zshippable
  edition: tkg`,
		},
	}
	tkgMetadataConfigmapWorkloadCluster := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: constants.TKGSystemPublicNamespace,
			Name:      constants.TKGMetaConfigMapName,
		},
		Data: map[string]string{
			"metadata.yaml": `cluster:
  name: some-workload-cluster
  type: workload
  plan: dev
  kubernetesProvider: VMware Tanzu Kubernetes Grid
  tkgVersion: v1.6.0-zshippable
  edition: tkg`,
		},
	}

	conciergeDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "pinniped-concierge",
			Name:      "pinniped-concierge",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(1),
		},
		Status: appsv1.DeploymentStatus{
			Replicas:      int32(1),
			ReadyReplicas: int32(1),
		},
	}
	supervisorDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "pinniped-supervisor",
			Name:      "pinniped-supervisor",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(1),
		},
		Status: appsv1.DeploymentStatus{
			Replicas:      int32(1),
			ReadyReplicas: int32(1),
		},
	}

	oidcIdentityProviderGVR := idpv1alpha1.SchemeGroupVersion.WithResource("oidcidentityproviders")
	upstreamIdentityProvider := &supervisoridpv1alpha1.OIDCIdentityProvider{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "pinniped-supervisor",
			Name:      "upstream-oidc-identity-provider",
		},
		Spec: supervisoridpv1alpha1.OIDCIdentityProviderSpec{
			Issuer: "https://identityforall.com",
			Claims: supervisoridpv1alpha1.OIDCClaims{
				Username: "email",
			},
			Client: supervisoridpv1alpha1.OIDCClient{
				SecretName: "some-upstream-oidc-client",
			},
		},
		Status: supervisoridpv1alpha1.OIDCIdentityProviderStatus{
			Phase: supervisoridpv1alpha1.PhaseReady,
		},
	}

	// NOTE: these tests exercise what is essentially glue code between TKGs & TKGm to make uTKG functional for the Pinniped Package.
	// There is a combination of flags, defaulting values and querying a configmap on cluster that can ultimately lead to a functioning
	// pinniped installation.  These tests are valuable as this is difficult to reason about without knowledge if the external factors.
	tests := []struct {
		name                        string
		newKubeClient               func() *kubefake.Clientset
		newCertManagerClient        func() *certmanagerfake.Clientset
		newSupervisorClient         func() *pinnipedsupervisorfake.Clientset
		newConciergeClient          func() *pinnipedconciergefake.Clientset
		finalSetup                  func()
		cleanup                     func()
		wantError                   string
		wantConciergeClientActions  []kubetesting.Action
		wantSupervisorClientActions []kubetesting.Action
	}{
		{
			name: "TKGm Management Cluster Case: The JWTAuthenticator audience is set to the Supervisor Service Endpoint value and the cluster.type is derived from tkg-metadata configmap and the pinniped supervisor will be deployed",
			newKubeClient: func() *kubefake.Clientset {
				c := kubefake.NewSimpleClientset(
					tkgMetadataConfigmapManagementCluster,
					conciergeDeployment,
					supervisorDeployment,
					supervisorService,
					supervisorCertificateSecret,
				)
				c.PrependReactor("delete", "secrets", func(action kubetesting.Action) (bool, runtime.Object, error) {
					// When we delete the secret in the implementation, we expect the cert-manager controller
					// to recreate it. Since there is no cert-manager controller running, let's just tell the
					// fake kube client to not delete it (i.e., that we "handled" the delete ourselves).
					return actionIsOnObject(action, supervisorCertificateSecret), nil, nil
				})
				return c
			},
			newCertManagerClient: func() *certmanagerfake.Clientset {
				return certmanagerfake.NewSimpleClientset(supervisorCertificate)
			},
			newSupervisorClient: func() *pinnipedsupervisorfake.Clientset {
				return pinnipedsupervisorfake.NewSimpleClientset(
					upstreamIdentityProvider,
					federationDomain,
				)
			},
			newConciergeClient: func() *pinnipedconciergefake.Clientset {
				defaultJWTAuthenticator := jwtAuthenticator("initial-incorrect-audience")
				return pinnipedconciergefake.NewSimpleClientset(defaultJWTAuthenticator)
			},
			// In the TKGm use case on a Management Cluster, the Audience is set to the supervisor service endpoint value.
			// This is probably a little weird, it is definitely unexpected.
			// But it is likely a unique value amongst the fleet of clusters and therefore seems to work today.
			// If this is ever changed (in production code) to be a more expected audience (e.g. cluster.name-cluster.uid)
			// it would invalidate existing deployed kubeconfigs.  Proceed with caution!
			wantConciergeClientActions: []kubetesting.Action{
				kubetesting.NewRootGetAction(jwtAuthenticatorGVR, jwtAuthenticator("").Name),
				kubetesting.NewRootUpdateAction(jwtAuthenticatorGVR, jwtAuthenticator("https://1.2.3.4:12345")),
			},
			wantSupervisorClientActions: []kubetesting.Action{
				kubetesting.NewGetAction(oidcIdentityProviderGVR, upstreamIdentityProvider.Namespace, upstreamIdentityProvider.Name),
				kubetesting.NewGetAction(federationDomainGVR, federationDomain.Namespace, federationDomain.Name),
				kubetesting.NewUpdateAction(federationDomainGVR, federationDomain.Namespace, &configv1alpha1.FederationDomain{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "my-federation-domain",
						Namespace: "pinniped-supervisor",
					},
					Spec: configv1alpha1.FederationDomainSpec{
						Issuer: "https://1.2.3.4:12345",
						TLS:    nil, // when nil, Pinniped will use trusted CA's compiled into the container image
					},
				}),
			},
		},
		{
			name: "TKGm Workload Cluster Case: When the tkg-metadata configmap is found, the cluster.name and cluster.type is used from the configmap to configure pinniped and set the concierge jwtauthenticator audience and supervisor will not be deployed",
			newKubeClient: func() *kubefake.Clientset {
				c := kubefake.NewSimpleClientset(
					tkgMetadataConfigmapWorkloadCluster,
					conciergeDeployment,
					supervisorDeployment,
					supervisorService,
					supervisorCertificateSecret,
				)
				c.PrependReactor("delete", "secrets", func(action kubetesting.Action) (bool, runtime.Object, error) {
					// When we delete the secret in the implementation, we expect the cert-manager controller
					// to recreate it. Since there is no cert-manager controller running, let's just tell the
					// fake kube client to not delete it (i.e., that we "handled" the delete ourselves).
					return actionIsOnObject(action, supervisorCertificateSecret), nil, nil
				})
				return c
			},
			newCertManagerClient: func() *certmanagerfake.Clientset {
				return certmanagerfake.NewSimpleClientset(supervisorCertificate)
			},
			newSupervisorClient: func() *pinnipedsupervisorfake.Clientset {
				return pinnipedsupervisorfake.NewSimpleClientset(
					upstreamIdentityProvider,
				)
			},
			newConciergeClient: func() *pinnipedconciergefake.Clientset {
				defaultJWTAuthenticator := jwtAuthenticator("initial-incorrect-audience")
				return pinnipedconciergefake.NewSimpleClientset(defaultJWTAuthenticator)
			},
			wantConciergeClientActions: []kubetesting.Action{
				kubetesting.NewRootGetAction(jwtAuthenticatorGVR, jwtAuthenticator("").Name),
				kubetesting.NewRootUpdateAction(jwtAuthenticatorGVR, jwtAuthenticator("some-workload-cluster")),
			},
			// this is not a mgmt cluster, so no supervisor actions should take place
			wantSupervisorClientActions: []kubetesting.Action{},
			finalSetup: func() {
				vars.SupervisorSvcEndpoint = "https://1.2.3.4:12345" // mgmt cluster figures this out by itself, but workload does not.
			},
			cleanup: func() {
				vars.SupervisorSvcEndpoint = ""
			},
		},
		{
			name: "TKGs Workload Cluster Case: When the --jwtauthenticator-audience flag is provided then its value becomes the jwtauthenticator audience and the cluster.type is assumed to be workload which means supervisor will not be deployed",
			newKubeClient: func() *kubefake.Clientset {
				c := kubefake.NewSimpleClientset(
					conciergeDeployment,
					supervisorDeployment,
					supervisorService,
					supervisorCertificateSecret,
				)
				c.PrependReactor("delete", "secrets", func(action kubetesting.Action) (bool, runtime.Object, error) {
					// When we delete the secret in the implementation, we expect the cert-manager controller
					// to recreate it. Since there is no cert-manager controller running, let's just tell the
					// fake kube client to not delete it (i.e., that we "handled" the delete ourselves).
					return actionIsOnObject(action, supervisorCertificateSecret), nil, nil
				})
				return c
			},
			newCertManagerClient: func() *certmanagerfake.Clientset {
				return certmanagerfake.NewSimpleClientset(supervisorCertificate)
			},
			newSupervisorClient: func() *pinnipedsupervisorfake.Clientset {
				// return a fake idp when requsted
				return pinnipedsupervisorfake.NewSimpleClientset(
					upstreamIdentityProvider,
				)
			},
			newConciergeClient: func() *pinnipedconciergefake.Clientset {
				defaultJWTAuthenticator := jwtAuthenticator("initial-incorrect-audience")
				return pinnipedconciergefake.NewSimpleClientset(defaultJWTAuthenticator)
			},
			finalSetup: func() {
				// this var emulate the set flag --jwtauthenticator-audience=amazing.cluster-1234567890
				// therefore:
				// - audience=amazing.cluster-1234567890 is set as Audience
				// - Cluster.Type is inferred to be "workload"
				vars.JWTAuthenticatorAudience = "amazing.cluster-1234567890"
				vars.SupervisorSvcEndpoint = "https://1.2.3.4:12345" // mgmt cluster figures this out by itself, but workload does not.
			},
			cleanup: func() {
				vars.JWTAuthenticatorAudience = ""
				vars.SupervisorSvcEndpoint = ""
			},
			wantConciergeClientActions: []kubetesting.Action{
				kubetesting.NewRootGetAction(jwtAuthenticatorGVR, jwtAuthenticator("").Name),
				kubetesting.NewRootUpdateAction(jwtAuthenticatorGVR, jwtAuthenticator("amazing.cluster-1234567890")),
			},
			// this is not a mgmt cluster, so no supervisor actions should take place
			wantSupervisorClientActions: []kubetesting.Action{},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			vars.JWTAuthenticatorName = "my-cool-authenticator"
			vars.FederationDomainName = "my-federation-domain"
			vars.SupervisorCABundleData = base64.StdEncoding.EncodeToString([]byte("some-ca-bundle-data"))
			t.Cleanup(func() {
				vars.JWTAuthenticatorName = ""
				vars.FederationDomainName = ""
				vars.SupervisorCABundleData = ""
			})
			fakeKubeClient := test.newKubeClient()
			fakeCertManagerClient := test.newCertManagerClient()
			supervisorClientset := test.newSupervisorClient()
			conciergeClientset := test.newConciergeClient()
			clients := Clients{
				K8SClientset:         fakeKubeClient,
				CertmanagerClientset: fakeCertManagerClient,
				SupervisorClientset:  supervisorClientset,
				ConciergeClientset:   conciergeClientset,
			}

			if test.finalSetup != nil {
				test.finalSetup()
			}
			t.Cleanup(func() {
				if test.cleanup != nil {
					test.cleanup()
				}
			})
			err := TKGAuthentication(clients)

			require.Equal(t, test.wantSupervisorClientActions, supervisorClientset.Actions())
			require.Equal(t, test.wantConciergeClientActions, conciergeClientset.Actions())

			if test.wantError != "" {
				require.EqualError(t, err, test.wantError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPinniped(t *testing.T) {
	const (
		supervisorNamespace = "pinniped-supervisor" // vars.SupervisorNamespace default
	)

	enableLogging() // Comment me out for less verbose test logs

	podGVR := corev1.SchemeGroupVersion.WithResource("pods")
	podGVK := corev1.SchemeGroupVersion.WithKind("Pod")
	supervisorPods := []*corev1.Pod{
		{ObjectMeta: metav1.ObjectMeta{Namespace: supervisorNamespace, Name: "pinniped-supervisor-abc"}},
		{ObjectMeta: metav1.ObjectMeta{Namespace: supervisorNamespace, Name: "pinniped-supervisor-def"}},
	}

	secretGVR := corev1.SchemeGroupVersion.WithResource("secrets")
	supervisorCertificateSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: supervisorNamespace,
			Name:      "some-supervisor-certificate-secret-name",
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			"ca.crt": []byte("some-ca-bundle-data"),
		},
	}

	serviceGVR := corev1.SchemeGroupVersion.WithResource("services")
	supervisorService := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: supervisorNamespace,
			Name:      "some-supervisor-service-name",
		},
		Spec: corev1.ServiceSpec{
			Type:  corev1.ServiceTypeLoadBalancer,
			Ports: []corev1.ServicePort{{Port: 12345}},
		},
		Status: corev1.ServiceStatus{
			LoadBalancer: corev1.LoadBalancerStatus{
				Ingress: []corev1.LoadBalancerIngress{{IP: "1.2.3.4"}},
			},
		},
	}

	configMapGVR := corev1.SchemeGroupVersion.WithResource("configmaps")
	pinnipedInfoConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "kube-public",
			Name:      "pinniped-info",
		},
		Data: map[string]string{
			"cluster_name":                "some-pinniped-info-management-cluster-name",
			"issuer":                      serviceHTTPSEndpoint(supervisorService),
			"issuer_ca_bundle_data":       base64.StdEncoding.EncodeToString(supervisorCertificateSecret.Data["ca.crt"]),
			"concierge_is_cluster_scoped": "true",
		},
	}

	certificateGVR := certmanagerv1.SchemeGroupVersion.WithResource("certificates")
	supervisorCertificate := &certmanagerv1.Certificate{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: supervisorNamespace,
			Name:      "some-supervisor-certificate-name",
		},
		Spec: certmanagerv1.CertificateSpec{
			SecretName: supervisorCertificateSecret.Name,
			IPAddresses: []string{
				supervisorService.Status.LoadBalancer.Ingress[0].IP,
			},
		},
	}

	federationDomainGVR := configv1alpha1.SchemeGroupVersion.WithResource("federationdomains")
	federationDomain := &configv1alpha1.FederationDomain{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: supervisorNamespace,
			Name:      "some-federation-domain-name",
		},
		Spec: configv1alpha1.FederationDomainSpec{
			Issuer: serviceHTTPSEndpoint(supervisorService),
		},
	}

	jwtAuthenticatorGVR := authv1alpha1.SchemeGroupVersion.WithResource("jwtauthenticators")
	jwtAuthenticator := &authv1alpha1.JWTAuthenticator{
		ObjectMeta: metav1.ObjectMeta{
			Name: "some-jwt-authenticator-name",
		},
		Spec: authv1alpha1.JWTAuthenticatorSpec{
			Issuer:   federationDomain.Spec.Issuer,
			Audience: federationDomain.Spec.Issuer,
			TLS: &authv1alpha1.TLSSpec{
				CertificateAuthorityData: base64.StdEncoding.EncodeToString(supervisorCertificateSecret.Data["ca.crt"]),
			},
		},
	}
	jwtAuthenticatorWithUUIDAudience := jwtAuthenticator.DeepCopy()
	jwtAuthenticatorWithUUIDAudience.ObjectMeta.Name = "jwt-authenticator-meow"
	jwtAuthenticatorWithUUIDAudience.Spec.Audience = "tiny angry kittens TINY ANGRY KITTENS!!!"

	tests := []struct {
		name                         string
		newKubeClient                func() *kubefake.Clientset
		newCertManagerClient         func() *certmanagerfake.Clientset
		newSupervisorClient          func() *pinnipedsupervisorfake.Clientset
		newConciergeClient           func() *pinnipedconciergefake.Clientset
		parameters                   Parameters
		wantError                    string
		wantKubeClientActions        []kubetesting.Action
		wantCertManagerClientActions []kubetesting.Action
		wantSupervisorClientActions  []kubetesting.Action
		wantConciergeClientActions   []kubetesting.Action
	}{
		{
			name: "management cluster configured from scratch",
			newKubeClient: func() *kubefake.Clientset {
				c := kubefake.NewSimpleClientset(
					supervisorService,
					supervisorCertificateSecret,
					supervisorPods[0],
					supervisorPods[1],
				)
				c.PrependReactor("delete", "secrets", func(action kubetesting.Action) (bool, runtime.Object, error) {
					// When we delete the secret in the implementation, we expect the cert-manager controller
					// to recreate it. Since there is no cert-manager controller running, let's just tell the
					// fake kube client to not delete it (i.e., that we "handled" the delete ourselves).
					return actionIsOnObject(action, supervisorCertificateSecret), nil, nil
				})
				return c
			},
			newCertManagerClient: func() *certmanagerfake.Clientset {
				return certmanagerfake.NewSimpleClientset(supervisorCertificate)
			},
			newSupervisorClient: func() *pinnipedsupervisorfake.Clientset {
				return pinnipedsupervisorfake.NewSimpleClientset()
			},
			newConciergeClient: func() *pinnipedconciergefake.Clientset {
				// The jwtauthenticator is usually deployed onto the management cluster with no spec fields
				// set; see:
				//   https://github.com/vmware-tanzu/community-edition/blob/1aa7936d88f5d9b04398d800bfe83a165619ee16/addons/packages/pinniped/0.4.4/bundle/config/overlay/pinniped-jwtauthenticator.yaml
				defaultJWTAuthenticator := jwtAuthenticator.DeepCopy()
				defaultJWTAuthenticator.Spec = authv1alpha1.JWTAuthenticatorSpec{}
				return pinnipedconciergefake.NewSimpleClientset(defaultJWTAuthenticator)
			},
			parameters: Parameters{
				ClusterType:              "management",
				ClusterName:              pinnipedInfoConfigMap.Data["cluster_name"],
				SupervisorSvcNamespace:   supervisorService.Namespace,
				SupervisorSvcName:        supervisorService.Name,
				FederationDomainName:     federationDomain.Name,
				SupervisorCertNamespace:  supervisorCertificate.Namespace,
				SupervisorCertName:       supervisorCertificate.Name,
				JWTAuthenticatorName:     jwtAuthenticator.Name,
				ConciergeIsClusterScoped: true,
			},
			wantKubeClientActions: []kubetesting.Action{
				// 1. Get the supervisor service endpoint to create the correct issuer
				kubetesting.NewGetAction(serviceGVR, supervisorService.Namespace, supervisorService.Name),
				// 4. Blow away the old supervisor certificate secret to force it to recreate
				kubetesting.NewDeleteAction(secretGVR, supervisorCertificateSecret.Namespace, supervisorCertificateSecret.Name),
				// 5. Read the new supervisor certificate secret that has the correct SAN(s) from step 3
				kubetesting.NewGetAction(secretGVR, supervisorCertificateSecret.Namespace, supervisorCertificateSecret.Name),
				// 7. Create the pinniped info configmap with the supervisor discovery information
				kubetesting.NewGetAction(configMapGVR, pinnipedInfoConfigMap.Namespace, pinnipedInfoConfigMap.Name),
				kubetesting.NewCreateAction(configMapGVR, pinnipedInfoConfigMap.Namespace, pinnipedInfoConfigMap),
				// 8. Kick the supervisor pods so that they reload new serving cert info
				kubetesting.NewListAction(podGVR, podGVK, supervisorNamespace, metav1.ListOptions{}),
				kubetesting.NewDeleteAction(podGVR, supervisorPods[0].Namespace, supervisorPods[0].Name),
				kubetesting.NewDeleteAction(podGVR, supervisorPods[1].Namespace, supervisorPods[1].Name),
			},
			wantCertManagerClientActions: []kubetesting.Action{
				// 3. We ensure the supervisor certificate has the correct SAN(s)
				kubetesting.NewGetAction(certificateGVR, supervisorCertificate.Namespace, supervisorCertificate.Name),
				kubetesting.NewGetAction(certificateGVR, supervisorCertificate.Namespace, supervisorCertificate.Name),
				kubetesting.NewUpdateAction(certificateGVR, supervisorCertificate.Namespace, supervisorCertificate),
			},
			wantSupervisorClientActions: []kubetesting.Action{
				// 2. We create the federationdomain with the correct issuer
				kubetesting.NewGetAction(federationDomainGVR, federationDomain.Namespace, federationDomain.Name),
				kubetesting.NewCreateAction(federationDomainGVR, federationDomain.Namespace, federationDomain),
			},
			wantConciergeClientActions: []kubetesting.Action{
				// 6. We update the jwtauthenticator with the correct supervisor issuer, CA data, and audience
				kubetesting.NewRootGetAction(jwtAuthenticatorGVR, jwtAuthenticator.Name),
				kubetesting.NewRootUpdateAction(jwtAuthenticatorGVR, jwtAuthenticator),
			},
		}, {
			name: "management cluster configured with explicit concierge audience should ignore the audience (not write it to the JWTAuthenticator)",
			newKubeClient: func() *kubefake.Clientset {
				c := kubefake.NewSimpleClientset(
					supervisorService,
					supervisorCertificateSecret,
					supervisorPods[0],
					supervisorPods[1],
				)
				c.PrependReactor("delete", "secrets", func(action kubetesting.Action) (bool, runtime.Object, error) {
					// When we delete the secret in the implementation, we expect the cert-manager controller
					// to recreate it. Since there is no cert-manager controller running, let's just tell the
					// fake kube client to not delete it (i.e., that we "handled" the delete ourselves).
					return actionIsOnObject(action, supervisorCertificateSecret), nil, nil
				})
				return c
			},
			newCertManagerClient: func() *certmanagerfake.Clientset {
				return certmanagerfake.NewSimpleClientset(supervisorCertificate)
			},
			newSupervisorClient: func() *pinnipedsupervisorfake.Clientset {
				return pinnipedsupervisorfake.NewSimpleClientset()
			},
			// even if a custom audience is provided (see parameters) the mgmt cluster should still return a generic authenticator
			newConciergeClient: func() *pinnipedconciergefake.Clientset {
				// The jwtauthenticator is usually deployed onto the management cluster with no spec fields
				// set; see:
				//   https://github.com/vmware-tanzu/community-edition/blob/1aa7936d88f5d9b04398d800bfe83a165619ee16/addons/packages/pinniped/0.4.4/bundle/config/overlay/pinniped-jwtauthenticator.yaml
				defaultJWTAuthenticator := jwtAuthenticator.DeepCopy()
				defaultJWTAuthenticator.Spec = authv1alpha1.JWTAuthenticatorSpec{}
				return pinnipedconciergefake.NewSimpleClientset(defaultJWTAuthenticator)
			},
			parameters: Parameters{
				ClusterType:              "management",
				ClusterName:              pinnipedInfoConfigMap.Data["cluster_name"],
				SupervisorSvcNamespace:   supervisorService.Namespace,
				SupervisorSvcName:        supervisorService.Name,
				FederationDomainName:     federationDomain.Name,
				SupervisorCertNamespace:  supervisorCertificate.Namespace,
				SupervisorCertName:       supervisorCertificate.Name,
				JWTAuthenticatorName:     jwtAuthenticator.Name,
				JWTAuthenticatorAudience: "I am a rebel and providing this even when I should not have done so.",
				ConciergeIsClusterScoped: true,
			},
			wantKubeClientActions: []kubetesting.Action{
				// 1. Get the supervisor service endpoint to create the correct issuer
				kubetesting.NewGetAction(serviceGVR, supervisorService.Namespace, supervisorService.Name),
				// 4. Blow away the old supervisor certificate secret to force it to recreate
				kubetesting.NewDeleteAction(secretGVR, supervisorCertificateSecret.Namespace, supervisorCertificateSecret.Name),
				// 5. Read the new supervisor certificate secret that has the correct SAN(s) from step 3
				kubetesting.NewGetAction(secretGVR, supervisorCertificateSecret.Namespace, supervisorCertificateSecret.Name),
				// 7. Create the pinniped info configmap with the supervisor discovery information
				kubetesting.NewGetAction(configMapGVR, pinnipedInfoConfigMap.Namespace, pinnipedInfoConfigMap.Name),
				kubetesting.NewCreateAction(configMapGVR, pinnipedInfoConfigMap.Namespace, pinnipedInfoConfigMap),
				// 8. Kick the supervisor pods so that they reload new serving cert info
				kubetesting.NewListAction(podGVR, podGVK, supervisorNamespace, metav1.ListOptions{}),
				kubetesting.NewDeleteAction(podGVR, supervisorPods[0].Namespace, supervisorPods[0].Name),
				kubetesting.NewDeleteAction(podGVR, supervisorPods[1].Namespace, supervisorPods[1].Name),
			},
			wantCertManagerClientActions: []kubetesting.Action{
				// 3. We ensure the supervisor certificate has the correct SAN(s)
				kubetesting.NewGetAction(certificateGVR, supervisorCertificate.Namespace, supervisorCertificate.Name),
				kubetesting.NewGetAction(certificateGVR, supervisorCertificate.Namespace, supervisorCertificate.Name),
				kubetesting.NewUpdateAction(certificateGVR, supervisorCertificate.Namespace, supervisorCertificate),
			},
			wantSupervisorClientActions: []kubetesting.Action{
				// 2. We create the federationdomain with the correct issuer
				kubetesting.NewGetAction(federationDomainGVR, federationDomain.Namespace, federationDomain.Name),
				kubetesting.NewCreateAction(federationDomainGVR, federationDomain.Namespace, federationDomain),
			},
			// Even though we are given parameters with a custom audience, we want to get back a generic authenticator
			// without the custom audience on the mgmt cluster.
			wantConciergeClientActions: []kubetesting.Action{
				// 6. We update the jwtauthenticator with the correct supervisor issuer, CA data, and audience
				kubetesting.NewRootGetAction(jwtAuthenticatorGVR, jwtAuthenticator.Name),
				kubetesting.NewRootUpdateAction(jwtAuthenticatorGVR, jwtAuthenticator),
			},
		},
		{
			name: "workload cluster configured from scratch",
			newKubeClient: func() *kubefake.Clientset {
				return kubefake.NewSimpleClientset()
			},
			newCertManagerClient: func() *certmanagerfake.Clientset {
				return certmanagerfake.NewSimpleClientset(supervisorCertificate)
			},
			newSupervisorClient: func() *pinnipedsupervisorfake.Clientset {
				return pinnipedsupervisorfake.NewSimpleClientset()
			},
			newConciergeClient: func() *pinnipedconciergefake.Clientset {
				// The jwtauthenticator is usually deployed onto the workload cluster with all fields set
				// correctly except for the audience; see:
				//   https://github.com/vmware-tanzu/community-edition/blob/1aa7936d88f5d9b04398d800bfe83a165619ee16/addons/packages/pinniped/0.4.4/bundle/config/overlay/pinniped-jwtauthenticator.yaml
				defaultJWTAuthenticator := jwtAuthenticator.DeepCopy()
				defaultJWTAuthenticator.Spec.Audience = "wrong-audience"
				return pinnipedconciergefake.NewSimpleClientset(defaultJWTAuthenticator)
			},
			parameters: Parameters{
				ClusterType:              "workload",
				ClusterName:              jwtAuthenticator.Spec.Audience,
				SupervisorSvcNamespace:   supervisorService.Namespace,
				SupervisorSvcEndpoint:    jwtAuthenticator.Spec.Issuer,
				SupervisorCABundleData:   jwtAuthenticator.Spec.TLS.CertificateAuthorityData,
				JWTAuthenticatorName:     jwtAuthenticator.Name,
				ConciergeIsClusterScoped: true,
			},
			wantKubeClientActions: []kubetesting.Action{
				// 2. Look for any supervisor pods to recreate (we do this on both management and workload clusters)
				kubetesting.NewListAction(podGVR, podGVK, supervisorNamespace, metav1.ListOptions{}),
			},
			wantCertManagerClientActions: []kubetesting.Action{},
			wantSupervisorClientActions:  []kubetesting.Action{},
			wantConciergeClientActions: []kubetesting.Action{
				// 1. We update the jwtauthenticator with the correct supervisor issuer, CA data, and audience
				kubetesting.NewRootGetAction(jwtAuthenticatorGVR, jwtAuthenticator.Name),
				kubetesting.NewRootUpdateAction(jwtAuthenticatorGVR, jwtAuthenticator),
			},
		},
		{
			name: "workload cluster configured from scratch with explicit concierge audience should generate a JWTAuthenticator with the customized audience value",
			newKubeClient: func() *kubefake.Clientset {
				return kubefake.NewSimpleClientset()
			},
			newCertManagerClient: func() *certmanagerfake.Clientset {
				return certmanagerfake.NewSimpleClientset(supervisorCertificate)
			},
			newSupervisorClient: func() *pinnipedsupervisorfake.Clientset {
				return pinnipedsupervisorfake.NewSimpleClientset()
			},
			// on the workload cluster, if given a custom audience, we should return a JWTAuthenticator with that audience
			newConciergeClient: func() *pinnipedconciergefake.Clientset {
				// The jwtauthenticator is usually deployed onto the workload cluster with all fields set
				// correctly except for the audience; see:
				//   https://github.com/vmware-tanzu/community-edition/blob/1aa7936d88f5d9b04398d800bfe83a165619ee16/addons/packages/pinniped/0.4.4/bundle/config/overlay/pinniped-jwtauthenticator.yaml
				jwtAuthenticatorWIthUUID := jwtAuthenticatorWithUUIDAudience.DeepCopy()
				return pinnipedconciergefake.NewSimpleClientset(jwtAuthenticatorWIthUUID)
			},
			parameters: Parameters{
				ClusterType:            "workload",
				SupervisorSvcNamespace: supervisorService.Namespace,
				SupervisorSvcEndpoint:  jwtAuthenticatorWithUUIDAudience.Spec.Issuer,
				SupervisorCABundleData: jwtAuthenticatorWithUUIDAudience.Spec.TLS.CertificateAuthorityData,
				JWTAuthenticatorName:   jwtAuthenticatorWithUUIDAudience.Name,
				// custom audience provided
				JWTAuthenticatorAudience: jwtAuthenticatorWithUUIDAudience.Spec.Audience,
				ConciergeIsClusterScoped: true,
			},
			wantKubeClientActions: []kubetesting.Action{
				// 2. Look for any supervisor pods to recreate (we do this on both management and workload clusters)
				kubetesting.NewListAction(podGVR, podGVK, supervisorNamespace, metav1.ListOptions{}),
			},
			wantCertManagerClientActions: []kubetesting.Action{},
			wantSupervisorClientActions:  []kubetesting.Action{},
			// on the workload cluster, if given a custom audience, we should return a JWTAuthenticator with that audience
			wantConciergeClientActions: []kubetesting.Action{
				// 1. We update the jwtauthenticator with the correct supervisor issuer, CA data, and audience
				kubetesting.NewRootGetAction(jwtAuthenticatorGVR, jwtAuthenticatorWithUUIDAudience.Name),
				kubetesting.NewRootUpdateAction(jwtAuthenticatorGVR, jwtAuthenticatorWithUUIDAudience),
			},
		},
		{
			name: "unknown cluster type",
			newKubeClient: func() *kubefake.Clientset {
				return kubefake.NewSimpleClientset()
			},
			newCertManagerClient: func() *certmanagerfake.Clientset {
				return certmanagerfake.NewSimpleClientset(supervisorCertificate)
			},
			newSupervisorClient: func() *pinnipedsupervisorfake.Clientset {
				return pinnipedsupervisorfake.NewSimpleClientset()
			},
			newConciergeClient: func() *pinnipedconciergefake.Clientset {
				defaultJWTAuthenticator := jwtAuthenticator.DeepCopy()
				return pinnipedconciergefake.NewSimpleClientset(defaultJWTAuthenticator)
			},
			parameters: Parameters{
				ClusterType: "penguin",
			},
			wantError:                    "unknown cluster type penguin",
			wantKubeClientActions:        []kubetesting.Action{},
			wantCertManagerClientActions: []kubetesting.Action{},
			wantSupervisorClientActions:  []kubetesting.Action{},
			wantConciergeClientActions:   []kubetesting.Action{},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			fakeKubeClient := test.newKubeClient()
			fakeCertManagerClient := test.newCertManagerClient()
			supervisorClientset := test.newSupervisorClient()
			conciergeClientset := test.newConciergeClient()
			clients := Clients{
				K8SClientset:         fakeKubeClient,
				CertmanagerClientset: fakeCertManagerClient,
				SupervisorClientset:  supervisorClientset,
				ConciergeClientset:   conciergeClientset,
			}
			inspector := inspect.Inspector{K8sClientset: fakeKubeClient, Context: context.Background()}
			err := Pinniped(context.Background(), clients, inspector, &test.parameters)
			if test.wantError != "" {
				require.EqualError(t, err, test.wantError)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, test.wantKubeClientActions, fakeKubeClient.Actions())
			require.Equal(t, test.wantCertManagerClientActions, fakeCertManagerClient.Actions())
			require.Equal(t, test.wantSupervisorClientActions, supervisorClientset.Actions())
			require.Equal(t, test.wantConciergeClientActions, conciergeClientset.Actions())
		})
	}
}

func enableLogging() {
	config := zap.NewDevelopmentConfig()
	logger, _ := config.Build()
	zap.ReplaceGlobals(logger)
}

func serviceHTTPSEndpoint(service *corev1.Service) string {
	return fmt.Sprintf("https://%s:%d", service.Status.LoadBalancer.Ingress[0].IP, service.Spec.Ports[0].Port)
}

func actionIsOnObject(action kubetesting.Action, object metav1.Object) bool {
	if action.GetNamespace() != object.GetNamespace() {
		return false
	}

	actionWithName, ok := action.(interface{ GetName() string })
	if !ok {
		return false
	}

	if actionWithName.GetName() != object.GetName() {
		return false
	}

	return true
}
