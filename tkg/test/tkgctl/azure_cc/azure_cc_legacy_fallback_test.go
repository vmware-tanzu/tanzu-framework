// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// nolint:typecheck,nolintlint
package azure_cc

import (
	"context"

	. "github.com/onsi/ginkgo"

	. "github.com/vmware-tanzu/tanzu-framework/tkg/test/tkgctl/shared"
)

var _ = Describe("Functional tests for fallback legacy creation handling", func() {
	E2ELegacyFallbackSpec(context.TODO(), func() E2ELegacyFallbackSpecInput {
		return E2ELegacyFallbackSpecInput{
			E2EConfig:       e2eConfig,
			ArtifactsFolder: artifactsFolder,
			Cni:             "antrea",
			Plan:            "dev",
			Namespace:       "tkg-system",
		}
	})
})
