// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package nodeutils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestReplaceNodes(t *testing.T) {
	tests := []struct {
		name     string
		dst      string
		src      string
		metadata map[string]string
		output   string
	}{
		{
			name:     "success replace nodes with empty patch strategy",
			metadata: make(map[string]string),
			dst: `name: test-mc
type: k8s
group: one
clusterOpts:
  isManagementCluster: true
  annotation: one
  required: true
  annotationStruct:
    one: one
  endpoint: test-endpoint
  path: test-path
  context: test-context
discoverySources:
  - gcp:
      name: test
      bucket: test-bucket
      manifestPath: test-manifest-path
      annotation: one
      required: true
    contextType: tmc
  - gcp:
      name: test-two
      bucket: test-bucket
      manifestPath: test-manifest-path
      annotation: two
      required: true
    contextType: tmc`,
			src: `name: test-mc
type: k8s
clusterOpts:
  isManagementCluster: true
  endpoint: test-endpoint
  context: test-context
discoverySources:
  - gcp:
      name: test
      bucket: test-bucket
      manifestPath: test-manifest-pat
    contextType: tmc
  - gcp:
      name: test-two
      bucket: test-bucket
      manifestPath: test-manifest-path
    contextType: tmc`,
			output: `name: test-mc
type: k8s
group: one
clusterOpts:
  isManagementCluster: true
  annotation: one
  required: true
  annotationStruct:
    one: one
  endpoint: test-endpoint
  path: test-path
  context: test-context
discoverySources:
  - gcp:
      name: test
      bucket: test-bucket
      manifestPath: test-manifest-path
      annotation: one
      required: true
    contextType: tmc
  - gcp:
      name: test-two
      bucket: test-bucket
      manifestPath: test-manifest-path
      annotation: two
      required: true
    contextType: tmc`,
		},

		{
			name: "success replace nodes with patch strategy",
			metadata: map[string]string{
				"contexts.group": "replace",
			},
			dst: `name: test-mc
type: k8s
group: one
clusterOpts:
  isManagementCluster: true
  annotation: one
  required: true
  annotationStruct:
    one: one
  endpoint: test-endpoint
  path: test-path
  context: test-context
discoverySources:
  - gcp:
      name: test
      bucket: test-bucket
      manifestPath: test-manifest-path
      annotation: one
      required: true
    contextType: tmc
  - gcp:
      name: test-two
      bucket: test-bucket
      manifestPath: test-manifest-path
      annotation: two
      required: true
    contextType: tmc`,
			src: `name: test-mc
type: k8s
clusterOpts:
  isManagementCluster: true
  endpoint: test-endpoint
  context: test-context
discoverySources:
  - gcp:
      name: test
      bucket: test-bucket
      manifestPath: test-manifest-pat
    contextType: tmc
  - gcp:
      name: test-two
      bucket: test-bucket
      manifestPath: test-manifest-path
    contextType: tmc`,
			output: `name: test-mc
type: k8s
clusterOpts:
  isManagementCluster: true
  annotation: one
  required: true
  annotationStruct:
    one: one
  endpoint: test-endpoint
  path: test-path
  context: test-context
discoverySources:
  - gcp:
      name: test
      bucket: test-bucket
      manifestPath: test-manifest-path
      annotation: one
      required: true
    contextType: tmc
  - gcp:
      name: test-two
      bucket: test-bucket
      manifestPath: test-manifest-path
      annotation: two
      required: true
    contextType: tmc`,
		},

		{
			name: "success replace nodes with patch strategies",
			metadata: map[string]string{
				"contexts.group":                        "replace",
				"contexts.clusterOpts.annotation":       "replace",
				"contexts.clusterOpts.annotationStruct": "replace",
				"contexts.clusterOpts.optional":         "replace",
			},
			dst: `name: test-mc
type: k8s
group: one
clusterOpts:
  isManagementCluster: true
  annotation: one
  required: true
  annotationStruct:
    one: one
  optional:
   - one
   - two
  endpoint: test-endpoint
  path: test-path
  context: test-context
discoverySources:
  - gcp:
      name: test
      bucket: test-bucket
      manifestPath: test-manifest-path
      annotation: one
      required: true
    contextType: tmc
  - gcp:
      name: test-two
      bucket: test-bucket
      manifestPath: test-manifest-path
      annotation: two
      required: true
    contextType: tmc`,
			src: `name: test-mc
type: k8s
clusterOpts:
  isManagementCluster: true
  endpoint: test-endpoint
  context: test-context
discoverySources:
  - gcp:
      name: test
      bucket: test-bucket
      manifestPath: test-manifest-pat
    contextType: tmc
  - gcp:
      name: test-two
      bucket: test-bucket
      manifestPath: test-manifest-path
    contextType: tmc`,
			output: `name: test-mc
type: k8s
clusterOpts:
  isManagementCluster: true
  required: true
  endpoint: test-endpoint
  path: test-path
  context: test-context
discoverySources:
  - gcp:
      name: test
      bucket: test-bucket
      manifestPath: test-manifest-path
      annotation: one
      required: true
    contextType: tmc
  - gcp:
      name: test-two
      bucket: test-bucket
      manifestPath: test-manifest-path
      annotation: two
      required: true
    contextType: tmc`,
		},
	}
	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			// Setup data
			var src yaml.Node
			var dst yaml.Node
			var output yaml.Node
			var err error
			err = yaml.Unmarshal([]byte(spec.src), &src)
			assert.NoError(t, err)
			err = yaml.Unmarshal([]byte(spec.dst), &dst)
			assert.NoError(t, err)
			err = yaml.Unmarshal([]byte(spec.output), &output)
			assert.NoError(t, err)

			// Perform action
			err = ReplaceNodes(&src, &dst, WithPatchStrategyKey("contexts"), WithPatchStrategies(spec.metadata))
			assert.NoError(t, err)

			// Assert outcome
			dstBytes, err := yaml.Marshal(&dst)
			assert.NoError(t, err)
			outputBytes, err := yaml.Marshal(&output)
			assert.NoError(t, err)
			assert.Equal(t, string(dstBytes), string(outputBytes))
		})
	}
}
