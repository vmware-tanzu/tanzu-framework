// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

func TestGetConfigMetadata(t *testing.T) {
	// setup
	func() {
		LocalDirName = TestLocalDirName
	}()

	defer func() {
		cleanupDir(LocalDirName)
	}()

	tests := []struct {
		name   string
		in     *configapi.Metadata
		out    string
		errStr string
	}{
		{
			name: "success k8s",
			in: &configapi.Metadata{
				ConfigMetadata: &configapi.ConfigMetadata{
					PatchStrategy: map[string]string{
						"contexts.clusterOpts":                      "replace",
						"contexts.discoverySources.gcp.annotations": "replace",
						"contexts.globalOpts.auth":                  "replace",
					},
				},
			},
		},
	}
	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			err := SetConfigMetadataPatchStrategies(spec.in.ConfigMetadata.PatchStrategy)
			if err != nil {
				fmt.Printf("SetConfigMetadataPatchStrategies errors: %v\n", err)
			}
			c, err := GetConfigMetadata()
			assert.Equal(t, c, spec.in)
			assert.NoError(t, err)
		})
	}
}
