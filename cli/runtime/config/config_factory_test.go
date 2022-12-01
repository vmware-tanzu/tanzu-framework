// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	cliapi "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

func setupCfgAndCfgNextGenData() (string, string, string, string) {
	cfg := `apiVersion: config.tanzu.vmware.com/v1alpha1
clientOptions:
  cli:
    bomRepo: projects.registry.vmware.com/tkg
    compatibilityFilePath: tkg-compatibility
    discoverySources:
      - contextType: k8s
        local:
          name: default-local
          path: standalone
      - local:
          name: admin-local
          path: admin
    edition: tkg
  features:
    cluster:
      custom-nameservers: 'false'
      dual-stack-ipv4-primary: 'false'
      dual-stack-ipv6-primary: 'false'
    global:
      context-aware-cli-for-plugins: 'true'
      context-target: 'false'
      tkr-version-v1alpha3-beta: 'false'
    management-cluster:
      aws-instance-types-exclude-arm: 'true'
      custom-nameservers: 'false'
      dual-stack-ipv4-primary: 'false'
      dual-stack-ipv6-primary: 'false'
      export-from-confirm: 'true'
      import: 'false'
      standalone-cluster-mode: 'false'
    package:
      kctrl-package-command-tree: 'true'
kind: ClientConfig
metadata:
  creationTimestamp: null
servers:
  - name: test-mc
    type: managementcluster
    managementClusterOpts:
      endpoint: test-endpoint
      path: test-path
      context: test-context
      annotation: one
      required: true
    discoverySources:
      - gcp:
          name: test
          bucket: test-bucket
          manifestPath: test-manifest-path
          annotation: one
          required: true
        contextType: tmc
current: test-mc
contexts:
  - name: test-mc
    target: kubernetes
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
        contextType: tmc
currentContext:
    kubernetes: test-mc
`
	expectedCfg := `apiVersion: config.tanzu.vmware.com/v1alpha1
clientOptions:
    cli:
        bomRepo: projects.registry.vmware.com/tkg
        compatibilityFilePath: tkg-compatibility
        discoverySources:
            - contextType: k8s
              local:
                name: default-local
                path: standalone
            - local:
                name: admin-local
                path: admin
        edition: tkg
    features:
        cluster:
            custom-nameservers: 'false'
            dual-stack-ipv4-primary: 'false'
            dual-stack-ipv6-primary: 'false'
        global:
            context-aware-cli-for-plugins: 'true'
            context-target: 'false'
            tkr-version-v1alpha3-beta: 'false'
        management-cluster:
            aws-instance-types-exclude-arm: 'true'
            custom-nameservers: 'false'
            dual-stack-ipv4-primary: 'false'
            dual-stack-ipv6-primary: 'false'
            export-from-confirm: 'true'
            import: 'false'
            standalone-cluster-mode: 'false'
        package:
            kctrl-package-command-tree: 'true'
kind: ClientConfig
metadata:
    creationTimestamp: null
servers:
    - name: test-mc
      type: managementcluster
      managementClusterOpts:
        endpoint: test-endpoint
        path: test-path
        context: test-context
        annotation: one
        required: true
      discoverySources:
        - gcp:
            name: test
            bucket: test-bucket
            manifestPath: test-manifest-path
            annotation: one
            required: true
          contextType: tmc
current: test-mc
contexts:
    - name: test-mc
      target: kubernetes
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
          contextType: tmc
currentContext:
    kubernetes: test-mc
`

	cfgNextGen := `
contexts:
  - name: test-mc
    target: kubernetes
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
        contextType: tmc
currentContext:
    kubernetes: test-mc
`
	expectedCfgNextGen := `contexts:
    - name: test-mc
      target: kubernetes
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
          contextType: tmc
currentContext:
    kubernetes: test-mc
`

	return cfg, cfgNextGen, expectedCfg, expectedCfgNextGen
}

