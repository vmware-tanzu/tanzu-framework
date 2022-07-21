// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package unit_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestUnit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cluster API Provider OCI Unit Suite")
}
