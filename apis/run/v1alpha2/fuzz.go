// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

import (
	fuzz "github.com/google/gofuzz"
	corev1 "k8s.io/api/core/v1"
)

func FuzzTKRSpec(tkrSpec *TanzuKubernetesReleaseSpec, c fuzz.Continue) {
	c.Fuzz(&tkrSpec.Version)
	c.Fuzz(&tkrSpec.KubernetesVersion)
	c.Fuzz(&tkrSpec.Repository)
	c.Fuzz(&tkrSpec.Images)
	tkrSpec.NodeImageRef = nil
	if c.RandBool() {
		tkrSpec.NodeImageRef = &corev1.ObjectReference{Name: c.RandString()}
	}
}
