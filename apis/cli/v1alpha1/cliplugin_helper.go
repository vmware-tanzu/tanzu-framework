// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

// StringToTarget converts string to Target type
func StringToTarget(target string) Target {
	if target == string(targetK8s) || target == string(TargetK8s) {
		return TargetK8s
	} else if target == string(targetTMC) || target == string(TargetTMC) {
		return TargetTMC
	} else if target == string(TargetNone) {
		return TargetNone
	}
	return TargetNone
}
