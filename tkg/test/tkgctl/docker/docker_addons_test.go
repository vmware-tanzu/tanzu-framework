// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// nolint:typecheck,nolintlint
package docker

import (
	"context"

	. "github.com/onsi/ginkgo"

	. "github.com/vmware-tanzu/tanzu-framework/tkg/test/tkgctl/shared"
)

var _ = Describe("Addon tests for capd clusters", func() {
	E2EAddonSpec(context.TODO(), func() E2EAddonSpecInput {
		return E2EAddonSpecInput{
			E2EConfig:       e2eConfig,
			ArtifactsFolder: artifactsFolder,
			Cni:             "antrea",
		}
	})
})
