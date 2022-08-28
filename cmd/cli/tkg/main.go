// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// This file allows building tkg-cli
// Note: This is added as some tests still relies on tkg cli.
// This will be removed once that dependency is removed
package main

import (
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/cmd"
)

func main() {
	cmd.Execute()
}
