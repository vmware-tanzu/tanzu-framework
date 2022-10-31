// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// nolint:typecheck,nolintlint
package azure_cc

import (
	"context"

	. "github.com/onsi/ginkgo"

	. "github.com/vmware-tanzu/tanzu-framework/tkg/test/tkgctl/shared"
)

var _ = Describe("Functional tests to create & upgrade Azure cluster with custom ClusterBootstrap", func() {
	E2ECommonCCSpec(context.TODO(), func() E2ECommonCCSpecInput {
		return E2ECommonCCSpecInput{
			E2EConfig:       e2eConfig,
			ArtifactsFolder: artifactsFolder,
			Cni:             "antrea",
			Plan:            "dev",
			Namespace:       "tkg-system",
			IsCustomCB:      true,
			DoUpgrade:       true,
		}
	})
})
