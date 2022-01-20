// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	crClient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/component"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/sdk/features/client"
)

var featuregate, outputFormat string
var activated, deactivated, unavailable, extended bool

// FeatureListCmd is for activating Features
var FeatureListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Features",
	Args:  cobra.NoArgs,
	Example: `
	# List a clusters Features
	tanzu feature list --activated
	tanzu feature list --unavailable
	tanzu feature list --deactivated`,
	RunE: featureList,
}

func init() {
	FeatureListCmd.Flags().BoolVarP(&extended, "extended", "e", false, "Include extended output. Higher latency as it requires more API calls.")
	FeatureListCmd.Flags().StringVarP(&featuregate, "featuregate", "f", "tkg-system", "List Features gated by a particular FeatureGate")
	FeatureListCmd.Flags().BoolVarP(&activated, "activated", "a", false, "List only activated Features")
	FeatureListCmd.Flags().BoolVarP(&deactivated, "deactivated", "d", false, "List only deactivated Features")
	FeatureListCmd.Flags().BoolVarP(&unavailable, "unavailable", "u", false, "List only Features specified in the gate but missing from cluster")
	FeatureListCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "", "Output format (yaml|json|table)")
}

// FeatureInfo is a struct that holds Feature information
type FeatureInfo struct {
	Name        string
	Maturity    string
	Description string
	Activated   bool
	Available   bool
	Immutable   bool
}

func featureList(cmd *cobra.Command, _ []string) error {
	featureGateClient, err := client.NewFeatureGateClient()
	if err != nil {
		return fmt.Errorf("couldn't get featureGateRunner: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	gate, err := featureGateClient.GetFeatureGate(ctx, featuregate)
	if crClient.IgnoreNotFound(err) != nil {
		return err
	}
	if err != nil {
		return fmt.Errorf("featuregate %s not found: %w", featuregate, err)
	}

	var features []*FeatureInfo
	switch {
	case activated:
		for _, a := range gate.Status.ActivatedFeatures {
			features = append(features, &FeatureInfo{
				Name:      a,
				Activated: true,
				Available: true,
			})
		}
	case deactivated:
		for _, a := range gate.Status.DeactivatedFeatures {
			features = append(features, &FeatureInfo{
				Name:      a,
				Activated: false,
				Available: true,
			})
		}
	case unavailable:
		for _, a := range gate.Status.UnavailableFeatures {
			features = append(features, &FeatureInfo{
				Name:      a,
				Activated: false,
				Available: false,
			})
		}
	default:
		for _, a := range gate.Status.ActivatedFeatures {
			features = append(features, &FeatureInfo{
				Name:      a,
				Activated: true,
				Available: true,
			})
		}
		for _, a := range gate.Status.DeactivatedFeatures {
			features = append(features, &FeatureInfo{
				Name:      a,
				Activated: false,
				Available: true,
			})
		}
		for _, a := range gate.Status.UnavailableFeatures {
			features = append(features, &FeatureInfo{
				Name:      a,
				Activated: false,
				Available: false,
			})
		}
	}
	if extended {
		err := joinFeatures(ctx, featureGateClient, features)
		if err != nil {
			return fmt.Errorf("couldn't get extended information about features: %w", err)
		}

		// Many API calls to gather and join on Features
		return listExtended(cmd, features)
	}

	return listBasic(cmd, features)
}

// listExtended renders the output with extended feature information
func listExtended(cmd *cobra.Command, features []*FeatureInfo) error {
	var t component.OutputWriterSpinner
	t, err := component.NewOutputWriterWithSpinner(cmd.OutOrStdout(), outputFormat,
		"Retrieving Features...", true, "NAME", "ACTIVATION STATE", "AVAILABLE", "MATURITY", "DESCRIPTION", "IMMUTABLE")
	if err != nil {
		return fmt.Errorf("couldn't get OutputWriterSpinner: %w", err)
	}

	for _, info := range features {
		t.AddRow(
			info.Name,
			info.Activated,
			info.Available,
			info.Maturity,
			info.Description,
			info.Immutable)
	}
	t.RenderWithSpinner()

	return nil
}

// listBasic renders the output with basic feature information
func listBasic(cmd *cobra.Command, features []*FeatureInfo) error {
	var t component.OutputWriterSpinner
	t, err := component.NewOutputWriterWithSpinner(cmd.OutOrStdout(), outputFormat,
		"Retrieving Features...", true, "NAME", "ACTIVATION STATE", "AVAILABLE")
	if err != nil {
		return fmt.Errorf("couldn't get OutputWriterSpinner: %w", err)
	}

	for _, info := range features {
		t.AddRow(
			info.Name,
			info.Activated,
			info.Available)
	}
	t.RenderWithSpinner()

	return nil
}

// joinFeatures retrieves additional information of a Feature and merges it with existing information.
func joinFeatures(ctx context.Context, featureGateClient *client.FeatureGateClient, features []*FeatureInfo) error {
	for _, info := range features {
		feature, err := featureGateClient.GetFeature(ctx, info.Name)
		if err != nil {
			// skip returning the error for unavailable feature when not found
			// to fetch others
			if !info.Available && apierrors.IsNotFound(err) {
				continue
			}
			return err
		}
		info.Maturity = feature.Spec.Maturity
		info.Description = feature.Spec.Description
		info.Immutable = feature.Spec.Immutable
	}
	return nil
}
