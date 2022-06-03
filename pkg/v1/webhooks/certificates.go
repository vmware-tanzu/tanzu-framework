// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
	"k8s.io/client-go/util/retry"
	"knative.dev/pkg/webhook/certificates/resources"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func addCertsToWebhookConfigs(ctx context.Context, k8sclient kubernetes.Interface, labelSelector string, secret *corev1.Secret) error {
	allValidatingWebhookConfigurations, err := k8sclient.AdmissionregistrationV1().ValidatingWebhookConfigurations().List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return err
	}
	for idx := range allValidatingWebhookConfigurations.Items {
		configName := allValidatingWebhookConfigurations.Items[idx].Name
		validatingWebhookConfiguration, err := k8sclient.AdmissionregistrationV1().ValidatingWebhookConfigurations().Get(ctx, configName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		for idx := range validatingWebhookConfiguration.Webhooks {
			validatingWebhookConfiguration.Webhooks[idx].ClientConfig.CABundle = secret.Data[resources.CACert]
		}
		if _, err := k8sclient.AdmissionregistrationV1().ValidatingWebhookConfigurations().Update(ctx, validatingWebhookConfiguration, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("error updating CA cert of ValidatingWebhookConfiguration %s: %v", configName, err)
		}
	}

	allMutatingWebhookConfigurations, err := k8sclient.AdmissionregistrationV1().MutatingWebhookConfigurations().List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return err
	}
	for idx := range allMutatingWebhookConfigurations.Items {
		configName := allMutatingWebhookConfigurations.Items[idx].Name
		mutatingWebhookConfiguration, err := k8sclient.AdmissionregistrationV1().MutatingWebhookConfigurations().Get(ctx, configName, metav1.GetOptions{})
		if err != nil {
			return err
		}

		for idx := range mutatingWebhookConfiguration.Webhooks {
			mutatingWebhookConfiguration.Webhooks[idx].ClientConfig.CABundle = secret.Data[resources.CACert]
		}
		if _, err := k8sclient.AdmissionregistrationV1().MutatingWebhookConfigurations().Update(ctx, mutatingWebhookConfiguration, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("error updating CA cert of MutatingWebhookConfiguration %s: %v", configName, err)
		}
	}
	return nil
}

// WriteServerTLSToFileSystem writes servers certificate and key in provided secret to the filesystem paths provided.
func WriteServerTLSToFileSystem(ctx context.Context, certPath, keyPath string, secret *corev1.Secret) error {
	if err := certutil.WriteCert(certPath, secret.Data[resources.ServerCert]); err != nil {
		return err
	}
	if err := keyutil.WriteKey(keyPath, secret.Data[resources.ServerKey]); err != nil {
		return err
	}
	return nil
}

// InstallNewCertificates creates a new set of keys and certificates and saves them to the filesystem paths provided.
// Adds the CA certificate to webhook configurations matching label selector.
// Returns a secret containing the server key, sever certificate and CA certificate.
func InstallNewCertificates(ctx context.Context, k8sConfig *rest.Config, certPath, keyPath, secretName, namespace, serviceName, labelSelector string) (*corev1.Secret, error) {
	if labelSelector == "" {
		return nil, fmt.Errorf("label selector not provided for webhook configurations udpate")
	}
	secret, err := resources.MakeSecret(ctx, secretName, namespace, serviceName)
	if err != nil {
		return nil, err
	}

	err = InstallCertificates(ctx, k8sConfig, secret, certPath, keyPath, labelSelector)
	if err != nil {
		return nil, err
	}
	return secret, nil
}

// InstallCertificates saves server certificate and key in provided secret to the filesystem paths provided.
// Adds the CA certificate to webhook configuration matching label selector.
func InstallCertificates(ctx context.Context, k8sConfig *rest.Config, secret *corev1.Secret, certPath, keyPath, labelSelector string) error {
	if err := WriteServerTLSToFileSystem(ctx, certPath, keyPath, secret); err != nil {
		return err
	}
	k8sclient, err := client.New(k8sConfig, client.Options{})
	if err != nil {
		return err
	}
	err = updateOrCreateTLSSecret(ctx, k8sclient, secret)
	if err != nil {
		return err
	}

	clientSet, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return err
	}
	err = addCertsToWebhookConfigs(ctx, clientSet, labelSelector, secret)
	if err != nil {
		return err
	}

	return nil
}

// ValidateTLSSecret checks secret has all required keys and certificates.
// Checks certificate lifetime is valid.
func ValidateTLSSecret(tlsSecret *corev1.Secret, certGracePeriod time.Duration) error {
	if tlsSecret == nil {
		return fmt.Errorf("webhook certificate secret is empty")
	}
	if _, haskey := tlsSecret.Data[resources.ServerKey]; !haskey {
		return fmt.Errorf("webhook certificate secret is missing key %q", resources.ServerKey)
	}
	if _, haskey := tlsSecret.Data[resources.ServerCert]; !haskey {
		return fmt.Errorf("webhook certificate secret is missing key %q", resources.ServerCert)
	}
	if _, haskey := tlsSecret.Data[resources.CACert]; !haskey {
		return fmt.Errorf("webhook certificate secret is missing key %q", resources.CACert)
	}

	cert, err := tls.X509KeyPair(tlsSecret.Data[resources.ServerCert], tlsSecret.Data[resources.ServerKey])
	if err != nil {
		return err
	}
	certData, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return err
	}
	if time.Now().Add(certGracePeriod).After(certData.NotAfter) {
		return fmt.Errorf("webhook certificate expired on %q", certData.NotAfter)
	}
	return nil
}

func updateOrCreateTLSSecret(ctx context.Context, k8sclient client.Client, tlsSecret *corev1.Secret) error {
	if tlsSecret == nil {
		return nil
	}
	currentSecret := &corev1.Secret{}
	err := k8sclient.Get(ctx, client.ObjectKey{
		Namespace: tlsSecret.Namespace,
		Name:      tlsSecret.Name}, currentSecret)
	if apierrors.IsNotFound(err) {
		err = k8sclient.Create(ctx, tlsSecret)
		if err != nil {
			return err
		}
		return nil
	}
	if err != nil {
		return err
	}

	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		return k8sclient.Update(ctx, tlsSecret)
	})
	if err != nil {
		return err
	}
	return nil
}
