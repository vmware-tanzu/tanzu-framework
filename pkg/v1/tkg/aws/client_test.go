// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package aws_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	awscreds "sigs.k8s.io/cluster-api-provider-aws/cmd/clusterawsadm/credentials"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/aws"
)

func TestClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AWS Client Suite")
}

const (
	// encoded credentials generated manually by runnging "clusterawsadm alpha bootstrap encode-aws-credentials"
	awsEncodedCredentials = "W2RlZmF1bHRdCmF3c19hY2Nlc3Nfa2V5X2lkID0gZmFrZS1hY2Nlc3Mva2V5K2lkCmF3c19zZWNyZXRfYWNjZXNzX2tleSA9IGZha2Utc2VjcmV0LWFjY2Vzcy1rZXkKcmVnaW9uID0gZmFrZS1yZWdpb24KCg=="

	fakeRegion          = "fake-region"
	fakeAccessKeyID     = "fake-access/key+id"
	fakeSecretAccessKey = "fake-secret-access-key" // #nosec
)

var _ = Describe("Unit tests for aws client", func() {
	var (
		err             error
		region          string
		accessKeyID     string
		secretAccessKey string
		awsClient       aws.Client
	)

	Describe("Encode aws credentials", func() {
		var encodedCreds string
		JustBeforeEach(func() {
			awsCreds := awscreds.AWSCredentials{
				AccessKeyID:     accessKeyID,
				SecretAccessKey: secretAccessKey,
				Region:          region,
			}
			awsClient, err = aws.New(awsCreds)
			Expect(err).ToNot(HaveOccurred())
			encodedCreds, err = awsClient.EncodeCredentials()
		})

		Context("When the same credentials are provided", func() {
			BeforeEach(func() {
				region = fakeRegion
				accessKeyID = fakeAccessKeyID
				secretAccessKey = fakeSecretAccessKey
			})
			It("should return the same encoded string for the same credentials", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(encodedCreds).To(Equal(awsEncodedCredentials))
			})
		})

		Context("When different credentials are provided", func() {
			BeforeEach(func() {
				region = "another-fake-region"
				accessKeyID = fakeAccessKeyID
				secretAccessKey = fakeSecretAccessKey
			})
			It("should return the different encoded string for the same credentials", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(encodedCreds).ToNot(Equal(awsEncodedCredentials))
			})
		})
	})
})
