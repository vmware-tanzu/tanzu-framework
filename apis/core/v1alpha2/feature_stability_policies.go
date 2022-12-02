// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

// Policy represents Stability level policy
type Policy struct {
	// DefaultActivation is the default activation state of the Feature. When a new Feature resource is added to the
	// cluster, the state of that Feature is expected to be in the default activation state defined by the stability
	// level that Feature uses.
	DefaultActivation bool
	// Immutable defines the Feature state immutability. When set to true for a stability level, the Feature with that
	// stability level cannot be toggled.
	Immutable bool
	// VoidsWarranty is for defining the warranty support for the environment where a Feature with that stability level
	// is activated. When set to true, activating the Feature using that stability level will void the warranty for that
	// environment.
	VoidsWarranty bool
	// Discoverable is for defining the Feature discoverability. When set to true, the Feature with that stability level
	// is discoverable
	Discoverable bool
}

// StabilityPolicies is map that holds the policies for different stability levels
var StabilityPolicies = map[StabilityLevel]Policy{
	WorkInProgress: {
		DefaultActivation: false,
		Immutable:         false,
		VoidsWarranty:     true,
		Discoverable:      false,
	},
	Experimental: {
		DefaultActivation: false,
		Immutable:         false,
		VoidsWarranty:     true,
		Discoverable:      true,
	},
	TechnicalPreview: {
		DefaultActivation: false,
		Immutable:         false,
		VoidsWarranty:     false,
		Discoverable:      true,
	},
	Stable: {
		DefaultActivation: true,
		Immutable:         true,
		VoidsWarranty:     false,
		Discoverable:      true,
	},
	Deprecated: {
		DefaultActivation: true,
		Immutable:         false,
		VoidsWarranty:     false,
		Discoverable:      true,
	},
}

// GetPolicyForStabilityLevel returns policy for stability level
func GetPolicyForStabilityLevel(stability StabilityLevel) Policy {
	return StabilityPolicies[stability]
}
