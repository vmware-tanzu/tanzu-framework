// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/featuregates/client/pkg/featuregateclient"
)

// FeatureDeactivateCmd is for deactivating Features
var FeatureDeactivateCmd = &cobra.Command{
	Use:   "deactivate <feature>",
	Short: "Deactivate Features",
	Args:  cobra.ExactArgs(1),
	Example: `
	# Deactivate a cluster Feature
	tanzu feature deactivate myfeature`,
	RunE: func(cmd *cobra.Command, args []string) error {
		featureName := args[0]

		fgClient, err := featuregateclient.NewFeatureGateClient()
		if err != nil {
			return fmt.Errorf("could not get FeatureGateClient: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
		defer cancel()

		gateName, err := fgClient.DeactivateFeature(ctx, featureName)
		if err != nil {
			return fmt.Errorf("could not deactivate Feature %s gated by FeatureGate %s: %w", featureName, gateName, err)
		}

		cmd.Printf("Feature %s gated by FeatureGate %s is deactivated.\n", featureName, gateName)
		return nil
	},
}
