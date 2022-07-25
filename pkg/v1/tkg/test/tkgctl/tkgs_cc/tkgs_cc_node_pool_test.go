// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgs_cc

import (
	"context"

	. "github.com/onsi/ginkgo"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	. "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/test/tkgctl/shared"
)

var _ = Describe("Node pool tests for tkgs - ClusterClass", func() {
	E2ENodePoolSpec(context.TODO(), func() E2ENodePoolSpecInput {
		return E2ENodePoolSpecInput{
			E2ECommonSpecInput: E2ECommonSpecInput{
				E2EConfig:       e2eConfig,
				ArtifactsFolder: artifactsFolder,
				Cni:             "antrea",
				Plan:            "devcc",
			},
			NodePool: client.NodePool{
				Name:        "np-1",
				WorkerClass: "node-pool",
				TKRResolver: "os-arch=amd64",
				Replicas:    func(i int32) *int32 { return &i }(1),
				Labels: &map[string]string{
					"provider": "tkg-service-vsphere",
				},
			},
		}
	})
})
