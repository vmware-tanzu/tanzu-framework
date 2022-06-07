// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"context"
	"os"
	"path"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	cert2 "k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
	"knative.dev/pkg/webhook/certificates/resources"
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

var _ = BeforeSuite(func() {
	var err error
	tmpDir, err = os.MkdirTemp("/tmp", "webhooktest")
	Expect(err).ToNot(HaveOccurred())
	certPath = path.Join(tmpDir, "tls.cert")
	keyPath = path.Join(tmpDir, "tls.key")
})

var _ = AfterSuite(func() {
	By("remove test resources")
	_ = os.RemoveAll(tmpDir) // ignore errors since we check directory status next

	_, err := os.Stat(tmpDir)
	Expect(os.IsNotExist(err))

}, 60)

var _ = Describe("Webhook", func() {

	Context("server's certificate and key", func() {
		It("should be generated and written to the webhook server CertDir", func() {
			secret, err := resources.MakeSecret(ctx, webhookScrtName, addonNamespace, webhookServiceName)
			Expect(err).ToNot(HaveOccurred())
			err = WriteServerTLSToFileSystem(ctx, certPath, keyPath, secret)
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
		It("should only be written to file system if content is different", func() {
			secret, err := resources.MakeSecret(ctx, webhookScrtName, addonNamespace, webhookServiceName)
			Expect(err).ToNot(HaveOccurred())
			err = WriteServerTLSToFileSystem(ctx, certPath, keyPath, secret)
			Expect(err).ToNot(HaveOccurred())
			Expect(secret).NotTo(BeNil())
			certPathFile, err := os.Stat(certPath)
			Expect(err).ToNot(HaveOccurred())
			firstCertPathModifiedTime := certPathFile.ModTime()
			keyPathFile, err := os.Stat(keyPath)
			Expect(err).ToNot(HaveOccurred())
			firstKeyPathModifiedTime := keyPathFile.ModTime()

			By("files should not be modified if content will remain unchanged")
			err = WriteServerTLSToFileSystem(ctx, certPath, keyPath, secret)
			Expect(err).ToNot(HaveOccurred())
			Expect(secret).NotTo(BeNil())
			certPathFile, err = os.Stat(certPath)
			Expect(err).ToNot(HaveOccurred())
			secondCertPathModifiedTime := certPathFile.ModTime()
			keyPathFile, err = os.Stat(keyPath)
			Expect(err).ToNot(HaveOccurred())
			secondKeyPathModifiedTime := keyPathFile.ModTime()
			Expect(firstCertPathModifiedTime.Equal(secondCertPathModifiedTime))
			Expect(firstKeyPathModifiedTime.Equal(secondKeyPathModifiedTime))
		})
		It("Should be rewritten to file system if content will be different", func() {
			secret, err := resources.MakeSecret(ctx, webhookScrtName, addonNamespace, webhookServiceName)
			Expect(err).ToNot(HaveOccurred())
			err = WriteServerTLSToFileSystem(ctx, certPath, keyPath, secret)
			Expect(err).ToNot(HaveOccurred())
			Expect(secret).NotTo(BeNil())

			time.Sleep(time.Millisecond) // lets wait just enough to make sure modtime is different enough
			f, err := os.Create(certPath)
			Expect(err).ToNot(HaveOccurred())
			defer f.Close()
			_, err = f.WriteString("garbled data")
			Expect(err).ToNot(HaveOccurred())

			Expect(err).ToNot(HaveOccurred())
			err = WriteServerTLSToFileSystem(ctx, certPath, keyPath, secret)
			Expect(err).ToNot(HaveOccurred())
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
		It("Should be rewritten to file system if files are missing", func() {
			secret, err := resources.MakeSecret(ctx, webhookScrtName, addonNamespace, webhookServiceName)
			Expect(err).ToNot(HaveOccurred())
			err = WriteServerTLSToFileSystem(ctx, certPath, keyPath, secret)
			Expect(err).ToNot(HaveOccurred())
			Expect(secret).NotTo(BeNil())

			err = os.Remove(certPath)
			Expect(err).ToNot(HaveOccurred())
			_, err = os.Stat(certPath)
			Expect(os.IsNotExist(err)).To(BeTrue())
			time.Sleep(time.Millisecond) // wait long enough for mod time to be different
			err = WriteServerTLSToFileSystem(ctx, certPath, keyPath, secret)
			Expect(err).ToNot(HaveOccurred())

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
			secret, err := resources.MakeSecret(ctx, webhookScrtName, addonNamespace, webhookServiceName)
			Expect(err).ToNot(HaveOccurred())
			Expect(secret).NotTo(BeNil())
			err = ValidateTLSSecret(secret, time.Hour*24) // valid cert life is one week. One day should not make it invalid
			Expect(err).ShouldNot(HaveOccurred())
			err = ValidateTLSSecret(secret, 8*time.Hour*24) // in 8 days certificate should be  invalid
			Expect(err).Should(HaveOccurred())

		})
	})
})
