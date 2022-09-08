// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package packageclient

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestPackageClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PackageClient Suite")
}
