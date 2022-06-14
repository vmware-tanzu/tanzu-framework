// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// nolint:typecheck,nolintlint
package aws_cc

import (
	"context"

	. "github.com/onsi/ginkgo"

	. "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/test/tkgctl/shared"
)

var _ = Describe("Functional tests for aws - Antrea", func() {
	E2ECommonSpec(context.TODO(), func() E2ECommonSpecInput {
		return E2ECommonSpecInput{
			E2EConfig:       e2eConfig,
			ArtifactsFolder: artifactsFolder,
			Cni:             "antrea",
			Plan:            "devcc",
			OtherConfigs: map[string]string{
				"AWS_VPC_ID":            "vpc-0f6da14d1b6f94a73",
				"AWS_PUBLIC_SUBNET_ID":  "subnet-06a866e3f2285b915",
				"AWS_PRIVATE_SUBNET_ID": "subnet-0ac292b48e3cceb4d",
			},
		}
	})
})
