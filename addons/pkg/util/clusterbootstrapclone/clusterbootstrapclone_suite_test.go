// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package clusterbootstrapclone

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestClusterbootstrapclone(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ClusterbootstrapClone Suite")
}
