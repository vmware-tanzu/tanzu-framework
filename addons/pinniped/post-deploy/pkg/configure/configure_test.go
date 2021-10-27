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
	authv1alpha1 "go.pinniped.dev/generated/1.19/apis/concierge/authentication/v1alpha1"
	configv1alpha1 "go.pinniped.dev/generated/1.19/apis/supervisor/config/v1alpha1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	kubedynamicfake "k8s.io/client-go/dynamic/fake"
	kubefake "k8s.io/client-go/kubernetes/fake"
	kubetesting "k8s.io/client-go/testing"

	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/post-deploy/pkg/inspect"
	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/post-deploy/pkg/pinnipedclientset"
)

// nolint:funlen
func TestPinniped(t *testing.T) {
	const (
		apiGroupSuffix      = "tuna.io"
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
			"cluster_name":                         "some-pinniped-info-management-cluster-name",
			"issuer":                               serviceHTTPSEndpoint(supervisorService),
			"issuer_ca_bundle_data":                base64.StdEncoding.EncodeToString(supervisorCertificateSecret.Data["ca.crt"]),
			"pinniped_api_group_suffix":            apiGroupSuffix,
			"pinniped_concierge_is_cluster_scoped": "false",
		},
	}

	pinnipedInfoConfigMapWorkloadCluster := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "kube-public",
			Name:      "pinniped-info",
		},
		Data: map[string]string{
			"pinniped_api_group_suffix":            apiGroupSuffix,
			"pinniped_concierge_is_cluster_scoped": "false",
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

	authv1alpha1GV := authv1alpha1.SchemeGroupVersion
	authv1alpha1GV.Group = "authentication.concierge." + apiGroupSuffix

	configv1alpha1GV := configv1alpha1.SchemeGroupVersion
	configv1alpha1GV.Group = "config.supervisor." + apiGroupSuffix

	scheme := runtime.NewScheme()
	scheme.AddKnownTypes(configv1alpha1GV, &configv1alpha1.FederationDomain{}, &configv1alpha1.FederationDomainList{})
	scheme.AddKnownTypes(authv1alpha1GV, &authv1alpha1.JWTAuthenticator{}, &authv1alpha1.JWTAuthenticatorList{})

	federationDomainGVR := configv1alpha1GV.WithResource("federationdomains")
	federationDomain := &configv1alpha1.FederationDomain{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: supervisorNamespace,
			Name:      "some-federation-domain-name",
		},
		Spec: configv1alpha1.FederationDomainSpec{
			Issuer: serviceHTTPSEndpoint(supervisorService),
		},
	}
	federationDomain.APIVersion, federationDomain.Kind = configv1alpha1GV.WithKind("FederationDomain").ToAPIVersionAndKind()

	jwtAuthenticatorGVR := authv1alpha1GV.WithResource("jwtauthenticators")
	jwtAuthenticator := &authv1alpha1.JWTAuthenticator{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "pinniped-concierge", // vars.ConciergeNamespace default
			Name:      "some-jwt-authenticator-name",
		},
		Spec: authv1alpha1.JWTAuthenticatorSpec{
			Issuer:   federationDomain.Spec.Issuer,
			Audience: federationDomain.Spec.Issuer,
			TLS: &authv1alpha1.TLSSpec{
				CertificateAuthorityData: base64.StdEncoding.EncodeToString(supervisorCertificateSecret.Data["ca.crt"]),
			},
		},
	}
	jwtAuthenticator.APIVersion, jwtAuthenticator.Kind = authv1alpha1GV.WithKind("JWTAuthenticator").ToAPIVersionAndKind()

	tests := []struct {
		name                         string
		newKubeClient                func() *kubefake.Clientset
		newCertManagerClient         func() *certmanagerfake.Clientset
		newKubeDynamicClient         func() *kubedynamicfake.FakeDynamicClient
		parameters                   Parameters
		wantError                    string
		wantKubeClientActions        []kubetesting.Action
		wantCertManagerClientActions []kubetesting.Action
		wantKubeDynamicClientActions []kubetesting.Action
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
			newKubeDynamicClient: func() *kubedynamicfake.FakeDynamicClient {
				// The jwtauthenticator is usually deployed onto the management cluster with no spec fields
				// set; see:
				//   https://github.com/vmware-tanzu/community-edition/blob/1aa7936d88f5d9b04398d800bfe83a165619ee16/addons/packages/pinniped/0.4.4/bundle/config/overlay/pinniped-jwtauthenticator.yaml
				defaultJWTAuthenticator := jwtAuthenticator.DeepCopy()
				defaultJWTAuthenticator.Spec = authv1alpha1.JWTAuthenticatorSpec{}
				return kubedynamicfake.NewSimpleDynamicClient(scheme, defaultJWTAuthenticator)
			},
			parameters: Parameters{
				ClusterType:             "management",
				ClusterName:             pinnipedInfoConfigMap.Data["cluster_name"],
				SupervisorSvcNamespace:  supervisorService.Namespace,
				SupervisorSvcName:       supervisorService.Name,
				FederationDomainName:    federationDomain.Name,
				SupervisorCertNamespace: supervisorCertificate.Namespace,
				SupervisorCertName:      supervisorCertificate.Name,
				JWTAuthenticatorName:    jwtAuthenticator.Name,
				PinnipedAPIGroupSuffix:  apiGroupSuffix,
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
			wantKubeDynamicClientActions: []kubetesting.Action{
				// 2. We create the federationdomain with the correct issuer
				kubetesting.NewGetAction(federationDomainGVR, federationDomain.Namespace, federationDomain.Name),
				kubetesting.NewCreateAction(federationDomainGVR, federationDomain.Namespace, toUnstructured(federationDomain, true)),
				// 6. We update the jwtauthenticator with the correct supervisor issuer, CA data, and audience
				kubetesting.NewGetAction(jwtAuthenticatorGVR, jwtAuthenticator.Namespace, jwtAuthenticator.Name),
				kubetesting.NewUpdateAction(jwtAuthenticatorGVR, jwtAuthenticator.Namespace, toUnstructured(jwtAuthenticator, false)),
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
			newKubeDynamicClient: func() *kubedynamicfake.FakeDynamicClient {
				// The jwtauthenticator is usually deployed onto the workload cluster with all fields set
				// correctly except for the audience; see:
				//   https://github.com/vmware-tanzu/community-edition/blob/1aa7936d88f5d9b04398d800bfe83a165619ee16/addons/packages/pinniped/0.4.4/bundle/config/overlay/pinniped-jwtauthenticator.yaml
				defaultJWTAuthenticator := jwtAuthenticator.DeepCopy()
				defaultJWTAuthenticator.Spec.Audience = "wrong-audience"
				return kubedynamicfake.NewSimpleDynamicClient(scheme, defaultJWTAuthenticator)
			},
			parameters: Parameters{
				ClusterType:            "workload",
				ClusterName:            jwtAuthenticator.Spec.Audience,
				SupervisorSvcNamespace: supervisorService.Namespace,
				SupervisorSvcEndpoint:  jwtAuthenticator.Spec.Issuer,
				SupervisorCABundleData: jwtAuthenticator.Spec.TLS.CertificateAuthorityData,
				JWTAuthenticatorName:   jwtAuthenticator.Name,
				PinnipedAPIGroupSuffix: apiGroupSuffix,
			},
			wantKubeClientActions: []kubetesting.Action{
				// 2. Create the Pinniped info configmap
				kubetesting.NewGetAction(configMapGVR, pinnipedInfoConfigMapWorkloadCluster.Namespace, pinnipedInfoConfigMapWorkloadCluster.Name),
				kubetesting.NewCreateAction(configMapGVR, pinnipedInfoConfigMapWorkloadCluster.Namespace, pinnipedInfoConfigMapWorkloadCluster),
				// 3. Look for any supervisor pods to recreate (we do this on both management and workload clusters)
				kubetesting.NewListAction(podGVR, podGVK, supervisorNamespace, metav1.ListOptions{}),
			},
			wantCertManagerClientActions: []kubetesting.Action{},
			wantKubeDynamicClientActions: []kubetesting.Action{
				// 1. We update the jwtauthenticator with the correct supervisor issuer, CA data, and audience
				kubetesting.NewGetAction(jwtAuthenticatorGVR, jwtAuthenticator.Namespace, jwtAuthenticator.Name),
				kubetesting.NewUpdateAction(jwtAuthenticatorGVR, jwtAuthenticator.Namespace, toUnstructured(jwtAuthenticator, false)),
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
			newKubeDynamicClient: func() *kubedynamicfake.FakeDynamicClient {
				defaultJWTAuthenticator := jwtAuthenticator.DeepCopy()
				return kubedynamicfake.NewSimpleDynamicClient(scheme, defaultJWTAuthenticator)
			},
			parameters: Parameters{
				ClusterType: "penguin",
			},
			wantError:                    "unknown cluster type penguin",
			wantKubeClientActions:        []kubetesting.Action{},
			wantCertManagerClientActions: []kubetesting.Action{},
			wantKubeDynamicClientActions: []kubetesting.Action{},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			fakeKubeClient := test.newKubeClient()
			fakeCertManagerClient := test.newCertManagerClient()
			fakeKubeDynamicClient := test.newKubeDynamicClient()
			clients := Clients{
				K8SClientset:         fakeKubeClient,
				CertmanagerClientset: fakeCertManagerClient,
				SupervisorClientset:  pinnipedclientset.NewSupervisor(fakeKubeDynamicClient, apiGroupSuffix),
				ConciergeClientset:   pinnipedclientset.NewConcierge(fakeKubeDynamicClient, apiGroupSuffix, false),
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
			require.Equal(t, test.wantKubeDynamicClientActions, fakeKubeDynamicClient.Actions())
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
