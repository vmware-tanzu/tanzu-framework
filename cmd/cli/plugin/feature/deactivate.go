// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/sdk/features/client"
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
		featureGateClient, err := client.NewFeatureGateClient()
		if err != nil {
			return fmt.Errorf("couldn't get a featureGateRunner: %w", err)
		}

		if err := featureGateClient.DeactivateFeature(cmd.Context(), featureName, featuregate); err != nil {
			return fmt.Errorf("couldn't deactivate feature %s: %w", featureName, err)
		}
		cmd.Printf("Feature %s Deactivated", featureName)
		return nil
	},
}

func init() {
	FeatureDeactivateCmd.Flags().StringVarP(&featuregate, "featuregate", "f", "tkg-system", "Deactivate Feature gated by a particular FeatureGate")
}