func TestGetClientConfigWithLockAndWithoutLock(t *testing.T) {
	// Setup config test data
	cfg, cfgNextGen, _, _ := setupCfgAndCfgNextGenData()
	_, cleanUp := setupTestConfig(t, &CfgTestData{cfg: cfg, cfgNextGen: cfgNextGen})

	defer func() {
		cleanUp()
	}()

	//Actions
	nodeWithLock, err := getClientConfigNode()
	assert.NoError(t, err)
	//Actions
	nodeWithoutLocK, err := getClientConfigNodeNoLock()
	assert.NoError(t, err)

	nodes := []*yaml.Node{nodeWithLock, nodeWithoutLocK}
	for _, node := range nodes {
		// Assertions
		assert.NotNil(t, node)
		assert.NoError(t, err)

		expectedCtx := &configapi.Context{
			Name:   "test-mc",
			Target: cliapi.TargetK8s,
			ClusterOpts: &configapi.ClusterServer{
				Endpoint:            "test-endpoint",
				Path:                "test-path",
				Context:             "test-context",
				IsManagementCluster: true,
			},
			DiscoverySources: []configapi.PluginDiscovery{
				{
					GCP: &configapi.GCPDiscovery{
						Name:         "test",
						Bucket:       "test-bucket",
						ManifestPath: "test-manifest-path",
					},
				},
				{
					GCP: &configapi.GCPDiscovery{
						Name:         "test-two",
						Bucket:       "test-bucket",
						ManifestPath: "test-manifest-path",
					},
				},
			},
		}

		ctx, err := getContext(node, "test-mc")
		assert.NoError(t, err)
		assert.Equal(t, expectedCtx, ctx)

		expectedServer := &configapi.Server{
			Name: "test-mc",
			Type: "managementcluster",
			ManagementClusterOpts: &configapi.ManagementClusterServer{
				Endpoint: "test-endpoint",
				Path:     "test-path",
				Context:  "test-context",
			},
			DiscoverySources: []configapi.PluginDiscovery{
				{
					GCP: &configapi.GCPDiscovery{
						Name:         "test",
						Bucket:       "test-bucket",
						ManifestPath: "test-manifest-path",
					},
				},
			},
		}

		server, err := getServer(node, "test-mc")
		assert.NoError(t, err)
		assert.Equal(t, expectedServer, server)
	}
}

func TestGetClientConfigWithLockAndMigratedToNewConfig(t *testing.T) {
	// Setup config test data
	cfg, cfgNextGen, _, _ := setupCfgAndCfgNextGenData()
	_, cleanUp := setupTestConfig(t, &CfgTestData{cfg: cfg, cfgNextGen: cfgNextGen, cfgMetadata: setupConfigMetadataWithMigrateToNewConfig()})

	defer func() {
		cleanUp()
	}()

	//Actions
	node, err := getClientConfigNode()

	// Assertions
	assert.NotNil(t, node)
	assert.NoError(t, err)

	expectedCtx := &configapi.Context{
		Name:   "test-mc",
		Target: cliapi.TargetK8s,
		ClusterOpts: &configapi.ClusterServer{
			Endpoint:            "test-endpoint",
			Path:                "test-path",
			Context:             "test-context",
			IsManagementCluster: true,
		},
		DiscoverySources: []configapi.PluginDiscovery{
			{
				GCP: &configapi.GCPDiscovery{
					Name:         "test",
					Bucket:       "test-bucket",
					ManifestPath: "test-manifest-path",
				},
			},
			{
				GCP: &configapi.GCPDiscovery{
					Name:         "test-two",
					Bucket:       "test-bucket",
					ManifestPath: "test-manifest-path",
				},
			},
		},
	}

	// Migrated To new config hence servers are not in the config-ng.yaml yet.
	ctx, err := getContext(node, "test-mc")
	assert.NoError(t, err)
	assert.Equal(t, expectedCtx, ctx)

	_, err = getServer(node, "test-mc")
	assert.Equal(t, "could not find server \"test-mc\"", err.Error())
}

