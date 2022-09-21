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
		featureGateClient, err := featuregateclient.NewFeatureGateClient()
		if err != nil {
			return fmt.Errorf("couldn't get a featureGateRunner: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
		defer cancel()

		if err := featureGateClient.DeactivateFeature(ctx, featureName, featuregate); err != nil {
			return fmt.Errorf("couldn't deactivate feature %s: %w", featureName, err)
		}
		cmd.Printf("Feature %s Deactivated", featureName)
		return nil
	},
}

func init() {
	FeatureDeactivateCmd.Flags().StringVarP(&featuregate, "featuregate", "f", "tkg-system", "Deactivate Feature gated by a particular FeatureGate")
}
