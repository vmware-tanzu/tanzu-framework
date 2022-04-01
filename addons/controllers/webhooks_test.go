// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	cert2 "k8s.io/client-go/util/cert"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/webhooks"
)

const (
	oneDay   = time.Hour * 24
	oneWeek  = time.Hour * 24 * 7
	zeroTime = time.Second * 0
)

var setupLog = ctrl.Log.WithName("controllers").WithName("Addon")

var _ = Describe("when webhook TLS is being continuously managed", func() {

	Context("if check frequency time is reached", func() {
		It("tls certificates should be validated and rotated if invalid", func() {
			privCtx := context.Background()
			cacelCtx, cancelFn := context.WithCancel(privCtx)
			rotationTime := time.Second
			managementFrequency := rotationTime / 2
			webhookTLS := webhooks.WebhookTLS{
				Ctx:           cacelCtx,
				K8sConfig:     k8sConfig,
				CertPath:      certPath,
				KeyPath:       keyPath,
				Name:          webhookScrtName,
				ServiceName:   webhookServiceName,
				LabelSelector: constants.AddonWebhookLabelKey,
				Logger:        setupLog,
				Namespace:     addonNamespace,
				RotationTime:  rotationTime,
			}

			err := webhookTLS.ManageCertificates(managementFrequency)
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(managementFrequency + rotationTime) // we wait longer than one rotation to allow for certificate installation
			firstCert, err := cert2.CertsFromFile(webhookTLS.CertPath)
			Expect(err).ToNot(HaveOccurred())
			firstCertPEM, err := cert2.EncodeCertificates(firstCert[0])
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(2 * rotationTime) // wait enough time to ensure the certificate has been rotated.

			secondCert, err := cert2.CertsFromFile(webhookTLS.CertPath)
			Expect(err).ToNot(HaveOccurred())
			secondCertPEM, err := cert2.EncodeCertificates(secondCert[0])
			Expect(err).ToNot(HaveOccurred())

			cancelFn()

			// The certificates should be different due to rotation
			Expect(secondCertPEM).ToNot(Equal(firstCertPEM))

		})
	})
	Context("if rotation time is longer than one week", func() {
		It("should set rotation to one week", func() {
			rotationTime := oneWeek + oneDay
			managementFrequency := rotationTime / 2
			webhookTLS := webhooks.WebhookTLS{
				Ctx:           context.Background(),
				K8sConfig:     k8sConfig,
				CertPath:      certPath,
				KeyPath:       keyPath,
				Name:          webhookScrtName,
				ServiceName:   webhookServiceName,
				LabelSelector: constants.AddonWebhookLabelKey,
				Logger:        setupLog,
				Namespace:     addonNamespace,
				RotationTime:  rotationTime,
			}
			err := webhookTLS.ManageCertificates(managementFrequency)
			Expect(err).ToNot(HaveOccurred())
			Expect(webhookTLS.RotationTime).To(Equal(oneWeek))

		})
	})
	Context("if rotation time is 0 or less than 0", func() {
		It("should set rotation to one day", func() {
			rotationTime := zeroTime
			managementFrequency := oneDay // any value larger than 0 will work here
			webhookTLS := webhooks.WebhookTLS{
				Ctx:           context.Background(),
				K8sConfig:     k8sConfig,
				CertPath:      certPath,
				KeyPath:       keyPath,
				Name:          webhookScrtName,
				ServiceName:   webhookServiceName,
				LabelSelector: constants.AddonWebhookLabelKey,
				Logger:        setupLog,
				Namespace:     addonNamespace,
				RotationTime:  rotationTime,
			}
			err := webhookTLS.ManageCertificates(managementFrequency)
			Expect(err).ToNot(HaveOccurred())
			Expect(webhookTLS.RotationTime).To(Equal(oneDay))

		})
	})
	Context("if rotation time is one week", func() {
		It("tls certificates should be validated and not rotated until one week", func() {
			privCtx := context.Background()
			cacelCtx, cancelFn := context.WithCancel(privCtx)
			rotationTime := oneWeek
			managementFrequency := time.Second / 2 // manage the certificates every 1/2 second for testing purposes
			webhookTLS := webhooks.WebhookTLS{
				Ctx:           cacelCtx,
				K8sConfig:     k8sConfig,
				CertPath:      certPath,
				KeyPath:       keyPath,
				Name:          webhookScrtName,
				ServiceName:   webhookServiceName,
				LabelSelector: constants.AddonWebhookLabelKey,
				Logger:        setupLog,
				Namespace:     addonNamespace,
				RotationTime:  rotationTime,
			}

			err := webhookTLS.ManageCertificates(managementFrequency)
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(managementFrequency * 3) // manage the certificates three times
			firstCert, err := cert2.CertsFromFile(webhookTLS.CertPath)
			Expect(err).ToNot(HaveOccurred())
			firstCertPEM, err := cert2.EncodeCertificates(firstCert[0])
			Expect(err).ToNot(HaveOccurred())

			time.Sleep(managementFrequency * 3) // manage the certificate three times
			secondCert, err := cert2.CertsFromFile(webhookTLS.CertPath)
			Expect(err).ToNot(HaveOccurred())
			secondCertPEM, err := cert2.EncodeCertificates(secondCert[0])
			Expect(err).ToNot(HaveOccurred())

			cancelFn()

			// The certificates should be the same since we did not hit one week rotation
			Expect(secondCertPEM).To(Equal(firstCertPEM))

		})
	})

})
