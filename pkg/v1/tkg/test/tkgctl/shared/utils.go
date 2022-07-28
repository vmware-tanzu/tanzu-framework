package shared

import (
	"context"

	v1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	admissionregistrationv1 "k8s.io/client-go/kubernetes/typed/admissionregistration/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetWebhookCACert(ctx context.Context, k8sClient client.Client, secretName, secretNamespace string) (string, error) {
	secret := corev1.Secret{}
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: secretName, Namespace: secretNamespace}, &secret); err != nil {
		return "", err
	}
	if value, exists := secret.Data["ca-cert.pem"]; exists {
		return string(value), nil
	}
	return "", nil
}

func GetLabeledWebhooks(ctx context.Context, admissionRegistrationClient admissionregistrationv1.AdmissionregistrationV1Interface, labelSelector string) (map[string][]v1.ValidatingWebhook, map[string][]v1.MutatingWebhook, error) {
	validatingWebhooks := map[string][]v1.ValidatingWebhook{}

	validatingWhConfigs, err := admissionRegistrationClient.ValidatingWebhookConfigurations().List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return nil, nil, err
	}
	for idx := range validatingWhConfigs.Items {
		configName := validatingWhConfigs.Items[idx].Name
		validatingWhConfig, err := admissionRegistrationClient.ValidatingWebhookConfigurations().Get(ctx, configName, metav1.GetOptions{})
		if err != nil {
			return nil, nil, err
		}
		validatingWebhooks[validatingWhConfig.Name] = validatingWhConfig.Webhooks
	}

	mutatingWebhooks := map[string][]v1.MutatingWebhook{}
	mutatingWhConfigs, err := admissionRegistrationClient.MutatingWebhookConfigurations().List(ctx, metav1.ListOptions{LabelSelector: labelSelector})
	if err != nil {
		return nil, nil, err
	}
	for idx := range mutatingWhConfigs.Items {
		configName := mutatingWhConfigs.Items[idx].Name
		mutatingWhConfig, err := admissionRegistrationClient.MutatingWebhookConfigurations().Get(ctx, configName, metav1.GetOptions{})
		if err != nil {
			return nil, nil, err
		}
		mutatingWebhooks[mutatingWhConfig.Name] = mutatingWhConfig.Webhooks
	}

	return validatingWebhooks, mutatingWebhooks, nil
}

func VerifyCABundleInWebhooks(validatingWebhooks map[string][]v1.ValidatingWebhook, mutatingWebhooks map[string][]v1.MutatingWebhook, caCert string) (bool, error) {
	for _, webhooks := range validatingWebhooks {
		for idx := range webhooks {
			if string(webhooks[idx].ClientConfig.CABundle) != caCert {
				return false, nil
			}
		}
	}

	for _, webhooks := range mutatingWebhooks {
		for idx := range webhooks {
			if string(webhooks[idx].ClientConfig.CABundle) != caCert {
				return false, nil
			}
		}
	}

	return true, nil
}

func VerifyCABundleInLabeledWebhooks(ctx context.Context, admissionRegistrationClient admissionregistrationv1.AdmissionregistrationV1Interface, labelSelector, caCert string) (bool, error) {
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
