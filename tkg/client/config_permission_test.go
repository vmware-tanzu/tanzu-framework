// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client_test

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"sigs.k8s.io/cluster-api-provider-aws/v2/cmd/clusterawsadm/credentials"

	. "github.com/vmware-tanzu/tanzu-framework/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
)

var err error

var _ = Describe("Unit tests for get AWS credentials", func() {
	var (
		tkgConfigPath  string
		err            error
		creds          *credentials.AWSCredentials
		client         *TkgClient
		curHome        string
		curUserProfile string
	)
	BeforeEach(func() {
		createTempDirectory("template_test")
		curHome = os.Getenv("HOME")
		curUserProfile = os.Getenv("USERPROFILE")
		_ = os.Setenv("HOME", testingDir)
		_ = os.Setenv("USERPROFILE", testingDir)
	})

	JustBeforeEach(func() {
		client, err = CreateTKGClient(tkgConfigPath, testingDir, defaultTKGBoMFileForTesting, 2*time.Second)
		Expect(err).ToNot(HaveOccurred())
		creds, err = client.GetAWSCreds()
	})

	Context("When AWS_REGION is not set in the environment variable nor tkgconfig", func() {
		BeforeEach(func() {
			tkgConfigPath = configFile
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("cannot find AWS region"))
		})
	})

	Context("When credentials are set in tkgconfig", func() {
		BeforeEach(func() {
			tkgConfigPath = "../fakes/config/config4.yaml"
		})
		It("should get the credentials without error", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(creds.Region).To(Equal("us-east-2"))
			Expect(creds.AccessKeyID).To(Equal("QWRETYUIOPLKJHGFDSAZ"))
			Expect(creds.SecretAccessKey).To(Equal("uNncCatIvWu1e$rqwerkg35qU7dswfEa4rdXJk/E"))
		})
	})

	Context("When credentials are not set in tkgconfig and default aws cred chain", func() {
		BeforeEach(func() {
			tkgConfigPath = configFile
			os.Setenv(constants.ConfigVariableAWSRegion, "us-east-2")
			os.Unsetenv(constants.ConfigVariableAWSAccessKeyID)
			os.Unsetenv(constants.ConfigVariableAWSSecretAccessKey)
		})
		It("should return an error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(HavePrefix("failed to gather credentials"))
		})
	})

	Context("When credentials are not set in tkgconfig and but in default aws cred chain", func() {
		BeforeEach(func() {
			tkgConfigPath = configFile
			os.Setenv(constants.ConfigVariableAWSRegion, "us-east-2")
			os.Unsetenv(constants.ConfigVariableAWSAccessKeyID)
			os.Unsetenv(constants.ConfigVariableAWSSecretAccessKey)
			err = os.Mkdir(filepath.Join(testingDir, ".aws"), os.ModePerm)

			Expect(err).ToNot(HaveOccurred())

			input, err := os.ReadFile("../fakes/config/aws_credentials")
			Expect(err).ToNot(HaveOccurred())
			dest := filepath.Join(testingDir, ".aws", "credentials")
			err = os.WriteFile(dest, input, 0o644)
			Expect(err).ToNot(HaveOccurred())
		})
		It("should get the credentials without error ", func() {
			Expect(err).ToNot(HaveOccurred())
			Expect(creds.Region).To(Equal("us-east-2"))
			Expect(creds.AccessKeyID).To(Equal("my_aws_access_key_id"))
			Expect(creds.SecretAccessKey).To(Equal("my_aws_secret_access_key"))
		})
	})

	AfterEach(func() {
		deleteTempDirectory()
		_ = os.Setenv("HOME", curHome)
		_ = os.Setenv("USERPROFILE", curUserProfile)
	})
})

func createTempDirectory(prefix string) {
	testingDir, err = os.MkdirTemp("", prefix)
	if err != nil {
		fmt.Println("Error TempDir: ", err.Error())
	}
}

func deleteTempDirectory() {
	os.Remove(testingDir)
}
