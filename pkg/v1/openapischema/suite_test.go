// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package openapischema

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestOpenapi(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Openapi Suite")
}
