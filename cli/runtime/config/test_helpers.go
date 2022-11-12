// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

func setupConfigMetadataWithMigrateToNewConfig() string {
	metadata := `configMetadata:
  settings:
    useUnifiedConfig: true`

	return metadata
}

//nolint:funlen
func setupMultiCfgData() (string, string) {
	cfg := `servers:
  - name: test-mc
    type: managementcluster
    managementClusterOpts:
      endpoint: test-ctx-endpoint
      path: test-ctx-path
      context: test-ctx-context
    discoverySources:
      - gcp:
          name: test
          bucket: test-ctx-bucket
          manifestPath: test-ctx-manifest-path
        contextType: tmc
  - name: test-mc2
    type: managementcluster
    managementClusterOpts:
      endpoint: test-ctx-endpoint
      path: test-ctx-path
      context: test-ctx-context
    discoverySources:
      - gcp:
          name: test
          bucket: test-ctx-bucket
          manifestPath: test-ctx-manifest-path
        contextType: tmc
      - gcp:
          name: test2
          bucket: test-ctx-bucket
          manifestPath: test-ctx-manifest-path
        contextType: tmc
  - name: test-mc3
    type: managementcluster
    managementClusterOpts:
      endpoint: test-ctx-endpoint
      path: test-ctx-path
      context: test-ctx-context
    discoverySources:
      - gcp:
          name: test
          bucket: test-ctx-bucket
          manifestPath: test-ctx-manifest-path
        contextType: tmc
current: test-mc
`
	cfg2 := `currentContext:
  k8s: test-mc
contexts:
  - name: test-mc
    ctx-field: new-ctx-field
    optional: true
    type: k8s
    clusterOpts:
      isManagementCluster: true
      endpoint: test-endpoint
      annotation: one
      required: true
      annotationStruct:
        one: one
    discoverySources:
      - gcp:
          name: test
          bucket: test-bucket
          manifestPath: test-manifest-path
          annotation: one
          required: true
        contextType: tmc
      - gcp:
          name: test2
          bucket: test-bucket
          manifestPath: test-manifest-path
          annotation: one
          required: true
        contextType: tmc
  - name: test-mc2
    ctx-field: new-ctx-field
    optional: true
    type: k8s
    clusterOpts:
      isManagementCluster: true
      endpoint: test-endpoint
      annotation: one
      required: true
      annotationStruct:
        one: one
    discoverySources:
      - gcp:
          name: test
          bucket: test-bucket
          manifestPath: test-manifest-path
          annotation: one
          required: true
        contextType: tmc
      - gcp:
          name: test2
          bucket: test-bucket
          manifestPath: test-manifest-path
          annotation: one
          required: true
        contextType: tmc
  - name: test-mc3
    ctx-field: new-ctx-field
    optional: true
    type: k8s
    clusterOpts:
      isManagementCluster: true
      endpoint: test-endpoint
      annotation: one
      required: true
      annotationStruct:
        one: one
    discoverySources:
      - gcp:
          name: test
          bucket: test-bucket
          manifestPath: test-manifest-path
          annotation: one
          required: true
        contextType: tmc
      - gcp:
          name: test2
          bucket: test-bucket
          manifestPath: test-manifest-path
          annotation: one
          required: true
        contextType: tmc
`
	return cfg, cfg2
}
