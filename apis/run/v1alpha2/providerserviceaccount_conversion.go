// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

// Hub marks ProviderServiceAccount as a conversion hub.
func (*ProviderServiceAccount) Hub() {}

// Hub marks ProviderServiceAccountList as a conversion hub.
func (*ProviderServiceAccountList) Hub() {}
