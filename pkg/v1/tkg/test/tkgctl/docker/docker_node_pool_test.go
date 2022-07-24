// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// nolint:typecheck,nolintlint
package docker

import (
	"context"

	. "github.com/onsi/ginkgo"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	. "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/test/tkgctl/shared"
)

var _ = Describe("Node pool tests for capd clusters", func() {
	E2ENodePoolSpec(context.TODO(), func() E2ENodePoolSpecInput {
		replicas := int32(1)
		return E2ENodePoolSpecInput{
			E2ECommonSpecInput: E2ECommonSpecInput{
				E2EConfig:       e2eConfig,
				ArtifactsFolder: artifactsFolder,
				Cni:             "antrea",
				Plan:            "dev",
				Namespace:       "default",
			},
			NodePool: client.NodePool{
				Name:     "np-1",
				Replicas: &replicas,
				Labels: &map[string]string{
					"provider": "docker",
				},
			},
		}
	})
})
