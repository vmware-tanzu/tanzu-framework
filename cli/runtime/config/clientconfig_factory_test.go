// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"golang.org/x/sync/errgroup"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

func TestClientConfigNodeUpdateInParallel(t *testing.T) {
	addServer := func(mcName string) error {
		_, err := getClientConfigNode()
		if err != nil {
			return err
		}

		s := &configapi.Server{
			Name: mcName,
			Type: configapi.ManagementClusterServerType,
			ManagementClusterOpts: &configapi.ManagementClusterServer{
				Context: "fake-context",
				Path:    "fake-path",
			},
		}
		err = SetServer(s, true)
		if err != nil {
			return err
		}
		_, err = getClientConfigNode()
		return err
	}
	// Run the parallel tests of reading and updating the configuration file
	// multiple times to make sure all the attempts are successful
	for testCounter := 1; testCounter <= 5; testCounter++ {
		func() {
			// Get the temp tanzu config file
			f, err := os.CreateTemp("", "tanzu_config*")
			assert.Nil(t, err)
			defer func(name string) {
				err = os.Remove(name)
				assert.NoError(t, err)
			}(f.Name())
			err = os.Setenv("TANZU_CONFIG", f.Name())
			assert.NoError(t, err)
			// run addServer in parallel
			parallelExecutionCounter := 100
			group, _ := errgroup.WithContext(context.Background())
			for i := 1; i <= parallelExecutionCounter; i++ {
				id := i
				group.Go(func() error {
					return addServer(fmt.Sprintf("mc-%v", id))
				})
			}
			_ = group.Wait()
			// Make sure that the configuration file is not corrupted
			node, err := getClientConfigNode()
			assert.Nil(t, err)
			// Make sure all expected servers are added to the knownServers list
			assert.Equal(t, parallelExecutionCounter, len(node.Content[0].Content[1].Content))
		}()
	}
}
