// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"testing"

	"github.com/stretchr/testify/assert"

	cliv1alpha1 "github.com/vmware-tanzu-private/core/apis/cli/v1alpha1"
)

func Test_newRootCmd(t *testing.T) {
	assert := assert.New(t)

	descriptor := cliv1alpha1.PluginDescriptor{
		Name:        "Test Plugin",
		Description: "Description of the plugin",
		Version:     "1.2.3",
		BuildSHA:    "cafecafe",
		Group:       "TestGroup",
		DocURL:      "https://docs.example.com",
		Hidden:      false,
	}

	cmd := newRootCmd(&descriptor)
	assert.Equal("Test Plugin", cmd.Use)
	assert.Equal(("Description of the plugin"), cmd.Short)
}
