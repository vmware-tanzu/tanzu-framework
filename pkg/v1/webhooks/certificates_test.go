// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"context"
	"os"
	"path"
	"testing"
	"time"

	cert2 "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
	"knative.dev/pkg/webhook/certificates/resources"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	addonNamespace     = "tkg-system"
	webhookServiceName = "webhook-service"
	webhookScrtName    = "webhook-tls"
)

var (
	ctx      = context.TODO()
	certPath string
	keyPath  string
	tmpDir   string
)

func TestWebhookCertificates(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Webhook certificate",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func() {
	var err error
	tmpDir, err = os.MkdirTemp("/tmp", "webhooktest")
	Expect(err).ToNot(HaveOccurred())
	certPath = path.Join(tmpDir, "tls.cert")
	keyPath = path.Join(tmpDir, "tls.key")
})

var _ = AfterSuite(func() {
	By("remove test resources")
	os.RemoveAll(tmpDir)
	_, err := os.Stat(tmpDir)
	Expect(os.IsNotExist(err))

}, 60)

var _ = Describe("Webhook", func() {

	Context("server's certificate and key", func() {
		It("should be generated and written to the webhook server CertDir", func() {
			secret, err := NewTLSSecret(ctx, webhookScrtName, webhookServiceName, certPath, keyPath, addonNamespace)
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
		It("should become invalid after one week", func() {
			secret, err := NewTLSSecret(ctx, webhookScrtName, webhookServiceName, certPath, keyPath, addonNamespace)
			Expect(err).ToNot(HaveOccurred())
			Expect(secret).NotTo(BeNil())
			err = ValidateTLSSecret(secret, time.Hour*24) // valid cert life is one week. One day should not make it invalid
			Expect(err).ShouldNot(HaveOccurred())
			err = ValidateTLSSecret(secret, 8*time.Hour*24) // in 8 days certificate should be  invalid
			Expect(err).Should(HaveOccurred())

		})
	})

})
