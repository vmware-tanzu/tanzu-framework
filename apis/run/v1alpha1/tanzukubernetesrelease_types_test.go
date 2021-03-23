// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainerImageString(t *testing.T) {
	assert := assert.New(t)

	ci := ContainerImage{
		Repository: "repo",
		Name:       "nginx",
		Tag:        "1.2.3",
	}

	assert.Equal(ci.String(), "repo/nginx:1.2.3")
}

func TestContainerImageString_Tagless(t *testing.T) {
	assert := assert.New(t)

	ci := ContainerImage{
		Repository: "repo",
		Name:       "nginx",
	}

	assert.Equal(ci.String(), "repo/nginx")
}

func TestContainerImageString_Repoless(t *testing.T) {
	assert := assert.New(t)

	ci := ContainerImage{
		Name: "nginx",
		Tag:  "1.2.3",
	}

	assert.Equal(ci.String(), "nginx:1.2.3")
}
