// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package aws_test

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/awslabs/goformation/v4/cloudformation"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sergi/go-diff/diffmatchpatch"
	"sigs.k8s.io/cluster-api-provider-aws/cmd/clusterawsadm/cloudformation/bootstrap"
	awscreds "sigs.k8s.io/cluster-api-provider-aws/cmd/clusterawsadm/credentials"
	"sigs.k8s.io/yaml"

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

	JustBeforeEach(func() {
		awsCreds := awscreds.AWSCredentials{
			AccessKeyID:     accessKeyID,
			SecretAccessKey: secretAccessKey,
			Region:          region,
		}
		awsClient, err = aws.New(awsCreds)
	})
	Describe("Encode aws credentials", func() {
		var encodedCreds string
		JustBeforeEach(func() {
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

	Describe("Generating Cloudformation", func() {
		awsClient, err = aws.New(awscreds.AWSCredentials{
			AccessKeyID:     fakeRegion,
			SecretAccessKey: fakeAccessKeyID,
			Region:          fakeSecretAccessKey,
		})
		Expect(err).ToNot(HaveOccurred())
	})
	Context("Defaults", func() {
		template, err := awsClient.GenerateBootstrapTemplate(aws.GenerateBootstrapTemplateInput{})
		Expect(err).NotTo(HaveOccurred())
		testBootstrapTemplate("default", template)
	})
	Context("With TMC Permissions", func() {
		template, err := awsClient.GenerateBootstrapTemplate(aws.GenerateBootstrapTemplateInput{
			EnableTanzuMissionControlPermissions: true,
		})
		Expect(err).NotTo(HaveOccurred())
		testBootstrapTemplate("with-tmc", template)
	})
})

func testBootstrapTemplate(name string, template *bootstrap.Template) {
	defer GinkgoRecover()
	cfn := cloudformation.Template{}
	data, err := os.ReadFile(path.Join("fixtures", "cloudformation-"+name+".yaml"))
	Expect(err).ToNot(HaveOccurred())
	err = yaml.Unmarshal(data, cfn)
	Expect(err).ToNot(HaveOccurred())
	tData, err := template.RenderCloudFormation().YAML()
	Expect(err).ToNot(HaveOccurred())
	err = os.WriteFile("/tmp/tmp1", tData, 0600)
	Expect(err).ToNot(HaveOccurred())
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(string(tData), string(data), false)
	out := dmp.DiffPrettyText(diffs)
	Expect(string(tData)).To(Equal(string(data)), fmt.Sprintf("Differing output (%s):\n%s", name, out))
}
