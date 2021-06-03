// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

func postInstallHook() error {
	// To initialize configuration we do not need to do anything
	// except creating tkgctl client. As initializing the client
	// initializes the configuration to the tanzu config directory.
	_, err := newTKGCtlClient()
	if err != nil {
		return err
	}
	return nil
}
