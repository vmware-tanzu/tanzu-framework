// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// nolint:typecheck,nolintlint
package azure_cc

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"

	. "github.com/vmware-tanzu/tanzu-framework/tkg/test/tkgctl/shared"
)

var _ = Describe("Functional tests for azure (clusterclass) - Antrea", func() {
	E2ECommonCCSpec(context.TODO(), func() E2ECommonCCSpecInput {
		return E2ECommonCCSpecInput{
			E2EConfig:             e2eConfig,
			ArtifactsFolder:       artifactsFolder,
			Cni:                   "antrea",
			Plan:                  "dev",
			Namespace:             "tkg-system",
			CheckAdmissionWebhook: true,
			Timeout:               time.Minute * 120,
			OtherConfig: map[string]string{
				"ENABLE_MHC":                 "true",
				"MHC_UNKNOWN_STATUS_TIMEOUT": "30m",
				"MHC_FALSE_STATUS_TIMEOUT":   "60m",
				"NODE_STARTUP_TIMEOUT":       "120m",
			},
		}
	})
})
