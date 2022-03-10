// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"os"

	adminregv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	cert2 "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
	"knative.dev/pkg/webhook/certificates/resources"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/addons/testutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/webhooks"
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
			secret, err := webhooks.NewTLSSecret(ctx, webhookScrtName, webhookServiceName, certPath, keyPath, addonNamespace)
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
			var err error

			labelMatch, _ := labels.NewRequirement("webhook-cert", selection.Equals, []string{"self-managed"})
			labelSelector := labels.NewSelector()
			labelSelector = labelSelector.Add(*labelMatch)

			orgvwcfgs := &adminregv1.ValidatingWebhookConfigurationList{}
			err = k8sClient.List(ctx, orgvwcfgs, &client.ListOptions{LabelSelector: labelSelector})
			Expect(err).ToNot(HaveOccurred())
			Expect(orgvwcfgs.Items).ToNot(BeEmpty())

			orgmwcfgs := &adminregv1.MutatingWebhookConfigurationList{}
			err = k8sClient.List(ctx, orgmwcfgs, &client.ListOptions{LabelSelector: labelSelector})
			Expect(err).ToNot(HaveOccurred())
			Expect(orgmwcfgs.Items).ToNot(BeEmpty())

			secret, err := webhooks.InstallNewCertificates(ctx, k8sConfig, certPath, keyPath, webhookScrtName, addonNamespace, webhookServiceName, "webhook-cert=self-managed")
			Expect(err).ToNot(HaveOccurred())

			vwcfgs := &adminregv1.ValidatingWebhookConfigurationList{}
			err = k8sClient.List(ctx, vwcfgs, &client.ListOptions{LabelSelector: labelSelector})
			Expect(err).ToNot(HaveOccurred())
			Expect(vwcfgs.Items).ToNot(BeEmpty())
			for _, wcfg := range vwcfgs.Items {
				for _, whook := range wcfg.Webhooks {
					Expect(whook.ClientConfig.CABundle).To(Equal(secret.Data[resources.CACert]))
				}
			}

			mwcfgs := &adminregv1.MutatingWebhookConfigurationList{}
			err = k8sClient.List(ctx, mwcfgs, &client.ListOptions{LabelSelector: labelSelector})
			Expect(err).ToNot(HaveOccurred())
			Expect(mwcfgs.Items).ToNot(BeEmpty())
			for _, wcfg := range mwcfgs.Items {
				for _, whook := range wcfg.Webhooks {
					Expect(whook.ClientConfig.CABundle).To(Equal(secret.Data[resources.CACert]))
				}
			}

		})
		It("should fail fail update when  label selector is not provided", func() {
			_, err := webhooks.InstallNewCertificates(ctx, k8sConfig, certPath, keyPath, webhookScrtName, addonNamespace, webhookServiceName, "")
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
