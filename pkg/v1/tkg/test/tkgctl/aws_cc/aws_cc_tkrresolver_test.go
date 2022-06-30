// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// nolint:typecheck,nolintlint
package aws_cc

import (
	"context"

	. "github.com/onsi/ginkgo"

	. "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/test/tkgctl/shared"
)

var _ = Describe("Functional tests for aws - TKRResolver", func() {
	E2ETKRResolverValidationForClusterCRUDSpec(context.TODO(), func() E2ETKRResolverValidationForClusterCRUDSpecInput {
		return E2ETKRResolverValidationForClusterCRUDSpecInput{
			E2EConfig:       e2eConfig,
			ArtifactsFolder: artifactsFolder,
			Cni:             "antrea",
			Plan:            "devcc",
			Namespace:       "tkg-system",
		}
	})
})
