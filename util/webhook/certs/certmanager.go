// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package certs

import (
	"context"
	"fmt"
	"os"
	"path"
	"strconv"
	"time"

	adminregv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	"knative.dev/pkg/webhook/certificates/resources"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// New returns a new instance of a CertificateManager.
func New(options Options) (*CertificateManager, error) {
	if err := options.defaultOpts(); err != nil {
		return nil, err
	}
	return &CertificateManager{opts: options}, nil
}

// CertificateManager creates and rotates certificates required by a controller manager's webhook server.
type CertificateManager struct {
	opts Options
}

// Start starts certificate management for a controller manager's webhooks.
// This method calls os.Exit when an error is encountered.
func (cm *CertificateManager) Start(ctx context.Context) error {
	go func() {
		if err := cm.start(ctx); err != nil {
			cm.opts.Logger.Error(err, "unexpected failure in starting certificate manager")
			os.Exit(1)
		}
	}()
	return nil
}

// start does the actual work of cert rotation.
func (cm *CertificateManager) start(ctx context.Context) error {
	var rotationTime time.Time
	for {
		nextRotationTime, err := cm.rotateCerts(ctx, rotationTime)
		if err != nil {
			return err
		}
		rotationTime = nextRotationTime
		select {
		case <-time.After(time.Until(nextRotationTime)):
		case <-ctx.Done():
			return nil
		}
	}
}

// rotateCerts rotates certificates at the scheduled rotation time, writes them to the secret and returns the next
// scheduled rotation time.
func (cm *CertificateManager) rotateCerts(ctx context.Context, scheduledNextRotationTime time.Time) (time.Time, error) {
	now := time.Now()
	cm.opts.Logger.Info("Rotating certificates", "now", now.String(), "scheduledNextRotationTime", scheduledNextRotationTime.String())

	// Get the webhook server's secret. Certificate manager expects an empty secret to have been created with the
	// controller deployment that intends to use cert management.
	// Any error in getting that secret is cause for exiting this program.
	secret := &corev1.Secret{}
	secretKey := client.ObjectKey{Name: cm.opts.SecretName, Namespace: cm.opts.SecretNamespace}
	cm.opts.Logger.Info("Getting secret", "namespacedName", secretKey.String())
	if err := cm.opts.Client.Get(ctx, secretKey, secret); err != nil {
		return time.Time{}, err
	}

	if len(secret.Annotations) == 0 {
		secret.Annotations = map[string]string{}
	}

	// If this is the first time rotate is called since starting certificate manager, then initialize the webhooks with
	// any CA data (if any) in the secret. This is for the case when a package is redeployed (for any reason) and the
	// webhook config properties are reset.
	if scheduledNextRotationTime.IsZero() {
		if caCertData := secret.Data[CACertName]; len(caCertData) > 0 {
			if err := cm.updateWebhookConfigs(ctx, caCertData); err != nil {
				return time.Time{}, err
			}
		}
	}

	// If the webhook secret already has a scheduled next rotation time that does not occur in the past, then defer this
	// rotation until that time. This usually happens when another certificate manager has updated the certs (perhaps
	// from an earlier process that was exited).
	nextRotationTime, _ := cm.getNextRotationTime(secret)
	if !nextRotationTime.IsZero() && !nextRotationTime.Before(now) {
		if !scheduledNextRotationTime.Equal(nextRotationTime) {
			cm.opts.Logger.Info("Rescheduling next rotation", "nextRotationTime", nextRotationTime.String())
			return nextRotationTime, nil
		}
	}

	// Determine the rotation interval.
	var rotationInterval time.Duration
	value := secret.Annotations[cm.opts.RotationIntervalAnnotationKey]
	// If key doesn't exist or the value is "", time.ParseDuration will return a zero time.
	if rotationInterval, _ = time.ParseDuration(value); rotationInterval == 0 {
		rotationInterval = defaultRotationInterval
	}

	// Update next rotation time and write it to the NextRotationAnnotationKey in the secret.
	nextRotationTime = now.Add(rotationInterval)
	secret.Annotations[cm.opts.NextRotationAnnotationKey] = strconv.FormatInt(nextRotationTime.Unix(), 10)

	// Increment the rotation count. If an error occurs just ignore it since
	// the rotation count is optional, primarily used for testing.
	var rotationCount int64
	if value := secret.Annotations[cm.opts.RotationCountAnnotationKey]; value != "" {
		rotationCount, _ = strconv.ParseInt(value, 10, 64)
	}
	rotationCount++
	secret.Annotations[cm.opts.RotationCountAnnotationKey] = strconv.FormatInt(rotationCount, 10)

	// Update the secret data.
	notAfter := nextRotationTime.Add(certExpirtationBuffer)
	cm.opts.Logger.Info("Generating certificates")
	secretData, err := cm.generateWebhookCertSecretData(ctx, notAfter)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to generate webhook server certificate secret data: %w", err)
	}
	secret.Data = secretData
	if err := cm.updateWebhookSecret(ctx, secret); err != nil {
		return time.Time{}, err
	}

	// Update webhook configurations with new cert data.
	if err := cm.updateWebhookConfigs(ctx, secretData[CACertName]); err != nil {
		return time.Time{}, err
	}

	cm.opts.Logger.Info("Rotated certificates", "nextRotation", nextRotationTime.String(), "totalRotationCount", rotationCount)
	return nextRotationTime, nil
}

