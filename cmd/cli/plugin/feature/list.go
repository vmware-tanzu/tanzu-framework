// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	corev1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/core/v1alpha2"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/component"
	"github.com/vmware-tanzu/tanzu-framework/featuregates/client/pkg/featuregateclient"
)

var (
	featuregate, outputFormat     string
	activated, deactivated        bool
	extended, includeExperimental bool
)

// FeatureListCmd is for listing features in (a) featuregate(s).
var FeatureListCmd = &cobra.Command{
	Use:   "list",
	Short: "List features",
	Args:  cobra.NoArgs,
	Example: `
	# List feature(s) in the cluster.
	tanzu feature list --activated
	tanzu feature list --deactivated`,
	RunE: printFeatures,
}

func init() {
	FeatureListCmd.Flags().BoolVarP(&extended, "extended", "e", false, "Include extended output")
	FeatureListCmd.Flags().StringVarP(&featuregate, "featuregate", "f", "", "List features gated by specified FeatureGate")
	FeatureListCmd.Flags().BoolVarP(&activated, "activated", "a", false, "List only activated features")
	FeatureListCmd.Flags().BoolVarP(&deactivated, "deactivated", "d", false, "List only deactivated features")
	FeatureListCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "", "Output format (yaml|json|table)")
	FeatureListCmd.Flags().BoolVarP(&includeExperimental, "include-experimental", "x", false, "Include experimental features in list")
}

// FeatureInfo is a struct that holds Feature information an if it can be listed by plugin.
type FeatureInfo struct {
	Name         string
	Description  string
	Stability    corev1alpha2.StabilityLevel
	Immutable    bool
	Discoverable bool
	FeatureGate  string
	Activated    bool
	ShowInList   bool
}

func printFeatures(cmd *cobra.Command, _ []string) error {
	fgClient, err := featuregateclient.NewFeatureGateClient()
	if err != nil {
		return fmt.Errorf("could not get FeatureGateClient: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	infos, err := featureInfoList(ctx, fgClient, featuregate)
	if err != nil {
		return fmt.Errorf("could not gather features' information: %w", err)
	}

	if extended {
		return listExtended(cmd, infos)
	}
	return listBasic(cmd, infos)
}

// featureInfoList will determine which features' information will be listed. If a feature's info
// is displayed or not depends on the flags passed in by the user (e.g., activated, deactivated),
// as well as the features' discoverable settings and stability levels.
func featureInfoList(ctx context.Context, cl *featuregateclient.FeatureGateClient, featuregate string) ([]FeatureInfo, error) {
	clusterFeatures, err := cl.GetFeatureList(ctx)
	if err != nil {
		return nil, err
	}

	gateList, err := cl.GetFeatureGateList(ctx)
	if err != nil {
		return nil, err
	}

	featureInfos := collectFeaturesInfo(gateList.Items, clusterFeatures.Items)

	setShowInList(featureInfos, includeExperimental, featuregate)

	filteredList := featuresFilteredByFlags(featureInfos, activated, deactivated)
	return filteredList, nil
}

// collectFeaturesInfo will create a map of features and their information from
// FeatureGate references and features.
func collectFeaturesInfo(gates []corev1alpha2.FeatureGate, features []corev1alpha2.Feature) map[string]*FeatureInfo {
	infos := map[string]*FeatureInfo{}

	for i := range features {
		policy := corev1alpha2.GetPolicyForStabilityLevel(features[i].Spec.Stability)

		infos[features[i].Name] = &FeatureInfo{
			Name:         features[i].Name,
			Description:  features[i].Spec.Description,
			Stability:    features[i].Spec.Stability,
			Activated:    features[i].Status.Activated,
			Immutable:    policy.Immutable,
			Discoverable: policy.Discoverable,
			FeatureGate:  "--",
		}
	}

	for i := range gates {
		for _, featRef := range gates[i].Spec.Features {
			info, ok := infos[featRef.Name]
			if ok {
				// FeatureGate referenced Feature is in cluster.
				info.FeatureGate = gates[i].Name
			}

			if !ok {
				// FeatureGate referenced Feature is not in cluster. Since the Discoverable policy
				// cannot be known until the Feature shows up in cluster, set it to true for now.
				infos[featRef.Name] = &FeatureInfo{
					Name:         featRef.Name,
					Discoverable: true,
					FeatureGate:  gates[i].Name,
				}
			}
		}
	}

	return infos
}

// setShowInList will determine if a Feature will be listed based on Features that can
// be listed by default and whether or not a FeatureGate was specified by the user.
func setShowInList(infos map[string]*FeatureInfo, inclExperimental bool, gateName string) {
	for _, v := range infos {
		// Only discoverable features can be listed.
		v.ShowInList = v.Discoverable

		// Experimental features are not listed by default, but can be listed via flag.
		if v.Stability == corev1alpha2.Experimental {
			v.ShowInList = inclExperimental
		}

		// If FeatureGate is specified, delist the Features not gated.
		if gateName != "" && v.FeatureGate != gateName {
			v.ShowInList = false
		}
	}
}

// featuresFilteredByFlags will determine which features will be listed based on the flags provided.
// If none of flags were set, then all features will be selected to display.
func featuresFilteredByFlags(infos map[string]*FeatureInfo, activated, deactivated bool) []FeatureInfo {
	var filteredList []FeatureInfo
	for _, v := range infos {
		if activated && v.Activated && v.ShowInList {
			filteredList = append(filteredList, *v)
		}

		if deactivated && !v.Activated && v.ShowInList {
			filteredList = append(filteredList, *v)
		}

		// No flags were provided, so only filter out features that shouldn't be listed.
		if !activated && !deactivated && v.ShowInList {
			filteredList = append(filteredList, *v)
		}
	}
	return filteredList
}

// listExtended renders the output with extended feature information
func listExtended(cmd *cobra.Command, features []FeatureInfo) error {
	var t component.OutputWriterSpinner
	t, err := component.NewOutputWriterWithSpinner(cmd.OutOrStdout(), outputFormat,
		"Retrieving Features...", true, "NAME", "ACTIVATION STATE", "STABILITY", "DESCRIPTION", "IMMUTABLE", "FEATUREGATE")
	if err != nil {
		return fmt.Errorf("could not get OutputWriterSpinner: %w", err)
	}

	for _, info := range features {
		t.AddRow(info.Name, info.Activated, info.Stability, info.Description, info.Immutable, info.FeatureGate)
	}
	t.RenderWithSpinner()

	return nil
}

// listBasic renders the output with basic feature information
func listBasic(cmd *cobra.Command, features []FeatureInfo) error {
	var t component.OutputWriterSpinner
	t, err := component.NewOutputWriterWithSpinner(cmd.OutOrStdout(), outputFormat,
		"Retrieving Features...", true, "NAME", "ACTIVATION STATE", "FEATUREGATE")
	if err != nil {
		return fmt.Errorf("could not get OutputWriterSpinner: %w", err)
	}

	for _, info := range features {
		t.AddRow(info.Name, info.Activated, info.FeatureGate)
	}
	t.RenderWithSpinner()

	return nil
}
