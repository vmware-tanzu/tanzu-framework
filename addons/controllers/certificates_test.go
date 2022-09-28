// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"os"

	cert2 "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
	"knative.dev/pkg/webhook/certificates/resources"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/webhooks"
	"github.com/vmware-tanzu/tanzu-framework/addons/test/testutil"
)

var _ = Describe("Webhook", func() {
	Context("Webhook Manifests", func() {
		It("Should create webhook manifests for tests", func() {
			f, err := os.Open("testdata/test-webhook-manifests.yaml")
			Expect(err).ToNot(HaveOccurred())
			defer f.Close()
			err = testutil.CreateResources(f, cfg, dynamicClient)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("server's certificate and key", func() {
		It("should be generated and written to the webhook server CertDir", func() {
			print(certPath, keyPath, constants.WebhookScrtName, addonNamespace, webhookServiceName, webhookSelectorString)
			secret, err := webhooks.InstallNewCertificates(ctx, k8sConfig, certPath, keyPath, constants.WebhookScrtName, addonNamespace, webhookServiceName, webhookSelectorString)
			Expect(err).ToNot(HaveOccurred())
			Expect(secret).NotTo(BeNil())
			cert, err := cert2.CertsFromFile(certPath)
			Expect(err).ToNot(HaveOccurred())
			certPEM, err := cert2.EncodeCertificates(cert[0])
			Expect(err).ToNot(HaveOccurred())
			Expect(certPEM).To(Equal(secret.Data[resources.ServerCert]))
			key, err := keyutil.PrivateKeyFromFile(keyPath)
			Expect(err).ToNot(HaveOccurred())
			orgKey, err := keyutil.ParsePrivateKeyPEM(secret.Data[resources.ServerKey])
			Expect(err).ToNot(HaveOccurred())
			Expect(key).To(Equal(orgKey))
		})
	})

	Context("Mutating and Validating Configurations", func() {
		It("should be updated with CA bundle ", func() {
			err := testutil.SetupWebhookCertificates(ctx, k8sClient, k8sConfig, &webhookCertDetails)
			Expect(err).ToNot(HaveOccurred())
		})
		It("should fail update when label selector is not provided", func() {
			_, err := webhooks.InstallNewCertificates(ctx, k8sConfig, certPath, keyPath, constants.WebhookScrtName, addonNamespace, webhookServiceName, "")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Delete webhook Manifests", func() {
		It("Should delete webhook manifests after tests", func() {
			f, err := os.Open("testdata/test-webhook-manifests.yaml")
			Expect(err).ToNot(HaveOccurred())
			defer f.Close()
			err = testutil.DeleteResources(f, cfg, dynamicClient, true)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
