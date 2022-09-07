// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package yamlprocessor_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestYamlprocessor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Yamlprocessor Suite")
}
