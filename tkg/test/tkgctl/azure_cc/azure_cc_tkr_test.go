// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package azure_cc

import (
	"context"

	. "github.com/onsi/ginkgo"

	. "github.com/vmware-tanzu/tanzu-framework/tkg/test/tkgctl/shared"
)

var _ = Describe("Functional tests for azure - TKR Compatibility tests", func() {
	TKRCompatibilityValidationSpec(context.TODO(), func() TKRCompatibilityValidationSpecInput {
		return TKRCompatibilityValidationSpecInput{
			E2EConfig: e2eConfig,
		}
	})
})
