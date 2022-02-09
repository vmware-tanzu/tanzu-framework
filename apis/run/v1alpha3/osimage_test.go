// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha3

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMachineImageInfo_DeepCopy(t *testing.T) {
	imageInfo := &MachineImageInfo{
		Type: "aws",
		Ref: map[string]interface{}{
			"id":    "ami-0f2e5eec7ae0a1986",
			"isFoo": true,
		},
	}
	assert.Equal(t, imageInfo, imageInfo.DeepCopy())
}
