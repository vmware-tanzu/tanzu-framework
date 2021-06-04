// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// nolint:typecheck,nolintlint
package docker

import (
	"context"

	. "github.com/onsi/ginkgo"

	. "github.com/vmware-tanzu-private/core/pkg/v1/tkg/test/tkgctl/shared"
)

var _ = Describe("Upgrade tests for capd clusters - antrea", func() {
	E2EUpgradeSpec(context.TODO(), func() E2EUpgradeSpecInput {
		return E2EUpgradeSpecInput{
			E2EConfig:       e2eConfig,
			ArtifactsFolder: artifactsFolder,
			Cni:             "antrea",
		}
	})
})
