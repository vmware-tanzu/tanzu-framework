// Copyright 2020 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package aws

import (
	"context"

	. "github.com/onsi/ginkgo"

	. "github.com/vmware-tanzu-private/core/pkg/v1/tkg/test/tkgctl/shared"

	. "github.com/vmware-tanzu-private/core/pkg/v1/tkg/test/tkgctl/shared"
)

var _ = Describe("Scale tests for aws clusters", func() {
	E2EScaleSpec(context.TODO(), func() E2EScaleSpecInput {
		return E2EScaleSpecInput{
			E2EConfig:       e2eConfig,
			ArtifactsFolder: artifactsFolder,
			Cni:             "calico",
		}
	})
})
