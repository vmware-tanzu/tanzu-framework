// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package constants

import "time"

const (
	ContextTimeout                       = 60 * time.Second
	ServiceAccountWithDefaultPermissions = "tanzu-capabilities-manager-default-sa"
	CapabilitiesControllerNamespace      = "tkg-system"
)
