package shared

import (
	"context"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	admissionregistrationv1 "k8s.io/client-go/kubernetes/typed/admissionregistration/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	webhookScrtName   = "webhook-tls"
	systemNamespace   = "tkg-system"
	webhookLabelKey   = "tkg.tanzu.vmware.com/addon-webhooks"
	webhookLabelValue = "addon-webhooks"
)

// getWebhookCACert returns the value of CA certificate "ca-cert.pem" for the "webhook-tls" secret
func getWebhookCACert(ctx context.Context, k8sClient client.Client, secretName, secretNamespace string) (string, error) {
	secret := corev1.Secret{}
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: secretName, Namespace: secretNamespace}, &secret); err != nil {
		return "", err
	}
	if value, exists := secret.Data["ca-cert.pem"]; exists {
		return string(value), nil
	}
	return "", errors.New("\"ca-cert.pem\" should exist and have non empty value")
}

// verifyCABundleInLabeledWebhooks verifies all validating/mutating webhooks which match the provided label selector have a caBundle field matching the value of the provided CA certificate
func verifyCABundleInLabeledWebhooks(ctx context.Context, admissionRegistrationClient admissionregistrationv1.AdmissionregistrationV1Interface, labelSelector, caCert string) (bool, error) {
	validatingWhConfigs, err := admissionRegistrationClient.ValidatingWebhookConfigurations().List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return false, err
	}
	for idx := range validatingWhConfigs.Items {
		configName := validatingWhConfigs.Items[idx].Name
		validatingWhConfig, err := admissionRegistrationClient.ValidatingWebhookConfigurations().Get(ctx, configName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		for idx := range validatingWhConfig.Webhooks {
			if string(validatingWhConfig.Webhooks[idx].ClientConfig.CABundle) != caCert {
				return false, nil
			}
		}
	}

	mutatingWhConfigs, err := admissionRegistrationClient.MutatingWebhookConfigurations().List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return false, err
	}
	for idx := range mutatingWhConfigs.Items {
		configName := mutatingWhConfigs.Items[idx].Name
		mutatingWhConfig, err := admissionRegistrationClient.MutatingWebhookConfigurations().Get(ctx, configName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		for idx := range mutatingWhConfig.Webhooks {
			if string(mutatingWhConfig.Webhooks[idx].ClientConfig.CABundle) != caCert {
				return false, nil
			}
		}
	}

	return true, nil
}
