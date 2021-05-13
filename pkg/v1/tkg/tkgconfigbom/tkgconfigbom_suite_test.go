// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgconfigbom_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestTkgconfigbom(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tkgconfigbom Suite")
}
