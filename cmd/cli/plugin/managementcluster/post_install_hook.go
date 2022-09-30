// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

func postInstallHook() error {
	// Creating a tkgctl client in turn initializes the TKG configuration in the tanzu config directory.
	forceUpdateTKGCompatibilityImage := true
	_, err := newTKGCtlClient(forceUpdateTKGCompatibilityImage)
	return err
}