// getNextRotationTime fetches the next rotation time from the webhook secret's annotation.
func (cm *CertificateManager) getNextRotationTime(secret *corev1.Secret) (time.Time, error) {
	// Get the next-rotation annotation
	szValue := secret.Annotations[cm.opts.NextRotationAnnotationKey]
	if szValue == "" {
		return time.Time{}, fmt.Errorf("missing next-rotation annotation")
	}

	// The error is ignored since the value will always be >0.
	iValue, _ := strconv.ParseInt(szValue, 10, 64)
	if iValue <= 0 {
		return time.Time{}, fmt.Errorf("next-rotation annotation value is invalid")
	}

	// Return the next-rotation value.
	return time.Unix(iValue, 0), nil
}

// updateWebhookConfigs updates the caBundle in the webhooks selected by the label selector.
func (cm *CertificateManager) updateWebhookConfigs(ctx context.Context, caCertData []byte) error {
	cm.opts.Logger.Info("Updating webhook configs with CA bundle")

	// Update validating webhooks with new CA data.
	labelSelector, err := metav1.ParseToLabelSelector(cm.opts.WebhookConfigLabel)
	if err != nil {
		return fmt.Errorf("failed to parse webhook label: %w", err)
	}
	matchLabels := client.MatchingLabels(labelSelector.MatchLabels)
	validatingWebhookList := &adminregv1.ValidatingWebhookConfigurationList{}
	if err := cm.opts.Client.List(ctx, validatingWebhookList, matchLabels); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to list validating webhooks: %w", err)
	}

	for _, webhookConfig := range validatingWebhookList.Items {
		for i := range webhookConfig.Webhooks {
			webhookConfig.Webhooks[i].ClientConfig.CABundle = caCertData
		}

		if err := cm.opts.Client.Update(ctx, &webhookConfig); err != nil {
			if !apierrors.IsConflict(err) {
				return fmt.Errorf("failed to update validating webhook configuration %s/%s: %w", webhookConfig.Namespace, webhookConfig.Name, err)
			}
		}
	}

	// Update mutating webhooks with new CA data.
	mutatingWebhookList := &adminregv1.MutatingWebhookConfigurationList{}
	if err := cm.opts.Client.List(ctx, mutatingWebhookList, matchLabels); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to list mutating webhooks: %w", err)
	}

	for _, webhookConfig := range mutatingWebhookList.Items {
		for i := range webhookConfig.Webhooks {
			webhookConfig.Webhooks[i].ClientConfig.CABundle = caCertData
		}

		if err := cm.opts.Client.Update(ctx, &webhookConfig); err != nil {
			if !apierrors.IsConflict(err) {
				return fmt.Errorf("failed to update mutating webhook configuration %s/%s: %w", webhookConfig.Namespace, webhookConfig.Name, err)
			}
		}
	}

	return nil
}

// generateWebhookCertSecretData generates the cert data to be written to the secret.
func (cm *CertificateManager) generateWebhookCertSecretData(ctx context.Context, notAfter time.Time) (map[string][]byte, error) {
	serverKey, serverCert, caCert, err := resources.CreateCerts(ctx, cm.opts.ServiceName, cm.opts.ServiceNamespace, notAfter)
	if err != nil {
		return nil, fmt.Errorf("failed to create certs: %w", err)
	}
	data := map[string][]byte{
		CACertName:     caCert,
		ServerCertName: serverCert,
		ServerKeyName:  serverKey,
	}
	return data, nil
}

// updateWebhookSecret updates the webhook secret with the cert data.
func (cm *CertificateManager) updateWebhookSecret(ctx context.Context, secret *corev1.Secret) error {
	namespacedName := fmt.Sprintf("%s/%s", secret.Namespace, secret.Name)
	cm.opts.Logger.Info("Updating secret with certificate data", "namespacedName", namespacedName)
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		return cm.opts.Client.Update(ctx, secret)
	})
	if err != nil {
		return fmt.Errorf("failed to update webhook secret %s: %w", namespacedName, err)
	}
	return nil
}

// WaitForCertDirReady blocks until certs are written to the cert directory or until a timeout occurs.
func (cm *CertificateManager) WaitForCertDirReady() error {
	timeout := time.Minute * 5
	select {
	case <-cm.certDirReady():
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("timed out after %s", timeout.String())
	}
}

// certDirReady returns a channel that is closed when certs are found in the directory.
func (cm *CertificateManager) certDirReady() <-chan struct{} {
	done := make(chan struct{})
	go func() {
		crtPath := path.Join(cm.opts.CertDir, ServerCertName)
		keyPath := path.Join(cm.opts.CertDir, ServerKeyName)
		for {
			certPathFound, keyPathFound := false, false
			if file, err := os.Stat(crtPath); err == nil && file.Size() > 0 {
				certPathFound = true
			}
			if file, err := os.Stat(keyPath); err == nil && file.Size() > 0 {
				keyPathFound = true
			}
			if certPathFound && keyPathFound {
				close(done)
				return
			}
			time.Sleep(time.Second * 1)
		}
	}()
	return done
}
