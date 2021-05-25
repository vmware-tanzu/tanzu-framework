// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"context"

	. "github.com/onsi/ginkgo"

	. "github.com/vmware-tanzu-private/core/pkg/v1/tkg/test/tkgctl/shared"
)

var _ = Describe("Functional tests for docker - calico", func() {
	E2ECommonSpec(context.TODO(), func() E2ECommonSpecInput {
		return E2ECommonSpecInput{
			E2EConfig:       e2eConfig,
			ArtifactsFolder: artifactsFolder,
			Cni:             "calico",
		}
	})
})
