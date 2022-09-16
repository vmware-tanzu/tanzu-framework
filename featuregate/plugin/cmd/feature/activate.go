// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/featuregate/client/pkg/featuregateclient"
)

// FeatureActivateCmd is for activating Features
var FeatureActivateCmd = &cobra.Command{
	Use:   "activate <feature>",
	Short: "Activate Features",
	Args:  cobra.ExactArgs(1),
	Example: `
	# Activate a cluster Feature
	tanzu feature activate myfeature`,
	RunE: featureActivate,
}

func init() {
	FeatureActivateCmd.Flags().StringVarP(&featuregate, "featuregate", "f", "tkg-system", "Activate a Feature gated by a particular FeatureGate")
}

func featureActivate(cmd *cobra.Command, args []string) error {
	featureName := args[0]
	featureGateClient, err := featuregateclient.NewFeatureGateClient()
	if err != nil {
		return fmt.Errorf("couldn't get featureGateRunner: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	if err := featureGateClient.ActivateFeature(ctx, featureName, featuregate); err != nil {
		return fmt.Errorf("couldn't activate feature %s: %w", featureName, err)
	}
	cmd.Printf("Feature %s Activated", featureName)
	return nil
}