func TestGetClientConfigWithoutLockAndMigratedToNewConfig(t *testing.T) {
	// Setup config test data
	cfg, cfgNextGen, _, _ := setupCfgAndCfgNextGenData()
	_, cleanUp := setupTestConfig(t, &CfgTestData{cfg: cfg, cfgNextGen: cfgNextGen, cfgMetadata: setupConfigMetadataWithMigrateToNewConfig()})

	defer func() {
		cleanUp()
	}()

	//Actions
	AcquireTanzuConfigNextGenLock()
	node, err := getClientConfigNodeNoLock()
	ReleaseTanzuConfigNextGenLock()

	// Assertions
	assert.NotNil(t, node)
	assert.NoError(t, err)

	expectedCtx := &configapi.Context{
		Name:   "test-mc",
		Target: cliapi.TargetK8s,
		ClusterOpts: &configapi.ClusterServer{
			Endpoint:            "test-endpoint",
			Path:                "test-path",
			Context:             "test-context",
			IsManagementCluster: true,
		},
		DiscoverySources: []configapi.PluginDiscovery{
			{
				GCP: &configapi.GCPDiscovery{
					Name:         "test",
					Bucket:       "test-bucket",
					ManifestPath: "test-manifest-path",
				},
			},
			{
				GCP: &configapi.GCPDiscovery{
					Name:         "test-two",
					Bucket:       "test-bucket",
					ManifestPath: "test-manifest-path",
				},
			},
		},
	}

	// Migrated To new config hence servers are not in the config-ng.yaml yet.
	ctx, err := getContext(node, "test-mc")
	assert.NoError(t, err)
	assert.Equal(t, expectedCtx, ctx)

	_, err = getServer(node, "test-mc")
	assert.Equal(t, "could not find server \"test-mc\"", err.Error())
}

func TestPersistConfig(t *testing.T) {
	// Setup config test data
	cfg, cfgNextGen, expectedCfg, expectedCfgNextGen := setupCfgAndCfgNextGenData()
	cfgTestFiles, cleanUp := setupTestConfig(t, &CfgTestData{cfg: cfg, cfgNextGen: cfgNextGen})

	defer func() {
		cleanUp()
	}()

	// Actions
	node, err := getClientConfigNode()
	assert.NotNil(t, node)
	assert.NoError(t, err)

	err = persistConfig(node)
	assert.NoError(t, err)

	cfgFileData, err := os.ReadFile(cfgTestFiles[0].Name())
	assert.NoError(t, err)
	assert.Equal(t, expectedCfg, string(cfgFileData))

	cfgNextGenFileData, err := os.ReadFile(cfgTestFiles[1].Name())
	assert.NoError(t, err)
	assert.Equal(t, expectedCfgNextGen, string(cfgNextGenFileData))
}

func TestPersistConfigWithMigrateToNewConfig(t *testing.T) {
	// Setup data
	cfg, cfgNextGen, _, _ := setupCfgAndCfgNextGenData()
	expected := `contexts:
    - name: test-mc
      target: kubernetes
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
          contextType: tmc
currentContext:
    kubernetes: test-mc
`

	cfgTestFiles, cleanUp := setupTestConfig(t, &CfgTestData{cfg: cfg, cfgNextGen: cfgNextGen, cfgMetadata: setupConfigMetadataWithMigrateToNewConfig()})

	defer func() {
		cleanUp()
	}()

	// Actions
	node, err := getClientConfigNode()
	assert.NotNil(t, node)
	assert.NoError(t, err)

	err = persistConfig(node)
	assert.NoError(t, err)

	cfgFileData, err := os.ReadFile(cfgTestFiles[0].Name())
	assert.NoError(t, err)
	assert.Equal(t, cfg, string(cfgFileData))

	cfgNextGenFileData, err := os.ReadFile(cfgTestFiles[1].Name())
	assert.NoError(t, err)
	assert.Equal(t, expected, string(cfgNextGenFileData))
}

