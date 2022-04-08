// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// TKRName is the name of TanzuKubernetesRelease object for the provided version
func TKRName(version string) string {
	if version = strings.TrimSpace(version); version == "" {
		return ""
	}
	return "v" + strings.ReplaceAll(strings.TrimPrefix(version, "v"), "+", "---")
}

// TKRRef returns an ObjectReference to the TanzuKubernetesRelease object for the provided version.
func TKRRef(version string) *corev1.ObjectReference {
	tkrName := TKRName(version)
	// TKR Reference should be nil if version == ""
	if tkrName == "" {
		return nil
	}
	return &corev1.ObjectReference{Name: tkrName}
}

// TKRVersion returns the TKR version given its name
func TKRVersion(s string) string {
	return strings.ReplaceAll(strings.TrimPrefix(s, "v"), "---", "+")
}

func TKRRefVersion(tkrReference *corev1.ObjectReference) string {
	if tkrReference == nil || tkrReference.Name == "" {
		return ""
	}
	return TKRVersion(tkrReference.Name)
}
