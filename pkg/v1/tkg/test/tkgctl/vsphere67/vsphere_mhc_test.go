// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// nolint:typecheck,nolintlint
package vsphere67

import (
	"context"

	. "github.com/onsi/ginkgo"

	. "github.com/vmware-tanzu-private/core/pkg/v1/tkg/test/tkgctl/shared"
)

var _ = Describe("MHC tests for vSphere clusters", func() {
	E2EMhcSpec(context.TODO(), func() E2EMhcSpecInput {
		return E2EMhcSpecInput{
			E2EConfig:       e2eConfig,
			ArtifactsFolder: artifactsFolder,
			Cni:             "calico",
		}
	})
})
