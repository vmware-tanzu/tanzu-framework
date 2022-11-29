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
)

func TestConfigMetadataNodeUpdateInParallel(t *testing.T) {
	addPatchStrategy := func(key, value string) error {
		// Get config metadata node
		_, err := getMetadataNode()
		if err != nil {
			return err
		}

		// Set patch strategy
		err = SetConfigMetadataPatchStrategy(key, value)
		if err != nil {
			return err
		}

		// Get config metadata node
		_, err = getMetadataNode()
		return err
	}

	// Run the parallel tests of reading and updating the configuration file
	// multiple times to make sure all the attempts are successful
	for testCounter := 1; testCounter <= 5; testCounter++ {
		func() {
			// Get the temp tanzu config file
			f, err := os.CreateTemp("", "tanzu_config_metadata*")
			assert.Nil(t, err)
			defer func(name string) {
				err = os.Remove(name)
				assert.NoError(t, err)
			}(f.Name())
			err = os.Setenv("TANZU_CONFIG_METADATA", f.Name())
			assert.NoError(t, err)

			// run addPatchStrategy in parallel
			parallelExecutionCounter := 100
			group, _ := errgroup.WithContext(context.Background())
			for i := 1; i <= parallelExecutionCounter; i++ {
				id := i
				group.Go(func() error {
					return addPatchStrategy(fmt.Sprintf("p%v", id), "replace")
				})
			}
			_ = group.Wait()

			// Make sure that the configuration file is not corrupted
			node, err := getMetadataNode()
			assert.Nil(t, err)
			// Make sure all expected patch strategies are added to the patchStrategy list
			assert.Equal(t, parallelExecutionCounter, len(node.Content[0].Content[1].Content[1].Content)/2)
		}()
	}
}
