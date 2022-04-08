// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

const (
	TkcConversionAnnotationKey = "run.tanzu.vmware.com/v1alpha2_TanzuKubernetesCluster"
	// in the past, this key was inappropriately used to store tkc yaml of
	// the hub (v1alpha2) in the spoke (v1alpha1 at the time)
	PreviousTkcConversionAnnotationKey = "cluster.x-k8s.io/conversion-data"
)

// Hub marks TanzuKubernetesCluster as a conversion hub.
func (*TanzuKubernetesCluster) Hub() {}

// Hub marks TanzuKubernetesClusterList as a conversion hub.
func (*TanzuKubernetesClusterList) Hub() {}
