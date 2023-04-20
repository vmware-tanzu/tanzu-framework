// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package certs

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	admissionregv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Rotating certificates", func() {
	var (
		options                  Options
		rotationInterval         time.Duration
		secretKey                client.ObjectKey
		webhookConfigKey         client.ObjectKey
		totalRotationDuration    time.Duration
		waitForCertManagerToStop sync.WaitGroup
	)

	BeforeEach(func() {
		rotationInterval = time.Second * 5
		totalRotationDuration = time.Second * 20
		waitForCertManagerToStop = sync.WaitGroup{}
	})

	JustBeforeEach(func() {
		By("installing the secret", func() {
			secret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    "default",
					GenerateName: "test-",
					Annotations: map[string]string{
						"certs.tanzu.vmware.com/rotation-interval": rotationInterval.String(),
					},
				},
			}
			Expect(k8sClient.Create(ctx, secret)).To(Succeed())
			secretKey = client.ObjectKey{Namespace: secret.Namespace, Name: secret.Name}
		})

		By("installing the webhook", func() {
			failurePolicy := admissionregv1.Ignore
			sideEffectNone := admissionregv1.SideEffectClassNone
			webhookURL := "https://k8s.svc.local:9443/validate"
			webhookConfig := &admissionregv1.ValidatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Namespace:    "default",
					GenerateName: "test-",
					Labels: map[string]string{
						"certs.tanzu.vmware.com/managed-certs": "true",
					},
				},
				Webhooks: []admissionregv1.ValidatingWebhook{
					{
						AdmissionReviewVersions: []string{"v1beta1"},
						Name:                    "certs.tanzu.vmware.com",
						ClientConfig: admissionregv1.WebhookClientConfig{
							CABundle: []byte{'\n'},
							URL:      &webhookURL,
						},
						FailurePolicy: &failurePolicy,
						Rules:         []admissionregv1.RuleWithOperations{},
						SideEffects:   &sideEffectNone,
					},
				},
			}
			Expect(k8sClient.Create(ctx, webhookConfig)).To(Succeed())
			webhookConfigKey = client.ObjectKey{Namespace: webhookConfig.Namespace, Name: webhookConfig.Name}
		})

		// Define the certificate manager options.
		options = Options{
			Client:                        k8sClient,
			Logger:                        ctrl.Log.WithName("certmanager-test"),
			WebhookConfigLabel:            "certs.tanzu.vmware.com/managed-certs=true",
			NextRotationAnnotationKey:     "certs.tanzu.vmware.com/next-rotation",
			RotationCountAnnotationKey:    "certs.tanzu.vmware.com/rotation-count",
			RotationIntervalAnnotationKey: "certs.tanzu.vmware.com/rotation-interval",
			SecretNamespace:               secretKey.Namespace,
			SecretName:                    secretKey.Name,
			ServiceNamespace:              "default",
			ServiceName:                   "webhook-service",
		}

		// Start cert manager.
		waitForCertManagerToStop.Add(1)
		// Certificate Manager needs a separate context than the other k8s operations here.
		cmCtx, cancel := context.WithCancel(context.Background())
		cm, err := New(options)
		Expect(err).ShouldNot(HaveOccurred())
		go func() {
			defer GinkgoRecover()
			defer waitForCertManagerToStop.Done()
			Expect(cm.start(cmCtx)).To(Succeed())
		}()

		// Background cancel the context after some time.
		time.AfterFunc(totalRotationDuration, func() { cancel() })
	})

	JustAfterEach(func() {
		// Wait for the certificate manager to stop.
		waitForCertManagerToStop.Wait()

		// Get the secret.
		secret := &corev1.Secret{}
		Expect(k8sClient.Get(ctx, secretKey, secret)).To(Succeed())

		// ASSERT that the number of rotations is non-zero.
		// We can't really assert for the exact number of rotation in a given period as it depends on multiple things
		// such as cancellations, errors, load etc. So we just check for a non-zero count.
		Expect(secret.Annotations).ToNot(BeNil())
		Expect(secret.Annotations[options.RotationCountAnnotationKey]).ToNot(BeEmpty())
		actualRotationCount, err := strconv.Atoi(secret.Annotations[options.RotationCountAnnotationKey])
		Expect(err).ShouldNot(HaveOccurred())
		Expect(actualRotationCount).To(BeNumerically(">", 0))

		// Get the webhook config.
		webhookConfig := &admissionregv1.ValidatingWebhookConfiguration{}
		Expect(k8sClient.Get(ctx, webhookConfigKey, webhookConfig)).To(Succeed())

		// ASSERT that the webhook config was updated.
		Expect(webhookConfig.Webhooks).To(HaveLen(1))
		Expect(webhookConfig.Webhooks[0].ClientConfig.CABundle).ToNot(BeNil())
		Expect(webhookConfig.Webhooks[0].ClientConfig.CABundle).To(Equal(secret.Data[CACertName]))
	})

	AfterEach(func() {
		rotationInterval = 0
		totalRotationDuration = 0
		options = Options{}
		secretKey = client.ObjectKey{}
		webhookConfigKey = client.ObjectKey{}
		waitForCertManagerToStop = sync.WaitGroup{}
	})

	When("a certificate manager is started", func() {
		It(fmt.Sprintf("should rotate the certificates non-zero times in %s", totalRotationDuration.String()), func() {
			// ASSERT happens in JustAfterEach above.
		})
	})

	When("the webhook YAML is reapplied and a new certificate manager starts", func() {
		It("should update the webhook configurations outside the normal interval when webhook properties are reset", func() {
			// Wait for the secret's initial rotation.
			Eventually(func() (bool, error) {
				secret := &corev1.Secret{}
				if err := k8sClient.Get(ctx, secretKey, secret); err != nil {
					return false, err
				}
				if len(secret.Data[CACertName]) == 0 {
					return false, nil
				}
				webhookConfig := &admissionregv1.ValidatingWebhookConfiguration{}
				if err := k8sClient.Get(ctx, webhookConfigKey, webhookConfig); err != nil {
					return false, err
				}
				if len(webhookConfig.Webhooks) == 0 {
					return false, nil
				}
				fmt.Println("Bytes equal", bytes.Equal(webhookConfig.Webhooks[0].ClientConfig.CABundle, secret.Data[CACertName]))
				return bytes.Equal(webhookConfig.Webhooks[0].ClientConfig.CABundle, secret.Data[CACertName]), nil
			}, time.Second*15, time.Second*1).Should(BeTrue(), "webhook config CABundle should be set to CA from secret")

			// GET the webhook config.
			webhookConfig := &admissionregv1.ValidatingWebhookConfiguration{}
			Expect(k8sClient.Get(ctx, webhookConfigKey, webhookConfig)).To(Succeed())

			// UPDATE the webhook config's CABundle to a newline char '\n'.
			webhookConfig.Webhooks[0].ClientConfig.CABundle = []byte{'\n'}
			Expect(k8sClient.Update(ctx, webhookConfig)).To(Succeed())

			// ASSERT the webhook config is eventually changed back to CA bundle in the secret.
			// This check happens in JustAfterEach above.
		})
	})
})
