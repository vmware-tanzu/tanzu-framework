// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package docker

import (
	"context"

	. "github.com/onsi/ginkgo"

	. "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/test/tkgctl/shared"
)

var _ = Describe("CEIP tests for capd clusters", func() {
	E2ECEIPSpec(context.TODO(), func() E2ECEIPSpecInput {
		return E2ECEIPSpecInput{
			E2EConfig:       e2eConfig,
			ArtifactsFolder: artifactsFolder,
			Cni:             "antrea",
		}
	})
})