func TestMultiConfig(t *testing.T) {
	tests := []struct {
		name   string
		cfg    string
		cfg2   string
		output string
	}{
		{
			name: "success concat src into empty dst node",
			cfg: `apiVersion: config.tanzu.vmware.com/v1alpha1
clientOptions:
  cli:
    bomRepo: projects.registry.vmware.com/tkg
    compatibilityFilePath: tkg-compatibility
    discoverySources:
      - contextType: k8s
        local:
          name: default-local
          path: standalone
      - local:
          name: admin-local
          path: admin
    edition: tkg
  features:
    cluster:
      custom-nameservers: 'false'
      dual-stack-ipv4-primary: 'false'
      dual-stack-ipv6-primary: 'false'
    global:
      context-aware-cli-for-plugins: 'true'
      context-target: 'false'
      tkr-version-v1alpha3-beta: 'false'
    management-cluster:
      aws-instance-types-exclude-arm: 'true'
      custom-nameservers: 'false'
      dual-stack-ipv4-primary: 'false'
      dual-stack-ipv6-primary: 'false'
      export-from-confirm: 'true'
      import: 'false'
      standalone-cluster-mode: 'false'
    package:
      kctrl-package-command-tree: 'true'
contexts:
  - name: test-mc
    target: kubernetes
    group: one
    clusterOpts:
      isManagementCluster: true
      annotation: one
      required: true
      annotationStruct:
        one: one
      endpoint: cfg-test-endpoint
      path: cfg-test-path
      context: cfg-test-context
    discoverySources:
      - gcp:
          name: test
          bucket: cfg-test-bucket
          manifestPath: cfg-test-manifest-path
          annotation: one
          required: true
        contextType: tmc
      - gcp:
          name: test-two
          bucket: cfg-test-bucket
          manifestPath: cfg-test-manifest-path
          annotation: two
          required: true
        contextType: tmc
currentContext:
  kubernetes: test-mc
kind: ClientConfig
metadata:
  creationTimestamp: null`,
			cfg2: ``,
			output: `apiVersion: config.tanzu.vmware.com/v1alpha1
kind: ClientConfig
metadata:
    creationTimestamp: null
clientOptions:
    cli:
        bomRepo: projects.registry.vmware.com/tkg
        compatibilityFilePath: tkg-compatibility
        discoverySources:
            - contextType: k8s
              local:
                name: default-local
                path: standalone
            - local:
                name: admin-local
                path: admin
        edition: tkg
    features:
        cluster:
            custom-nameservers: 'false'
            dual-stack-ipv4-primary: 'false'
            dual-stack-ipv6-primary: 'false'
        global:
            context-aware-cli-for-plugins: 'true'
            context-target: 'false'
            tkr-version-v1alpha3-beta: 'false'
        management-cluster:
            aws-instance-types-exclude-arm: 'true'
            custom-nameservers: 'false'
            dual-stack-ipv4-primary: 'false'
            dual-stack-ipv6-primary: 'false'
            export-from-confirm: 'true'
            import: 'false'
            standalone-cluster-mode: 'false'
        package:
            kctrl-package-command-tree: 'true'
`,
		},
		{
			name: "success concat src into dst node if contexts and currentContexts already present in cfg2 no override",
			cfg: `apiVersion: config.tanzu.vmware.com/v1alpha1
clientOptions:
  cli:
    bomRepo: projects.registry.vmware.com/tkg
    compatibilityFilePath: tkg-compatibility
    discoverySources:
      - contextType: k8s
        local:
          name: default-local
          path: standalone
      - local:
          name: admin-local
          path: admin
    edition: tkg
  features:
    cluster:
      custom-nameservers: 'false'
      dual-stack-ipv4-primary: 'false'
      dual-stack-ipv6-primary: 'false'
    global:
      context-aware-cli-for-plugins: 'true'
      context-target: 'false'
      tkr-version-v1alpha3-beta: 'false'
    management-cluster:
      aws-instance-types-exclude-arm: 'true'
      custom-nameservers: 'false'
      dual-stack-ipv4-primary: 'false'
      dual-stack-ipv6-primary: 'false'
      export-from-confirm: 'true'
      import: 'false'
      standalone-cluster-mode: 'false'
    package:
      kctrl-package-command-tree: 'true'
contexts:
  - name: test-mc
    target: kubernetes
    group: one
    clusterOpts:
      isManagementCluster: true
      annotation: one
      required: true
      annotationStruct:
        one: one
      endpoint: cfg-test-endpoint
      path: cfg-test-path
      context: cfg-test-context
    discoverySources:
      - gcp:
          name: test
          bucket: cfg-test-bucket
          manifestPath: cfg-test-manifest-path
          annotation: one
          required: true
        contextType: tmc
      - gcp:
          name: test-two
          bucket: cfg-test-bucket
          manifestPath: cfg-test-manifest-path
          annotation: two
          required: true
        contextType: tmc
currentContext:
  kubernetes: test-mc
kind: ClientConfig
metadata:
  creationTimestamp: null`,
			cfg2: `
contexts:
  - name: test-mc
    target: kubernetes
    group: one
    clusterOpts:
      isManagementCluster: true
      annotation: one
      required: true
      annotationStruct:
        one: one
      endpoint: cfg2-test-endpoint
      path: cfg2-test-path
      context: cfg2-test-context
    discoverySources:
      - gcp:
          name: test
          bucket: cfg2-test-bucket
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
        contextType: tmc
currentContext:
  kubernetes: test-mc`,
			output: `apiVersion: config.tanzu.vmware.com/v1alpha1
kind: ClientConfig
metadata:
    creationTimestamp: null
clientOptions:
    cli:
        bomRepo: projects.registry.vmware.com/tkg
        compatibilityFilePath: tkg-compatibility
        discoverySources:
            - contextType: k8s
              local:
                name: default-local
                path: standalone
            - local:
                name: admin-local
                path: admin
        edition: tkg
    features:
        cluster:
            custom-nameservers: 'false'
            dual-stack-ipv4-primary: 'false'
            dual-stack-ipv6-primary: 'false'
        global:
            context-aware-cli-for-plugins: 'true'
            context-target: 'false'
            tkr-version-v1alpha3-beta: 'false'
        management-cluster:
            aws-instance-types-exclude-arm: 'true'
            custom-nameservers: 'false'
            dual-stack-ipv4-primary: 'false'
            dual-stack-ipv6-primary: 'false'
            export-from-confirm: 'true'
            import: 'false'
            standalone-cluster-mode: 'false'
        package:
            kctrl-package-command-tree: 'true'
contexts:
    - name: test-mc
      target: kubernetes
      group: one
      clusterOpts:
        isManagementCluster: true
        annotation: one
        required: true
        annotationStruct:
            one: one
        endpoint: cfg2-test-endpoint
        path: cfg2-test-path
        context: cfg2-test-context
      discoverySources:
        - gcp:
            name: test
            bucket: cfg2-test-bucket
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
          contextType: tmc
currentContext:
    kubernetes: test-mc
`,
		},
		{
			name: "success concat src into dst node with difference currentContext",
			cfg: `apiVersion: config.tanzu.vmware.com/v1alpha1
clientOptions:
  cli:
    bomRepo: projects.registry.vmware.com/tkg
    compatibilityFilePath: tkg-compatibility
    discoverySources:
      - contextType: k8s
        local:
          name: default-local
          path: standalone
      - local:
          name: admin-local
          path: admin
    edition: tkg
  features:
    cluster:
      custom-nameservers: 'false'
      dual-stack-ipv4-primary: 'false'
      dual-stack-ipv6-primary: 'false'
    global:
      context-aware-cli-for-plugins: 'true'
      context-target: 'false'
      tkr-version-v1alpha3-beta: 'false'
    management-cluster:
      aws-instance-types-exclude-arm: 'true'
      custom-nameservers: 'false'
      dual-stack-ipv4-primary: 'false'
      dual-stack-ipv6-primary: 'false'
      export-from-confirm: 'true'
      import: 'false'
      standalone-cluster-mode: 'false'
    package:
      kctrl-package-command-tree: 'true'
contexts:
  - name: test-mc
    target: kubernetes
    group: one
    clusterOpts:
      isManagementCluster: true
      annotation: one
      required: true
      annotationStruct:
        one: one
      endpoint: cfg-test-endpoint
      path: cfg-test-path
      context: cfg-test-context
    discoverySources:
      - gcp:
          name: test
          bucket: cfg-test-bucket
          manifestPath: cfg-test-manifest-path
          annotation: one
          required: true
        contextType: tmc
      - gcp:
          name: test-two
          bucket: cfg-test-bucket
          manifestPath: cfg-test-manifest-path
          annotation: two
          required: true
        contextType: tmc
currentContext:
  kubernetes: test-mc
kind: ClientConfig
metadata:
  creationTimestamp: null`,
			cfg2: `
contexts:
  - name: test-mc
    target: kubernetes
    group: one
    clusterOpts:
      isManagementCluster: true
      annotation: one
      required: true
      annotationStruct:
        one: one
      endpoint: cfg2-test-endpoint
      path: cfg2-test-path
      context: cfg2-test-context
    discoverySources:
      - gcp:
          name: test
          bucket: cfg2-test-bucket
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
        contextType: tmc
currentContext:
  mission-control: test-tmc`,
			output: `apiVersion: config.tanzu.vmware.com/v1alpha1
kind: ClientConfig
metadata:
    creationTimestamp: null
clientOptions:
    cli:
        bomRepo: projects.registry.vmware.com/tkg
        compatibilityFilePath: tkg-compatibility
        discoverySources:
            - contextType: k8s
              local:
                name: default-local
                path: standalone
            - local:
                name: admin-local
                path: admin
        edition: tkg
    features:
        cluster:
            custom-nameservers: 'false'
            dual-stack-ipv4-primary: 'false'
            dual-stack-ipv6-primary: 'false'
        global:
            context-aware-cli-for-plugins: 'true'
            context-target: 'false'
            tkr-version-v1alpha3-beta: 'false'
        management-cluster:
            aws-instance-types-exclude-arm: 'true'
            custom-nameservers: 'false'
            dual-stack-ipv4-primary: 'false'
            dual-stack-ipv6-primary: 'false'
            export-from-confirm: 'true'
            import: 'false'
            standalone-cluster-mode: 'false'
        package:
            kctrl-package-command-tree: 'true'
contexts:
    - name: test-mc
      target: kubernetes
      group: one
      clusterOpts:
        isManagementCluster: true
        annotation: one
        required: true
        annotationStruct:
            one: one
        endpoint: cfg2-test-endpoint
        path: cfg2-test-path
        context: cfg2-test-context
      discoverySources:
        - gcp:
            name: test
            bucket: cfg2-test-bucket
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
          contextType: tmc
currentContext:
    mission-control: test-tmc
`,
		},
		{
			name: "success concat src into dst node with empty contexts and currentContexts",
			cfg: `apiVersion: config.tanzu.vmware.com/v1alpha1
clientOptions:
  cli:
    bomRepo: projects.registry.vmware.com/tkg
    compatibilityFilePath: tkg-compatibility
    discoverySources:
      - contextType: k8s
        local:
          name: default-local
          path: standalone
      - local:
          name: admin-local
          path: admin
    edition: tkg
  features:
    cluster:
      custom-nameservers: 'false'
      dual-stack-ipv4-primary: 'false'
      dual-stack-ipv6-primary: 'false'
    global:
      context-aware-cli-for-plugins: 'true'
      context-target: 'false'
      tkr-version-v1alpha3-beta: 'false'
    management-cluster:
      aws-instance-types-exclude-arm: 'true'
      custom-nameservers: 'false'
      dual-stack-ipv4-primary: 'false'
      dual-stack-ipv6-primary: 'false'
      export-from-confirm: 'true'
      import: 'false'
      standalone-cluster-mode: 'false'
    package:
      kctrl-package-command-tree: 'true'
contexts:
  - name: test-mc
    target: kubernetes
    group: one
    clusterOpts:
      isManagementCluster: true
      annotation: one
      required: true
      annotationStruct:
        one: one
      endpoint: cfg-test-endpoint
      path: cfg-test-path
      context: cfg-test-context
    discoverySources:
      - gcp:
          name: test
          bucket: cfg-test-bucket
          manifestPath: cfg-test-manifest-path
          annotation: one
          required: true
        contextType: tmc
      - gcp:
          name: test-two
          bucket: cfg-test-bucket
          manifestPath: cfg-test-manifest-path
          annotation: two
          required: true
        contextType: tmc
currentContext:
  kubernetes: test-mc
kind: ClientConfig
metadata:
  creationTimestamp: null`,
			cfg2: `
contexts: []
currentContext: {}`,
			output: `apiVersion: config.tanzu.vmware.com/v1alpha1
kind: ClientConfig
metadata:
    creationTimestamp: null
clientOptions:
    cli:
        bomRepo: projects.registry.vmware.com/tkg
        compatibilityFilePath: tkg-compatibility
        discoverySources:
            - contextType: k8s
              local:
                name: default-local
                path: standalone
            - local:
                name: admin-local
                path: admin
        edition: tkg
    features:
        cluster:
            custom-nameservers: 'false'
            dual-stack-ipv4-primary: 'false'
            dual-stack-ipv6-primary: 'false'
        global:
            context-aware-cli-for-plugins: 'true'
            context-target: 'false'
            tkr-version-v1alpha3-beta: 'false'
        management-cluster:
            aws-instance-types-exclude-arm: 'true'
            custom-nameservers: 'false'
            dual-stack-ipv4-primary: 'false'
            dual-stack-ipv6-primary: 'false'
            export-from-confirm: 'true'
            import: 'false'
            standalone-cluster-mode: 'false'
        package:
            kctrl-package-command-tree: 'true'
contexts: []
currentContext: {}
`,
		},
		{
			name: "success concat scalar nodes with no entry in cfg2",
			cfg: `current: test-server
`,
			cfg2: `
contexts: []
currentContext: {}`,
			output: `current: test-server
contexts: []
currentContext: {}
`,
		},
		{
			name: "success concat scalar nodes with empty entry in cfg2",
			cfg: `currentContext: test-ctx
`,
			cfg2: `contexts: []
currentContext: {}
`,
			output: `contexts: []
currentContext: {}
`,
		},
		{
			name: "success concat scalar nodes with empty entry in cfg2",
			cfg: `apiVersion: config.tanzu.vmware.com/v1alpha1
clientOptions:
  cli:
    bomRepo: projects.registry.vmware.com/tkg
    compatibilityFilePath: tkg-compatibility
    discoverySources:
      - contextType: k8s
        local:
          name: default-local
          path: standalone
      - local:
          name: admin-local
          path: admin
    edition: tkg
  features:
    cluster:
      custom-nameservers: 'false'
      dual-stack-ipv4-primary: 'false'
      dual-stack-ipv6-primary: 'false'
    global:
      context-aware-cli-for-plugins: 'true'
      context-target: 'false'
      tkr-version-v1alpha3-beta: 'false'
    management-cluster:
      aws-instance-types-exclude-arm: 'true'
      custom-nameservers: 'false'
      dual-stack-ipv4-primary: 'false'
      dual-stack-ipv6-primary: 'false'
      export-from-confirm: 'true'
      import: 'false'
      standalone-cluster-mode: 'false'
    package:
      kctrl-package-command-tree: 'true'
contexts:
  - name: test-mc
    target: kubernetes
    group: one
    clusterOpts:
      isManagementCluster: true
      annotation: one
      required: true
      annotationStruct:
        one: one
      endpoint: cfg-test-endpoint
      path: cfg-test-path
      context: cfg-test-context
    discoverySources:
      - gcp:
          name: test
          bucket: cfg-test-bucket
          manifestPath: cfg-test-manifest-path
          annotation: one
          required: true
        contextType: tmc
      - gcp:
          name: test-two
          bucket: cfg-test-bucket
          manifestPath: cfg-test-manifest-path
          annotation: two
          required: true
        contextType: tmc
currentContext:
  kubernetes: test-mc
kind: ClientConfig
metadata:
  creationTimestamp: null`,
			cfg2: `contexts: []
currentContext: {}
current: 
`,
			output: `apiVersion: config.tanzu.vmware.com/v1alpha1
kind: ClientConfig
metadata:
    creationTimestamp: null
clientOptions:
    cli:
        bomRepo: projects.registry.vmware.com/tkg
        compatibilityFilePath: tkg-compatibility
        discoverySources:
            - contextType: k8s
              local:
                name: default-local
                path: standalone
            - local:
                name: admin-local
                path: admin
        edition: tkg
    features:
        cluster:
            custom-nameservers: 'false'
            dual-stack-ipv4-primary: 'false'
            dual-stack-ipv6-primary: 'false'
        global:
            context-aware-cli-for-plugins: 'true'
            context-target: 'false'
            tkr-version-v1alpha3-beta: 'false'
        management-cluster:
            aws-instance-types-exclude-arm: 'true'
            custom-nameservers: 'false'
            dual-stack-ipv4-primary: 'false'
            dual-stack-ipv6-primary: 'false'
            export-from-confirm: 'true'
            import: 'false'
            standalone-cluster-mode: 'false'
        package:
            kctrl-package-command-tree: 'true'
contexts: []
currentContext: {}
current:
`,
		},
		{
			name: "success empty data",
			cfg:  ``,
			cfg2: ``,
			output: `{}
`,
		},
	}

	for _, spec := range tests {
		t.Run(spec.name, func(t *testing.T) {
			_, cleanUp := setupTestConfig(t, &CfgTestData{cfg: spec.cfg, cfgNextGen: spec.cfg2})

			defer func() {
				cleanUp()
			}()

			multiNode, err := getMultiConfig()
			assert.NoError(t, err)

			multiBytes, err := yaml.Marshal(multiNode)
			assert.NoError(t, err)

			assert.Equal(t, spec.output, string(multiBytes))
		})
	}
}
