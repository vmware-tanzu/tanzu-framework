// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// nolint:typecheck,nolintlint
package aws_cc

import (
	"context"

	. "github.com/onsi/ginkgo"

	. "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/test/tkgctl/shared"
)

var _ = Describe("Functional tests for aws (clusterclass) - Antrea", func() {
	E2ECommonSpec(context.TODO(), func() E2ECommonSpecInput {
		return E2ECommonSpecInput{
			E2EConfig:       e2eConfig,
			ArtifactsFolder: artifactsFolder,
			Cni:             "antrea",
			Plan:            "dev",
			Namespace:       "tkg-system",
			OtherConfigs:    map[string]string{"clusterclass": "true"},
		}
	})
})
