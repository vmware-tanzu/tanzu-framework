// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

// Hub marks TanzuKubernetesAddon as a conversion hub.
func (*TanzuKubernetesAddon) Hub() {}

// Hub marks TanzuKubernetesAddonList as a conversion hub.
func (*TanzuKubernetesAddonList) Hub() {}
