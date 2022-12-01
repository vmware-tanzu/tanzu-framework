// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	corev1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/core/v1alpha2"
	"github.com/vmware-tanzu/tanzu-framework/featuregates/client/pkg/featuregateclient"
)

var (
	userAllowsVoidingWarranty bool
)

// FeatureActivateCmd is for activating Features
var FeatureActivateCmd = &cobra.Command{
	Use:   "activate <feature>",
	Short: "Activate features",
	Args:  cobra.ExactArgs(1),
	Example: `
	# Activate a cluster Feature
	tanzu feature activate myfeature`,
	RunE: featureActivate,
}

func init() {
	FeatureActivateCmd.Flags().BoolVar(&userAllowsVoidingWarranty, "permanentlyVoidAllSupportGuarantees", false, "Allow for the permanent voiding of all support guarantees for this environment. For some features, e.g. experimental features, if a user sets the activation status to one that does not match the default activation, all support guarantees for this environment will be permanently voided.")
}

func featureActivate(cmd *cobra.Command, args []string) error {
	featureName := args[0]

	fgClient, err := featuregateclient.NewFeatureGateClient()
	if err != nil {
		return fmt.Errorf("could not get FeatureGate client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	feature, err := fgClient.GetFeature(ctx, featureName)
	if err != nil {
		return fmt.Errorf("could not get Feature %s: %w", featureName, err)
	}

	gates, err := fgClient.GetFeatureGateList(ctx)
	if err != nil {
		return fmt.Errorf("could not get FeatureGate List: %w", err)
	}

	gateName, featRef := featuregateclient.FeatureRefFromGateList(gates, featureName)

	var proceedWithVoidingWarranty bool
	if willWarrantyBeVoided(featRef, feature) {
		// The warranty will be voided with the request, so check that user allows it.
		var userAllows *bool
		if cmd.Flags().Changed("permanentlyVoidAllSupportGuarantees") {
			userAllows = &userAllowsVoidingWarranty
		}
		proceedWithVoidingWarranty, err = userGivesPermissionToVoidWarranty(feature, userAllows)
		if err != nil {
			return fmt.Errorf("could not get user permission to void warranty for Feature %s: %w", featureName, err)
		}
	}

	err = fgClient.ActivateFeature(ctx, featureName, proceedWithVoidingWarranty)
	if err != nil {
		return fmt.Errorf("could not activate Feature %s gated by FeatureGate %s: %w", featureName, gateName, err)
	}

	displayActivationWarnings(feature)

	cmd.Printf("Feature %s gated by FeatureGate %s is activated.\n", featureName, gateName)
	return nil
}

// displayActivationWarnings warns the user that technical preview features are
// unstable and lack support.
func displayActivationWarnings(feature *corev1alpha2.Feature) {
	if feature.Spec.Stability == corev1alpha2.TechnicalPreview {
		fmt.Printf("Warning: Technical preview features are not ready, but are not believed to be dangerous. The feature itself is unsupported, but activating technical preview features does not affect the support status of the environment. Use at your own risk.\n\n")
	}
}

// willWarrantyBeVoided checks that activating the Feature will cause warranty to be voided.
// Warranty will be voided only if all the following conditions are met:
//   - The stability policy's VoidsWarranty field value is true. This means changing the activation
//     set point away from default activation (in this case, default activate will have to be false)
//     will void the warranty.
//   - The warranty is not already void. If it's void, then it cannot be voided as it's already void.
//   - The current Feature activation set point in the FeatureGate reference is false. If the opposite is
//     true--that the Feature is already set to be activated--then there is nothing to change and
//     the warranty will not be voided.
//   - Activating the Feature will set it different than its default activation setting. This means the
//     default activation of a Feature must be false for the warranty to be voided. On the other hand,
//     if the default activation is true, then activating the Feature will not void the warranty.
//
// If all of the above are true, then the warranty will be voided and the function returns true.
// Otherwise, if any are false, the warranty will not be voided and the function returns false.
func willWarrantyBeVoided(ref corev1alpha2.FeatureReference, feature *corev1alpha2.Feature) bool {
	stability := feature.Spec.Stability
	policy := corev1alpha2.GetPolicyForStabilityLevel(stability)
	return policy.VoidsWarranty && !ref.PermanentlyVoidAllSupportGuarantees && !ref.Activate && !policy.DefaultActivation
}

func userGivesPermissionToVoidWarranty(feature *corev1alpha2.Feature, userAllowsByFlag *bool) (bool, error) {
	if userAllowsByFlag == nil {
		// Request user permission interactively to void the warranty if not already set by flag.
		fmt.Printf("Warning: activating %q, a %s Feature will irrevocably void all support guarantees for this environment. You will need to recreate the environment to return to a supported state.\nWould you like to continue [y/N]?", feature.Name, feature.Spec.Stability)
		return userAllowsByInteractiveCLI()
	}
	// If the user did pass a value to the flag, use the user's input. No interactive user input is needed.
	return *userAllowsByFlag, nil
}

func userAllowsByInteractiveCLI() (bool, error) {
	var userAnswer string
	if n, err := fmt.Scanln(&userAnswer); err != nil {
		if n == 0 {
			return false, nil
		}
		return false, fmt.Errorf("could not read user input: %w", err)
	}

	userAnswer = strings.TrimSpace(strings.ToLower(userAnswer))
	switch {
	case userAnswer == "yes" || userAnswer == "y":
		return true, nil
	case userAnswer == "" || userAnswer == "no" || userAnswer == "n":
		return false, nil
	default:
		fmt.Printf("Please respond with %q or %q\n", "y", "n")
		return userAllowsByInteractiveCLI()
	}
}
