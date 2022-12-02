// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"

	cliapi "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

func TestGetClientConfigNextGenNode(t *testing.T) {
	// Setup config test data
	_, cleanUp := setupTestConfig(t, &CfgTestData{})

	defer func() {
		cleanUp()
	}()

	//Action
	node, err := getClientConfigNextGenNode()

	//Assertions
	assert.NoError(t, err)
	c := &configapi.ClientConfig{}
	expectedNode, err := convertClientConfigToNode(c)
	expectedNode.Content[0].Style = 0
	assert.NoError(t, err)
	assert.Equal(t, expectedNode, node)
}

func TestClientConfigNextGenNodeUpdateInParallel(t *testing.T) {
	addContext := func(mcName string) error {
		_, err := getClientConfigNextGenNode()
		if err != nil {
			return err
		}

		ctx := &configapi.Context{
			Name:   mcName,
			Target: cliapi.TargetK8s,
			ClusterOpts: &configapi.ClusterServer{
				Path:                "test-path",
				Context:             "test-context",
				IsManagementCluster: true,
			},
			DiscoverySources: []configapi.PluginDiscovery{
				{
					GCP: &configapi.GCPDiscovery{
						Name:         "test",
						Bucket:       "updated-test-bucket",
						ManifestPath: "test-manifest-path",
					},
				},
			},
		}
		err = SetContext(ctx, true)
		if err != nil {
			return err
		}
		_, err = getClientConfigNextGenNode()
		return err
	}
	// Run the parallel tests of reading and updating the configuration file
	// multiple times to make sure all the attempts are successful
	for testCounter := 1; testCounter <= 1; testCounter++ {
		func() {
			// Setup config test data
			_, cleanUp := setupTestConfig(t, &CfgTestData{})

			defer func() {
				cleanUp()
			}()

			// run addContext in parallel
			parallelExecutionCounter := 100
			group, _ := errgroup.WithContext(context.Background())
			for i := 1; i <= parallelExecutionCounter; i++ {
				id := i
				group.Go(func() error {
					return addContext(fmt.Sprintf("mc-%v", id))
				})
			}
			_ = group.Wait()
			// Make sure that the configuration file is not corrupted
			node, err := getClientConfigNextGenNode()
			assert.Nil(t, err)
			// Make sure all expected servers are added to the knownServers list
			assert.Equal(t, parallelExecutionCounter, len(node.Content[0].Content[1].Content))
		}()
	}
}
